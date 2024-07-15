package scraper

import (
	"context"
	"strings"
	"time"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	"github.com/bonavadeur/miporin/pkg/libs"
	"github.com/bonavadeur/miporin/pkg/miporin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KodomoScraper struct {
	Name      string
	Metrics   Metrics
	Window    string // seconds
	sleepTime int8   // seconds
	Okasan    *OkasanScraper
	PodOnNode map[string]int32
	Weight    [][]int32
	Quit	  chan bool
}

type Metrics struct {
	servt [][]int32
	respt [][]int32
}

func NewKodomoScraper(name string, window string, sleepTime int8) *KodomoScraper {
	atarashiiKodomoScraper := &KodomoScraper{
		Name:      name,
		Metrics:   *NewMetrics(),
		Window:    window,
		sleepTime: sleepTime,
		Okasan:    nil,
		PodOnNode: map[string]int32{},
		Weight:    make([][]int32, len(NODENAMES)),
		Quit:      make(chan bool),
	}

	for _, nodename := range NODENAMES {
		atarashiiKodomoScraper.PodOnNode[nodename] = int32(0)
	}

	return atarashiiKodomoScraper
}

func NewMetrics() *Metrics {
	newMetrics := &Metrics{
		servt: [][]int32{},
		respt: [][]int32{},
	}

	return newMetrics
}

func (k *KodomoScraper) Scrape() {
	for {
        select {
        case <- k.Quit:
            return
        default:
			k.scrapeServingTime()
			k.scrapePodOnNode()
			bonalib.Info("KodomoScraper", k.Name, k.Metrics.servt, k.Okasan.Latency, k.PodOnNode)
			k.Metrics.respt = libs.AddMatrix(k.Metrics.servt, k.Okasan.Latency)

			w := make([][]int32, len(NODENAMES))
			for i, row := range k.Metrics.respt {
				w[i] = WeightedNegative(row)
			}

			_sumPods := int32(0)
			for nodename := range k.PodOnNode {
				_sumPods += k.PodOnNode[nodename]
			}

			if _sumPods == 0 { // PoN == [0, 0, 0]
				w = [][]int32{
					{100, 0, 0},
					{0, 100, 0},
					{0, 0, 100},
				}
			} else {
				for i := range w {
					for j := range w[i] {
						if k.PodOnNode[NODENAMES[j]] == 0 {
							w[i][j] = 0
						}
						if k.PodOnNode[NODENAMES[j]] != 0 && w[i][j] == 0 {
							w[i][j] = 1
						}
					}
				}
				for i, row := range w {
					w[i] = WeightedPositive(row)
				}
			}

			k.Weight = w
			// RESPONSETIME = estimatedResponseTime
			bonalib.Succ("WEIGHTED", k.Weight)

			time.Sleep(time.Duration(k.sleepTime) * time.Second)
        }
    }
	for {
		k.scrapeServingTime()
		k.scrapePodOnNode()
		bonalib.Info("KodomoScraper", k.Name, k.Metrics.servt, k.Okasan.Latency, k.PodOnNode)
		k.Metrics.respt = libs.AddMatrix(k.Metrics.servt, k.Okasan.Latency)

		w := make([][]int32, len(NODENAMES))
		for i, row := range k.Metrics.respt {
			w[i] = WeightedNegative(row)
		}

		_sumPods := int32(0)
		for nodename := range k.PodOnNode {
			_sumPods += k.PodOnNode[nodename]
		}

		if _sumPods == 0 { // PoN == [0, 0, 0]
			w = [][]int32{
				{100, 0, 0},
				{0, 100, 0},
				{0, 0, 100},
			}
		} else {
			for i := range w {
				for j := range w[i] {
					if k.PodOnNode[NODENAMES[j]] == 0 {
						w[i][j] = 0
					}
					if k.PodOnNode[NODENAMES[j]] != 0 && w[i][j] == 0 {
						w[i][j] = 1
					}
				}
			}
			for i, row := range w {
				w[i] = WeightedPositive(row)
			}
		}

		k.Weight = w
		// RESPONSETIME = estimatedResponseTime
		bonalib.Succ("WEIGHTED", k.Weight)

		time.Sleep(time.Duration(k.sleepTime) * time.Second)
	}
}

func (k *KodomoScraper) scrapeServingTime() {
	servingTimeRaw := Query("rate(revision_request_latencies_sum{service_name=\"" + k.Name + "\"}[" + k.Window + "s])/rate(revision_request_latencies_count{service_name=\"" + k.Name + "\"}[" + k.Window + "s])")
	servingTimeResult := servingTimeRaw["data"].(map[string]interface{})["result"].([]interface{})

	servingTimeLine := make([][]int32, len(NODENAMES))

	for _, stResult := range servingTimeResult {
		ip := strings.Split(stResult.(map[string]interface{})["metric"].(map[string]interface{})["instance"].(string), ":")[0]
		_servingTime := libs.String2RoundedInt(stResult.(map[string]interface{})["value"].([]interface{})[1].(string))
		_inNode := miporin.CheckIPInNode(ip)
		for i, node := range NODENAMES {
			if _inNode == node {
				servingTimeLine[i] = append(servingTimeLine[i], _servingTime)
			}
		}
	}

	servingTimeRow := make([]int32, len(NODENAMES))
	for i, stl := range servingTimeLine {
		servingTimeRow[i] = libs.Average(stl)
	}

	servingTime := make([][]int32, len(NODENAMES))
	for i := range servingTime {
		servingTime[i] = servingTimeRow
	}

	k.Metrics.servt = servingTime
}

func (k *KodomoScraper) scrapePodOnNode() {
	pods, err := miporin.CLIENTSET.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	podOnNode := map[string]int32{}
	for _, node := range NODENAMES {
		podOnNode[node] = 0
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != "Terminating" && pod.Status.Phase != "Pending" && strings.Contains(pod.Name, "hello") {
			podOnNode[pod.Spec.NodeName]++
		}
	}

	k.PodOnNode = podOnNode
}
