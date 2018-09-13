FROM golang:alpine as builder
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN apk add git
RUN go get -v -t -d ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o trend-analyzer .

FROM scratch
COPY --from=builder /build/trends-analyzer /app/
WORKDIR /app
ENTRYPOINT ["./trends-analyzer"]
