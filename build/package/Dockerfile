FROM golang:alpine as builder

RUN apk add git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/cmd/koochooloo
RUN go build -o /koochooloo

FROM alpine:latest as release

LABEL maintainer="Parham Alvani <parham.alvani@gmail.com>"

WORKDIR /app/

COPY --from=builder /koochooloo .

EXPOSE 1378

ENTRYPOINT ["./koochooloo"]

CMD ["server"]
