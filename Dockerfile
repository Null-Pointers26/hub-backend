FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hub-backend .

FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=builder /app/hub-backend /hub-backend

EXPOSE 80 443 8080

ENV APP_ENV=development
ENV DOMAIN=localhost
ENV CERT_EMAIL=mail@skola.cz
ENV CERT_CACHE_DIR=/certs
ENV FRONTEND_TARGET=http://frontend:3000
ENV DEV_LISTEN_ADDR=:8080
ENV HTTP_ADDR=:80
ENV HTTPS_ADDR=:443

ENTRYPOINT ["/hub-backend"]
