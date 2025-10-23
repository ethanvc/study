#!/bin/bash

# 创建SSL目录（如果不存在）
mkdir -p ./nginx/ssl

# 生成自签名SSL证书
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout ./nginx/ssl/server.key \
    -out ./nginx/ssl/server.crt \
    -subj "/C=US/ST=State/L=City/O=Organization/OU=Department/CN=localhost"

echo "SSL证书已生成在 ./nginx/ssl 目录下"
    