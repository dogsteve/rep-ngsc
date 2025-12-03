# Use a lightweight Go base image
FROM golang:1.25 as builder

# Set working directory
WORKDIR /app

# Use build argument to set service name
ARG BUILD_SERVICE_NAME

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN mkdir should_copy_data
RUN go build -o /app/should_copy_data/main ./main.go
RUN find . -maxdepth 1 -type f -exec cp {} /app/should_copy_data/ \;

# Use a minimal base image for deployment
FROM alpine:latest

ENV SERVICE_NAME=none

# Set working directory
WORKDIR /root/

RUN apk add --no-cache tzdata
RUN apk add --no-cache libc6-compat
RUN apk add --no-cache curl

# 3. THIẾT LẬP biến môi trường TZ
# Đảm bảo bạn sử dụng tên múi giờ chuẩn IANA (ví dụ: Asia/Ho_Chi_Minh)
ENV TZ=Asia/Ho_Chi_Minh

# Copy the built binary from the builder stage
COPY --from=builder /app/should_copy_data/ .
RUN pwd
RUN ls -la

# Expose default Go app port (change as needed)
EXPOSE 8080

# Run the service
CMD ["sh", "-c", "/root/main"]