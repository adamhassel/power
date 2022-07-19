package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/adamhassel/power"
	"github.com/adamhassel/power/interfaces"
	"github.com/adamhassel/power/repos/eloverblik"
	"github.com/adamhassel/power/repos/energidataservice"
)

func writeReply(w http.ResponseWriter, body string, status int) {
	w.WriteHeader(status)
	w.Write([]byte(body))
}

func GetPowerPricesConfigHandler(c interfaces.Configurator, ignoreMissingTariffs bool) func(http.ResponseWriter, *http.Request) {
	return GetPowerPrices(c, ignoreMissingTariffs)
}

// GetPowerPrices is a handler to fetch and display power prices
// * handler to return power data
// * cache tariffs in mem to not have to get them all the time.
func GetPowerPrices(c interfaces.Configurator, ignoreMissingTariffs bool) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		// default, get 12 hours
		h := 12

		// refresh tariffs unconditionally if they're more than 24 hrs old
		if time.Now().Sub(eloverblik.FullTariffsCached.UpdatedAt()) > 24*time.Hour {
			if err := eloverblik.PreloadTariffs(c); err != nil {
				if !ignoreMissingTariffs {
					writeReply(w, err.Error(), http.StatusBadGateway)
					return
				}
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
		now := time.Now().Truncate(time.Hour)
		p, err := power.Prices(now, now.Add(time.Duration(h)*time.Hour), c, ignoreMissingTariffs)

		//p, err := getSpotPrices(h)
		if err != nil {
			writeReply(w, err.Error(), http.StatusBadGateway)
			return
		}
		/*
			combined := power.Summarize(p, eloverblik.FullTariffsCached)

		*/
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
			o := make([]Simple, len(p))
			for i, p := range p {
				var suffix string
				if len(p.Taxes) == 0 {
					suffix = "*"
				}
				o[i].Period = fmt.Sprintf("%s - %s", p.ValidFrom.Format("15:04"), p.ValidTo.Format("15:04"))
				o[i].Price = fmt.Sprintf("%0.2f%s kr.", p.TotalIncVAT, suffix)
			}
			renderJson(w, o)
			return
		}
		renderJson(w, p)
		return
	}
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
