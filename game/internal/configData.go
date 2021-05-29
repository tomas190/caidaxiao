package internal

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	RECODE_CHAOCHUXIANHONG  = "4444"
	RECODE_DOWNBETMONEYFULL = "1001" // 房间限红
	RECODE_DOWNBETLIMITBET  = "1002" // 投注无效

)

func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.6f", value), 64)
	return value
}

func RandInRange(min int, max int) int {
	time.Sleep(1 * time.Nanosecond)
	return rand.Intn(max-min) + min
}

func RandFloatNum() float64 {
	slice := []float64{0.5, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45,
		0.5, 0.55, 0.6, 0.65, 0.7, 0.75, 0.8, 0.85, 0.9, 0.95}
	n := rand.Intn(len(slice))
	return slice[n]
}

func getNextTime() string {
	timeLayout := "2006-01-02 15:04" //转化所需模板
	timestamp := time.Now().Unix()
	datetime := time.Unix(timestamp, 0).Format(timeLayout)
	return datetime + ":00"
}

func SetPackageTaxM(packageT uint16, tax float64) {
	packageTax[packageT] = tax

	for _, v := range hall.roomList {
		if v != nil {
			v.PackageId = packageT
		}
	}
}

func getMinute(timeStr string) int {
	a1 := strings.Split(timeStr, " ")
	a2 := strings.Split(a1[1], ":")
	m, e := strconv.Atoi(a2[1])
	if e != nil {
		return -1
	}
	return m
}
