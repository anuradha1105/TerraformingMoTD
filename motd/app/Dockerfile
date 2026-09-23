FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /motd-web .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /motd-web /usr/local/bin/motd-web
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/motd-web"]
