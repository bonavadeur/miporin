package scraper

import (
	"context"
	"fmt"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	"github.com/bonavadeur/miporin/pkg/miporin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

var _ = bonalib.Baka()

var (
	PROMSERVER   = "http://prometheus-kube-prometheus-prometheus:9090/api/v1/query?query="
	NODENAMES    []string
	WEIGHT       [][]int
	SLEEPTIME    = 2
	RESPONSETIME [][]int
	DYNCLIENT    = miporin.GetDynamicClient()
)

func init() {
	NODENAMES = miporin.GetNodenames()
	WEIGHT = make([][]int, len(NODENAMES))
}

func Scraper(OKASAN_SCRAPER map[string]*OkasanScraper) {
	okasan := NewOkasanScraper("okaasan", "10", int8(2))
	OKASAN_SCRAPER["okaasan"] = okasan
	go watchEventCreateKsvc(okasan)
	// child := NewKodomoScraper("hello", "10", int8(2))
	// okasan.AddKodomo(child)
	go okasan.Scrape()
}

func watchEventCreateKsvc(okasan *OkasanScraper) {
	namespace := "default"

	ksvcGVR := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services",
	}

	watcher, err := DYNCLIENT.Resource(ksvcGVR).Namespace(namespace).Watch(context.TODO(), metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		fmt.Println(err)
		panic(err.Error())
	}

	for event := range watcher.ResultChan() {
		if event.Type == watch.Added {
			ksvc, _ := event.Object.(*unstructured.Unstructured)
			ksvcName, _, _ := unstructured.NestedString(ksvc.Object, "metadata", "name")
			child := NewKodomoScraper(ksvcName, "10", int8(2))
			okasan.AddKodomo(child)
			bonalib.Warn("Ksvc has been created:", ksvcName)
		}
		if event.Type == watch.Deleted {
			ksvc, _ := event.Object.(*unstructured.Unstructured)
			ksvcName, _, _ := unstructured.NestedString(ksvc.Object, "metadata", "name")
			okasan.DeleteKodomo(ksvcName)
			bonalib.Warn("Ksvc has been deleted:", ksvcName)
		}
	}
}

// func ScrapeMetrics() {
// 	for {
// 		servingTime := scrapeServingTime()
// 		podOnNode := scrapePodOnNode()
// 		latency := scrapeLatency()
// 		estimatedResponseTime := libs.AddMatrix(servingTime, latency)

// 		w := make([][]int, 3)
// 		for i, row := range estimatedResponseTime {
// 			w[i] = WeightedNegative(row)
// 		}

// 		_sumPods := 0
// 		for nodename := range podOnNode {
// 			_sumPods += podOnNode[nodename]
// 		}

// 		if _sumPods == 0 { // PoN == [0, 0, 0]
// 			w = [][]int{
// 				{100, 0, 0},
// 				{0, 100, 0},
// 				{0, 0, 100},
// 			}
// 		} else {
// 			for i := range w {
// 				for j := range w[i] {
// 					if podOnNode[NODENAMES[j]] == 0 {
// 						w[i][j] = 0
// 					}
// 					if podOnNode[NODENAMES[j]] != 0 && w[i][j] == 0 {
// 						w[i][j] = 1
// 					}
// 				}
// 			}
// 			for i, row := range w {
// 				w[i] = WeightedPositive(row)
// 			}
// 		}

// 		WEIGHT = w
// 		RESPONSETIME = estimatedResponseTime
// 		// bonalib.Succ("WEIGHT", WEIGHT)

// 		time.Sleep(time.Duration(SLEEPTIME) * time.Second)
// 	}
// }

// func scrapeServingTime() [][]int {
// 	servingTimeRaw := Query("rate(revision_request_latencies_sum[10s])/rate(revision_request_latencies_count[10s])")
// 	servingTimeResult := servingTimeRaw["data"].(map[string]interface{})["result"].([]interface{})

// 	servingTimeLine := make([][]int, len(NODENAMES))

// 	for _, stResult := range servingTimeResult {
// 		ip := strings.Split(stResult.(map[string]interface{})["metric"].(map[string]interface{})["instance"].(string), ":")[0]
// 		_servingTime := libs.String2RoundedInt(stResult.(map[string]interface{})["value"].([]interface{})[1].(string))
// 		_inNode := miporin.CheckIPInNode(ip)
// 		for i, node := range NODENAMES {
// 			if _inNode == node {
// 				servingTimeLine[i] = append(servingTimeLine[i], _servingTime)
// 			}
// 		}
// 	}

// 	servingTimeRow := make([]int, len(NODENAMES))
// 	for i, stl := range servingTimeLine {
// 		servingTimeRow[i] = libs.Average(stl)
// 	}

// 	servingTime := make([][]int, len(NODENAMES))
// 	for i := range servingTime {
// 		servingTime[i] = servingTimeRow
// 	}
// 	return servingTime
// }

// func scrapePodOnNode() map[string]int {
// 	pods, err := miporin.CLIENTSET.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
// 	if err != nil {
// 		panic(err)
// 	}

// 	podOnNode := map[string]int{}
// 	for _, node := range NODENAMES {
// 		podOnNode[node] = 0
// 	}

// 	for _, pod := range pods.Items {
// 		if pod.Status.Phase != "Terminating" && pod.Status.Phase != "Pending" && strings.Contains(pod.Name, "hello") {
// 			podOnNode[pod.Spec.NodeName]++
// 		}
// 	}

// 	return podOnNode
// }

// func scrapeLatency() [][]int {
// 	latencyRaw := Query("avg_over_time(latency_between_nodes[10s])")
// 	latencyResult := latencyRaw["data"].(map[string]interface{})["result"].([]interface{})

// 	latency := make([][]int, 3)
// 	for i := range latency {
// 		latency[i] = []int{0, 0, 0}
// 	}

// 	nodeIndex := map[string]int{}
// 	for i, node := range NODENAMES {
// 		nodeIndex[node] = i
// 	}

// 	for _, lr := range latencyResult {
// 		lrMetric := lr.(map[string]interface{})["metric"].(map[string]interface{})
// 		lrValue := libs.String2RoundedInt(lr.(map[string]interface{})["value"].([]interface{})[1].(string))
// 		latency[nodeIndex[lrMetric["from"].(string)]][nodeIndex[lrMetric["to"].(string)]] = lrValue
// 	}

// 	return latency
// }
