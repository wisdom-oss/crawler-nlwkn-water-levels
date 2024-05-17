-- name: create-station-locations
CREATE TABLE geodata.water_level_stations
(
    website_id text NOT NULL UNIQUE,
    public_id  text NOT NULL UNIQUE,
    name       text NOT NULL,
    operator   text,
    location   geometry('POINT', 4326)
);

-- name: create-base-table
CREATE TABLE timeseries.nlwkn_water_levels
(
    station         text NOT NULL REFERENCES geodata.water_level_stations (website_id) MATCH SIMPLE,
    date            date NOT NULL DEFAULT NOW()::date,
    classification  text,
    water_level_nhn numeric,
    water_level_gok numeric
);

CREATE UNIQUE INDEX idx_one_measurement_per_date ON timeseries.nlwkn_water_levels (station, date);


-- name: convert-to-hypertable
SELECT create_hypertable('timeseries.nlwkn_water_levels', by_range('date', INTERVAL '1 month'));

-- name: insert-measurement
INSERT INTO timeseries.nlwkn_water_levels
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO NOTHING;

-- name: null-measurement-exists
SELECT EXISTS(SELECT station, date
              FROM timeseries.nlwkn_water_levels
              WHERE station = $1
                AND date = $2
                AND (water_level_gok IS NULL OR water_level_nhn IS NULL OR classification IS NULL));

-- name: update-measurement
UPDATE timeseries.nlwkn_water_levels
SET water_level_gok = $1,
    water_level_nhn = $2,
    classification  = $3
WHERE station = $4
  AND date = $5;

-- name: insert-station
INSERT INTO geodata.water_level_stations
VALUES ($1, $2, $3, $4, geomfromewkb($5))
ON CONFLICT DO NOTHING;

-- name: station-exists
SELECT EXISTS(SELECT website_id
              FROM geodata.water_level_stations
              WHERE website_id = $1);

-- name: last-entry-for-station
SELECT date
FROM timeseries.nlwkn_water_levels
WHERE station = $1
ORDER BY date DESC
LIMIT 1;
