FROM golang:latest as dev

RUN apt-get install --no-install-recommends -y neovim && \
  go install -v golang.org/x/tools/gopls@latest

FROM golang:alpine as builder

# Install git command required for building with go1.18
RUN apk add git

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
WORKDIR /app/cmd/koochooloo
RUN go build -o /koochooloo

FROM alpine:latest as release

# Add Maintainer Info
LABEL maintainer="Parham Alvani <parham.alvani@gmail.com>"

WORKDIR /app/

COPY --from=builder /koochooloo .

# Expose port 1378 to the outside world
EXPOSE 1378

ENTRYPOINT ["./koochooloo"]

# Run server
CMD ["server"]
