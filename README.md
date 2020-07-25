# koochooloo :baby:
> a Persian word which means "small" or "li'l". It is often used to refer to a girl when flirting, (with the meaning, li'l girl)
>
> Urban Dictionary

[![Drone (cloud)](https://img.shields.io/drone/build/1995parham/koochooloo.svg?style=flat-square&logo=drone)](https://cloud.drone.io/1995parham/koochooloo)

## Introduction
Here is a mini project for shortening your URLs.
This sweet project shows how to write a simple lovely Golang's project that contains the Database, Configuration,
and, etc. You can use this project as a guidance to write ReST applications.

I want to dedicate this project to my love :heart:.

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
