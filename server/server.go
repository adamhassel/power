package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/adamhassel/power"
	"github.com/adamhassel/power/entities/config"
	"github.com/adamhassel/power/repos/eloverblik"
	"github.com/adamhassel/power/repos/energidataservice"
)

var confFile string
var port int

var FullTariffs eloverblik.FullTariffs
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
	if err := preloadTariffs(); err != nil {
		log.Fatalf("error preloading tariffs: %s", err)
	}
	http.HandleFunc("/powerPrices", getPowerPrices)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func preloadTariffs() error {
	var t eloverblik.Eloverblik
	conf := config.GetConf()
	if err := t.Authenticate([]byte(conf.Token())); err != nil {
		return err
	}
	if err := t.Identify([]byte(conf.MID())); err != nil {
		return err
	}
	ft, err := t.Query()
	if err != nil {
		return err
	}
	FullTariffs = ft.(eloverblik.FullTariffs)
	return nil
}

func writeReply(w http.ResponseWriter, body string, status int) {
	w.WriteHeader(status)
	w.Write([]byte(body))
}

 // * handler to return power data
 // * cache tariffs in mem to not have to get them all the time.
func getPowerPrices(w http.ResponseWriter, req *http.Request) {
	// default, get 12 hours
	h := 12

	// refresh tariffs once a day
	if time.Now().Sub(FullTariffs.UpdatedAt()) > 24*time.Hour {
		if err := preloadTariffs(); err != nil {
			writeReply(w, err.Error(), http.StatusBadGateway)
			return
		}
	}
	params := req.URL.Query()
	if hours, ok := params["hours"]; ok {
		if len(hours) > 0 {
			var err error
			if h, err = strconv.Atoi(hours[0]); err != nil {
				writeReply(w, fmt.Sprintf("Error parsing %s as integer", hours[0]), http.StatusBadRequest)
				return
			}
		}
	}
	p, err := getSpotPrices(h)
	if err != nil {
		writeReply(w, err.Error(), http.StatusBadGateway)
		return
	}
	combined := power.Summarize(p, FullTariffs)
	var simple bool
	if s, ok := params["simple"]; ok {
		if len(s) > 0 && s[0] != "" {
			simple = true
		}
	}
	if simple {
			type Simple struct {
			Period string `json:"period"`
			Price  string `json:"price"`
		}
		data := combined
		o := make([]Simple, len(data))
		for i, p := range data {
			o[i].Period = fmt.Sprintf("%s - %s", p.ValidFrom.Format("15:04"), p.ValidTo.Format("15:04"))
			o[i].Price = fmt.Sprintf("%0.2f kr.", p.TotalIncVAT)
		}
		renderJson(w, o)
		return
	}
	renderJson(w, combined)
	return
}

func renderJson(w http.ResponseWriter, data interface{}) {
	output, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(output)
}

func getSpotPrices(noOfHours int) (energidataservice.Prices, error) {
	var e energidataservice.EnergiDataService
	e.Area(energidataservice.AreaDKEast)
	e.Timer(time.Now(), time.Now().Add(time.Duration(noOfHours)*time.Hour))
	p, err := e.Query()
	if err != nil {
		return energidataservice.Prices{}, err
	}
	return p.(energidataservice.Prices), nil
}