package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/poynting/purpleair-api/purpleair"
	"github.com/urfave/cli/v2"
)

type InfluxDbClient struct {
	host       string
	port       int
	database   string
	username   string
	password   string
	HTTPClient *http.Client
}

func NewInfluxClient(host string, port int, database string, username string, password string) *InfluxDbClient {
	return &InfluxDbClient{
		host:     host,
		port:     port,
		database: database,
		username: username,
		password: password,
		HTTPClient: &http.Client{
			Timeout: time.Second * 2,
		},
	}
}

func GetEnvToParams(cCtx *cli.Context) (string, string, map[string]string, error) {
	readkey := os.Getenv("PURPLEAIR_READ_KEY")
	writekey := os.Getenv("PURPLEAIR_WRITE_KEY")
	latstr := os.Getenv("PURPLEAIR_LATITUDE")
	lonstr := os.Getenv("PURPLEAIR_LONGITUDE")
	rangestr := os.Getenv("PURPLEAIR_RANGE_KM")
	if readkey == "" {
		return "", "", nil, fmt.Errorf("read key is required. Set env PURPLEAIR_READ_KEY")
	}
	if latstr == "" || lonstr == "" {
		return "", "", nil, fmt.Errorf("lat,lon is required. Set env PURPLEAIR_LATITUDE, PURPLEAIR_LONGITUDE")
	}
	if rangestr == "" {
		return "", "", nil, fmt.Errorf("range in km is required. Set env PURPLEAIR_RANGE_KM")
	}
	lat_deg, err := strconv.ParseFloat(latstr, 64)
	if err != nil {
		return "", "", nil, fmt.Errorf("could not parse PURPLEAIR_LATITUDE into float")
	}
	lon_deg, err := strconv.ParseFloat(lonstr, 64)
	if err != nil {
		return "", "", nil, fmt.Errorf("could not parse PURPLEAIR_LATITUDE into float")
	}
	range_km, err := strconv.ParseFloat(rangestr, 64)
	if err != nil {
		return "", "", nil, fmt.Errorf("could not parse PURPLEAIR_RANGE_KM into float")
	}
	lat_rad := purpleair.Radians(lat_deg)
	lon_rad := purpleair.Radians(lon_deg)
	nwlat, nwlng := purpleair.PointFromLocRadial(lat_rad, lon_rad, range_km, purpleair.Radians(-45))
	selat, selng := purpleair.PointFromLocRadial(lat_rad, lon_rad, range_km, purpleair.Radians(135))
	b, err := purpleair.NewBounds(
		float32(purpleair.Degrees(nwlng)),
		float32(purpleair.Degrees(nwlat)),
		float32(purpleair.Degrees(selng)),
		float32(purpleair.Degrees(selat)))
	if err != nil {
		return "", "", nil, err
	}
	params := map[string]string{
		"fields":        "humidity,temperature,voc,pm1.0,pm2.5,pm10.0,pm2.5_alt",
		"location_type": "0",
	}
	params = purpleair.AppendBoundsParams(params, b)
	_, nwlng_valid := params["nwlng"]
	_, nwlat_valid := params["nwlat"]
	_, selng_valid := params["selng"]
	_, selat_valid := params["selat"]

	if !(nwlng_valid && nwlat_valid && selng_valid && selat_valid) {
		return "", "", nil, fmt.Errorf("error: must provide a location")
	}

	return readkey, writekey, params, err
}

func getSamples(cCtx *cli.Context) ([]purpleair.Sample, error) {
	readkey, writekey, params, err := GetEnvToParams(cCtx)
	if err != nil {
		return nil, err
	}
	c, err := purpleair.NewClient(readkey, writekey)
	if err != nil {
		return nil, err
	}
	r, err := c.GetSensors(params)
	if err != nil {
		return nil, err
	}
	samples := c.SensorsToSamples(r.DataTimeStamp, r.Fields, r.Data)
	return samples, nil
}

func samplesToPrettyJson(samples []purpleair.Sample) (string, error) {
	samples_json, err := purpleair.SamplesJson(samples)
	if err != nil {
		return "", err
	}
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, samples_json, "", "    ")
	return fmt.Sprint(prettyJSON.String()), nil
}

func getSensorsToJson(cCtx *cli.Context) error {
	samples, err := getSamples(cCtx)
	if err != nil {
		return err
	}
	js, err := samplesToPrettyJson(samples)
	if err != nil {
		return err
	}
	fmt.Println(js)
	return nil
}

func publishInfluxDb(influx *InfluxDbClient, measurement string, tags map[string]string, samples []purpleair.Sample) error {
	for _, s := range samples {
		line := measurement
		for key, val := range tags {
			line += fmt.Sprintf(",%s=%s", key, val)
		}
		line += fmt.Sprintf(",sensor_index=%d ", int(math.Round(float64(s.Sampledata["sensor_index"]))))
		j := 0
		for key, val := range s.Sampledata {
			if key != "sensor_index" {
				if j != 0 {
					line += ","
				}
				line += fmt.Sprintf("%s=%.2f", key, val)
				j++
			}
		}
		if val, ok := s.Sampledata["pm2.5_alt"]; ok {
			line += fmt.Sprintf(",aqi_epa=%d", purpleair.Pm25ToAqi(float64(val)))
		}
		if val, ok := s.Sampledata["pm2.5"]; ok {
			line += fmt.Sprintf(",aqi_raw=%d", purpleair.Pm25ToAqi(float64(val)))
		}

		line += fmt.Sprintf(" %d", int(s.Timestamp*1e9))
		fmt.Println(line)
		url := fmt.Sprintf("http://%s:%d/write?db=%s", influx.host, influx.port, influx.database)
		reader := strings.NewReader(line)
		req, err := http.NewRequest("POST", url, reader)
		if err != nil {
			fmt.Println(err)
		}
		influx.HTTPClient.Do(req)
	}
	return nil
}

func getSensorsToInflux(cCtx *cli.Context) error {
	host := os.Getenv("INFLUXDB_HOST")
	portstr := os.Getenv("INFLUXDB_PORT")
	db := os.Getenv("INFLUXDB_DB")
	measurement := os.Getenv("INFLUX_MEASUREMENT_NAME")
	loc := os.Getenv("INFLUX_LOCATION_TAG")
	if measurement == "" {
		return fmt.Errorf("measurement name must be set via env INFLUX_MEASUREMENT_NAME")
	}
	if loc == "" {
		return fmt.Errorf("location tag value must be set via env INFLUX_LOCATION_TAG")
	}
	tags := map[string]string{"location": loc}

	if host == "" || portstr == "" || db == "" {
		return fmt.Errorf("host, port, and database are required. Set env INFLUXDB_HOST, INFLUXDB_PORT, INFLUXDB_DB")
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		return err
	}
	influxClient := NewInfluxClient(host, port, db, "", "")
	sleep_time := 1 * time.Second
	for 1 < 2 {
		samples, err := getSamples(cCtx)
		if err != nil {
			fmt.Println("error getting sensors", err)
			sleep_time = time.Duration(rand.Float32()*20.0+5.0) * time.Second
		} else {
			err := publishInfluxDb(influxClient, measurement, tags, samples)
			if err != nil {
				fmt.Println("error publishing to InfluxdB")
				sleep_time = time.Duration(rand.Float32()*20.0+5.0) * time.Second
			} else {
				sleep_time = 1 * time.Minute
			}

		}
		fmt.Println(time.Now().Format(time.RFC3339) + fmt.Sprintf(" sleeping %s", sleep_time))
		time.Sleep(sleep_time)
	}
	return nil
}

func main() {

	app := &cli.App{
		Name:  "purpleair",
		Usage: "interact with the purpleair api",
		Commands: []*cli.Command{
			{
				Name:    "influx",
				Aliases: []string{"s"},
				Usage:   "get sensors from the purpleair api and post to influx",
				Action:  getSensorsToInflux,
			},
			{
				Name:    "sensors",
				Aliases: []string{"s"},
				Usage:   "get sensors from the purpleair api and print JSON",
				Action:  getSensorsToJson,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "readkey", Aliases: []string{"r"}},
			&cli.StringFlag{Name: "writekey", Aliases: []string{"w"}},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
