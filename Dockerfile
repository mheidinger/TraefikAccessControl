FROM golang:1.13 as builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-linkmode external -extldflags -static" -o TraefikAccessControl ./cmd/TraefikAccessControl

FROM alpine:3.10

WORKDIR /app
COPY --from=builder /build/TraefikAccessControl .
COPY static ./static
COPY templates ./templates

COPY tac_data.json .

ENV GIN_MODE release

EXPOSE 4181
CMD ["./TraefikAccessControl", "-import_name", "tac_data.json", "-force_import"]