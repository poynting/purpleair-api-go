package purpleair

import "math"

func lerp(ylo float64, yhi float64, xlo float64, xhi float64, x float64) float64 {
	return ((x-xlo)/(xhi-xlo))*(yhi-ylo) + ylo
}

func Pm25ToAqi(pm25 float64) int {
	c := math.Floor(10.0*pm25) / 10
	var a float64
	if c < 0 {
		a = 0
	} else if c < 12.1 {
		a = lerp(0, 50, 0.0, 12.0, c)
	} else if c < 35.5 {
		a = lerp(51, 100, 12.1, 35.4, c)
	} else if c < 55.5 {
		a = lerp(101, 150, 35.5, 55.4, c)
	} else if c < 150.5 {
		a = lerp(151, 200, 55.5, 150.4, c)
	} else if c < 250.5 {
		a = lerp(201, 300, 150.5, 250.4, c)
	} else if c < 350.5 {
		a = lerp(301, 400, 250.5, 350.4, c)
	} else if c < 500.5 {
		a = lerp(401, 500, 350.5, 500.4, c)
	} else {
		a = 500 // values above 500 are considered beyond AQI
	}
	return int(math.Round(a))
}

func Radians(angle_degrees float64) float64 {
	return (angle_degrees * math.Pi / 180.)
}

func Degrees(angle_radians float64) float64 {
	return (angle_radians * 180. / math.Pi)
}

func KmToNm(km float64) float64 {
	return km / 1.852
}

func NmToKm(nm float64) float64 {
	return nm * 1.852
}

func distanceKmToRadians(d_km float64) float64 {
	return (math.Pi / (180. * 60.)) * KmToNm(d_km)
}

func distanceRadiansToKm(d_rad float64) float64 {
	return NmToKm((180. * 60. / math.Pi) * d_rad)
}

// http://edwilliams.org/avform147.htm
// A point {lat,lon} is a distance distance_km out on the bearing_rad radial from point 1
// updated this formula to have west longitudes be negative
func PointFromLocRadial(lat1 float64, lon1 float64, distance_km float64, bearing_rad float64) (float64, float64) {
	d_rad := distanceKmToRadians(distance_km) // radial distance
	lat := math.Asin(math.Sin(lat1)*math.Cos(d_rad) + math.Cos(lat1)*math.Sin(d_rad)*math.Cos(bearing_rad))
	dlon := math.Atan2(math.Sin(bearing_rad)*math.Sin(d_rad)*math.Cos(lat1), math.Cos(d_rad)-math.Sin(lat1)*math.Sin(lat))
	lon := math.Mod(lon1+dlon+math.Pi, 2.*math.Pi) - math.Pi
	return lat, lon
}
