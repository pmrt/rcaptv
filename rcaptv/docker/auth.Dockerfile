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
  then go build -tags RELEASE -o /build/auth ./cmd/auth; \
  else go build -o /build/auth ./cmd/auth; \
  fi
RUN mkdir -p /build/migrations/postgres /build/x509 &&\
  cp /src/database/postgres/migrations/* /build/migrations/postgres/ &&\
  cp /src/certs/x509/* /build/x509/ &&\
  chmod +x /build/auth

FROM gcr.io/distroless/static-debian11:latest-amd64 AS final
WORKDIR /app
COPY --from=build /build .
EXPOSE 4001
ENTRYPOINT ["/app/auth"]
