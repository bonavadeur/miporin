package libs

import (
	"math"
	"strconv"
	"time"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	"github.com/bonavadeur/miporin/pkg/miporin"
)

var (
	CLIENTSET = miporin.GetClientSet()
)

func init() {

}

func License() {
	for {
		targetDate, _ := time.Parse("02-01-2006", "15-11-2024")
		now := time.Now()
		if !now.Before(targetDate) {
			bonalib.Warn("This image is expired, contact to daodaihiep22ussr@gmail.com for extending license")
			panic("This image is expired, contact to daodaihiep22ussr@gmail.com for extending license")
		}
		time.Sleep(86400 * time.Second)
	}
}

func AddMatrix(MatA [][]int32, MatB [][]int32) [][]int32 {
	MatC := make([][]int32, 3)
	for i := range MatC {
		MatC[i] = make([]int32, 3)
		for j := range MatC[i] {
			MatC[i][j] = MatA[i][j] + MatB[i][j]
		}

	}
	return MatC
}

func String2RoundedInt(s string) int32 {
	floatValue, _ := strconv.ParseFloat(s, 32)
	if math.IsNaN(floatValue) {
		floatValue = 0.0
	}
	intValue := int32(math.Round(floatValue))
	return intValue
}

func Average(slice []int32) int32 {
	sum := int32(0)
	for _, value := range slice {
		sum += value
	}
	ret := math.Round(float64(sum) / float64(len(slice)))
	if math.IsNaN(ret) {
		return int32(0.0)
	} else {
		return int32(ret)
	}
}
