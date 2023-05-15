FROM golang:1.16.0-alpine3.13 as builder

RUN apk update
RUN apk add gcc g++

WORKDIR /build

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY go.mod .
COPY go.sum .

RUN cd cmd/kronos && go mod download

RUN cd cmd/kronos && go build -o kronos .

FROM alpine:3.13

WORKDIR /app

COPY --from=builder /build/cmd/kronos/kronos .

ENTRYPOINT ["/app/kronos"]

#RUN "dbus-run-session"

CMD ["run"]
