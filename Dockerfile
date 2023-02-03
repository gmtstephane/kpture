FROM golang:1.19-alpine as base


RUN apk update
RUN apk add build-base libpcap-dev libpcap
WORKDIR /app

COPY . .
RUN go mod download

# Build the Go app
RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o /app/kpture cmd/agent/main.go 

FROM scratch
# Copy our static executable.
COPY --from=base /app/kpture /kpture
# Run the hello binary.
CMD ["/kpture"]