FROM golang:1.22-rc-alpine3.19 AS build

WORKDIR /app

COPY . .
RUN go mod download;  go build -a -o main .

FROM alpine:3.19

WORKDIR /app

COPY --from=build /app/main .
COPY static/ static/

EXPOSE 8080

CMD ["./main"]