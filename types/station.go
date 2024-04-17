package types

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	_ "github.com/twpayne/pgx-geom"
	"golang.org/x/net/html"
)

type Station struct {
	WebsiteID pgtype.Text `json:"websiteID" db:"website_id"`
	PublicID  pgtype.Text `json:"publicID" db:"public_id"`
	Name      pgtype.Text `json:"name" db:"name"`
	Operator  pgtype.Text `json:"operator" db:"operator"`
	Location  *ewkb.Point `json:"location" db:"location"`
}

func (s *Station) FromDataFields(dataFields []*html.Node) error {
	err := s.Name.Scan(dataFields[0].FirstChild.Data)
	if err != nil {
		return err
	}
	err = s.WebsiteID.Scan(dataFields[1].FirstChild.Data)
	if err != nil {
		return err
	}
	err = s.PublicID.Scan(dataFields[2].FirstChild.Data)
	if err != nil {
		return err
	}
	err = s.Operator.Scan(dataFields[4].FirstChild.Data)
	if err != nil {
		return err
	}
	var lat, long float64
	lat, err = strconv.ParseFloat(dataFields[9].FirstChild.Data, 64)
	if err != nil {
		return err
	}
	long, err = strconv.ParseFloat(dataFields[10].FirstChild.Data, 64)
	if err != nil {
		return err
	}
	point := geom.NewPointFlat(geom.XY, []float64{lat, long})
	s.Location = &ewkb.Point{Point: point}
	return nil

}
