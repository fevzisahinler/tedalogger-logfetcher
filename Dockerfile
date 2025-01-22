FROM golang:1.23.3-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o tedalogger-logfetcher cmd/app/main.go

FROM alpine:edge

WORKDIR /app

COPY --from=build /app/tedalogger-logfetcher .

COPY .env /app

RUN apk --no-cache add tzdata

EXPOSE 4000

ENTRYPOINT ["/app/tedalogger-logfetcher"]