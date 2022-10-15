# purpleair-api-go
golang client for getting data from the purpleair api

# Setup

Set up the environment. First set up the purpleair API information, and the location and range you want to query.
The location and range is converted to a bounding box inscribed within a circle of the radius, centered on the location.

```
PURPLEAIR_READ_KEY = "YOUR-READ-KEY"
PURPLEAIR_LATITUDE = "33.333"
PURPLEAIR_LONGITUDE = "-96.666"
PURPLEAIR_RANGE_KM = "3"
```

To log to influxdb, set up the information

```
INFLUXDB_HOST = "localhost"
INFLUXDB_PORT = "8086"
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

# Dockerfile for balena service
I run a telegraf-influx-grafana stack on a raspberry pi, and use this dockerfile to build a very small image that saves the local PurpleAir data to influxdb, and then graphs it in grafana. 

```
FROM balenalib/raspberrypi4-64-debian-golang:latest AS build-env
WORKDIR /app
ENV GOBIN /app
RUN CGO_ENABLED=0 go install -v -ldflags '-extldflags "-static" -s -w' github.com/poynting/purpleair-api-go@latest 

FROM arm64v8/alpine
COPY --from=build-env /app/purpleair-api-go /purpleair-api-go
CMD ["/purpleair-api-go","influx"]
```
Or, if not using balena, this works as well and builds tiny containers:
```
FROM golang:1.19-alpine AS build-env
ENV GOBIN /app
RUN CGO_ENABLED=0 go install -v -ldflags '-extldflags "-static" -s -w' github.com/poynting/purpleair-api-go@latest 

FROM alpine
COPY --from=build-env /app/purpleair-api-go /purpleair-api-go
CMD ["/purpleair-api-go","influx"]
```
## Example of aggregating and plotting local sensor data in Grafana
![image](https://user-images.githubusercontent.com/6610131/196003936-e27a7be8-32ee-4fa9-b816-21a60c80a358.png)
