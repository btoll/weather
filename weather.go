package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	FORECAST_URL = "https://api.open-meteo.com/v1/forecast?"
	GEO_URL      = "https://geocoding-api.open-meteo.com/v1/search?"
)

var (
	client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			ResponseHeaderTimeout: 3 * time.Second,
			IdleConnTimeout:       30 * time.Second,
			MaxIdleConns:          40,
			MaxConnsPerHost:       20,
			MaxIdleConnsPerHost:   20,
			TLSHandshakeTimeout:   5 * time.Second,
		},
	}
)

type GeoResult struct {
	City        string   `json:"name"`
	State       string   `json:"admin1"`
	County      string   `json:"admin2"`
	CountryCode string   `json:"country_code"`
	PostCodes   []string `json:"postcodes"`
	Timezone    string   `json:"timezone"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
}

type GeoResults struct {
	Results []*GeoResult `json:"results"`
	State   string
	Err     error
}

type Forecast struct {
	City     string
	State    string
	Timezone string `json:"timezone"`
	Current  struct {
		Temperature float64 `json:"temperature_2m"`
		WindSpeed   float64 `json:"wind_speed_10m"`
	} `json:"current"`
	Err error
}

func GetForecast(g GeoResult) *Forecast {
	params := url.Values{}
	params.Add("latitude", strconv.FormatFloat(g.Latitude, 'g', -1, 64))
	params.Add("longitude", strconv.FormatFloat(g.Longitude, 'g', -1, 64))
	params.Add("current", "weather_code,temperature_2m,wind_speed_10m")
	params.Add("timezone", g.Timezone)
	params.Add("temperature_unit", "fahrenheit")
	params.Add("wind_speed_unit", "mph")
	params.Add("precipitation_unit", "inch")

	fresults := &Forecast{
		City:  g.City,
		State: g.State,
	}
	resp, err := client.Get(FORECAST_URL + params.Encode())
	if err != nil {
		fresults.Err = err
		return fresults
	}
	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fresults.Err = err
		return fresults
	}
	err = json.Unmarshal(b, fresults)
	if err != nil {
		fresults.Err = err
		return fresults
	}
	return fresults
}

func GetLocation(city, state string) *GeoResults {
	params := url.Values{}
	params.Add("count", "50")
	params.Add("name", city)
	params.Add("language", "en")
	params.Add("countryCode", "US")

	gresults := &GeoResults{State: state}
	resp, err := client.Get(GEO_URL + params.Encode())
	if err != nil {
		gresults.Err = err
		return gresults
	}
	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		gresults.Err = err
		return gresults
	}
	err = json.Unmarshal(b, gresults)
	if err != nil {
		gresults.Err = err
		return gresults
	}
	return gresults
}
