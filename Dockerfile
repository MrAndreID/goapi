FROM golang:1.24-alpine3.22

ENV TZ=Asia/Jakarta

RUN apk update && \
    apk add --no-cache nano curl gcc g++ make libwebp-dev

RUN mkdir /app

ADD . /app

WORKDIR /app

COPY .env.example .env

COPY  go.mod  .

RUN go mod tidy

RUN go build -o engine ./

CMD ["./engine"]