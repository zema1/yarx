FROM golang:1.17-alpine as builder

LABEL Author=Koalr(https://github.com/zema1)

WORKDIR /app

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w -extldflags=-static" cmd/yarx



FROM alpine

WORKDIR /app

COPY --from=builder /app/yarx /app/yarx
ADD ./pocs /app/pocs
ADD ./assets/html /app/html

EXPOSE 8080

ENTRYPOINT ["./yarx"]

CMD ["-p", "./pocs", "-l", "0.0.0.0:8080", "-r", "./html"]