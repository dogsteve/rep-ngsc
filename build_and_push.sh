#!/bin/bash

# 1. Khai báo các biến
REPO="harbor.ngsd.vn/chamcong/chamcong"
TAG=$1

# 2. Kiểm tra tham số đầu vào
if [ -z "$TAG" ]; then
    echo "Lỗi: Bạn chưa nhập version tag!"
    echo "Sử dụng: ./build_and_push.sh v1.0.0"
    exit 1
fi

FULL_IMAGE_NAME="$REPO:$TAG"

echo "=========================================="
echo "Bắt đầu build image (linux/amd64): $FULL_IMAGE_NAME"
echo "=========================================="

# 3. Thực hiện Build Docker Image với platform linux/amd64
# --load: đảm bảo image sau khi build sẽ được nạp vào docker local để sẵn sàng push
docker build --platform linux/amd64 -t "$FULL_IMAGE_NAME" .

# Kiểm tra lỗi build
if [ $? -ne 0 ]; then
    echo "Lỗi: Build Docker image thất bại!"
    exit 1
fi

echo "------------------------------------------"
echo "Build thành công hệ linux/amd64. Đang chuẩn bị push..."
echo "------------------------------------------"

# 4. Push image lên Harbor
docker push "$FULL_IMAGE_NAME"

# 5. Kiểm tra kết quả push
if [ $? -eq 0 ]; then
    echo "=========================================="
    echo "Thành công! Image linux/amd64 đã có trên repo:"
    echo "$FULL_IMAGE_NAME"
    echo "=========================================="
else
    echo "Lỗi: Không thể push image."
    exit 1
fi