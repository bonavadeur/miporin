package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	"github.com/bonavadeur/miporin/pkg/libs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

var _ = bonalib.Baka()

type OkasanScraper struct {
	Name      string
	Latency   [][]int32
	Window    string
	sleepTime int8
	Kodomo    map[string]*KodomoScraper
}

func NewOkasanScraper(
	name string,
	window string,
	sleepTime int8,
) *OkasanScraper {

	atarashiiOkasanScraper := &OkasanScraper{
		Name:      name,
		Latency:   [][]int32{},
		Window:    window,
		sleepTime: sleepTime,
		Kodomo:    map[string]*KodomoScraper{},
	}
	atarashiiOkasanScraper.init()

	// okasan scrape common metrics like: latency,
	go atarashiiOkasanScraper.scrape()

	// okasan watch ksvc create event to add or remove kodomo
	go atarashiiOkasanScraper.watchKsvcCreateEvent()

	return atarashiiOkasanScraper
}

func (o *OkasanScraper) init() {
	ksvcGVR := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services",
	}
	ksvcList, err := DYNCLIENT.Resource(ksvcGVR).Namespace("default").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		bonalib.Warn("Error listing Knative services", err)
	}
	for _, ksvc := range ksvcList.Items {
		ksvcName := ksvc.GetName()
		child := NewKodomoScraper(ksvcName, o.Window, o.sleepTime)
		o.addKodomo(child)
	}
}

func (o *OkasanScraper) scrape() {
	go o.scrapeLatency()
}

func (o *OkasanScraper) scrapeLatency() [][]int32 {
	for {
		latencyRaw := Query("avg_over_time(latency_between_nodes[" + o.Window + "s])")
		latencyResult := latencyRaw["data"].(map[string]interface{})["result"].([]interface{})

		latency := make([][]int32, len(NODENAMES))
		for i := range latency {
			latency[i] = make([]int32, len(NODENAMES))
		}

		nodeIndex := map[string]int32{}
		for i, node := range NODENAMES {
			nodeIndex[node] = int32(i)
		}

		for _, lr := range latencyResult {
			lrMetric := lr.(map[string]interface{})["metric"].(map[string]interface{})
			lrValue := libs.String2RoundedInt(lr.(map[string]interface{})["value"].([]interface{})[1].(string))
			latency[nodeIndex[lrMetric["from"].(string)]][nodeIndex[lrMetric["to"].(string)]] = lrValue
		}

		o.Latency = latency

		time.Sleep(time.Duration(o.sleepTime) * time.Second)
	}
}

func (o *OkasanScraper) watchKsvcCreateEvent() {
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
		ksvc, _ := event.Object.(*unstructured.Unstructured)
		ksvcName, _, _ := unstructured.NestedString(ksvc.Object, "metadata", "name")
		if event.Type == watch.Added {
			child := NewKodomoScraper(ksvcName, o.Window, int8(2))
			o.addKodomo(child)
			createServiceMonitor(ksvcName)
		}
		if event.Type == watch.Deleted {
			deleteServiceMonitor(ksvcName)
			o.deleteKodomo(ksvcName)
		}
	}
}

func (o *OkasanScraper) addKodomo(kodomo *KodomoScraper) {
	kodomo.Okasan = o
	o.Kodomo[kodomo.Name] = kodomo
}

func (o *OkasanScraper) deleteKodomo(kodomo string) {
	o.Kodomo[kodomo].ScrapeStop.Stop()
	o.Kodomo[kodomo] = nil
	delete(o.Kodomo, kodomo)
}
