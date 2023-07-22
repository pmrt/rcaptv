FROM golang:1.20-alpine AS build
WORKDIR /src
# deps
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# source code
COPY . .

# Build. Don't use libc, the resulting binary will be statically linked agaisnt
# the libraries
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ARG IS_PROD
RUN if [ "$IS_PROD" = "1" ]; \
  then go build -tags RELEASE -o /build/api ./cmd/api; \
  else go build -o /build/api ./cmd/api; \
  fi
RUN mkdir -p /build/migrations/postgres &&\
  cp /src/database/postgres/migrations/* /build/migrations/postgres/ &&\
  chmod +x /build/api

FROM gcr.io/distroless/static-debian11:latest-amd64 AS final
WORKDIR /app
COPY --from=build /build .
EXPOSE 3021
ENTRYPOINT ["/app/api"]
