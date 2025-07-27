FROM golang:1.24.4

WORKDIR /api-wallet

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# RUN CGO_ENABLED=0 GOOS=linux go build -o ./api

EXPOSE 8080

CMD ["go", "run", "./api"]