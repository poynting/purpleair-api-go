package purpleair

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoundsUrlString(t *testing.T) {
	b, err := NewBounds(12.300000, 45.599998, 78.900000, -1.200000)
	assert.NotNil(t, b)
	assert.Nil(t, err)
	assert.Equal(t, b.UrlString(), "nwlng=12.30000&nwlat=45.60000&selng=78.90000&selat=-1.20000")
}

func TestBadLatLon(t *testing.T) {
	b, err := NewBounds(300, 45.599998, 78.900000, -1.200000)
	assert.Nil(t, b)
	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("nwlng out of bounds"))

	b, err = NewBounds(12.300000, -95, 78.900000, -1.200000)
	assert.Nil(t, b)
	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("nwlat out of bounds"))

	b, err = NewBounds(12.300000, 45.599998, -189, -1.200000)
	assert.Nil(t, b)
	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("selng out of bounds"))

	b, err = NewBounds(12.300000, 45.599998, 78.900000, 180)
	assert.Nil(t, b)
	assert.NotNil(t, err)
	assert.Equal(t, err, fmt.Errorf("selat out of bounds"))
}

func TestGoodParams(t *testing.T) {
	params := map[string]string{
		"fields":        "humidity,temperature,voc,pm1.0,pm2.5,pm10.0",
		"location_type": "0",
	}
	err := validateParams(params)
	assert.Nil(t, err)
}

func TestBadParams(t *testing.T) {
	params := map[string]string{
		"fields":        "bad_field,temperature,voc,pm1.0,pm2.5,pm10.0",
		"location_type": "0",
	}
	err := validateParams(params)
	assert.NotNil(t, err)
}

func TestBoundsParams(t *testing.T) {
	params := map[string]string{
		"fields":        "humidity",
		"location_type": "0",
	}
	b, err := NewBounds(12.300000, 45.599998, 78.900002, -1.200000)
	assert.Nil(t, err)
	params = AppendBoundsParams(params, b)
	assert.Equal(t, params["nwlng"], "12.300000")
	assert.Equal(t, params["nwlat"], "45.599998")
	assert.Equal(t, params["selng"], "78.900002")
	assert.Equal(t, params["selat"], "-1.200000")
}
