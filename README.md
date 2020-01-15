# koochooloo :baby:
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
curl -X POST -d '{"url": "www.google.com"}' -H 'Content-Type: application/json' 127.0.0.1:8080/api/urls
curl -L 127.0.0.1:8080/api/CKaniA
```
