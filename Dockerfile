FROM golang:1.10.1-alpine

WORKDIR /go/src/app
COPY ./src .

RUN apk add --no-cache git

RUN go get -d -v ./...
RUN go install -v ./...

RUN go build

CMD ["app"]