# koochooloo :baby:
[![Drone (cloud)](https://img.shields.io/drone/build/1995parham/koochooloo.svg?style=flat-square)](https://cloud.drone.io/1995parham/koochooloo)

## Introduction
Here is a mini project for shortening your URLs.
This sweet project shows how to write a simple lovely Golang project that contains the Database, Configuration, and etc.

## Up and Running

```sh
go build
docker-compose up -d
./koochooloo
```

```sh
curl -X POST -d '{"url": "https://elahe-dastan.github.io"}' -H 'Content-Type: application/json' 127.0.0.1:1378/api/urls
curl -L 127.0.0.1:1378/api/CKaniA
```
