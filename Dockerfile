# golang alpine 1.13.x
FROM golang:1.16-alpine as builder

ARG TARGETARCH
ARG TARGETOS

LABEL maintainer="info@reinkrul.nl"

ENV GO111MODULE on
ENV GOPATH /

RUN mkdir /app && cd /app
COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /app/nuts-explorer

# alpine 3.12.x
FROM alpine:3.12
COPY --from=builder /app/nuts-explorer /usr/bin/nuts-explorer

HEALTHCHECK --start-period=5s --timeout=5s --interval=5s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

EXPOSE 8080
ENTRYPOINT ["/usr/bin/nuts-explorer"]
