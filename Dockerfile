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

ENTRYPOINT ["/hub-backend"]
