package main

import (
	owm "github.com/briandowns/openweathermap"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"net/http"
)

func newClient(c *Config) (*Client, error) {
	var cc Client

	cc.C = *c
	cc.HttpClient = http.DefaultClient

	return &cc, nil
}

func (c *Client) newOWMClient() error {
	w, err := owm.NewCurrent("F", "EN", c.C.OWMApiKey, owm.WithHttpClient(c.HttpClient))

	if err != nil {
		return err
	}

	c.W = w

	return nil
}

func (c *Client) newInfluxClient(co *Config) error {
	c.InfluxClient = influxdb2.NewClient(co.Influx.URL, co.Influx.Token)
	return nil
}
