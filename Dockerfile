FROM golang:1.17-alpine as builder

WORKDIR /app 

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o wg-http-proxy . 

FROM scratch

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/wg-http-proxy /usr/bin/

ENTRYPOINT ["wg-http-proxy"]
