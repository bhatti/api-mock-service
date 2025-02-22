FROM golang:1.23.6-alpine as go-builder
COPY . /src
WORKDIR /src
ENV GO111MODULE=on
RUN apk add --no-cache git make bash
RUN apk add build-base
RUN make build

FROM alpine:3.21
ENTRYPOINT ["/api-mock-service"]

ARG HTTP_PORT
ARG DATA_DIR
ARG HISTORY_DIR
ARG ASSET_DIR
ENV \
  HTTP_PORT=${HTTP_PORT} \
  DATA_DIR=${DATA_DIR} \
  HISTORY_DIR=${HISTORY_DIR} \
  ASSET_DIR=${ASSET_DIR}

COPY --from=go-builder /src/out/bin/api-mock-service /api-mock-service
RUN apk add --no-cache ca-certificates
RUN adduser -S -D -H -h /api-mock-service svc-user
USER svc-user
