# power
Utilities for querying power prices

This is a proof of concept (and probably way to abstracted for that, I got a
little carried away, so it might change in time) on how to fetch by-the-hour
power prices from your electrical company in Denmark.

Use it as inspiration, or simply run (`go run cmd/main.go`) or build and run
(`go build -o power cmd/main.go && ./power`) to get some JSON out that you can
feed to Influx or whatever.

## Example utility

The example program in `cmd` will fetch upcoming by-the-hour prices and
correlate them with official tariffs for your power meter (which varies a lot
in DK, and depends on who delivers your power to your house), and spit out some
JSON with your by-the-hour power price, broken into taxes and tariffs, etc.

This assumes, of course, that your power plan has by the hour pricing. I have
no idea what happens if you don't, but I'd think you'd just get a list of
identical data :)

### Prerequisites

Go to https://eloverblik.dk, and log in with NemID. Obtain an **API Token** from their web
interface, and copy that to the config file (check `power.conf.example`). While
you're there, note your **measurement point ID**, an 18-digit number identifying
your power meter, in the same config file.

### Running

The app accepts a `-c` option, which will tell it which config file to read.
Default is `power.conf`. It also accepts a `-h` option, which is the number of
hours in the future to get prices for. The default is to get for the next 12
hours.

## Possible further features

* Consumption summary in kWH and DKK
