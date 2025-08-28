FROM golang:1.24-alpine AS builder
COPY . /go/src/postgre_api
WORKDIR /go/src/postgre_api
RUN go build -o postgre_api ./cmd/

FROM alpine AS runtime
WORKDIR /app
COPY --from=builder /go/src/postgre_api/postgre_api /app/
RUN chmod +x ./postgre_api
EXPOSE 8080/tcp
ENTRYPOINT ./postgre_api