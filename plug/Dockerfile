FROM golang:alpine as builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN GOPROXY=goproxy.cn go mod download
COPY main.go /app/
COPY chaincode /app/chaincode

RUN GOPROXY=goproxy.cn go build --ldflags "-s -w" -o server

FROM alpine:latest
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories
RUN apk upgrade --update-cache --available
RUN apk add openssl && rm -rf /var/cache/apk/*

COPY --from=builder /app/server /app/server
CMD ["/app/server"]
