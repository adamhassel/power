package entities

/*

GET https://www.nationalbanken.dk/_vti_bin/DN/DataService.svc/CurrencyRatesXML?lang=en yields price per 100 DKK:

<exchangerates xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" type="Exchange rates" author="Danmarks Nationalbank" refcur="DKK" refamt="1">
<dailyrates id="2022-02-07">
<currency code="AUD" desc="Australian dollars" rate="462.47"/>
<currency code="BGN" desc="Bulgarian lev" rate="380.63"/>
<currency code="BRL" desc="Brazilian real" rate="122.96"/>
<currency code="CAD" desc="Canadian dollars" rate="511.78"/>
<currency code="CHF" desc="Swiss francs" rate="704.22"/>
<currency code="CNY" desc="Chinese yuan renminbi" rate="102.25"/>
<currency code="CZK" desc="Czech koruny" rate="30.73"/>
<currency code="EUR" desc="Euro" rate="744.43"/>
<currency code="GBP" desc="Pounds sterling" rate="879.06"/>
<currency code="HKD" desc="Hong Kong dollars" rate="83.45"/>
<currency code="HRK" desc="Croatian kuna" rate="98.99"/>
<currency code="HUF" desc="Hungarian forints" rate="2.106"/>
<currency code="IDR" desc="Indonesian rupiah" rate="0.0452"/>
<currency code="ILS" desc="Israeli shekel" rate="203.69"/>
<currency code="INR" desc="Indian rupee" rate="8.70"/>
<currency code="ISK" desc="Icelandic kronur *" rate="5.191"/>
<currency code="JPY" desc="Japanese yen" rate="5.6572"/>
<currency code="KRW" desc="South Korean won" rate="0.5427"/>
<currency code="MXN" desc="Mexican peso" rate="31.58"/>
<currency code="MYR" desc="Malaysian ringgit" rate="155.38"/>
<currency code="NOK" desc="Norwegian kroner" rate="73.96"/>
<currency code="NZD" desc="New Zealand dollars" rate="430.85"/>
<currency code="PHP" desc="Philippine peso" rate="12.62"/>
<currency code="PLN" desc="Polish zlotys" rate="163.86"/>
<currency code="RON" desc="Romanian leu" rate="150.51"/>
<currency code="RUB" desc="Russian rouble" rate="8.60"/>
<currency code="SEK" desc="Swedish kronor" rate="71.25"/>
<currency code="SGD" desc="Singapore dollars" rate="483.74"/>
<currency code="THB" desc="Thai baht" rate="19.73"/>
<currency code="TRY" desc="Turkish lira" rate="47.96"/>
<currency code="USD" desc="US dollars" rate="650.33"/>
<currency code="XDR" desc="SDR (Calculated **)" rate="913.58"/>
<currency code="ZAR" desc="South African rand" rate="42.06"/>
</dailyrates>
</exchangerates>
*/
