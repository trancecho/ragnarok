package util

func Round(value float64, precision int) float64 {
	if precision < 0 {
		return value
	}
	pow := 1.0
	for i := 0; i < precision; i++ {
		pow *= 10
	}
	return float64(int(value*pow+0.5)) / pow
}
