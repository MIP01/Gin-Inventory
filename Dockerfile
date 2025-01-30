# Gunakan base image Golang untuk tahap build
FROM golang:1.23.4 AS builder

# Set working directory di dalam container
WORKDIR /app

# Salin module files (jika ada) dan unduh dependencies
COPY go.mod go.sum ./
RUN go mod download || go mod init Gin-Inventory && go mod tidy

# Salin seluruh isi proyek ke dalam container
COPY . /app

# Build aplikasi (statis)
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Gunakan base image yang lebih kecil untuk tahap runtime
FROM debian:trixie-slim

# Salin binary dari tahap build
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

# Jalankan aplikasi
CMD ["./main"]