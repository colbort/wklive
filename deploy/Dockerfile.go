ARG GO_VERSION=1.26.4
FROM golang:${GO_VERSION}-alpine AS builder

ARG SERVICE_DIR
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY . .
WORKDIR /src/${SERVICE_DIR}
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/service .

FROM alpine:3.23

RUN apk add --no-cache busybox-extras ca-certificates tzdata \
    && addgroup -S wklive \
    && adduser -S -G wklive wklive

WORKDIR /app
COPY --from=builder /out/service /app/service

USER wklive
ENTRYPOINT ["/app/service"]
