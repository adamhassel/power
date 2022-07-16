package power

import (
	"log"
	"time"

	"github.com/adamhassel/power/entities"
	"github.com/adamhassel/power/interfaces"
	"github.com/adamhassel/power/repos/eloverblik"
	"github.com/adamhassel/power/repos/energidataservice"

	"github.com/adamhassel/errors"
)

type FullPrices struct {
	Contents []entities.FullPrice
	From     time.Time
	To       time.Time
}

var FullPricesCached FullPrices

// InRange returns true if fb contains data in the full range from - to
func (fp FullPrices) InRange(from, to time.Time) bool {
	if to.Before(from) {
		return false
	}
	return !fp.From.After(from) && !fp.To.Before(to)
}

// Range returns the subset of fp that are between from and to. If out of range, returns empty
func (fp FullPrices) Range(from, to time.Time) FullPrices {
	var rv FullPrices
	var fset bool // Flag to signal that "From" has been set in the return value.
	rv.Contents = make([]entities.FullPrice, 0)
	for _, f := range fp.Contents {
		// if this entry is not within the range specified, skip it
		if !f.InWindow(from, to) {
			continue
		}
		if !fset && rv.From.Before(f.ValidFrom) {
			rv.From = f.ValidFrom
			fset = true
		}
		if rv.To.Before(f.ValidTo) {
			rv.To = f.ValidTo
		}
		rv.Contents = append(rv.Contents, f)
	}
	return rv
}

// Summarize will combine the information in spot and t into a list of FullPrices
func Summarize(spot interfaces.SpotPricer, t interfaces.Indexer) FullPrices {
	var fp = make([]entities.FullPrice, len(spot.SpotPrices()))
	idx := t.Index()
	var (
		to   time.Time
		from time.Time = time.Now().Add(365 * 24 * time.Hour)
	)
	var fromset bool
	for i, p := range spot.SpotPrices() {
		h := time.Time(p.HourUTC).Local().Hour()
		taxes := idx.AtPos(h).Taxes()
		taxesSubTotal := taxes.Total()
		// Price data is per MWh, so let's make that per kWh
		rawPrice := *p.SpotPriceDKK / 1000
		fp[i] = entities.FullPrice{
			Taxes:         taxes,
			ValidFrom:     time.Time(p.HourUTC).Local(),
			ValidTo:       time.Time(p.HourUTC).Add(time.Hour).Local(),
			Estimated:     p.DKKEstimated,
			EstimatedRate: p.EstimatedRate,
			RawPrice:      rawPrice,
			TaxesSubTotal: taxesSubTotal,
			Total:         taxesSubTotal + rawPrice,
			TotalIncVAT:   (taxesSubTotal + rawPrice) * 1.25,
		}
		if !fromset && fp[i].ValidFrom.Before(from) {
			from = fp[i].ValidFrom
			fromset = true
		}
		if fp[i].ValidTo.After(to) {
			to = fp[i].ValidTo
		}
	}

	return FullPrices{
		Contents: fp,
		From:     from,
		To:       to,
	}
}

var ErrEloverblik = errors.New("error getting data from eloverblik.dk")

// Prices fetches price data from `from` and as far ahead as they're available, for the given `mid` using the
// `token` for auth. If 'IgnoreMissingTariffs' is true, just return spot prices
// without tariffs, if they can't be fetched.
func Prices(from, to time.Time, c interfaces.Configurator, ignoreMissingTariffs bool) ([]entities.FullPrice, error) {
	// return cached prices if available
	if FullPricesCached.InRange(from, to) {
		return FullPricesCached.Range(from, to).Contents, nil
	}
	var e energidataservice.EnergiDataService
	e.Area(energidataservice.AreaDKEast)
	// always fetch until tomorrow at midnight. If they're not ready yet, the service will return as much as is can.
	end := time.Now().Truncate(24 * time.Hour).Add(48 * time.Hour)
	e.Timer(from, end)
	p, err := e.Query()
	if err != nil {
		return nil, errors.Wrap(err, ErrEloverblik)
	}

	// refresh tariffs once a day
	if time.Now().Sub(eloverblik.FullTariffsCached.UpdatedAt()) > 24*time.Hour {
		if err := eloverblik.PreloadTariffs(c); err != nil {
			if ignoreMissingTariffs {
				log.Printf("encountered error %s, but ignoring", err)
				err = nil
			} else {
				return nil, err
			}
		}
	}
	if err != nil {
		return nil, err
	}

	FullPricesCached = Summarize(p.(energidataservice.Prices), eloverblik.FullTariffsCached)
	return FullPricesCached.Range(from, to).Contents, nil
}
