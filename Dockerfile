FROM golang:1.10.1 AS build-env

WORKDIR /go/src/app
COPY ./src .

RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mutterblack-discord .

# Build runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=build-env /go/src/app/mutterblack-discord .

CMD ["./mutterblack-discord"]