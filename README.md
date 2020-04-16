# koochooloo :baby:
> a Persian word which means "small" or "li'l". It is often used to refer to a girl when flirting, (with the meaning, li'l girl)
>
> Urban Dictionary

[![Drone (cloud)](https://img.shields.io/drone/build/1995parham/koochooloo.svg?style=flat-square)](https://cloud.drone.io/1995parham/koochooloo)

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
