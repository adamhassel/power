package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/adamhassel/power/entities/config"
	"github.com/adamhassel/power/httpapi"
	"github.com/adamhassel/power/repos/eloverblik"
)

var confFile string
var port int

func init() {
	flag.StringVar(&confFile, "c", "power.conf", "location of configuration file.")
	flag.IntVar(&port, "p", 8080, "port to listen on")
}
func main() {
	flag.Parse()
	c, err := config.LoadConfig(confFile)
	if err != nil {
		log.Fatalf("error reading conf: %s", err)
	}
	if c.MID() == "" || c.Token() == "" {
		log.Fatal("MID or Token invalid")
	}
	if err := eloverblik.PreloadTariffs(c); err != nil {
		log.Fatalf("error preloading tariffs: %s", err)
	}
	http.HandleFunc("/powerPrices", httpapi.GetPowerPrices)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
