#!/bin/bash

# Complete External Share Distribution and Decryption Flow
# This script runs all steps in sequence

set -e

echo "🚀 Complete External Share Distribution and Decryption Flow"
echo "=========================================================="
echo ""
echo "This script will run the complete flow:"
echo "1. Generate keys for Genesis Boot"
echo "2. Start TEE and Host services"
echo "3. Run Genesis Boot (generate encrypted shares)"
echo "4. Distribute shares to members"
echo "5. Decrypt shares using member keys"
echo "6. Inject decrypted shares back to TEE"
echo "7. Verify complete flow"
echo ""

# Make all scripts executable
chmod +x scripts/*.sh

# Run all steps
echo "🔄 Starting complete flow..."
echo ""

echo "Step 1/7: Generating keys..."
./scripts/01-generate-keys.sh
echo ""

echo "Step 2/7: Starting services..."
./scripts/02-start-services.sh
echo ""

echo "Step 3/7: Running Genesis Boot..."
./scripts/03-genesis-boot.sh
echo ""

echo "Step 4/7: Distributing shares..."
./scripts/04-distribute-shares.sh
echo ""

echo "Step 5/7: Decrypting shares..."
./scripts/05-decrypt-shares.sh
echo ""

echo "Step 6/7: Injecting shares..."
./scripts/06-inject-shares.sh
echo ""

echo "Step 7/7: Verifying results..."
./scripts/07-verify-results.sh
echo ""

echo "🎉 COMPLETE FLOW SUCCESSFUL!"
echo "============================"
echo "All steps completed successfully. The external share distribution"
echo "and decryption flow has been fully implemented and tested."
echo ""
echo "The quorum key has been reconstructed from distributed shares"
echo "and is now available in the TEE for use."

