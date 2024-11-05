package main

import (
	"context"
	"net/http"

	"github.com/bonavadeur/miporin/pkg/bonalib"
	"github.com/bonavadeur/miporin/pkg/miporin"
	"github.com/bonavadeur/miporin/pkg/scraper"
	"github.com/bonavadeur/miporin/pkg/yukari"
	"github.com/labstack/echo/v4"
)

var (
	KUBECONFIG        = miporin.Kubeconfig()
	OKASAN_SCRAPERS   = map[string]*scraper.OkasanScraper{}
	OKASAN_SCHEDULERS = map[string]*yukari.OkasanScheduler{}
)

func init() {
	scraper.OKASAN_SCRAPERS = OKASAN_SCRAPERS
	yukari.OKASAN_SCRAPERS = OKASAN_SCRAPERS
	yukari.OKASAN_SCHEDULERS = OKASAN_SCHEDULERS
}

func main() {
	bonalib.Log("Konnichiwa, Miporin-chan desu")
	ctx := context.Background()

	// start scraper
	go scraper.Scraper(OKASAN_SCRAPERS)

	// start scheduler
	if miporin.Cm2Bool("ikukantai-miporin-enable-yukari") {
		// go yukari.Scheduler(OKASAN_SCHEDULERS)
	}

	// start echo server
	go server()

	// hangout forever
	<-ctx.Done()
}

func server() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Konnichiwa, Miporin-chan desu\n")
	})

	e.GET("/api/weight/okasan/:okasan/kodomo/:kodomo", func(c echo.Context) error {
		okasanScraper, ok := OKASAN_SCRAPERS[c.Param("okasan")]
		if ok {
			kodomoScraper, ok := okasanScraper.Kodomo[c.Param("kodomo")]
			if ok {
				return c.JSON(http.StatusOK, kodomoScraper.Weight)
			} else {
				return c.JSON(http.StatusNotFound, "NotFound")
			}
		} else {
			return c.JSON(http.StatusNotFound, "NotFound")
		}
	})

	e.Logger.Fatal(e.Start(":18080"))
}
