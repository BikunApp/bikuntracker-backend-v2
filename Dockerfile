FROM golang:1.23-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /work

RUN apk add --no-cache make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o exec .

FROM alpine:3.20

RUN apk --no-cache add tzdata make go

WORKDIR /work

COPY --from=builder /work/exec .
COPY --from=builder /work/Makefile .
COPY --from=builder /work/db/migrations ./db/migrations

CMD make migrate && ./exec
