package main

import "fmt"

// this is a half-hearted attempt to calculate an at least somewhat useful rating for the day's weather
// the logic is not sound, and it's some of the most shameful code I've written
// it lives in its own file so i don't have to look at it unnecessarily
func calculateWeatherScore(temp float64, wind float64, cloud int, z string) int {
	var v int

	// these variable names are awful
	minBound := 40.0
	maxBound := 75.0
	windBreezeMax := 5.0
	cloudOnHotDayMin := 40
	cloudOnColdDayMax := 20

	cloudOnBetweenMin := 0
	cloudOnBetweenMax := 60

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
