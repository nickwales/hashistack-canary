#
# builder image
#
FROM golang:1.13-alpine3.11 as builder
ENV GOBIN $GOPATH/go/
RUN mkdir /build
ADD *.go go.* /build/
WORKDIR /build
RUN apk add --no-cache git
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -o hashistack-canary .
#
# final image
#
FROM alpine:3.11.3

# copy binary into container
COPY --from=builder /build/hashistack-canary .

ENTRYPOINT [ "./hashistack-canary" ]