FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/topq ./cmd/topq

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /out/topq /app/topq

EXPOSE 8081
ENTRYPOINT ["/app/topq"]
