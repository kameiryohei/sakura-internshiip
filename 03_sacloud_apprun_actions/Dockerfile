# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN apk add --no-cache gcc musl-dev sqlite-dev \
    && go mod download

COPY . 

RUN CGO_ENABLED=1 GOOS=linux go build -o app

# Run stage
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/app /app/app

RUN apk add --no-cache bash curl ca-certificates sqlite-libs

ARG PORT=8080
ENV PORT=$PORT
EXPOSE $PORT

ENTRYPOINT ["/app/app"]
