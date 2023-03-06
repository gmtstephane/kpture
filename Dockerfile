FROM golang:1.19-alpine as base

ARG BUILDTAG
RUN apk update
RUN apk add build-base 
RUN if [ "$BUILDTAG" = "agent" ]; then apk add libpcap-dev libpcap; fi

COPY . /app/service
WORKDIR /app/service
# Build the Go app
RUN go build --tags $BUILDTAG -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o /app/service/kpture .

FROM scratch
# Copy our static executable.
COPY --from=base /app/service/kpture /kpture
ENTRYPOINT ["/kpture"]