package main

import (
	log "github.com/sirupsen/logrus"
)

// this is a half-hearted attempt to calculate an at least somewhat useful rating for the day's weather
// the logic is not sound, and it's some of the most shameful code I've written
// it lives in its own file so i don't have to look at it unnecessarily
func calculateWeatherScore(temp float64, wind float64, cloud int, z string, logger *log.Logger) int {
	var v int

	// these variable names are awful
	minBound := 40.0
	maxBound := 75.0
	windBreezeMax := 5.0
	cloudOnHotDayMin := 40
	cloudOnColdDayMax := 20

	cloudOnBetweenMin := 0
	cloudOnBetweenMax := 40
	/*
		A pleasant day = 5, it must be between N and M degrees.

		If over M, look at cloud coverage, more is better. Look at wind speed for slight breeze (< 5 MPH)

		If under N, cloud are bad, wind is also bad

		If between, wind over W is bad

	*/

	scoreLogger := logger.WithFields(
		log.Fields{
			"zip":   z,
			"temp":  temp,
			"cloud": cloud,
			"wind":  wind,
		},
	)

	if minBound <= temp && temp <= maxBound {
		scoreLogger.Debug("between")
		ws := 0
		// add a point if there's a nice breeze
		if wind < windBreezeMax {
			scoreLogger.Debug("breeze +1")
			ws += 1
		}
		// deduct points for bad wind
		if wind > 15 {
			scoreLogger.Debug("wind -1")
			ws -= 1
		}
		if wind > 20 {
			scoreLogger.Debug("wind -1")
			ws -= 1
		}

		cs := 0
		if cloud < cloudOnBetweenMin || cloud > cloudOnBetweenMax {
			cs -= 1
			scoreLogger.Debug("cloud -1")
		} else {
			scoreLogger.Debug("cloud +1")
			cs += 1
		}
		// deduct points for being too cloudy when it's cold
		if temp < 50 {
			if cloud > 60 {
				scoreLogger.Debug("cloud + temp -1")
				cs -= 1
			}
		}

		ww := 0
		if temp >= 50 && temp <= 70 {
			// add points for a really nice day
			scoreLogger.Debug("temp +2")
			ww += 2
		}

		// 5 points for being in between
		v = 5 + ws + cs + ww
	} else if temp < minBound {
		// cold day logic
		scoreLogger.Debug("cold day")
		ws := 0
		// add points for no wind on cold day
		if wind < windBreezeMax {
			scoreLogger.Debug("wind +1")
			ws += 1
		}

		// deduct points for bad wind
		if wind > 12 {
			scoreLogger.Debug("wind -1")
			ws -= 1
		}
		if wind > 20 {
			scoreLogger.Debug("wind -1")
			ws -= 1
		}

		cs := 0
		if cloud <= cloudOnColdDayMax {
			scoreLogger.Debug("cloud +1")
			cs += 1
		} else {
			scoreLogger.Debug("cloud +1")
			cs -= 1
		}

		// deduct points for being too cloudy when it's cold
		if cloud > 60 {
			scoreLogger.Debug("cloud -1")
			cs -= 1
		}

		ww := 0
		// deduct an extra point for it being too cold
		if temp < 30 {
			scoreLogger.Debug("temp -1")
			ww -= 1
		}

		// 3 points for being cold + 1 for not too windy + 1 for not too cloudy
		v = 3 + ws + cs + ww
	} else {
		// warm day logic
		scoreLogger.Debug("hot day")
		ws := 0
		// moderate breeze, add
		if wind <= 12 {
			scoreLogger.Debug("wind +1")
			ws += 1
		} else {
			scoreLogger.Debug("wind -1")
			ws -= 1
		}

		// deduct points for bad wind
		if wind > 20 {
			scoreLogger.Debug("wind -1")
			ws -= 1
		}
		if wind > 25 {
			scoreLogger.Debug("wind -1")
			ws -= 1
		}

		cs := 0
		if cloud > cloudOnHotDayMin {
			scoreLogger.Debug("cloud +1")
			cs += 1
		} else {
			scoreLogger.Debug("cloud -1")
			cs -= 1
		}
		// deduct points for not enough clouds when hot
		if temp > maxBound {
			if cloud < 30 {
				scoreLogger.Debug("wind and temp -1")
				cs -= 1
			}
		}

		ww := 0
		// deduct an extra point for it being too hot
		if temp > 85 {
			scoreLogger.Debug("temp -1")
			ww -= 1
		}
		if temp > 95 {
			scoreLogger.Debug("temp -1")
			ww -= 1
		}

		// 3 points for being hot + 1 for not too windy + 1 for not too cloudy
		v = 3 + ws + cs + ww
	}

	return v
}
