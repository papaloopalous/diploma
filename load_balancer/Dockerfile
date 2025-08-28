FROM golang:1.24-alpine AS builder
COPY . /go/src/load_balancer
WORKDIR /go/src/load_balancer
RUN go build -o load_balancer ./cmd/

FROM alpine AS runtime
WORKDIR /app
COPY --from=builder /go/src/load_balancer/load_balancer /app/
RUN chmod +x ./load_balancer
EXPOSE 8080/tcp
ENTRYPOINT ./load_balancer