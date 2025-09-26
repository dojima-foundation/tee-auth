#!/bin/bash

# Step 6: Share Injection
# This script injects decrypted shares back to TEE for reconstruction

set -e

echo "💉 Step 6: Share Injection"
echo "=========================="

# Check if share injection request exists
if [ ! -f "share_injection_request.json" ]; then
    echo "❌ Share injection request file not found: share_injection_request.json"
    echo "   Please run Step 5 first: ./scripts/05-decrypt-shares.sh"
    exit 1
fi

# Check if services are still running
if ! curl -s http://localhost:3000/health > /dev/null 2>&1; then
    echo "❌ Host service is not responding"
    echo "   Please run Step 2 first: ./scripts/02-start-services.sh"
    exit 1
fi

echo "💉 Injecting decrypted shares back to TEE..."
echo "   Request file: share_injection_request.json"
echo "   API endpoint: http://localhost:3000/enclave/inject-shares"

# Make the Share Injection API call
curl -X POST http://localhost:3000/enclave/inject-shares \
  -H "Content-Type: application/json" \
  -d @share_injection_request.json \
  -o share_injection_response.json \
  -w "HTTP Status: %{http_code}\n" \
  -s

if [ $? -ne 0 ]; then
    echo "❌ Share injection API call failed"
    exit 1
fi

# Check if response file was created
if [ ! -f "share_injection_response.json" ]; then
    echo "❌ Share injection response file not created"
    exit 1
fi

# Check HTTP status
HTTP_STATUS=$(curl -X POST http://localhost:3000/enclave/inject-shares \
  -H "Content-Type: application/json" \
  -d @share_injection_request.json \
  -w "%{http_code}" \
  -o /dev/null \
  -s)

if [ "$HTTP_STATUS" != "200" ]; then
    echo "❌ Share injection failed with HTTP status: $HTTP_STATUS"
    echo "Response:"
    cat share_injection_response.json
    exit 1
fi

echo "✅ Share injection completed successfully!"

# Display the response
echo ""
echo "📊 Share Injection Response:"
echo "============================"
cat share_injection_response.json | jq '{
  success: .success,
  reconstructed_quorum_key: (.reconstructed_quorum_key | length)
}'

# Verify the reconstruction
if [ "$(cat share_injection_response.json | jq -r '.success')" = "true" ]; then
    echo ""
    echo "🎉 SUCCESS: Quorum key reconstructed successfully!"
    echo "   - Reconstructed key: $(cat share_injection_response.json | jq -r '.reconstructed_quorum_key | length') bytes"
    echo "   - Key is now available in TEE"
else
    echo ""
    echo "❌ FAILED: Quorum key reconstruction failed"
    echo "   Check the response for error details"
fi

echo ""
echo "📁 Files Created:"
echo "  - share_injection_response.json (Share injection API response)"
echo ""
echo "✅ Step 6 Complete: Share injection successful!"
echo "   Next: Run Step 7 - Verification"
echo "   Command: ./scripts/07-verify-results.sh"

