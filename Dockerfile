FROM golang:1.22-rc-alpine3.19 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o main .

FROM alpine:3.19

WORKDIR /app

COPY --from=build /app/main .
COPY static/ static/

EXPOSE 8080

CMD ["./main"]