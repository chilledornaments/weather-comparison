# weather-comparison

This is silly project to gather long(ish)term data on weather across different areas. I want to spend some time in different parts of the country, and weather is an important factor.

Because I'm lazy, I've put together some basic logic to give each day a rating on a scale of ?-9 (:shrug:).  

## Requirements

### OpenWeather Map

Sign up for a free [OpenWeather Map](https://openweathermap.org/api) account and generate an API token.

### InfluxDB

Sign up for a free [InfluxDB cloud account](https://cloud2.influxdata.com/signup) (or run it yourself, I don't care). Create a bucket, grab its name, your org ID, an API token, and the URL.

## Config file

```yaml
zip_codes:
  - 11111
  - 22222
owm_api_key: "zzz"
influx:
  token: "foo"
  org: "bar"
  bucket: "e.g. weather-comparison"
  url: "https://us-west-2-1.aws.cloud2.influxdata.com"
```

## Tests

`go test -v ./...`
