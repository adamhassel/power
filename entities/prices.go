package entities

import "time"

// FullPrice is the final price per kWh, when all taxes are taken into account.
type FullPrice struct {
	Taxes         []Tax     `json:"taxes"`
	ValidFrom     time.Time `json:"valid_from"`
	ValidTo       time.Time `json:"valid_to"`
	Estimated     bool      `json:"dkk_estimated"`
	EstimatedRate float64   `json:"rate,omitempty"`
	RawPrice      float64   `json:"spot_price_ex_vat"`
	TaxesSubTotal float64   `json:"taxes_subtotal_ex_vat"`
	Total         float64   `json:"total_ex_vat"`
	TotalIncVAT   float64   `json:"total_inc_vat"`
}

// Elspotprice is the raw per price data
type Elspotprice struct {
	HourUTC       pTime    `json:"HourUTC"`
	HourDK        pTime    `json:"HourDK"`
	PriceArea     string   `json:"PriceArea"`
	SpotPriceDKK  *float64 `json:"SpotPriceDKK"`
	DKKEstimated  bool     `json:"dkk_estimated"`
	EstimatedRate float64  `json:"rate,omitempty"`
	SpotPriceEUR  float64  `json:"SpotPriceEUR"`
	Typename      string   `json:"__typename"`
}

type pTime time.Time

func (t *pTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)

	// Get rid of the quotes "" around the value.
	// A second option would be to include them
	// in the date format string instead, like so below:
	//   time.Parse(`"`+time.RFC3339Nano+`"`, s)
	s = s[1 : len(s)-1]

	o, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		o, err = time.Parse("2006-01-02T15:04:05", s)
	}
	*t = pTime(o)
	return
}
