FROM golang:alpine AS builder

RUN apk --update --no-cache add git protobuf

WORKDIR /go/src/app

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

RUN ls /go/bin

# Next stage
FROM alpine

RUN apk --update --no-cache add ca-certificates openssl

COPY --from=builder /go/bin/app /usr/local/bin

# Add credentials to image for now
COPY secrets/nlp.credentials.json /usr/local/app/
ENV GOOGLE_APPLICATION_CREDENTIALS="/usr/local/app/nlp.credentials.json"

CMD ["/usr/local/bin/app"]
