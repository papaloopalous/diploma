FROM golang:1.24-alpine AS builder
COPY . /go/src/logger
WORKDIR /go/src/logger
RUN go build -o logger ./cmd/

FROM alpine AS runtime
WORKDIR /app
COPY --from=builder /go/src/logger/logger /app/
RUN chmod +x ./logger
EXPOSE 8080/tcp
ENTRYPOINT ./logger