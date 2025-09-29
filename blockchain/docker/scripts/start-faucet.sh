#!/bin/bash

set -e

echo "🚰 Starting Geth Faucet Service..."

# Wait for Geth to be ready
echo "⏳ Waiting for Geth to be ready..."
until curl -f http://geth-dev:8545 >/dev/null 2>&1; do
    echo "   Waiting for Geth RPC..."
    sleep 5
done

echo "✅ Geth is ready, starting faucet..."

# Faucet configuration
FAUCET_PRIVATE_KEY=${FAUCET_PRIVATE_KEY:-"0x4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"}
FAUCET_AMOUNT=${FAUCET_AMOUNT:-"1000000000000000000000"}  # 1000 ETH in wei
RPC_URL=${RPC_URL:-"http://geth-dev:8545"}

# Get faucet address from private key
FAUCET_ADDRESS=$(node -e "
const Web3 = require('web3');
const web3 = new Web3();
const account = web3.eth.accounts.privateKeyToAccount('$FAUCET_PRIVATE_KEY');
console.log(account.address);
")

echo "💰 Faucet Address: $FAUCET_ADDRESS"
echo "💧 Faucet Amount: $(echo $FAUCET_AMOUNT | sed 's/000000000000000000//') ETH per request"

# Create faucet server
cat > /blockchain/scripts/faucet-server.js << 'EOF'
const express = require('express');
const Web3 = require('web3');
const cors = require('cors');

const app = express();
const PORT = 3000;

// Middleware
app.use(cors());
app.use(express.json());

// Configuration
const RPC_URL = process.env.RPC_URL || 'http://geth-dev:8545';
const FAUCET_PRIVATE_KEY = process.env.FAUCET_PRIVATE_KEY || '0x4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d';
const FAUCET_AMOUNT = process.env.FAUCET_AMOUNT || '1000000000000000000000';

// Initialize Web3
const web3 = new Web3(RPC_URL);

// Create account from private key
const faucetAccount = web3.eth.accounts.privateKeyToAccount(FAUCET_PRIVATE_KEY);
web3.eth.accounts.wallet.add(faucetAccount);

console.log('🚰 Faucet initialized with address:', faucetAccount.address);

// Rate limiting (simple in-memory store)
const requests = new Map();
const RATE_LIMIT_WINDOW = 60000; // 1 minute
const MAX_REQUESTS_PER_WINDOW = 5;

function isRateLimited(address) {
    const now = Date.now();
    const userRequests = requests.get(address) || [];
    
    // Remove old requests
    const recentRequests = userRequests.filter(time => now - time < RATE_LIMIT_WINDOW);
    requests.set(address, recentRequests);
    
    return recentRequests.length >= MAX_REQUESTS_PER_WINDOW;
}

function addRequest(address) {
    const now = Date.now();
    const userRequests = requests.get(address) || [];
    userRequests.push(now);
    requests.set(address, userRequests);
}

// Health check endpoint
app.get('/health', (req, res) => {
    res.json({ 
        status: 'healthy', 
        faucetAddress: faucetAccount.address,
        networkId: 1337
    });
});

// Faucet endpoint
app.post('/faucet', async (req, res) => {
    try {
        const { address } = req.body;
        
        if (!address) {
            return res.status(400).json({ error: 'Address is required' });
        }
        
        if (!web3.utils.isAddress(address)) {
            return res.status(400).json({ error: 'Invalid Ethereum address' });
        }
        
        if (isRateLimited(address)) {
            return res.status(429).json({ 
                error: 'Rate limit exceeded. Maximum 5 requests per minute per address.' 
            });
        }
        
        // Check faucet balance
        const faucetBalance = await web3.eth.getBalance(faucetAccount.address);
        if (web3.utils.toBN(faucetBalance).lt(web3.utils.toBN(FAUCET_AMOUNT))) {
            return res.status(503).json({ 
                error: 'Faucet is out of funds',
                faucetBalance: web3.utils.fromWei(faucetBalance, 'ether')
            });
        }
        
        // Send transaction
        const tx = {
            from: faucetAccount.address,
            to: address,
            value: FAUCET_AMOUNT,
            gas: 21000,
            gasPrice: '0'
        };
        
        console.log(`💰 Sending ${web3.utils.fromWei(FAUCET_AMOUNT, 'ether')} ETH to ${address}`);
        
        const receipt = await web3.eth.sendTransaction(tx);
        
        addRequest(address);
        
        res.json({
            success: true,
            transactionHash: receipt.transactionHash,
            amount: web3.utils.fromWei(FAUCET_AMOUNT, 'ether'),
            recipient: address,
            faucetAddress: faucetAccount.address
        });
        
    } catch (error) {
        console.error('Faucet error:', error);
        res.status(500).json({ 
            error: 'Failed to send transaction',
            details: error.message 
        });
    }
});

// Get faucet info
app.get('/faucet/info', async (req, res) => {
    try {
        const balance = await web3.eth.getBalance(faucetAccount.address);
        res.json({
            faucetAddress: faucetAccount.address,
            balance: web3.utils.fromWei(balance, 'ether'),
            amountPerRequest: web3.utils.fromWei(FAUCET_AMOUNT, 'ether'),
            networkId: 1337,
            rpcUrl: RPC_URL
        });
    } catch (error) {
        res.status(500).json({ error: 'Failed to get faucet info' });
    }
});

app.listen(PORT, '0.0.0.0', () => {
    console.log(`🚰 Faucet server running on port ${PORT}`);
    console.log(`📡 RPC URL: ${RPC_URL}`);
    console.log(`💰 Faucet Address: ${faucetAccount.address}`);
});
EOF

# Install required Node.js packages
npm install express web3 cors

# Start the faucet server
exec node /blockchain/scripts/faucet-server.js
