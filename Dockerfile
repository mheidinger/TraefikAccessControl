FROM golang:1.13-alpine as builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o TraefikAccessControl ./cmd/TraefikAccessControl

FROM alpine:3.10

WORKDIR /app
COPY --from=builder /build/TraefikAccessControl .

ENV GIN_MODE release

EXPOSE 4181
CMD ["./TraefikAccessControl"]