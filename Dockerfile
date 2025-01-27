# Gunakan base image Golang yang sesuai
FROM golang:1.23.4

# Set working directory di dalam container
WORKDIR /app

# Salin module files (jika ada) dan unduh dependencies
COPY go.mod go.sum ./
RUN go mod download || go mod init Gin-Inventory && go mod tidy

# Salin seluruh isi proyek ke dalam container
COPY . /app

# Build aplikasi
RUN go build -o main .

# Expose port
EXPOSE 8080

# Jalankan aplikasi
CMD ["./main"]