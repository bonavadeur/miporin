package miporin

import (
	"context"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Cm2Bool(data string) bool {
	configMap, err := CLIENTSET.
		CoreV1().
		ConfigMaps("default").
		Get(context.TODO(), "config-ikukantai", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	dataFromCm := configMap.Data[data]
	ret, err := strconv.ParseBool(dataFromCm)
	if err != nil {
		panic(err.Error())
	}
	return ret
}
