# power
Utilities for querying power prices

This is a proof of concept (and probably way to abstracted for that) on how to fetch by-the-hour power prices from your electrical company in Denmark.

Use it as inspiration, or simply run (`go run cmd/main.go`) or build and run
(`go build -o power cmd/main.go && ./power`) to get some JSON out that you can
feed to Influx or whatever.

## Possible further features

* Consumption summary in kWH and DKK
