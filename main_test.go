package main

import (
	"os"
	"reflect"
	"testing"
)

const (
	testConfigFileName = "weather_comparison_test.yaml"
)

func Test_parseConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		generateConfig(t)

		t.Cleanup(func() {
			os.Unsetenv("WC_CONFIG_FILE_NAME")
		})

		os.Setenv("WC_CONFIG_FILE_NAME", testConfigFileName)

		got := parseConfig()
		var want Config

		want.ZipCodes = []string{"80303", "80404"}
		want.Database.Host = "foo"
		want.OWMApiKey = "hello"
		if !reflect.DeepEqual(got, want) {
			t.Errorf("parseConfig() = %v, want %v", got, want)
		}

	})

}

func generateConfig(t *testing.T) {
	t.Helper()

	s := `zip_codes:
  - "80303"
  - "80404"
db:
  host: "foo"
owm_api_key: "hello"
influx:
  token: "xxx"
  org: ""
  bucket: ""
  url: ""`

	fh, err := os.Create(testConfigFileName)

	if err != nil {
		t.Error(err)
	}

	_, err = fh.Write([]byte(s))
	if err != nil {
		t.Error(err)
	}

}
