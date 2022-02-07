package entities

// Tariff is a flattened version of a tariff from eloverblik
type Tariff struct {
	TariffId      interface{} `json:"tariffId"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Owner         string      `json:"owner"`
	PeriodType    string      `json:"periodType"`
	ValidFromDate string      `json:"validFromDate"`
	ValidToDate   *string     `json:"validToDate"`
	Price         float64     `json:"price"`
}

// Tariffs is a slice of Tariff
type Tariffs []Tariff

// Tax is a named tax with an amount (in DKK) per kWh
type Tax struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

// Taxes is a slice of Tax
type Taxes []Tax

// TariffMetadata contains information needed to properly duplicate data when default positions needs repeating alongside position specific data
type TariffMetadata struct {
	PositionCount int
	Positions     IntSet
}

type TariffIndex map[int][]Tariff

// Tax converts a Tariff to a Tax
func (t Tariff) Tax() Tax {
	return Tax{
		Name:   t.Name,
		Amount: t.Price,
	}
}

// Taxes returns a list of taxes contained in ts
func (ts Tariffs) Taxes() Taxes {
	rv := make([]Tax, 0, len(ts))
	for _, tariff := range ts {
		rv = append(rv, tariff.Tax())
	}
	return rv
}

// Total summarizes the values in ts
func (ts Taxes) Total() float64 {
	var rv float64
	for _, v := range ts {
		rv += v.Amount
	}
	return rv
}

// AtPos returns the Tariff at index p, if it exists. Otherwise returns at the default position (0)
func (t TariffIndex) AtPos(p int) Tariffs {
	if v, ok := t[p]; ok {
		return Tariffs(v)
	}
	if v, ok := t[0]; ok {
		return Tariffs(v)
	}
	return nil
}
