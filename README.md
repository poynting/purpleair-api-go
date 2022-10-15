# purpleair-api-go
golang client for getting data from the purpleair api

# Usage
The api purpleair module can be standalone. There is also a very basic cli client that shows how to use the data and can output json or push data to an influxdb server.

```
import github.com/poynting/purpleair-api-go@/purpleair
```

Or install the (very basic) cli using
```
go install github.com/poynting/purpleair-api-go@latest 
```
Which will put the purpleair-api-go executable in your GOPATH.

## Log from Purpleair to influx

```
purpleair-api-go influx
```

## Print out sensors measurements from the PA api as json
```
purpleair-api-go influx
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

## Example of aggregating and plotting local sensor data in Grafana
![image](https://user-images.githubusercontent.com/6610131/196003936-e27a7be8-32ee-4fa9-b816-21a60c80a358.png)

# Comments
But the purpleair api is just JSON, why not use telegraf? 

This is how I aggregate a lot of other services, but I found that in this use case the data coming out of the PurpleAir API isn't well formed for (my understanding of) the telegraf transformers. In addition, there are a couple of extra transformation functions the PurpleAir data needs to go through before it's useful to me, nameley, the calibration described in [this paper](https://www.sciencedirect.com/science/article/abs/pii/S135223102100251X?via%3Dihub) to get calibrated ug/m^3 and then the particulate to AQI transformation, which is peicewise linear.
