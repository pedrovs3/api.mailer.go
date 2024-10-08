FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o mailer

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=builder /app/mailer .

CMD ["./mailer"]
