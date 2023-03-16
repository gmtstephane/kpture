FROM golang:1.20-alpine as base

ARG BUILDTAG
ARG UID=1000
ARG GID=1000

RUN apk update && \
  apk add --no-cache build-base shadow && \
  groupadd -g $GID appgroup && \
  useradd -u $UID -g $GID -s /bin/sh -m kpture

RUN if [ "$BUILDTAG" = "agent" ]; then apk add libpcap-dev libpcap libcap; fi

COPY . /app/service
WORKDIR /app/service
RUN go build --tags $BUILDTAG -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o /app/service/kpture .
RUN if [ "$BUILDTAG" = "agent" ]; then setcap 'cap_net_raw+ep' /app/service/kpture; fi

FROM scratch
USER kpture:appgroup
COPY --from=base /etc/passwd /etc/group /etc/
COPY --from=base --chown=kpture:appgroup /app/service/kpture /kpture
ENTRYPOINT ["/kpture"]