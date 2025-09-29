FROM ethereum/client-go:latest

# Install additional tools for development
RUN apk add --no-cache \
    curl \
    jq \
    bash \
    nodejs \
    npm

# Find and create symlink for geth
RUN find / -name "geth" -type f 2>/dev/null | head -1 | xargs -I {} ln -sf {} /usr/local/bin/geth

# Install web3 tools for faucet functionality
RUN npm install -g web3@4.3.0

# Create directories for blockchain data and scripts
RUN mkdir -p /blockchain/data /blockchain/scripts /blockchain/config

# Copy configuration and scripts
COPY config/ /blockchain/config/
COPY scripts/ /blockchain/scripts/

# Make scripts executable
RUN chmod +x /blockchain/scripts/*.sh

# Expose ports
EXPOSE 8545 8546 30303

# Set working directory
WORKDIR /blockchain

# Default command
CMD ["/blockchain/scripts/start-dev.sh"]
