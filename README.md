# purpleair-api-go
golang client for getting data from the purpleair api

# Setup

Set up yor environment. First set up the purpleair API information, and the location and range you want to query.
The location and range is converted to a bounding box inscribed within a circle of the radius, centered on the location.
```
PURPLEAIR_READ_KEY = 'YOUR-READ-KEY'
PURPLEAIR_LATITUDE = "33.333"
PURPLEAIR_LONGITUDE = "-96.666"
PURPLEAIR_RANGE_KM = "3"
```

To log to influxdb, set up the information
```
INFLUXDB_HOST = 'localhost'
INFLUXDB_PORT = '8086'
INFLUXDB_DB = "purpleair"
INFLUX_MEASUREMENT_NAME = "purpleair"
INFLUX_LOCATION_TAG = "home"
```

# Usage
Log from Purpleair to influx

```
go run .\main.go influx
```

Print out sensors measurements from the PA api as json:
```
go run .\main.go influx
```
