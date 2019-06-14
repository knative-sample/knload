# Build the manager binary
FROM golang:1.12 as builder

# Copy in the go src
WORKDIR /go/src/github.com/knative-samples/knload/
COPY cmd/   cmd/
COPY pkg/   pkg/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/httpload code.aliyun.com/knative-samples/http-client/cmd/

FROM alpine:3.9
WORKDIR /app/
RUN mkdir -p /app/bin/
COPY --from=builder /go/src/code.aliyun.com/knative-samples/http-client/bin/httpload /app/bin/httpload
ENTRYPOINT ["/app/bin/httpload"]


