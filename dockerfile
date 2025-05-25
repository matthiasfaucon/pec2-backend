FROM golang:1.23
RUN echo "Building at $(date)" > /tmp/build_time
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main

EXPOSE 8080

CMD ["/app/main"]