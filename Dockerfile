FROM golang:alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o config-chg .


FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/config-chg .
COPY ./fixtures ./fixtures
CMD ["./config-chg"]