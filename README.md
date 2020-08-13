# koochooloo :baby:
> a Persian word which means "small" or "li'l". It is often used to refer to a girl when flirting, (with the meaning, li'l girl)
>
> Urban Dictionary

[![Drone (cloud)](https://img.shields.io/drone/build/1995parham/koochooloo.svg?style=flat-square&logo=drone)](https://cloud.drone.io/1995parham/koochooloo)
[![Codecov](https://img.shields.io/codecov/c/gh/1995parham/koochooloo?logo=codecov&style=flat-square)](https://codecov.io/gh/1995parham/koochooloo)
![Docker Image Size (tag)](https://img.shields.io/docker/image-size/1995parham/koochooloo/latest?style=flat-square&logo=docker)
![Docker Pulls](https://img.shields.io/docker/pulls/1995parham/koochooloo?style=flat-square&logo=docker)

## Introduction
Here is a mini project for shortening your URLs.
This sweet project shows how to write a simple lovely Golang's project that contains the Database, Configuration,
and, etc. You can use this project as a guidance to write your ReST applications.
This project try to be strongly typed, easy to read and easy to maintain therefore there is no global variable, `init` function and etc.
We have used the singular name for package as a de-facto standard.

I want to dedicate this project to my love :heart:.

## Structure
### Binaries
First of all, `cmd` package contains the binaries of this project with use of [cobra](https://github.com/spf13/cobra).
It is good to have a simple binary for database migrations that can be run on initiation phase of project.
Each binary has its `main.go` in its package and register itself with a `Register` function.
In the `root.go` of `cmd` configuration and other shared things are initiated.

### Configuration
The main part of each application is its configuration. There are many ways for having configuration in the project from configuration file to environment variables. [koanf](https://github.com/knadh/koanf) has all of them. The main points here are:

- having a defined and typed structure for configuration
- don't use global configuration. each module has its configuration defined in `config` module and it will pass to it in its initiation.

P.S. [koanf](https://github.com/knadh/koanf) is way better than [viper](https://github.com/spf13/viper) for having typed configuration.
By typed configuration I mean you have a defined structure for configuration and then load configuration from many sources into it.

### Database
There is a `db` package that is responsible for connecting to the database. This package use the database configuration that is defined in `config` module and create a database instance. It is good to ping your database here to have fully confident to your database instance.

### Model
Project models are defined in `model` package. These models are used internally but the can be used in `response` or `request` package.
There is no structure for communicating with database in this package.

### Store
Stores are responsible for commnunicating with database to store or retrieve models. Stores are `interface` and there is an SQL and moked version for them.
SQL model is used in main code and mocked is used for tests. Please note that the tests for SQL stores are touchy and are done with actual database.

### Handler
HTTP handler are defined in `handler` package. [Echo](https://github.com/labstack/echo) is an awesome HTTP framework that has eveything you need. Each handler has its structure with a `Register` method that registers its route into a given route group. Route group is a concept from [Echo](https://github.com/labstack/echo) framework for grouping routes under a specific parent path. Each handler has what it needs into its structure. Handler structure are created in `main.go` then register on their group.

```go
type Healthz struct {
}

// Handle shows server is up and running.
func (h Healthz) Handle(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// Register registers the routes of healthz handler on given echo group.
func (h Healthz) Register(g *echo.Group) {
	g.GET("/healthz", h.Handle)
}
```

### Metrics
All metrics are gathered using [prometheus](https://prometheus.io/). Each package has its `metric.go` that defines a structure contains the metrics and have methods for changing them. For migrating from prmotheus you just need to change `metric.go`. Metrics aren't global and they created for each instance seperately and with the following code there is no issue with duplicate registration for prometheus metrics.


```go
// my_mertic is a prometheus.Counter

if err := prometheus.Register(my_metric); err != nil {
    if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
        my_metric = are.ExistingCollector.(prometheus.Counter)
    } else {
        panic(err)
    }
}
```

For having better controller on metrics endpoint there is another HTTP server that is defined in `metric` package for monitoring.

## Up and Running
This project only requires MongoDB, and you can run it with provided `docker-compose`.

```sh
go build
docker-compose up -d
./koochooloo
```

```sh
curl -X POST -d '{"url": "https://elahe-dastan.github.io"}' -H 'Content-Type: application/json' 127.0.0.1:1378/api/urls
curl -L 127.0.0.1:1378/api/CKaniA
```

## Load Testing

```
    checks.....................: 99.83% ✓ 2995  ✗ 5
    data_received..............: 2.0 MB 64 kB/s
    data_sent..................: 521 kB 17 kB/s
    group_duration.............: avg=649.18ms min=153.18µs med=265.45ms max=30.95s   p(90)=1.61s    p(95)=2.06s
    http_req_blocked...........: avg=14.12ms  min=0s       med=3µs      max=1.65s    p(90)=13µs     p(95)=147.04µs
    http_req_connecting........: avg=6.23ms   min=0s       med=0s       max=1.36s    p(90)=0s       p(95)=0s
    http_req_duration..........: avg=272.98ms min=0s       med=127.99ms max=4.81s    p(90)=830.93ms p(95)=1.29s
    http_req_receiving.........: avg=125.23µs min=0s       med=60µs     max=11.21ms  p(90)=228µs    p(95)=363µs
    http_req_sending...........: avg=50.78µs  min=0s       med=22µs     max=7.28ms   p(90)=86µs     p(95)=138µs
    http_req_tls_handshaking...: avg=7.86ms   min=0s       med=0s       max=653.63ms p(90)=0s       p(95)=0s
    http_req_waiting...........: avg=272.8ms  min=0s       med=127.71ms max=4.81s    p(90)=830.87ms p(95)=1.29s
    http_reqs..................: 4000   129.093962/s
    iteration_duration.........: avg=1.29s    min=142.34ms med=1.04s    max=30.97s   p(90)=2.18s    p(95)=2.64s
    iterations.................: 1000   32.273491/s
    vus........................: 100    min=100 max=100
    vus_max....................: 100    min=100 max=100
```
