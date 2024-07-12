package bonalib

import (
	"context"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"fmt"

	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Baka() string {
	return "Baka"
}

func RandNumber() string {
	return strconv.Itoa(rand.Intn(1000))
}

// 1;31: red
// 1;32: green
// 1;33: yellow
// 1;34: blue
// 1;35: purple

func Log(msg string, obj ...interface{}) {
	if msg == "" {
		msg = "-"
	}

	color := "\033[1;33m%v\033[0m" // yellow
	fmt.Printf("\033[1;33m0---bonaLog %v \033[0m", msg)

	for _, v := range obj {
		fmt.Printf("\033[1;33m%v \033[0m", v)
	}

	color = "\033[0m%v\033[0m" // reset
	fmt.Printf(color, "\n\n")
}

func Succ(msg string, obj ...interface{}) {
	if msg == "" {
		msg = "-"
	}

	color := "\033[1;32m%v\033[0m" // yellow
	fmt.Printf("\033[1;32m0---bonaLog %v \033[0m", msg)

	for _, v := range obj {
		fmt.Printf("\033[1;32m%v \033[0m", v)
	}

	color = "\033[0m%v\033[0m" // reset
	fmt.Printf(color, "\n\n")
}

func Warn(msg string, obj ...interface{}) {
	if msg == "" {
		msg = "-"
	}

	color := "\033[1;31m%v\033[0m" // yellow
	fmt.Printf("\033[1;31m0---bonaLog %v \033[0m", msg)

	for _, v := range obj {
		fmt.Printf("\033[1;31m%v \033[0m", v)
	}

	color = "\033[0m%v\033[0m" // reset
	fmt.Printf(color, "\n\n")
}

func Info(msg string, obj ...interface{}) {
	if msg == "" {
		msg = "-"
	}

	color := "\033[1;34m%v\033[0m" // yellow
	fmt.Printf("\033[1;34m0---bonaLog %v \033[0m", msg)

	for _, v := range obj {
		fmt.Printf("\033[1;34m%v \033[0m", v)
	}

	color = "\033[0m%v\033[0m" // reset
	fmt.Printf(color, "\n\n")
}

func Vio(msg string, obj ...interface{}) {
	if msg == "" {
		msg = "-"
	}

	color := "\033[1;35m%v\033[0m" // yellow
	fmt.Printf("\033[1;35m0---bonaLog %v \033[0m", msg)

	for _, v := range obj {
		fmt.Printf("\033[1;35m%v \033[0m", v)
	}

	color = "\033[0m%v\033[0m" // reset
	fmt.Printf(color, "\n\n")
}

func Line() {
	fmt.Printf("\n\n\n")
}

func Use(variable ...interface{}) {}

func Type(variable interface{}) string {
	return reflect.TypeOf(variable).String()
}

func Size(variable interface{}) int {
	size := unsafe.Sizeof(variable)
	return int(size)
}

func Logln(msg string, obj interface{}) {
	// 32: green 33: yellow
	if msg == "" {
		msg = "-"
	}
	if obj == "" {
		obj = "-"
	}
	color := "\033[1;33m%v\033[0m" // yellow
	str := spew.Sprintln("0---bonaLog", msg, obj)
	fmt.Printf(color, str)
	color = "\033[0m%v\033[0m" // reset
	fmt.Printf(color, "-------------------------------\n\n")
}

func Str2Int(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return -1
	}
	return num
}

func Cm2IntSlice(namespace string, configmapName string, data string) []int {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configmapName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	arrayData := configMap.Data[data]
	var returnSlice []int
	err = yaml.Unmarshal([]byte(arrayData), &returnSlice)
	if err != nil {
		panic(err.Error())
	}
	return returnSlice
}

func Cm2StringSlice(namespace string, configmapName string, data string) []string {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configmapName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	arrayData := configMap.Data[data]
	var returnSlice []string
	err = yaml.Unmarshal([]byte(arrayData), &returnSlice)
	if err != nil {
		panic(err.Error())
	}
	return returnSlice
}

func Cm2Int(data string) int {
	result, err := strconv.Atoi(os.Getenv(data))
	if err != nil {
		panic(err.Error())
	}
	return result
}

func Cm2String(data string) string {
	return os.Getenv(data)
}

func Cm2IntMatrix(namespace string, configmap string, data string) [][]int {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configmap, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	matrixData := strings.Split(configMap.Data[data], "\n")

	var matrix [][]int
	for _, row := range matrixData {
		var matrixRow []int
		numbers := strings.Split(strings.Trim(row, "- \"\t"), ",")
		if numbers[0] != "" {
			for _, number := range numbers {
				if len(number) != 0 {
					num, err := strconv.Atoi(number)
					if err != nil {
						panic(err.Error())
					}
					matrixRow = append(matrixRow, num)
				}
			}
			matrix = append(matrix, matrixRow)
		}
	}
	return matrix
}

func Cm2Bool(data string) bool {
	ret, err := strconv.ParseBool(Cm2String(data))
	if err != nil {
		panic(err.Error())
	}
	return ret
}
