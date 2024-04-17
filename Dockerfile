FROM golang:alpine AS builder

LABEL stage=gobuilder
ENV CGO_ENABLED 0
ENV GOOS linux

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /build/main ./cmd/main/main.go

FROM alpine
WORKDIR /app

COPY --from=builder /build/main /app/main

CMD ["/app/main"]