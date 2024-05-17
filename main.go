package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
	healthcheckServer "github.com/wisdom-oss/go-healthcheck/server"

	"microservice/globals"
)

const crawlUrl = "https://www.grundwasserstandonline.nlwkn.niedersachsen.de/Messwerte"
const tableID = "ctl00_MainContent_rgMesswerte_ctl00__"

// crawlFrequency sets the time which needs to have after the last crawl before
// accessing the page again
var crawlFrequency time.Duration

// tickerFrequency sets the interval at which the service checks if it may
// access the page again
var tickerFrequency time.Duration

var hcFunc = func() error {
	_, err := downloadTablePage()
	if err != nil {
		return err
	}
	return globals.Db.Ping(context.Background())
}

// the main function bootstraps the http server and handlers used for this
// microservice
func main() {
	// create a new logger for the main function
	l := log.With().Str("step", "main").Logger()
	l.Info().Msgf("starting %s service", globals.ServiceName)

	// create the healthcheck server
	hcServer := healthcheckServer.HealthcheckServer{}
	hcServer.InitWithFunc(hcFunc)
	err := hcServer.Start()
	if err != nil {
		l.Fatal().Err(err).Msg("unable to start healthcheck server")
	}
	go hcServer.Run()

	// setup graceful crawler shutdown
	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, os.Interrupt)

	// setup ticker for recurring data pulls
	ticker := time.NewTicker(tickerFrequency)

	// setup an http client

	// create a boolean to track if a crawl has been successful
	var lastCrawlCall time.Time

	for {
		select {
		case _ = <-cancelSignal:
			l.Info().Msg("shutting down gracefully")
			os.Exit(0)
		case <-ticker.C:
			if !lastCrawlCall.IsZero() && time.Now().Sub(lastCrawlCall) < crawlFrequency {
				log.Warn().Msgf("already crawled data in the last %s. skipping this run", crawlFrequency)
				break
			}

			log.Info().Msg("accessing measurement table page")
			page, err := downloadTablePage()
			if err != nil {
				l.Error().Err(err).Msg("unable to parse response into document")
				break
			}

			log.Info().Msg("reading measurement table")
			stations, measurements, err := readTable(page)
			if err != nil {
				log.Error().Err(err).Msg("unable to read downloaded table")
				break
			}
			log.Info().Msg("reading finished. writing found entries asynchronously")
			lastCrawlCall = time.Now()
			go func() {
				writeStations(stations)
				writeMeasurements(measurements)
			}()
		}
	}
}
