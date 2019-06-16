# Build the manager binary
FROM registry.cn-hangzhou.aliyuncs.com/knative-sample/golang:1.12 as builder

# Copy in the go src
WORKDIR /go/src/github.com/knative-sample/knload/
COPY cmd/   cmd/
COPY pkg/   pkg/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/knload github.com/knative-sample/knload/cmd/

FROM registry.cn-hangzhou.aliyuncs.com/knative-sample/alpine-sh:3.9-263c0d94

WORKDIR /app/
RUN mkdir -p /app/bin/
COPY --from=builder /go/src/github.com/knative-sample/knload/bin/knload /app/bin/knload
ENTRYPOINT ["/app/bin/knload"]


