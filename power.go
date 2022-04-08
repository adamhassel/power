package power

import (
	"time"

	"github.com/adamhassel/power/entities"
	"github.com/adamhassel/power/interfaces"
	"github.com/adamhassel/power/repos/eloverblik"
	"github.com/adamhassel/power/repos/energidataservice"
)

type FullPrices []entities.FullPrice

// Summarize will combine the information in spot and t into a list of FullPrices
func Summarize(spot interfaces.SpotPricer, t interfaces.Indexer) FullPrices {
	var fp = make([]entities.FullPrice, len(spot.SpotPrices()))
	idx := t.Index()
	for i, p := range spot.SpotPrices() {
		h := p.HourUTC.Local().Hour()
		taxes := idx.AtPos(h).Taxes()
		taxesSubTotal := taxes.Total()
		// Price data is per MWh, so let's make that per kWh
		rawPrice := *p.SpotPriceDKK / 1000
		fp[i] = entities.FullPrice{
			Taxes:         taxes,
			ValidFrom:     p.HourUTC.Local(),
			ValidTo:       p.HourUTC.Add(time.Hour).Local(),
			Estimated:     p.DKKEstimated,
			EstimatedRate: p.EstimatedRate,
			RawPrice:      rawPrice,
			TaxesSubTotal: taxesSubTotal,
			Total:         taxesSubTotal + rawPrice,
			TotalIncVAT:   (taxesSubTotal + rawPrice) * 1.25,
		}
	}
	return fp
}

// Prices fetches price data from `from` to `to`, for the given `mid` using the `token` for auth.
func Prices(from, to time.Time, mid, token string) (FullPrices, error){
	var e energidataservice.EnergiDataService
	e.Area(energidataservice.AreaDKEast)
	e.Timer(from, to)
	p, err := e.Query()
	if err != nil {
		return nil, err
	}

	var t eloverblik.Eloverblik
	if err := t.Authenticate([]byte(token)); err != nil {
		return nil, err
	}
	if err := t.Identify([]byte(mid)); err != nil {
		return nil, err
	}
	ft, err := t.Query()
	if err != nil {
		return nil, err
	}

	combined := Summarize(p.(energidataservice.Prices), ft.(eloverblik.FullTariffs))
	return combined, nil
}