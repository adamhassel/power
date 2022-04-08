package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/adamhassel/power"
	"github.com/adamhassel/power/entities/config"
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
	var conf config.Config
	if err := conf.Load(confFile); err != nil {
		log.Fatalf("error reading conf: %s", err)
	}
	if conf.MID() == "" || conf.Token() == "" {
		log.Fatal("MID or Token invalid")
	}


	prices, err := power.Prices(time.Now(), time.Now().Add(time.Duration(noOfHours)*time.Hour), conf.MID(), conf.Token())
	if err != nil {
		log.Fatal(err)
	}
	var data interface{}
	if simple {
		type Simple struct {
			Period string `json:"period"`
			Price  string `json:"price"`
		}
		o := make([]Simple, len(prices))
		for i, p := range prices {
			o[i].Period = fmt.Sprintf("%s - %s", p.ValidFrom.Format("15:04"), p.ValidTo.Format("15:04"))
			o[i].Price = fmt.Sprintf("%0.2f kr.", p.TotalIncVAT)
		}
		data = o
	} else {
		data = prices
	}

	var output []byte
	if pretty {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		log.Fatalf("error marshalling result: %s", err)
	}

	fmt.Print(string(output))
}
