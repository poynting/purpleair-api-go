# purpleair-api-go
golang client for getting data from the [purpleair api](https://api.purpleair.com/). To use this it's necessary to request an api key via email, which seems like an automated process and usually comes back in a day or so.

# Usage
The api purpleair module can be used standalone in your own code. There is also a very basic cli client that shows how to use the data and can output json or push data to an influxdb server.

```
go install github.com/poynting/purpleair-api-go@latest 
```

Then in your code:
```
import github.com/poynting/purpleair-api-go@/purpleair
```

## Log from Purpleair to influx
The install will install the purpleair-api-go executable in your GOPATH.  It's currently configured only using environment variables.

```
purpleair-api-go influx
```

## Print out sensors measurements from the PA api as json
```
purpleair-api-go influx
```

In VSCode/Powersheel I use a command line like this to test the cli
```
$env:PURPLEAIR_READ_KEY = 'MY-READ-KEY'; $env:INFLUXDB_HOST = 'localhost'; $env:INFLUXDB_PORT = '8086'; $env:INFLUXDB_DB = "purpleair"; $env:PURPLEAIR_LATITUDE = "33.3333"; $env:PURPLEAIR_LONGITUDE = "-96.6666"; $env:PURPLEAIR_RANGE_KM = "3"; $env:INFLUX_MEASUREMENT_NAME = "purpleair"; $env:INFLUX_LOCATION_TAG = "home"; go run .\main.go influx
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

Or, if not using balena, this builds tiny containers:
```
FROM golang:1.19-alpine AS build-env
ENV GOBIN /app
RUN CGO_ENABLED=0 go install -v -ldflags '-extldflags "-static" -s -w' github.com/poynting/purpleair-api-go@latest 

FROM alpine
COPY --from=build-env /app/purpleair-api-go /purpleair-api-go
CMD ["/purpleair-api-go","influx"]
```

With the dockerfile in an empty folder, I use 
```
docker build -t purpleair-influx .
```
Which results in a ~12MB container image.

## Container setup

To use the containerized version all of the parameters are set as environment variables. First set up the purpleair API information, and the location and range you want to query. The location and range is converted to a bounding box inscribed within a circle of the radius, centered on the location.

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

## Example of aggregating and plotting local sensor data in Grafana
![image](https://user-images.githubusercontent.com/6610131/196003936-e27a7be8-32ee-4fa9-b816-21a60c80a358.png)

# Comments
But the purpleair api is just JSON, why not use telegraf? 

This is how I aggregate a lot of other services, but I found that in this use case the data coming out of the PurpleAir API isn't well formed for (my understanding of) the telegraf transformers. In addition, there are a couple of extra transformation functions the PurpleAir data needs to go through before it's useful to me, nameley, the calibration described in [this paper](https://www.sciencedirect.com/science/article/abs/pii/S135223102100251X?via%3Dihub) to get calibrated ug/m^3 and then the particulate to AQI transformation, which is peicewise linear.
