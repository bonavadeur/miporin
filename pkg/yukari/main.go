package yukari

import (
	"github.com/bonavadeur/miporin/pkg/miporin"
	"github.com/bonavadeur/miporin/pkg/scraper"
)

var (
	NODENAMES         = miporin.GetNodenames()
	CLIENTSET         = miporin.GetClientSet()
	DYNCLIENT         = miporin.GetDynamicClient()
	MAXPON            = []int{10, 10, 3}
	OKASAN_SCRAPERS   = map[string]*scraper.OkasanScraper{}
	OKASAN_SCHEDULERS = map[string]*OkasanScheduler{}
)

func init() {

}

func Scheduler(OKASAN_SCHEDULERS map[string]*OkasanScheduler) {
	// create new OkasanScheduler
	okasan := NewOkasanScheduler("okaasan", int8(2))

	// add OkasanScheduler to OKASAN_SCHEDULERS
	OKASAN_SCHEDULERS["okaasan"] = okasan
}
