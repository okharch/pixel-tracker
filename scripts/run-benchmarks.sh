#!/bin/sh
set -e
echo "Running batch-size 10"
go run cmd/generate_pixels/pixel-generator.go -batch-size 10
echo "Running batch-size 100"
go run cmd/generate_pixels/pixel-generator.go -batch-size 100
echo "Running batch-size 200"
go run cmd/generate_pixels/pixel-generator.go -batch-size 200
echo "Running batch-size 400"
go run cmd/generate_pixels/pixel-generator.go -batch-size 400
