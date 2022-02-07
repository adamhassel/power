package energidataservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/adamhassel/power/entities"
)

const dataServiceUrl = "https://data-api.energidataservice.dk/v1/graphql"
const queryTemplate = `{"operationName":"Dataset","variables":{},"query":"query Dataset {\n  elspotprices(\n    where: {HourDK: {_gte: \"%s\", _lt: \"%s\"}, PriceArea: {_eq: \"%s\"}}\n    order_by: {HourUTC: asc}\n    limit: %d\n    offset: %d\n  ) {\n    HourUTC\n    HourDK\n    PriceArea\n    SpotPriceDKK\n    SpotPriceEUR\n    __typename\n  }\n}\n"}`

type area string

const (
	areaDKWest area = "DK1"
	areaDKEast area = "DK2"
)

const (
	defaultLimit  = 100
	defaultOffset = 0
)

type EnergiDataService struct {
	area     area
	from, to time.Time
	p        Prices
}

// Prices is the data returned from  energidataservice, containing raw power prices
type Prices struct {
	Data struct {
		Elspotprices []entities.Elspotprice `json:"elspotprices"`
	} `json:"data"`
}

func (p Prices) SpotPrices() []entities.Elspotprice {
	return p.Data.Elspotprices
}

func (e *EnergiDataService) Timer(from, to time.Time) {
	e.from = from
	e.to = to
}

func (e *EnergiDataService) Query() (interface{}, error) {
	e.p = Prices{}
	if e.from.IsZero() {
		e.from = time.Now()
	}
	if e.to.IsZero() || e.to.Before(e.from) {
		e.to = e.from.Add(12 * time.Hour)
	}
	if err := e.p.query(e.from, e.to); err != nil {
		return nil, err
	}
	return e.p, nil
}

func (p *Prices) query(from, to time.Time) error {
	if err := p.getRawSpotPrices(from, to); err != nil {
		return err
	}
	return p.fixupDKK()
}

func (p *Prices) getRawSpotPrices(from, to time.Time) error {
	body := makeSpotPriceQuery(from, to)

	resp, err := http.Post(dataServiceUrl, "application/json", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.Status)
	}
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(response, &p)
}

// on weekends, there are no DKK prices, since there are no exchange rates. After
// the weekend, they're set retroactively. But we don't have future vision, so
// for any prices with only a euro price, we'll use the last record with DKK and
// EUR from before the weekend, and derive an exchange rate from that, which
// we'll use.
func (p *Prices) fixupDKK() error {
	// find out if there are any missing DKK..
	var latestEUR, latestDKK float64
	emptyDKK := false
	for _, sp := range p.Data.Elspotprices {
		if sp.SpotPriceDKK == nil {
			emptyDKK = true
		} else if *sp.SpotPriceDKK != 0 && sp.SpotPriceEUR != 0 {
			latestEUR = sp.SpotPriceEUR
			latestDKK = *sp.SpotPriceDKK
		}
	}
	// All DKK prices filled,
	if !emptyDKK {
		return nil
	}

	// if we already have prices (like, a friday), use that
	if latestEUR == 0 && latestDKK == 0 {
		// We know the first price is the earliest, since that's how it's sorted
		earliest := p.Data.Elspotprices[0].HourUTC

		var target time.Time
		// Check if we're a saturday or a sunday, and find Friday @ 2300 local
		switch earliest.Weekday() {
		case time.Monday:
			target = earliest.Add(-72 * time.Hour)
		case time.Sunday:
			target = earliest.Add(-48 * time.Hour)
		case time.Saturday:
			target = earliest.Add(-24 * time.Hour)
		}
		if target.Weekday() != time.Friday {
			return errors.New("Not friday??")
		}
		// fetch the last price of the friday, set time to 2300 local
		target = time.Date(target.Year(), target.Month(), target.Day(), 23, 0, 0, 0, earliest.Location())
		var rateprice Prices
		rateprice.getRawSpotPrices(target, target.Add(time.Hour))
		if rateprice.Data.Elspotprices[0].SpotPriceDKK == nil {
			return errors.New("DKK price was nil")
		}
		latestDKK = *rateprice.Data.Elspotprices[0].SpotPriceDKK
		latestEUR = rateprice.Data.Elspotprices[0].SpotPriceEUR
	}

	rate := latestDKK / latestEUR

	// finally, fixup prices where appropriate
	for i, price := range p.Data.Elspotprices {
		if price.SpotPriceDKK == nil {
			p.Data.Elspotprices[i].SpotPriceDKK = new(float64)
			*p.Data.Elspotprices[i].SpotPriceDKK = price.SpotPriceEUR * rate
			p.Data.Elspotprices[i].DKKEstimated = true
			p.Data.Elspotprices[i].EstimatedRate = rate
		}
	}
	return nil
}

func makeSpotPriceQuery(start, end time.Time) string {
	return fmt.Sprintf(queryTemplate, start.Local().Format("2006-01-02T15:04:05"), end.Local().Format("2006-01-02T15:04:05"), areaDKEast, defaultLimit, defaultOffset)
}