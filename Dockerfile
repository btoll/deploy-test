FROM golang:1.16-alpine

WORKDIR /go/src/test-app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["test-app"]

