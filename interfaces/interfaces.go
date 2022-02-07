package interfaces

import "github.com/adamhassel/power/entities"

// Queryer will query in a well defined manner
type Queryer interface {
	Query() (interface{}, error)
}

// Authenticater allows configuration of authentication data
type Authenticater interface {
	Authenticate([]byte) error
}

// Identifier will associate an arbitrary identification
type Identifier interface {
	Identify([]byte) error
}

type Indexer interface {
	Index() entities.TariffIndex
}

type Taxer interface {
	Taxes() entities.Taxes
}

type AtPoser interface {
	AtPos(int) entities.Tariffs
}

type SpotPricer interface {
	SpotPrices() []entities.Elspotprice
}
