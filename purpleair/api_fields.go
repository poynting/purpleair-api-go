package purpleair

import (
	"fmt"
	"strings"
)

type Location int32

const (
	Outside Location = 0
	Inside
)

type Sensors struct {
	APIVersion             string       `json:"api_version"`
	TimeStamp              uint         `json:"time_stamp"`
	DataTimeStamp          uint         `json:"data_time_stamp"`
	LocationType           Location     `json:"location_type"`
	MaxAge                 uint         `json:"max_age"`
	FirmwareDefaultVersion string       `json:"firmware_default_version"`
	Fields                 []string     `json:"fields"`
	Data                   [][]*float32 `json:"data"`
}

type Bounds struct {
	nwlng float32
	nwlat float32
	selng float32
	selat float32
}

func latValid(lat float32) bool {
	if lat > 90 || lat < -90 {
		return false
	}
	return true
}

func lngValid(lon float32) bool {
	if lon > 180 || lon < -180 {
		return false
	}
	return true
}

func (b Bounds) toMap() map[string]string {
	return map[string]string{
		"nwlng": fmt.Sprintf("%.6f", b.nwlng),
		"nwlat": fmt.Sprintf("%.6f", b.nwlat),
		"selng": fmt.Sprintf("%.6f", b.selng),
		"selat": fmt.Sprintf("%.6f", b.selat),
	}
}

func AppendBoundsParams(params map[string]string, b *Bounds) map[string]string {
	for k, v := range b.toMap() {
		params[k] = v
	}
	return params
}

func NewBounds(nwlng float32, nwlat float32, selng float32, selat float32) (*Bounds, error) {
	if !lngValid(nwlng) {
		return nil, fmt.Errorf("nwlng out of bounds")
	}
	if !latValid(nwlat) {
		return nil, fmt.Errorf("nwlat out of bounds")
	}
	if !lngValid(selng) {
		return nil, fmt.Errorf("selng out of bounds")
	}
	if !latValid(selat) {
		return nil, fmt.Errorf("selat out of bounds")
	}
	if (nwlat - selat) < 0 {
		return nil, fmt.Errorf("selat must be less than nwlat")
	}
	if (selng - nwlng) < 0 {
		return nil, fmt.Errorf("nwlng must be less than selng")
	}
	return &Bounds{
		nwlng: nwlng,
		nwlat: nwlat,
		selng: selng,
		selat: selat,
	}, nil
}

func (b Bounds) UrlString() string {
	return fmt.Sprintf("nwlng=%3.5f&nwlat=%2.5f&selng=%3.5f&selat=%2.5f", b.nwlng, b.nwlat, b.selng, b.selat)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func validSensorFields() []string {
	return []string{
		"name",
		"icon",
		"model",
		"hardware",
		"location_type",
		"private",
		"latitude",
		"longitude",
		"altitude",
		"position_rating",
		"led_brightness",
		"firmware_version",
		"firmware_upgrade",
		"rssi",
		"uptime",
		"pa_latency",
		"memory",
		"last_seen",
		"last_modified",
		"date_created",
		"channel_state",
		"channel_flags",
		"channel_flags_manual",
		"channel_flags_auto",
		"confidence",
		"confidence_manual",
	}
}

func validEnvironmentalFields() []string {
	return []string{
		"humidity", "humidity_a", "humidity_b",
		"temperature", "temperature_a", "temperature_b",
		"pressure", "pressure_a", "pressure_b",
	}
}

func validMiscFields() []string {
	return []string{
		"voc", "voc_a", "voc_b",
		"ozone1", "analog_input",
	}
}

func validPm1Fields() []string {
	return []string{
		"pm1.0", "pm1.0_a", "pm1.0_b",
		"pm1.0_atm", "pm1.0_atm_a", "pm1.0_atm_b",
		"pm1.0_cf_1", "pm1.0_cf_1_a", "pm1.0_cf_1_b",
	}
}

func validPm25Fields() []string {
	return []string{
		"pm2.5_alt", "pm2.5_alt_a", "pm2.5_alt_b",
		"pm2.5", "pm2.5_a", "pm2.5_b",
		"pm2.5_atm", "pm2.5_atm_a", "pm2.5_atm_b",
		"pm2.5_cf_1", "pm2.5_cf_1_a", "pm2.5_cf_1_b",
	}
}

func validPm25AverageFields() []string {
	return []string{
		"pm2.5_10minute", "pm2.5_10minute_a", "pm2.5_10minute_b",
		"pm2.5_30minute", "pm2.5_30minute_a", "pm2.5_30minute_b",
		"pm2.5_60minute", "pm2.5_60minute_a", "pm2.5_60minute_b",
		"pm2.5_6hour", "pm2.5_6hour_a", "pm2.5_6hour_b",
		"pm2.5_24hour", "pm2.5_24hour_a", "pm2.5_24hour_b",
		"pm2.5_1week", "pm2.5_1week_a", "pm2.5_1week_b",
	}
}

func validPm10Fields() []string {
	return []string{
		"pm10.0", "pm10.0_a", "pm10.0_b",
		"pm10.0_atm", "pm10.0_atm_a", "pm10.0_atm_b",
		"pm10.0_cf_1", "pm10.0_cf_1_a", "pm10.0_cf_1_b",
	}
}

func allValidFields() []string {
	var v []string
	v = append(v, validSensorFields()...)
	v = append(v, validEnvironmentalFields()...)
	v = append(v, validMiscFields()...)
	v = append(v, validPm1Fields()...)
	v = append(v, validPm25Fields()...)
	v = append(v, validPm25AverageFields()...)
	v = append(v, validPm10Fields()...)
	return v
}

func validateParams(params map[string]string) error {
	// validate params
	for p, v := range params {
		if p == "fields" {
			for _, f := range strings.Split(v, ",") {
				if !contains(allValidFields(), f) {
					return fmt.Errorf("invalid field %s", f)
				}
			}
		} else if p == "location_type" {
			if v != "0" && v != "1" {
				return fmt.Errorf("invald location type %s", v)
			}
		} else if contains(validUrlParams(), p) {
			continue
		} else {
			return fmt.Errorf("unknown parameter %s", p)
		}

	}
	return nil
}

func validUrlParams() []string {
	return []string{
		"fields",
		"location_type",
		"read_keys",
		"show_only",
		"modified_since",
		"max_age",
		"nwlng",
		"nwlat",
		"selng",
		"selat",
	}
}
