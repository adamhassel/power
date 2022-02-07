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

var confFile string
var noOfHours uint
var pretty bool

func init() {
	flag.UintVar(&noOfHours, "h", 12, "Number of hours to get price data for.")
	flag.StringVar(&confFile, "c", "power.conf", "location of configuration file.")
	flag.BoolVar(&pretty, "p", false, "pretty-print (indent) JSON output.")
}
func main() {
	flag.Parse()
	var conf entities.Config
	if err := conf.Load(confFile); err != nil {
		log.Fatalf("error reading conf: %s", err)
	}
	if conf.MID == "" || conf.Token == "" {
		log.Fatal("MID or Token invalid")
	}

	var e energidataservice.EnergiDataService
	e.Area(energidataservice.AreaDKEast)
	e.Timer(time.Now(), time.Now().Add(time.Duration(noOfHours)*time.Hour))
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
	var output []byte
	if pretty {
		output, err = json.MarshalIndent(combined, "", "  ")
	} else {
		output, err = json.Marshal(combined)
	}
	if err != nil {
		log.Fatalf("error marshalling result: %s", err)
	}

	fmt.Print(string(output))
}
