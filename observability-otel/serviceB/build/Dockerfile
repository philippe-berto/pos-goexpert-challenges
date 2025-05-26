FROM golang:1.23.0 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cloudrun

FROM scratch
WORKDIR /app
COPY --from=build /app/cloudrun .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./cloudrun"]