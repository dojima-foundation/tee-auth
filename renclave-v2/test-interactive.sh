#!/bin/bash
echo "🧪 Interactive renclave-v2 Testing Guide"
echo "========================================"

echo "🐳 Starting Docker container with shell access..."
docker run -it --rm --name renclave-interactive -p 3001:3000 renclave-v2:latest /bin/bash
