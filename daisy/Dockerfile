FROM alpine

RUN apk add --no-cache ca-certificates apache2-utils

COPY gopath/bin/daisy /daisy
# COPY /workflows/ /workflows/

ENV GCE_METADATA_HOST 172.17.0.2

ENTRYPOINT ["/daisy"]

