package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
	healthcheckServer "github.com/wisdom-oss/go-healthcheck/server"

	"microservice/globals"
	"microservice/types"
)

const crawlUrl = "https://www.grundwasserstandonline.nlwkn.niedersachsen.de/Messwerte"
const tableID = "ctl00_MainContent_rgMesswerte_ctl00__"

// crawlFrequency sets the time which needs to have after the last crawl before
// accessing the page again
var crawlFrequency time.Duration

// tickerFrequency sets the interval at which the service checks if it may
// access the page again

// the main function bootstraps the http server and handlers used for this
// microservice
func main() {
	// create a new logger for the main function
	l := log.With().Str("step", "main").Logger()
	l.Info().Msgf("starting %s service", globals.ServiceName)

	// create the healthcheck server
	hcServer := healthcheckServer.HealthcheckServer{}
	hcServer.InitWithFunc(func() error {
		// test if the database is reachable
		return globals.Db.Ping(context.Background())
	})
	err := hcServer.Start()
	if err != nil {
		l.Fatal().Err(err).Msg("unable to start healthcheck server")
	}
	go hcServer.Run()

	// setup graceful crawler shutdown
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, os.Interrupt)

	// setup ticker for recurring data pulls
	ticker := time.NewTicker(1 * time.Minute)

	// setup an http client
	httpClient := http.Client{}

	// create a boolean to track if a crawl has been successful
	var lastCrawlCall time.Time

	for {
		select {
		case _ = <-cancelSignal:
			l.Info().Msg("shutting down gracefully")
			os.Exit(0)
		case <-ticker.C:
			if lastCrawlCall.IsZero() {
				goto crawling
			}
			if time.Now().Sub(lastCrawlCall) < 6*time.Hour {
				log.Warn().Msg("already called crawl in the last 6h, skipping run")
				break
			}
		crawling:
			log.Info().Msg("checking for new data on the webpage")
			res, err := httpClient.Get(crawlUrl)
			if err != nil {
				l.Error().Err(err).Msg("error while requesting measurement page")
				break
			}
			page, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				l.Error().Err(err).Msg("unable to parse response into document")
				break
			}

			rows := page.Find(`tr[id^="` + tableID + `"]`)

			var stations []types.Station
			var measurements []types.Measurement
			var errorOccurred bool

			rows.Each(func(i int, row *goquery.Selection) {
				dataFields := row.Find("td").Nodes
				station := types.Station{}
				err = station.FromDataFields(dataFields)
				if err != nil {
					log.Error().Err(err).Msg("unable to create station from page")
					errorOccurred = true
					return
				}
				stations = append(stations, station)
				measurement := types.Measurement{}
				err = measurement.FromDataFields(dataFields)
				if err != nil {
					log.Error().Err(err).Msg("unable to create measurement from page")
					errorOccurred = true
					return
				}
				measurements = append(measurements, measurement)
			})
			if errorOccurred {
				log.Error().Msg("an error occurred during the handling of the parsed page. please refer to the previous logs")
				break
			}
			lastCrawlCall = time.Now()
			log.Info().Msg("crawling finished. writing entries asynchronously")
			go func() {
				writeStations(stations)
				writeMeasurements(measurements)
			}()

		}
	}
}

func writeMeasurements(measurements []types.Measurement) {
	ctx := context.Background()
	insertQuery, err := globals.SqlQueries.Raw("insert-measurement")
	if err != nil {
		log.Error().Err(err).Msg("unable to prepare sql query for measurement insertion")
		return
	}
	nullCheckQuery, err := globals.SqlQueries.Raw("null-measurement-exists")
	if err != nil {
		log.Error().Err(err).Msg("unable to prepare sql query for measurement validity check")
		return
	}
	updateQuery, err := globals.SqlQueries.Raw("update-measurement")
	if err != nil {
		log.Error().Err(err).Msg("unable to prepare sql query for measurement update")
		return
	}
	for _, measurement := range measurements {
		log.Debug().Str("station", measurement.Station.String).Msg("checking for valid data")
		var dataInvalid bool
		err = pgxscan.Get(ctx, globals.Db, &dataInvalid, nullCheckQuery, measurement.Station, measurement.Date)
		if err != nil {
			log.Error().Str("station", measurement.Station.String).Err(err).Msg("unable check for incomplete data")
			continue
		}
		if dataInvalid {
			log.Warn().Str("station", measurement.Station.String).Msg("found incomplete data. completing data")
			_, err = globals.Db.Exec(ctx, updateQuery,
				measurement.WaterLevelGOK,
				measurement.WaterLevelNHN,
				measurement.Classification,
				measurement.Station,
				measurement.Date)
			if err != nil {
				log.Error().Str("station", measurement.Station.String).Err(err).Msg("unable to update incomplete data")
				continue
			}
			return
		}
		log.Debug().Str("station", measurement.Station.String).Msg("writing measurement data")
		_, err := globals.Db.Exec(ctx, insertQuery,
			measurement.Station,
			measurement.Date,
			measurement.Classification,
			measurement.WaterLevelNHN,
			measurement.WaterLevelGOK)
		if err != nil {
			log.Error().Err(err).Msg("unable to insert/update station")
			return
		}
	}
}

func writeStations(stations []types.Station) {
	ctx := context.Background()
	query, err := globals.SqlQueries.Raw("insert-station")
	if err != nil {
		log.Error().Err(err).Msg("unable to prepare sql query for station insertion")
		return
	}
	for _, station := range stations {
		log.Debug().Str("name", station.Name.String).Msg("writing station data")
		_, err := globals.Db.Exec(ctx, query,
			station.WebsiteID,
			station.PublicID,
			station.Name,
			station.Operator,
			station.Location)
		if err != nil {
			log.Error().Err(err).Msg("unable to insert/update station")
			return
		}
	}

}
