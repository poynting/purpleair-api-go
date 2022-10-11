package purpleair

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	c, err := NewClient("test-read-key", "test-write-key")
	assert.NotNil(t, c)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestNoWriteKey(t *testing.T) {
	c, err := NewClient("test-read-key", "")
	assert.NotNil(t, c)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestNoReadKey(t *testing.T) {
	c, err := NewClient("", "test-write-key")
	assert.Nil(t, c)
	assert.NotNil(t, err)
}

func TestBuildURL(t *testing.T) {
	c, _ := NewClient("test-read-key", "")
	params := map[string]string{"d": "e", "f": "1234"}
	url := c.BuildUrl("/abc", params)
	assert.Equal(t, url, "https://api.purpleair.com/v1/abc?d=e&f=1234", "URL is incorrect")
}

func TestKeysValid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/keys" {
			t.Errorf("Expected to request '/keys', got: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
			"api_version" : "V1.0.11-0.0.40",
			"time_stamp" : 1663477141,
			"api_key_type" : "READ"
		  }`))
	}))
	defer server.Close()
	c, _ := NewClient("test-read-key", "")
	c.BaseURL = server.URL
	keysValid, err := c.KeysValid()
	assert.Nil(t, err)
	assert.True(t, keysValid)
}

func setupServer(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sensors" {
			t.Errorf("Expected to request '/sensors', got: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
		"api_version" : "V1.0.11-0.0.40",
		"time_stamp" : 1664170828,
		"data_time_stamp" : 1664170800,
		"location_type" : 0,
		"max_age" : 604800,
		"firmware_default_version" : "7.00",
		"fields" : ["sensor_index","humidity","temperature","voc","pm1.0","pm2.5","pm10.0"],
		"data" : [
		  [15111,43,77,null,6.2,8.7,9.4],
		  [20755,55,69,null,7.3,9.9,10.3],
		  [90011,47,72,null,7.1,10.3,11.0],
		  [127397,51,69,null,4.1,7.3,7.8]
		]
	  }`))
	}))
	return server
}

func setupServerFewerParams(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sensors" {
			t.Errorf("Expected to request '/sensors', got: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{
		"api_version" : "V1.0.11-0.0.40",
		"time_stamp" : 1664170828,
		"data_time_stamp" : 1664170800,
		"location_type" : 0,
		"max_age" : 600,
		"firmware_default_version" : "7.00",
		"fields" : ["sensor_index","humidity","temperature"],
		"data" : [
		  [15111,43,77],
		  [20755,55,69]
		]
	  }`))
	}))
	return server
}

func TestGetSensors(t *testing.T) {
	server := setupServer(t)
	defer server.Close()
	b, _ := NewBounds(12.300000, 45.599998, 78.900002, -1.200000)
	params := map[string]string{
		"fields":        "humidity,temperature,voc,pm1.0,pm2.5,pm10.0",
		"location_type": "0",
	}
	params = AppendBoundsParams(params, b)
	c, _ := NewClient("test-read-key", "")
	c.BaseURL = server.URL
	s, err := c.GetSensors(params)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, s.APIVersion, "V1.0.11-0.0.40")
	assert.Equal(t, s.TimeStamp, uint(1664170828))
	assert.Equal(t, s.DataTimeStamp, uint(1664170800))
	assert.Equal(t, s.LocationType, Outside)
	assert.Equal(t, s.MaxAge, uint(604800))
	assert.Equal(t, s.FirmwareDefaultVersion, "7.00")
	assert.Equal(t, s.Fields, []string{"sensor_index", "humidity", "temperature", "voc", "pm1.0", "pm2.5", "pm10.0"})
}

func TestGetSensorsBadParams(t *testing.T) {
	server := setupServer(t)
	defer server.Close()
	params := map[string]string{
		"unknown_param": "unknown_value",
		"fields":        "humidity,temperature,voc,pm1.0,pm2.5,pm10.0",
		"location_type": fmt.Sprintf("%d", Outside),
	}
	c, _ := NewClient("test-read-key", "")
	c.BaseURL = server.URL
	s, err := c.GetSensors(params)
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("unknown parameter unknown_param"))
}

func TestGetSensorsBadFields(t *testing.T) {
	server := setupServer(t)
	defer server.Close()
	params := map[string]string{
		"fields":        "bad_field,temperature,voc,pm1.0,pm2.5,pm10.0",
		"location_type": fmt.Sprintf("%d", Outside),
	}
	c, _ := NewClient("test-read-key", "")
	c.BaseURL = server.URL
	s, err := c.GetSensors(params)
	assert.Nil(t, s)
	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("invalid field bad_field"))
}

func TestSensorsToSamples(t *testing.T) {
	server := setupServer(t)
	defer server.Close()
	b, _ := NewBounds(12.300000, 45.599998, 78.900002, -1.200000)
	params := map[string]string{
		"fields":        "humidity,temperature,voc,pm1.0,pm2.5,pm10.0",
		"location_type": "0",
	}
	params = AppendBoundsParams(params, b)
	c, _ := NewClient("test-read-key", "")
	c.BaseURL = server.URL
	s, err := c.GetSensors(params)
	assert.Nil(t, err)

	samples := c.SensorsToSamples(s.DataTimeStamp, s.Fields, s.Data)
	assert.Equal(t, len(samples), 4)
	assert.Equal(t, len(samples[3].Sampledata), 6)
	assert.Zero(t, samples[3].Sampledata["voc"]) // verify nil reponse for "voc" isn't included in map
	assert.InDelta(t, samples[3].Sampledata["humidity"], 51.0, 0.5)
}

func TestSensorsJson(t *testing.T) {
	server := setupServer(t)
	defer server.Close()
	b, _ := NewBounds(12.300000, 45.599998, 78.900002, -1.200000)
	params := map[string]string{
		"fields":        "humidity,temperature,voc,pm1.0,pm2.5,pm10.0",
		"location_type": "0",
	}
	params = AppendBoundsParams(params, b)
	c, _ := NewClient("test-read-key", "")
	c.BaseURL = server.URL
	s, err := c.GetSensors(params)
	assert.Nil(t, err)

	samples := c.SensorsToSamples(s.DataTimeStamp, s.Fields, s.Data)
	samples_json, err := SamplesJson(samples)
	assert.Nil(t, err)
	expected_response := []byte(`[
		{
			"time_stamp":1664170800,
			"data":{
				"humidity":43,
				"pm1.0":6.2,
				"pm10.0":9.4,
				"pm2.5":8.7,
				"sensor_index":15111,
				"temperature":77
			}
		},{
			"time_stamp":1664170800,
			"data":{
				"humidity":55,
				"pm1.0":7.3,
				"pm10.0":10.3,
				"pm2.5":9.9,
				"sensor_index":20755,
				"temperature":69
			}
		},{
			"time_stamp":1664170800,
			"data":{
				"humidity":47,
				"pm1.0":7.1,
				"pm10.0":11,
				"pm2.5":10.3,
				"sensor_index":90011,
				"temperature":72
			}
		},{
			"time_stamp":1664170800,
			"data":{
				"humidity":51,
				"pm1.0":4.1,
				"pm10.0":7.8,
				"pm2.5":7.3,
				"sensor_index":127397,
				"temperature":69
			}
		}
	]`)
	assert.JSONEq(t, string(samples_json), string(expected_response))
}

func TestSensorsJsonFewerParams(t *testing.T) {
	server := setupServerFewerParams(t)
	defer server.Close()
	b, _ := NewBounds(12.300000, 45.599998, 78.900002, -1.200000)
	params := map[string]string{
		"fields":        "humidity,temperature",
		"location_type": "0",
	}
	params = AppendBoundsParams(params, b)
	c, _ := NewClient("test-read-key", "")
	c.BaseURL = server.URL
	s, err := c.GetSensors(params)
	assert.Nil(t, err)

	samples := c.SensorsToSamples(s.DataTimeStamp, s.Fields, s.Data)
	samples_json, err := SamplesJson(samples)
	assert.Nil(t, err)
	expected_response := []byte(`[
		{
			"time_stamp":1664170800,
			"data":{
				"humidity":43,
				"sensor_index":15111,
				"temperature":77
			}
		},{
			"time_stamp":1664170800,
			"data":{
				"humidity":55,
				"sensor_index":20755,
				"temperature":69
			}
		}
	]`)
	assert.JSONEq(t, string(samples_json), string(expected_response))
}
