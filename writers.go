package main

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/rs/zerolog/log"

	"microservice/globals"
	"microservice/types"
)

func writeStations(stations []types.Station) {
	log := log.With().Str("writer", "stations").Logger()
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
			continue
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
		log := log.With().Str("writer", "measurements").Logger()
		log.Debug().Str("station", measurement.Station.String).Msg("checking for valid data")
		var dataInvalid bool
		err = pgxscan.Get(ctx, globals.Db, &dataInvalid, nullCheckQuery, measurement.Station, measurement.Date)
		if err != nil {
			log.Error().Str("station", measurement.Station.String).Err(err).Msg("unable check for incomplete data")
			continue
		}
		if dataInvalid {
			log.Warn().Str("station", measurement.Station.String).Msg("found incomplete measurement data. updating data with crawled data")
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
			continue
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
			continue
		}
	}
	log.Info().Msg("wrote measurement data into database")
}
