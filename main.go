package main

import (
	"context"
	"fmt"
	owm "github.com/briandowns/openweathermap"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"sync"
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
	ZipCodes  []string `mapstructure:"zip_codes"`
	OWMApiKey string   `mapstructure:"owm_api_key"`
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
		map[string]interface{}{"clouds": w.clouds, "temp": w.temp, "wind_speed": w.windSpeed, "weather_id": w.weatherIDs[0], "weather_score_v2": calculateWeatherScore(w.temp, w.windSpeed, w.clouds, z)},
		time.Now(),
	)

	err := a.WritePoint(context.TODO(), p)

	return err
}

func main() {
	conf := parseConfig()
	c, _ := newClient(&conf)

	if err := c.newOWMClient(); err != nil {
		panic(err)
	}

	if err := c.newInfluxClient(&conf); err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	for _, z := range conf.ZipCodes {
		wg.Add(1)

		go func(zz string) {
			w, err := c.getWeatherForZip(zz)
			if err != nil {
				fmt.Println("error getting weather for zip - ", err.Error())
			} else {
				// only run this logic if we've retrieved weather data
				if err = c.storeData(w, zz); err != nil {
					fmt.Println("error writing to influx - ", err.Error())
				} else {
					fmt.Println("success for ", zz)
				}
			}
			wg.Done()
		}(z)
	}
	wg.Wait()
}
