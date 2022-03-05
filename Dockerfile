FROM golang:1.13-alpine as builder

# Install our build tools
RUN apk add --update git make bash ca-certificates

WORKDIR /go/src/github.com/daspawnw/training-server
COPY . ./
RUN CGO_ENABLED=0 go build -o bin/training-server ./cmd/training-server/...

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/daspawnw/training-server/bin/training-server /training-server

ENTRYPOINT ["/training-server"]