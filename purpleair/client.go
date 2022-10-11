package purpleair

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

type Client struct {
	ReadKey    string
	WriteKey   string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(readkey string, writekey string) (*Client, error) {
	if readkey == "" {
		return nil, errors.New("must provide API read key")
	}
	return &Client{
		ReadKey:  readkey,
		WriteKey: writekey,
		BaseURL:  "https://api.purpleair.com/v1",
		HTTPClient: &http.Client{
			Timeout: time.Second * 5,
		},
	}, nil
}

func (c Client) BuildUrl(endpoint string, params map[string]string) string {
	s := c.BaseURL + endpoint
	if params != nil {
		s += "?"
	}

	keys := make([]string, 0)
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		if i != 0 {
			s += "&"
		}
		s += k + "=" + params[k]
		i++
	}
	return s
}

func (c Client) NewGetRequest(endpoint string, params map[string]string) *http.Request {
	req, _ := http.NewRequest("GET", c.BuildUrl(endpoint, params), nil)
	req.Header.Add("X-API-Key", c.ReadKey)
	req.Header.Add("Accept", "application/json")
	return req
}

func (c Client) KeysValid() (bool, error) {
	resp, err := c.HTTPClient.Do(c.NewGetRequest("/keys", nil))
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusForbidden {
		return false, nil
	}
	if resp.StatusCode != http.StatusCreated {
		return false, fmt.Errorf("expected return code 201 or 403 got %d", resp.StatusCode)
	}
	return true, nil
}

func (c Client) GetSensors(params map[string]string) (*Sensors, error) {
	// validate params
	err := validateParams(params)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(c.NewGetRequest("/sensors", params))
	if err != nil {
		return nil, fmt.Errorf("error getting /sensors: %s", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}
	var s Sensors
	if err := json.Unmarshal(body, &s); err != nil { // Parse []byte to go struct pointer
		return nil, fmt.Errorf("can not unmarshal response JSON")
	}
	return &s, err

}

type Sample struct {
	Timestamp  uint               `json:"time_stamp"`
	Sampledata map[string]float32 `json:"data"`
}

func NewSample(ts uint) *Sample {
	return &Sample{
		Timestamp:  ts,
		Sampledata: make(map[string]float32),
	}
}

// *float32 parameter is used to support null values in response json
// This happens when a field is requested from the api that a sensor doesn't provide
// so we don't include that field in our sensor data map for that sensor
func (c Client) SensorsToSamples(timestamp uint, fields []string, data [][]*float32) []Sample {

	samples := make([]Sample, len(data))
	for i, d := range data {
		samples[i].Timestamp = timestamp
		samples[i].Sampledata = make(map[string]float32)
		for j, v := range d {
			if v != nil {
				k := fields[j]
				samples[i].Sampledata[k] = *v
			}
		}
	}
	return samples
}

func SamplesJson(samples []Sample) ([]byte, error) {
	return json.Marshal(samples)
}
