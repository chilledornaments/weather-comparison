package main

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func Test_calculateWeatherScore(t *testing.T) {
	logger := newLogger(true)

	type args struct {
		temp   float64
		wind   float64
		cloud  int
		z      string
		logger log.Logger
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "great day",
			args: args{
				temp:   68,
				wind:   4,
				cloud:  20,
				z:      "11111",
				logger: *logger,
			},
			want: 9,
		},
		{
			name: "best cold day",
			args: args{
				temp:   39.0,
				wind:   4,
				cloud:  20,
				z:      "11111",
				logger: *logger,
			},
			want: 5,
		},
		{
			name: "worst cold day",
			args: args{
				temp:   25.0,
				wind:   25,
				cloud:  100,
				z:      "11111",
				logger: *logger,
			},
			want: -2,
		},
		{
			name: "best warm day",
			args: args{
				temp:   76.0,
				wind:   5.0,
				cloud:  45,
				z:      "11111",
				logger: *logger,
			},
			want: 5,
		},
		{
			name: "worst hot day",
			args: args{
				temp:   100.0,
				wind:   30,
				cloud:  0,
				z:      "11111",
				logger: *logger,
			},
			want: -4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateWeatherScore(tt.args.temp, tt.args.wind, tt.args.cloud, tt.args.z, &tt.args.logger); got != tt.want {
				t.Errorf("calculateWeatherScore() = %v, want %v", got, tt.want)
			}
		})
	}
}
