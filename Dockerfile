FROM golang:1.24-alpine AS build
WORKDIR /src
ENV GOTOOLCHAIN=auto
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /goai ./cmd/goai

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /goai /usr/local/bin/goai
ENV PORT=8088
EXPOSE 8088
ENTRYPOINT ["/usr/local/bin/goai"]
