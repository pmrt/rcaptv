FROM golang:1.20-alpine AS build

WORKDIR /src

# deps
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# source code
COPY . .

# Build. Don't use libc, the resulting binary will be statically linked agasint
# the libraries
ENV CGO_ENABLED=0
# RUN go build -tags RELEASE -o /usr/local/bin/tracker ./cmd/tracker
RUN go build -o /usr/local/bin/rcaptv ./cmd/rcaptv

EXPOSE 3021
ENTRYPOINT [ "rcaptv" ]