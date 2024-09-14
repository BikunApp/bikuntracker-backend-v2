# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /work

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o exec .

FROM alpine:3.20

RUN apk --no-cache add tzdata

WORKDIR /work

COPY --from=builder /work/exec .

ENTRYPOINT ["./exec"]
