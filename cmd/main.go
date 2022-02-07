package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/adamhassel/power"
	"github.com/adamhassel/power/entities"
	"github.com/adamhassel/power/repos/eloverblik"
	"github.com/adamhassel/power/repos/energidataservice"
)

var conffile string
var noofhours uint

func init() {
	flag.UintVar(&noofhours, "h", 12, "Number of hours to get price data for. Default: 12")
	flag.StringVar(&conffile, "c", "power.conf", "location of configuration file. Default \"power.conf\"")
}
func main() {
	flag.Parse()
	var conf entities.Config
	if err := conf.Load(conffile); err != nil {
		log.Fatal("error reading conf: %s", err)
	}
	if conf.MID == "" || conf.Token == "" {
		log.Fatal("MID or Token invalid")
	}

	var e energidataservice.EnergiDataService
	e.Timer(time.Now().Truncate(time.Hour), time.Now().Truncate(time.Hour).Add(time.Duration(noofhours)*time.Hour))
	p, err := e.Query()
	if err != nil {
		log.Fatal(err)
	}

	var t eloverblik.Eloverblik
	if err := t.Authenticate([]byte(conf.Token)); err != nil {
		log.Fatal(err)
	}
	if err := t.Identify([]byte(conf.MID)); err != nil {
		log.Fatal(err)
	}
	ft, err := t.Query()
	if err != nil {
		log.Fatal(err)
	}

	combined := power.Summarize(p.(energidataservice.Prices), ft.(eloverblik.FullTariffs))
	output, _ := json.MarshalIndent(combined, "", "  ")

	fmt.Println(string(output))
}
