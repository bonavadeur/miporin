package scraper

import (
	"github.com/bonavadeur/miporin/pkg/libs"
)

type OkasanScraper struct {
	Name      string
	PodOnNode map[string]int32
	Latency   [][]int32
	Window    string
	sleepTime int8
	Kodomo    map[string]*KodomoScraper
}

func NewOkasanScraper(name string, window string, sleepTime int8) *OkasanScraper {
	atarashiiOkasanScraper := &OkasanScraper{
		Name:      name,
		PodOnNode: map[string]int32{},
		Latency:   [][]int32{},
		Window:    window,
		sleepTime: sleepTime,
		Kodomo:    map[string]*KodomoScraper{},
	}

	for _, nodename := range NODENAMES {
		atarashiiOkasanScraper.PodOnNode[nodename] = int32(0)
	}

	return atarashiiOkasanScraper
}

func (o *OkasanScraper) Scrape() {
	go o.scrapeLatency()
	o.scrapeServingTime()
}

func (o *OkasanScraper) server() {

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
	}
}

func (o *OkasanScraper) scrapeServingTime() {
	for _, ks := range o.Kodomo {
		go ks.Scrape()
	}
}

func (o *OkasanScraper) AddKodomo(kodomo *KodomoScraper) {
	kodomo.Okasan = o
	go kodomo.Scrape()
	o.Kodomo[kodomo.Name] = kodomo
}

func (o *OkasanScraper) DeleteKodomo(kodomo string) {
	o.Kodomo[kodomo].Quit <- true
	o.Kodomo[kodomo] = nil
	delete(o.Kodomo, kodomo)
}
