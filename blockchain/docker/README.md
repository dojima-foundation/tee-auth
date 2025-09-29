# Geth Development Environment with Faucet

A complete Docker-based Ethereum development environment using Geth in developer mode with an integrated faucet service for easy token distribution.

## Features

- 🚀 **Geth in Developer Mode**: Local Ethereum testnet with instant block generation
- 🚰 **Integrated Faucet**: Automated token distribution service
- 🔧 **Pre-configured Accounts**: Pre-funded developer accounts
- 📡 **Full RPC Access**: HTTP and WebSocket JSON-RPC endpoints
- 🐳 **Docker Ready**: Easy deployment with Docker Compose
- 🧪 **Testing Tools**: Built-in scripts for testing and account creation

## Quick Start

### 1. Start the Development Environment

```bash
# Clone and navigate to the blockchain directory
cd blockchain/docker

# Start Geth and Faucet services
docker-compose up -d

# Check service status
docker-compose ps
```

### 2. Verify Services

```bash
# Check Geth is running
curl http://localhost:8545 -X POST \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Check Faucet is running
curl http://localhost:3000/health
```

### 3. Get Test Tokens

```bash
# Request tokens for a specific address
curl -X POST http://localhost:3000/faucet \
  -H "Content-Type: application/json" \
  -d '{"address": "0x742d35Cc6634C0532925a3b8D2C9C5C5E5C5E5C5"}'
```

## Configuration

### Pre-configured Accounts

The genesis block includes several pre-funded accounts:

| Address | Private Key | Balance |
|---------|-------------|---------|
| `0x7Aa16266Ba3d309e3cb278B452b1A6307E52Fb62` | Auto-generated | 1000 ETH |
| `0x4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d` | `0x4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d` | 1000 ETH |
| `0x90F8bf6A479f320ead074411a4B0e7944Ea8c9C1` | Auto-generated | 1000 ETH |
| `0xFFcf8FDEE72ac11b5c542428B35EEF5769C409f0` | Auto-generated | 1000 ETH |
| `0x22d491Bde2303f2f43325b2108D26f1eAbA1e32b` | Auto-generated | 1000 ETH |

### Network Configuration

- **Chain ID**: 1337
- **Network Name**: Local Development
- **Gas Price**: 0 (free transactions)
- **Block Time**: 3 seconds (configurable)
- **Genesis**: Custom configuration with pre-funded accounts

## Services

### Geth Node

- **HTTP RPC**: `http://localhost:8545`
- **WebSocket RPC**: `ws://localhost:8546`
- **P2P Port**: 30303
- **Data Directory**: `/blockchain/data` (persistent volume)

**Available RPC Methods**:
- `eth_*` - Ethereum protocol methods
- `web3_*` - Web3 utility methods
- `net_*` - Network information
- `personal_*` - Account management
- `txpool_*` - Transaction pool methods

### Faucet Service

- **HTTP API**: `http://localhost:3000`
- **Rate Limit**: 5 requests per minute per address
- **Amount**: 1000 ETH per request
- **Health Check**: `GET /health`
- **Faucet Request**: `POST /faucet`

**Faucet Endpoints**:

```bash
# Get faucet information
curl http://localhost:3000/faucet/info

# Request tokens
curl -X POST http://localhost:3000/faucet \
  -H "Content-Type: application/json" \
  -d '{"address": "0xYourAddress"}'
```

## Usage Examples

### 1. Connect with MetaMask

1. Open MetaMask
2. Add Custom Network:
   - **Network Name**: Local Dev
   - **RPC URL**: `http://localhost:8545`
   - **Chain ID**: 1337
   - **Currency Symbol**: ETH
3. Import one of the pre-configured accounts using their private keys

### 2. Connect with Remix IDE

1. Go to [Remix IDE](https://remix.ethereum.org)
2. Navigate to "Deploy & Run Transactions"
3. Select "Web3 Provider"
4. Enter: `http://localhost:8545`
5. Connect to your imported account

### 3. Connect with Hardhat

```javascript
// hardhat.config.js
module.exports = {
  networks: {
    local: {
      url: "http://localhost:8545",
      chainId: 1337,
      accounts: [
        "0x4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
      ]
    }
  }
};
```

### 4. Connect with Web3.js

```javascript
const Web3 = require('web3');
const web3 = new Web3('http://localhost:8545');

// Get network info
const networkId = await web3.eth.net.getId();
console.log('Network ID:', networkId);

// Get balance
const balance = await web3.eth.getBalance('0x742d35Cc6634C0532925a3b8D2C9C5C5E5C5E5C5');
console.log('Balance:', web3.utils.fromWei(balance, 'ether'), 'ETH');
```

## Scripts

### Available Scripts

```bash
# Create a new account
docker exec geth-dev-node /blockchain/scripts/create-account.sh

# Test faucet functionality
docker exec geth-dev-node /blockchain/scripts/test-faucet.sh

# Start development environment
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f geth-dev
docker-compose logs -f faucet

# Reset blockchain data
docker-compose down -v
docker-compose up -d
```

## Development Workflow

### 1. Smart Contract Development

```bash
# Deploy contract using Remix
# 1. Open Remix IDE
# 2. Connect to local network
# 3. Deploy your contract
# 4. Interact with contract functions
```

### 2. Testing

```bash
# Run faucet tests
docker exec geth-dev-node /blockchain/scripts/test-faucet.sh

# Check transaction history
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", true],"id":1}'
```

### 3. Account Management

```bash
# Create new account
docker exec geth-dev-node /blockchain/scripts/create-account.sh

# Get faucet tokens for new account
curl -X POST http://localhost:3000/faucet \
  -H "Content-Type: application/json" \
  -d '{"address": "NEW_ADDRESS_HERE"}'
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   ```bash
   # Check if ports are in use
   lsof -i :8545
   lsof -i :3000
   
   # Stop conflicting services or change ports in docker-compose.yml
   ```

2. **Permission Issues**
   ```bash
   # Fix script permissions
   chmod +x scripts/*.sh
   ```

3. **Container Won't Start**
   ```bash
   # Check container logs
   docker-compose logs geth-dev
   
   # Rebuild containers
   docker-compose down
   docker-compose build --no-cache
   docker-compose up -d
   ```

4. **Faucet Out of Funds**
   ```bash
   # Check faucet balance
   curl http://localhost:3000/faucet/info
   
   # Restart services to reset balances
   docker-compose restart
   ```

### Logs and Debugging

```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f geth-dev
docker-compose logs -f faucet

# Access container shell
docker exec -it geth-dev-node bash
docker exec -it geth-faucet bash
```

## Security Notes

⚠️ **This is for development only!**

- All accounts are pre-funded with test tokens
- Private keys are exposed in configuration
- No real value or security measures
- Do not use in production environments

## API Reference

### Geth RPC Methods

Standard Ethereum JSON-RPC methods are available. See [Geth documentation](https://geth.ethereum.org/docs/rpc) for complete reference.

### Faucet API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Check faucet health |
| GET | `/faucet/info` | Get faucet information |
| POST | `/faucet` | Request tokens |

**Request Format**:
```json
{
  "address": "0x742d35Cc6634C0532925a3b8D2C9C5C5E5C5E5C5"
}
```

**Response Format**:
```json
{
  "success": true,
  "transactionHash": "0x...",
  "amount": "1000",
  "recipient": "0x742d35Cc6634C0532925a3b8D2C9C5C5E5C5E5C5",
  "faucetAddress": "0x4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
