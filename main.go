package main

import (
	"context"
	"fmt"
	owm "github.com/briandowns/openweathermap"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"time"
)

const (
	configFileName = "weather_comparison.yaml"
)

type Client struct {
	C            Config
	W            *owm.CurrentWeatherData
	HttpClient   *http.Client
	InfluxClient influxdb2.Client
}

type Config struct {
	ZipCodes []string `mapstructure:"zip_codes"`
	Database struct {
		Host string `mapstructure:"host"`
	} `mapstructure:"db"`
	OWMApiKey string `mapstructure:"owm_api_key"`
	Influx    struct {
		Token  string `mapstructure:"token"`
		Bucket string `mapstructure:"bucket"`
		Org    string `mapstructure:"org"`
		URL    string `mapstructure:"url"`
	} `mapstructure:"influx"`
}

type weather struct {
	clouds     int
	weatherIDs []int
	temp       float64
	windSpeed  float64
}

func parseConfig() Config {
	o := os.Getenv("WC_CONFIG_FILE_NAME")
	if o == "" {
		o = configFileName
	}
	viper.SetConfigType("yaml")
	viper.SetConfigName(o)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	var c Config

	err = viper.Unmarshal(&c)
	if err != nil {
		panic(err)
	}

	return c
}

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

func (c *Client) getWeatherForZip(zip string) (weather, error) {
	var r weather

	err := c.W.CurrentByZipcode(zip, "US")
	if err != nil {
		return r, err
	}

	r.clouds = c.W.Clouds.All
	r.temp = c.W.Main.Temp
	r.windSpeed = c.W.Wind.Speed
	for _, i := range c.W.Weather {
		r.weatherIDs = append(r.weatherIDs, i.ID)
	}

	return r, nil
}

func (c *Client) storeData(w weather, z string) error {
	a := c.InfluxClient.WriteAPIBlocking(c.C.Influx.Org, c.C.Influx.Bucket)

	p := influxdb2.NewPoint("stat",
		map[string]string{"zip": z},
		map[string]interface{}{"clouds": w.clouds, "temp": w.temp, "wind_speed": w.windSpeed, "weather_id": w.weatherIDs[0]},
		time.Now(),
	)

	err := a.WritePoint(context.TODO(), p)

	return err
}

func main() {
	conf := parseConfig()
	c, _ := newClient(&conf)

	c.newOWMClient()
	c.newInfluxClient(&conf)

	for _, z := range conf.ZipCodes {
		w, err := c.getWeatherForZip(z)
		if err != nil {
			panic(err)
		}
		if err = c.storeData(w, z); err != nil {
			fmt.Println("error writing to influx - ", err.Error())
		} else {
			fmt.Println("success for ", z)
		}
	}
}
