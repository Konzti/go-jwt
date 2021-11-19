# syntax=docker/dockerfile:1

FROM golang:1.17-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
COPY . ./
RUN go build -o app

FROM alpine:3.14 as production

# Copy built binary from builder
COPY --from=builder app .
# Expose port
EXPOSE 3000
# Exec built binary
CMD ./app