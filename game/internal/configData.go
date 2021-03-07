package internal

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.6f", value), 64)
	return value
}

func RandInRange(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(1 * time.Nanosecond)
	return rand.Intn(max-min) + min
}

func RandFloatNum() float64 {
	slice := []float64{0.5, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45,
		0.5, 0.55, 0.6, 0.65, 0.7, 0.75, 0.8, 0.85, 0.9, 0.95}
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(slice))
	return slice[n]
}
