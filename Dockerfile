FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o mailer ./cmd/mailer

FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=builder /app/mailer .

EXPOSE 8080

CMD ["./mailer"]
