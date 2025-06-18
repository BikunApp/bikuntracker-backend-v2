FROM golang:1.23-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /work

RUN apk add --no-cache make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o exec .
RUN go build -o migrate ./db/migrations/migrate.go
RUN go build -o seeder ./scripts/seeder/main.go


FROM alpine:3.20

RUN apk --no-cache add tzdata

WORKDIR /work

COPY --from=builder /work/exec .
COPY --from=builder /work/migrate .
COPY --from=builder /work/seeder .
COPY --from=builder /work/db/migrations ./db/migrations
COPY --from=builder /work/fixtures ./fixtures
COPY --from=builder /work/startup.sh .

RUN chmod +x startup.sh

CMD ["./startup.sh"]
