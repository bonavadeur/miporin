package scraper

import (
	"github.com/bonavadeur/miporin/pkg/bonalib"
	"github.com/bonavadeur/miporin/pkg/miporin"
)

var _ = bonalib.Baka()

var (
	PROMSERVER      = "http://prometheus-kube-prometheus-prometheus:9090/api/v1/query?query="
	NODENAMES       = miporin.GetNodenames()
	WEIGHT          = make([][]int, len(NODENAMES))
	SLEEPTIME       = 2
	RESPONSETIME    [][]int
	CLIENTSET       = miporin.GetClientSet()
	DYNCLIENT       = miporin.GetDynamicClient()
	OKASAN_SCRAPERS = map[string]*OkasanScraper{}
)

func init() {
	WEIGHT = make([][]int, len(NODENAMES))
}

func Scraper(OKASAN_SCRAPERS map[string]*OkasanScraper) {
	// create new okasan
	okasan := NewOkasanScraper("okaasan", "10", int8(2))

	// add okasan to OKASAN_SCRAPERS
	OKASAN_SCRAPERS["okaasan"] = okasan
}
