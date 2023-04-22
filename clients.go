package main

import (
	owm "github.com/briandowns/openweathermap"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"net/http"
)

func newClient(c *Config) (*Runner, error) {
	var cc Runner

	cc.C = *c
	cc.HttpClient = http.DefaultClient

	cc.Logger = newLogger(c.Debug)

	return &cc, nil
}

func (c *Runner) newOWMClient() error {
	w, err := owm.NewCurrent("F", "EN", c.C.OWMApiKey, owm.WithHttpClient(c.HttpClient))

	if err != nil {
		return err
	}

	c.W = w

	return nil
}

func (c *Runner) newInfluxClient(co *Config) error {
	c.InfluxClient = influxdb2.NewClient(co.Influx.URL, co.Influx.Token)
	return nil
}
