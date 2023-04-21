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

func calculatePleasantness(temp float64, wind float64, cloud int, z string) int {
	var v int

	// these variable names are awful
	minBound := 40.0
	maxBound := 75.0
	windBreezeMax := 5.0
	cloudOnHotDayMin := 40
	cloudOnColdDayMax := 20
	// TODO this should scale based on temp
	// Min only makes sense when scaling
	cloudOnBetweenMin := 0
	cloudOnBetweenMax := 50

	between := false

	isTooWindy := false
	isTooCloudy := false
	/*
		A pleasant day = 5, it must be between N and M degrees.

		If over M, look at cloud coverage, more is better. Look at wind speed for slight breeze (< 5 MPH)

		If under N, cloud are bad, wind is also bad

		If between, wind over W is bad

	*/

	if minBound >= temp && temp <= maxBound {
		between = true
	}

	if between {
		if wind > windBreezeMax {
			isTooWindy = true
		}

		if cloud < cloudOnBetweenMin || cloud > cloudOnBetweenMax {
			isTooCloudy = true
		}

		ws := 0
		if !isTooWindy {
			ws += 1
		}
		// deduct points for bad wind
		if wind > 15 {
			ws -= 1
		}
		if wind > 20 {
			ws -= 1
		}

		cs := 0
		if !isTooCloudy {
			cs += 1
		}
		// deduct points for too cloudy when it's cold
		if temp < 50 {

			if cloud > 60 {
				cs -= 1
			}
		}

		ww := 0
		if temp >= 50 && temp <= 70 {
			ww += 2
		}

		// 5 points for being in between
		v = 5 + ws + cs + ww
	} else if temp < minBound {
		// cold day logic
		if cloud > cloudOnColdDayMax {
			isTooCloudy = true
		}

		if wind > windBreezeMax {
			isTooWindy = true
		}

		if cloud > cloudOnColdDayMax {
			isTooCloudy = true
		}

		ws := 0
		if !isTooWindy {
			ws += 1
		}
		// deduct points for bad wind
		if wind > 15 {
			ws -= 1
		}
		if wind > 20 {
			ws -= 1
		}

		cs := 0
		if !isTooCloudy {
			cs += 1
		}
		// deduct points for too cloudy when it's cold
		if temp < minBound {
			if cloud > 60 {
				cs -= 1
			}
		}

		ww := 0
		// deduct an extra point for it being too cold
		if temp < 30 {
			ww -= 1
		}

		// 3 points for being cold + 1 for not too windy + 1 for not too cloudy
		v = 3 + ws + cs + ww
	} else {
		// hot day logic
		if cloud < cloudOnHotDayMin {
			// it's not really too cloudy, but we deduct a point this way
			isTooCloudy = true
		}

		if wind > 12 || wind < windBreezeMax {
			// not windy enough or too windy, either way deduct a point
			isTooWindy = true
		}

		ws := 0
		if !isTooWindy {
			ws += 1
		}
		// deduct points for bad wind
		if wind > 20 {
			ws -= 1
		}
		if wind > 25 {
			ws -= 1
		}

		cs := 0
		if !isTooCloudy {
			cs += 1
		}
		// deduct points for not enough clouds when hot
		if temp > maxBound {
			if cloud < 30 {
				cs -= 1
			}
		}

		ww := 0
		// deduct an extra point for it being too hot
		if temp > 85 {
			ww -= 1
		}
		if temp > 95 {
			ww -= 1
		}

		// 3 points for being hot + 1 for not too windy + 1 for not too cloudy
		v = 3 + ws + cs + ww
	}

	// debugging
	fmt.Printf("%s: cloud=%d wind=%f temp=%f score=%d\n", z, cloud, wind, temp, v)

	return v
}

func (c *Client) storeData(w weather, z string) error {
	a := c.InfluxClient.WriteAPIBlocking(c.C.Influx.Org, c.C.Influx.Bucket)

	p := influxdb2.NewPoint("stat",
		map[string]string{"zip": z},
		map[string]interface{}{"clouds": w.clouds, "temp": w.temp, "wind_speed": w.windSpeed, "weather_id": w.weatherIDs[0], "weather_score_v2": calculatePleasantness(w.temp, w.windSpeed, w.clouds, z)},
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
				panic(err)
			}
			if err = c.storeData(w, zz); err != nil {
				fmt.Println("error writing to influx - ", err.Error())
			} else {
				fmt.Println("success for ", zz)
			}
			wg.Done()
		}(z)
	}
	wg.Wait()
}
