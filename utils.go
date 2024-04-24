package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"

	"microservice/types"
)

// doInsecurePull checks the error and evaluates if an insecure pull is possible
// and returns the result of the insecure pull
func doInsecurePull(err error) (*http.Response, error) {
	insecureHttpClient := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		var certErr *tls.CertificateVerificationError
		if errors.As(urlErr, &certErr) {
			var x509Err x509.CertificateInvalidError
			if errors.As(certErr.Err, &x509Err) {
				if x509Err.Reason == x509.Expired {
					log.Warn().Msg("server certificate expired. retrying access without certificate verification")
					res, err := insecureHttpClient.Get(crawlUrl)
					if err != nil {
						return nil, fmt.Errorf("unable to execute insecure pull: %w", err)
					}
					return res, nil
				} else {
					return nil, fmt.Errorf("certificate is invalid due to other reasons than expiry: %w", err)
				}
			}
		}
	}
	return nil, err
}

func downloadTablePage() (*goquery.Document, error) {
	httpClient := http.Client{}
	res, err := httpClient.Get(crawlUrl)
	if err != nil {
		res, err = doInsecurePull(err)
		if err != nil {
			log.Error().Err(err).Msg("unable to fetch measurement page")
			return nil, err
		}
	}
	return goquery.NewDocumentFromReader(res.Body)
}

func readTable(page *goquery.Document) (stations []types.Station, measurements []types.Measurement, err error) {
	rows := page.Find(`tr[id^="` + tableID + `"]`)
	var parseErrors []error
	rows.Each(func(i int, row *goquery.Selection) {
		dataFields := row.Find("td").Nodes
		station := types.Station{}
		err = station.FromDataFields(dataFields)
		if err != nil {
			parseErrors = append(parseErrors, err)
			return
		}
		stations = append(stations, station)
		measurement := types.Measurement{}
		err = measurement.FromDataFields(dataFields)
		if err != nil {
			parseErrors = append(parseErrors, err)
			return
		}
		measurements = append(measurements, measurement)
	})
	if len(parseErrors) != 0 {
		var concatErr error
		for _, e := range parseErrors {
			concatErr = fmt.Errorf("%w, %w", concatErr, e)
		}
		return nil, nil, concatErr
	}
	return stations, measurements, nil
}
