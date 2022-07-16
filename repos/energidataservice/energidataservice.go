package energidataservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/adamhassel/power/entities"
	"github.com/rickar/cal/v2"
	"github.com/rickar/cal/v2/dk"
)

const dataServiceUrl = "https://api.energidataservice.dk/dataset/elspotprices"

//const queryTemplate = `{"operationName":"Dataset","variables":{},"query":"query Dataset {\n  elspotprices(\n    where: {HourDK: {_gte: \"%s\", _lt: \"%s\"}, PriceArea: {_eq: \"%s\"}}\n    order_by: {HourUTC: asc}\n    limit: %d\n    offset: %d\n  ) {\n    HourUTC\n    HourDK\n    PriceArea\n    SpotPriceDKK\n    SpotPriceEUR\n    __typename\n  }\n}\n"}`
const queryTemplate = `start=%s&end=%s&filter={"PriceArea":"%s"}&limit=%d&offset=%d&sort=HourUTC`

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
	Elspotprices []entities.Elspotprice `json:"records"`
}

func (p Prices) SpotPrices() []entities.Elspotprice {
	return p.Elspotprices
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
		e.from = time.Now().Truncate(24 * time.Hour)
	}
	if e.to.IsZero() || e.to.Before(e.from) {
		e.to = e.from.Add(48 * time.Hour)
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
	params := makeSpotPriceQuery(from, to, a)
	u := dataServiceUrl + "?" + params
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	out, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(out))
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("energiDataService returned %s, '%s'", resp.Status, response)
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
	for _, sp := range p.Elspotprices {
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
		earliest := p.Elspotprices[0].HourUTC

		var target time.Time
		// Check if we're a non-working day, and find Previsou workday @ 2300 local
		c := cal.NewBusinessCalendar()
		c.Calendar.AddHoliday(dk.Holidays...)
		target = prevWorkDay(c, time.Time(earliest))
		// fetch the last price of the workday, set time to 2300 local
		target = time.Date(target.Year(), target.Month(), target.Day(), 23, 0, 0, 0, time.Time(earliest).Location())
		var rateprice Prices
		if err := rateprice.getRawSpotPrices(target, target.Add(time.Hour), a); err != nil {
			return err
		}
		if rateprice.Elspotprices[0].SpotPriceDKK == nil {
			return errors.New("DKK price was nil")
		}
		latestDKK = *rateprice.Elspotprices[0].SpotPriceDKK
		latestEUR = rateprice.Elspotprices[0].SpotPriceEUR
	}

	rate := latestDKK / latestEUR

	// finally, fixup prices where appropriate
	for i, price := range p.Elspotprices {
		if price.SpotPriceDKK == nil {
			p.Elspotprices[i].SpotPriceDKK = new(float64)
			*p.Elspotprices[i].SpotPriceDKK = price.SpotPriceEUR * rate
			p.Elspotprices[i].DKKEstimated = true
			p.Elspotprices[i].EstimatedRate = rate
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
	return fmt.Sprintf(queryTemplate, start.Local().Format("2006-01-02T15:04"), end.Local().Format("2006-01-02T15:04"), a, defaultLimit, defaultOffset)
}
