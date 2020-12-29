FROM golang:buster as build
WORKDIR /go/src/latest
COPY . /go/src/latest
RUN go get -d ./...
RUN go build -o /go/bin/latest ./cmd/latest

FROM gcr.io/distroless/base
COPY --from=build /go/bin/latest /
CMD ["/latest"]
