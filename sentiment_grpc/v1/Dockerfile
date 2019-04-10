FROM golang:alpine AS builder

RUN apk --update --no-cache add git protobuf

WORKDIR /go/src/sentiment

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

RUN ls /go/bin

# Next stage
FROM alpine

RUN apk --update --no-cache add ca-certificates openssl

COPY --from=builder /go/bin/sentiment /usr/local/bin

# Add credentials to image for now
COPY secrets/sentiment.credentials.json /usr/local/sentiment/
ENV GOOGLE_APPLICATION_CREDENTIALS="/usr/local/sentiment/sentiment.credentials.json"

CMD ["/usr/local/bin/sentiment"]
