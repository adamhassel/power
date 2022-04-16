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
	"github.com/rickar/cal/v2"
	"github.com/rickar/cal/v2/dk"
)

const dataServiceUrl = "https://data-api.energidataservice.dk/v1/graphql"
const queryTemplate = `{"operationName":"Dataset","variables":{},"query":"query Dataset {\n  elspotprices(\n    where: {HourDK: {_gte: \"%s\", _lt: \"%s\"}, PriceArea: {_eq: \"%s\"}}\n    order_by: {HourUTC: asc}\n    limit: %d\n    offset: %d\n  ) {\n    HourUTC\n    HourDK\n    PriceArea\n    SpotPriceDKK\n    SpotPriceEUR\n    __typename\n  }\n}\n"}`

type area string

const (
	// AreaDKWest is for anyone living west of Storebælt
	AreaDKWest area = "DK1"
	// AreaDKEast is for anyone living east of Storebælt
	AreaDKEast area = "DK2"
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

func (e *EnergiDataService) Area(a area) {
	e.area = a
}

func (e *EnergiDataService) Timer(from, to time.Time) {
	e.from = from.Truncate(time.Hour)
	e.to = to.Truncate(time.Hour)
}

func (e *EnergiDataService) Query() (interface{}, error) {
	e.p = Prices{}
	if e.from.IsZero() {
		e.from = time.Now()
	}
	if e.to.IsZero() || e.to.Before(e.from) {
		e.to = e.from.Add(12 * time.Hour)
	}
	if err := e.p.query(e.from, e.to, e.area); err != nil {
		return nil, err
	}
	return e.p, nil
}

func (p *Prices) query(from, to time.Time, a area) error {
	if err := p.getRawSpotPrices(from, to, a); err != nil {
		return err
	}
	return p.fixupDKK(a)
}

func (p *Prices) getRawSpotPrices(from, to time.Time, a area) error {
	body := makeSpotPriceQuery(from, to, a)

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
func (p *Prices) fixupDKK(a area) error {
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
		// Check if we're a non-working day, and find Previsou workday @ 2300 local
		c := cal.NewBusinessCalendar()
		c.Calendar.AddHoliday(dk.Holidays...)
		target = prevWorkDay(c, earliest)
		// fetch the last price of the workday, set time to 2300 local
		target = time.Date(target.Year(), target.Month(), target.Day(), 23, 0, 0, 0, earliest.Location())
		var rateprice Prices
		if err := rateprice.getRawSpotPrices(target, target.Add(time.Hour), a); err != nil {
			return err
		}
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

func prevWorkDay(c *cal.BusinessCalendar, date time.Time) time.Time {
	t := date
	if date.Before(c.WorkdayStart(date)) {
		t = t.Add(-24 * time.Hour)
	}

	for !c.IsWorkday(t) {
		t = t.Add(-24 * time.Hour)
	}
	return c.WorkdayStart(t)
}

func makeSpotPriceQuery(start, end time.Time, a area) string {
	if a == "" {
		a = AreaDKEast
	}
	return fmt.Sprintf(queryTemplate, start.Local().Format("2006-01-02T15:04:05"), end.Local().Format("2006-01-02T15:04:05"), a, defaultLimit, defaultOffset)
}
