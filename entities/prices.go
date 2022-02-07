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
	HourUTC       time.Time `json:"HourUTC"`
	HourDK        string    `json:"HourDK"`
	PriceArea     string    `json:"PriceArea"`
	SpotPriceDKK  *float64  `json:"SpotPriceDKK"`
	DKKEstimated  bool      `json:"dkk_estimated"`
	EstimatedRate float64   `json:"rate,omitempty"`
	SpotPriceEUR  float64   `json:"SpotPriceEUR"`
	Typename      string    `json:"__typename"`
}
