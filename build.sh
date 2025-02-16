#!/bin/bash

# Build the project
echo "Building the project..."

go build -o caching-proxy cmd/app/main.go

echo "Project built successfully!"
 
 