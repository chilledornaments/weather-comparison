package main

import (
	"context"
	owm "github.com/briandowns/openweathermap"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	configFileName = "weather_comparison.yaml"
)

type Runner struct {
	C            Config
	W            *owm.CurrentWeatherData
	HttpClient   *http.Client
	InfluxClient influxdb2.Client
	Logger       *log.Logger
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
	Debug bool `mapstructure:"debug"`
}

type weather struct {
	clouds     int
	weatherIDs []int
	temp       float64
	windSpeed  float64
}

func newLogger(debug bool) *log.Logger {
	l := log.New()

	l.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	l.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	l.SetLevel(log.InfoLevel)

	if debug {
		l.SetLevel(log.DebugLevel)
	}

	return l
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

func (c *Runner) getWeatherForZip(zip string) (weather, error) {
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

func (c *Runner) storeData(w weather, z string) error {
	a := c.InfluxClient.WriteAPIBlocking(c.C.Influx.Org, c.C.Influx.Bucket)

	p := influxdb2.NewPoint("stat",
		map[string]string{"zip": z},
		map[string]interface{}{"clouds": w.clouds, "temp": w.temp, "wind_speed": w.windSpeed, "weather_id": w.weatherIDs[0], "weather_score_v3": calculateWeatherScore(w.temp, w.windSpeed, w.clouds, z, c.Logger)},
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
				c.Logger.WithError(err).WithField("zip", zz).Error("error getting weather for zip")
			} else {
				// only run this logic if we've retrieved weather data
				if err = c.storeData(w, zz); err != nil {
					c.Logger.WithError(err).WithField("zip", zz).Error("error writing to influx")
				} else {
					c.Logger.WithField("zip", zz).Info("success")
				}
			}
			wg.Done()
		}(z)
	}
	wg.Wait()
}
