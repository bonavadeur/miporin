package scraper

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/url"
	"reflect"

	"github.com/bonavadeur/miporin/pkg/bonalib"
)

func Query(query string) map[string]interface{} {
	query = url.QueryEscape(query)

	resp, err := http.Get(PROMSERVER + query)
	if err != nil {
		bonalib.Warn("err", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		bonalib.Warn("err", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		bonalib.Warn("err", err)
	}

	return result
}

func WeightedNegative(array []int32) []int32 {
	weightedArray := make([]int32, len(array))
	var sum float64
	for _, value := range array {
		if value == 0 {
			sum += 1.0 / float64(0.1)
		} else {
			sum += 1.0 / float64(value)
		}
	}
	weightedArray[len(array)-1] = 100
	var _w float64
	for i := range weightedArray {
		if i != len(array)-1 {
			if array[i] == 0 {
				_w = math.Round((1.0 / float64(0.1)) / float64(sum) * 100)
			} else {
				_w = math.Round((1.0 / float64(array[i])) / float64(sum) * 100)
			}
			if math.IsNaN(_w) {
				weightedArray[i] = 0.0
			} else {
				weightedArray[i] = int32(_w)
			}
			weightedArray[len(array)-1] -= weightedArray[i]
		}
	}
	return weightedArray
}

func WeightedPositive(array []int32) []int32 {
	weightedArray := make([]int32, len(array))
	var sum float64
	for _, value := range array {
		sum += float64(value)
	}
	weightedArray[len(array)-1] = 100
	var _w float64
	for i := range weightedArray {
		if i != len(array)-1 {
			_w = math.Round((float64(array[i])) / float64(sum) * 100)
			weightedArray[i] = int32(_w)
			weightedArray[len(array)-1] -= weightedArray[i]
		}
	}
	if reflect.DeepEqual(array, make([]int32, len(NODENAMES))) {
		weightedArray = make([]int32, len(NODENAMES))
	}
	return weightedArray
}
