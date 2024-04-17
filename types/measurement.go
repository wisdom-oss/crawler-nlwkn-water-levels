package types

import (
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/net/html"
)

var ErrNoMeasurement = errors.New("no measurement available")

type Measurement struct {
	// Station contains the Station.WebsiteID of the water measurement station at
	// which the measurement has been taken
	Station pgtype.Text `json:"station" db:"station"`

	// Date contains the date at which the measurement has been taken
	Date pgtype.Date `json:"date" db:"date"`

	// Classification contains the textual representation of the water levels
	// classification.
	// The values are documented here:
	// https://www.grundwasserstandonline.nlwkn.niedersachsen.de/Hinweis#einstufungGrundwasserstandsklassen
	Classification pgtype.Text `json:"classification" db:"classification"`

	// WaterLevelNHN refers to the water level in reference to the sea level in
	// Germany
	WaterLevelNHN pgtype.Numeric `json:"waterLevelNHN" db:"water_level_nhn"`

	// WaterLevelGOK refers to the water level in reference to the terrain
	// height around the measurement station
	WaterLevelGOK pgtype.Numeric `json:"waterLevelGOK" db:"water_level_gok"`
}

func (m *Measurement) FromDataFields(dataFields []*html.Node) (err error) {
	if err = m.Station.Scan(dataFields[1].FirstChild.Data); err != nil {
		return err
	}
	if err = m.Classification.Scan(dataFields[8].FirstChild.Data); err != nil {
		return err
	}
	var date time.Time
	date, err = time.Parse("02.01.2006", dataFields[5].FirstChild.Data)
	if err != nil {
		return err
	}
	err = m.Date.Scan(date)
	if err != nil {
		return err
	}
	normalizedWaterLevelNHN := strings.TrimSpace(strings.ReplaceAll(dataFields[6].FirstChild.Data, ",", "."))
	normalizedWaterLevelGOK := strings.TrimSpace(strings.ReplaceAll(dataFields[7].FirstChild.Data, ",", "."))
	if normalizedWaterLevelNHN == "-" {
		_ = m.WaterLevelNHN.Scan(nil)
	} else {
		if err = m.WaterLevelNHN.Scan(normalizedWaterLevelNHN); err != nil {
			return err
		}
	}
	if normalizedWaterLevelGOK == "-" {
		_ = m.WaterLevelGOK.Scan(nil)
	} else {
		if err = m.WaterLevelGOK.Scan(normalizedWaterLevelGOK); err != nil {
			return err
		}

	}

	return nil
}
