FROM golang:1.25.11-bookworm AS builder

WORKDIR /app

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /main .

EXPOSE 8089

CMD ["./main"]