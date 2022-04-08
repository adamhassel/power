package eloverblik

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/adamhassel/power/entities"
	"github.com/tidwall/gjson"
)

const elOverblikUrl = "https://api.eloverblik.dk/CustomerApi/api"

var ErrAuth = errors.New("authorization error")

// Eloverblik implements the interfaces for getting tariffs
type Eloverblik struct {
	authToken    []byte
	refreshToken []byte
	mid          string
	ft           FullTariffs
	rg bool
}

// RequestData is the meetring data (mid) sent to eloverblik for extracting tariffs
type RequestData struct {
	MeteringPoints struct {
		MeteringPoint []string `json:"meteringPoint"`
	} `json:"meteringPoints"`
}

// FullTariffs is the data format returned from eloverblik, containg tariff information.
type FullTariffs struct {
	Result []struct {
		Result struct {
			MeteringPointId string `json:"meteringPointId"`
			Subscriptions   []struct {
				SubscriptionId interface{} `json:"subscriptionId"`
				Name           string      `json:"name"`
				Description    string      `json:"description"`
				Owner          string      `json:"owner"`
				ValidFromDate  string      `json:"validFromDate"`
				ValidToDate    interface{} `json:"validToDate"`
				Price          float64     `json:"price"`
				Quantity       int         `json:"quantity"`
			} `json:"subscriptions"`
			Fees    []interface{} `json:"fees"`
			Tariffs []struct {
				TariffId      interface{} `json:"tariffId"`
				Name          string      `json:"name"`
				Description   string      `json:"description"`
				Owner         string      `json:"owner"`
				PeriodType    string      `json:"periodType"`
				ValidFromDate string      `json:"validFromDate"`
				ValidToDate   *string     `json:"validToDate"`
				Prices        []struct {
					Position string  `json:"position"`
					Price    float64 `json:"price"`
				} `json:"prices"`
			} `json:"tariffs"`
		} `json:"result"`
		Success       bool        `json:"success"`
		ErrorCode     string      `json:"errorCode"`
		ErrorCodeEnum string      `json:"errorCodeEnum"`
		ErrorText     string      `json:"errorText"`
		Id            string      `json:"id"`
		StackTrace    interface{} `json:"stackTrace"`
	} `json:"result"`
	ts time.Time `json:"-"`
}

func (ft FullTariffs) UpdatedAt() time.Time {
	return ft.ts
}

func (e *Eloverblik) Authenticate(token []byte) error {
	e.authToken = token
	return nil
}

// ExecAuth performs the actual authentication step and stores/refreshes the refresh token
func (e *Eloverblik) ExecAuth() error {
	t, err := getRefreshToken(e.authToken)
	if err != nil {
		return err
	}
	e.refreshToken = []byte(t)
	return nil
}

func (e *Eloverblik) Identify(mid []byte) error {
	e.mid = string(mid)
	return nil
}

func (e *Eloverblik) Query() (interface{}, error) {
	if e.refreshToken == nil {
		if err := e.ExecAuth(); err != nil {
			return nil, err
		}
	}
	e.ft = FullTariffs{}
	if err := e.ft.query(e.refreshToken, e.mid); err != nil {
		if errors.Is(err, ErrAuth) && !e.rg {
			if err := e.ExecAuth(); err != nil {
				return nil, err
			}
			// set recurse guard
			e.rg = true
			return e.Query()
		}
		return nil, err
	}
	e.ft.ts = time.Now()
	return e.ft, nil
}

func (e Eloverblik) FullTariffs() FullTariffs {
	return e.ft
}

// Index tariffs by position (hour)
func (ft FullTariffs) Index() entities.TariffIndex {
	rv := make(entities.TariffIndex)
	var m entities.TariffMetadata
	m.Positions = entities.NewIntSet()
	for _, res := range ft.Result {
		for _, tar := range res.Result.Tariffs {
			m.PositionCount = max(m.PositionCount, len(tar.Prices))
			for _, p := range tar.Prices {
				pos, _ := strconv.Atoi(p.Position)
				m.Positions.Add(pos)
			}
		}
	}
	for _, res := range ft.Result {
		for _, tar := range res.Result.Tariffs {
			tariff := entities.Tariff{
				TariffId:      tar.TariffId,
				Name:          tar.Name,
				Description:   tar.Description,
				Owner:         tar.Owner,
				PeriodType:    tar.PeriodType,
				ValidFromDate: tar.ValidFromDate,
				ValidToDate:   tar.ValidToDate,
			}

			for i := 0; i < m.PositionCount; i++ {
				pos := 0
				if len(tar.Prices) > i {
					pos, _ = strconv.Atoi(tar.Prices[i].Position)
					pos--
				}
				ptariff := tariff
				ptariff.Price = tar.Prices[pos].Price
				rv[i] = append(rv[i], ptariff)

			}
		}
	}
	return rv
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (ft *FullTariffs) query(token []byte, mid string) error {
	u, _ := url.Parse(elOverblikUrl + "/meteringpoints/meteringpoint/getcharges")
	var h = make(http.Header)
	h.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	h.Add("Content-Type", "application/json")
	r := &http.Request{
		Method: "POST",
		URL:    u,
		Header: h,
		Body:   ioutil.NopCloser(strings.NewReader(makeMeteringPointBody(mid))),
	}
	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.Status)
		fmt.Println(string(response))
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return ErrAuth
	}

	if err := json.Unmarshal(response, &ft); err != nil {
		return err
	}

	return nil
}

func getRefreshToken(token []byte) (string, error) {
	u, _ := url.Parse(elOverblikUrl + "/Token")
	var h = make(http.Header)
	h.Add("Authorization", fmt.Sprintf("Bearer %s", string(token)))
	r := &http.Request{
		Method: "GET",
		URL:    u,
		Header: h,
	}

	c := http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	return gjson.GetBytes(response, "result").String(), nil
}

func makeMeteringPointBody(mid string) string {
	var m RequestData
	m.MeteringPoints.MeteringPoint = make([]string, 1)
	m.MeteringPoints.MeteringPoint[0] = mid
	rv, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(rv)
}
