FROM golang:alpine as builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN apk add -U --no-cache git ca-certificates
RUN go get -v -t -d ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o trends-analyzer .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/trends-analyzer /app/
WORKDIR /app
ENTRYPOINT ["./trends-analyzer"]
