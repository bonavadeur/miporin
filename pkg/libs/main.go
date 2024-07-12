package libs

import (
	"math"
	"strconv"
)

func init() {

}

func AddMatrix(MatA [][]int, MatB [][]int) [][]int {
	MatC := make([][]int, 3)
	for i := range MatC {
		MatC[i] = make([]int, 3)
		for j := range MatC[i] {
			MatC[i][j] = MatA[i][j] + MatB[i][j]
		}

	}
	return MatC
}

func String2RoundedInt(s string) int {
	floatValue, _ := strconv.ParseFloat(s, 32)
	if math.IsNaN(floatValue) {
		floatValue = 0.0
	}
	intValue := int(math.Round(floatValue))
	return intValue
}

func Average(slice []int) int {
	sum := 0
	for _, value := range slice {
		sum += value
	}
	ret := math.Round(float64(sum) / float64(len(slice)))
	if math.IsNaN(ret) {
		return int(0.0)
	} else {
		return int(ret)
	}
}
