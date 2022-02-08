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
var pretty, simple bool

func init() {
	flag.UintVar(&noOfHours, "h", 12, "Number of hours to get price data for.")
	flag.StringVar(&confFile, "c", "power.conf", "location of configuration file.")
	flag.BoolVar(&pretty, "p", false, "pretty-print (indent) JSON output.")
	flag.BoolVar(&simple, "s", false, "simple data output, only period and total price.")
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
	var combined interface{}
	combined = power.Summarize(p.(energidataservice.Prices), ft.(eloverblik.FullTariffs))

	if simple {
		type Simple struct {
			Period string `json:"period"`
			Price  string `json:"price"`
		}
		data := combined.([]entities.FullPrice)
		o := make([]Simple, len(data))
		for i, p := range data {
			o[i].Period = fmt.Sprintf("%s - %s", p.ValidFrom.Format("15:04"), p.ValidTo.Format("15:04"))
			o[i].Price = fmt.Sprintf("%0.2f kr.", p.TotalIncVAT)
		}
		combined = o
	}

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
