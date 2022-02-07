package power

import (
	"time"

	"github.com/adamhassel/power/entities"
	"github.com/adamhassel/power/interfaces"
)

// Summarize will combine the information in spot and t into a list of FullPrices
func Summarize(spot interfaces.SpotPricer, t interfaces.Indexer) []entities.FullPrice {
	var fp = make([]entities.FullPrice, 0, len(spot.SpotPrices()))
	i := t.Index()
	for _, p := range spot.SpotPrices() {
		h := p.HourUTC.Local().Hour()
		taxes := i.AtPos(h).Taxes()
		taxesSubTotal := taxes.Total()
		// Price data is per MWh, so let's make that per kWh
		rawPrice := *p.SpotPriceDKK / 1000
		fp = append(fp, entities.FullPrice{
			Taxes:         taxes,
			ValidFrom:     p.HourUTC.Local(),
			ValidTo:       p.HourUTC.Add(time.Hour).Local(),
			Estimated:     p.DKKEstimated,
			RawPrice:      rawPrice,
			TaxesSubTotal: taxesSubTotal,
			Total:         taxesSubTotal + rawPrice,
			TotalIncVAT:   (taxesSubTotal + rawPrice) * 1.25,
		})
	}

	return fp
}
