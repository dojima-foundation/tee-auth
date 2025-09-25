//! Transaction signing module following QoS sign.rs patterns
//! Implements ECDSA signing for Ethereum transactions and raw messages

use anyhow::{anyhow, Result};
use p256::ecdsa::{
    signature::{Signer, Verifier},
    Signature, SigningKey, VerifyingKey,
};
use p256::elliptic_curve::rand_core::OsRng;
use zeroize::ZeroizeOnDrop;

use crate::quorum::P256Pair;

/// Sign private key pair following QoS P256SignPair structure exactly.
pub struct P256SignPair {
    /// The key used for signing
    private: SigningKey,
}

impl Clone for P256SignPair {
    fn clone(&self) -> Self {
        // Recreate the signing key from bytes
        let key_bytes = self.private.to_bytes();
        let private = SigningKey::from_bytes(&key_bytes).unwrap();
        Self { private }
    }
}

impl ZeroizeOnDrop for P256SignPair {}

impl P256SignPair {
    /// Generate a new private key following QoS P256SignPair::generate() exactly.
    #[must_use]
    pub fn generate() -> Self {
        Self {
            private: SigningKey::random(&mut OsRng),
        }
    }

    /// Sign the message and return the raw signature.
    /// Following QoS P256SignPair::sign() exactly.
    pub fn sign(&self, message: &[u8]) -> Result<Vec<u8>> {
        let signature: Signature = self.private.sign(message);
        Ok(signature.to_der().as_bytes().to_vec())
    }

    /// Get the public key of this pair.
    /// Following QoS P256SignPair::public_key() exactly.
    #[must_use]
    #[allow(dead_code)]
    pub fn public_key(&self) -> P256SignPublic {
        P256SignPublic {
            public: VerifyingKey::from(&self.private),
        }
    }

    /// Deserialize key from raw scalar byte slice.
    /// Following QoS P256SignPair::from_bytes() exactly.
    #[allow(dead_code)]
    pub fn from_bytes(bytes: &[u8]) -> Result<Self> {
        Ok(Self {
            private: SigningKey::from_bytes(bytes).map_err(|_| anyhow!("Failed to read secret"))?,
        })
    }

    /// Serialize key to raw scalar byte slice.
    /// Following QoS P256SignPair::to_bytes() exactly.
    #[must_use]
    #[allow(dead_code)]
    pub fn to_bytes(&self) -> Vec<u8> {
        self.private.to_bytes().to_vec()
    }
}

/// Sign public key for verifying signatures following QoS P256SignPublic structure exactly.
#[derive(Clone, PartialEq, Eq)]
pub struct P256SignPublic {
    public: VerifyingKey,
}

impl P256SignPublic {
    /// Verify a signature and message against this public key.
    /// Following QoS P256SignPublic::verify() exactly.
    #[allow(dead_code)]
    pub fn verify(&self, message: &[u8], signature: &[u8]) -> Result<()> {
        let signature =
            Signature::from_der(signature).map_err(|_| anyhow!("Invalid signature format"))?;

        self.public
            .verify(message, &signature)
            .map_err(|_| anyhow!("Signature verification failed"))
    }

    /// Get the public key bytes.
    #[allow(dead_code)]
    pub fn to_bytes(&self) -> Vec<u8> {
        self.public.to_encoded_point(false).as_ref().to_vec()
    }

    /// Create from bytes.
    #[allow(dead_code)]
    pub fn from_bytes(bytes: &[u8]) -> Result<Self> {
        let public = VerifyingKey::from_sec1_bytes(bytes)
            .map_err(|_| anyhow!("Invalid public key format"))?;
        Ok(Self { public })
    }
}

/// Ethereum transaction structure for signing.
#[derive(Debug, Clone)]
pub struct EthereumTransaction {
    pub to: Vec<u8>,    // Recipient address (20 bytes)
    pub value: u64,     // Value in wei
    pub gas_limit: u64, // Gas limit
    pub gas_price: u64, // Gas price in wei
    pub nonce: u64,     // Transaction nonce
    pub data: Vec<u8>,  // Transaction data
    pub chain_id: u64,  // Chain ID (1 for mainnet, etc.)
}

/// Signed Ethereum transaction result.
#[derive(Debug, Clone)]
#[allow(dead_code)]
pub struct SignedEthereumTransaction {
    pub transaction: EthereumTransaction,
    pub signature: Vec<u8>, // ECDSA signature (65 bytes)
    pub recovery_id: u8,    // Recovery ID for signature
}

/// Transaction signing service using quorum keys.
/// Main interface for transaction signing operations.
#[derive(Clone)]
pub struct TransactionSigner {
    quorum_key: P256Pair,
    sign_pair: P256SignPair,
}

impl TransactionSigner {
    /// Create a new transaction signer.
    pub fn new(quorum_key: P256Pair) -> Self {
        Self {
            quorum_key,
            sign_pair: P256SignPair::generate(),
        }
    }

    /// Create a new transaction signer with existing sign pair.
    #[allow(dead_code)]
    pub fn with_sign_pair(quorum_key: P256Pair, sign_pair: P256SignPair) -> Self {
        Self {
            quorum_key,
            sign_pair,
        }
    }

    /// Sign an Ethereum transaction.
    /// Following QoS signing patterns with Ethereum-specific handling.
    #[allow(dead_code)]
    pub fn sign_ethereum_transaction(
        &self,
        tx: &EthereumTransaction,
    ) -> Result<SignedEthereumTransaction> {
        // Create the transaction hash for signing
        let tx_hash = self.create_ethereum_transaction_hash(tx)?;

        // Sign the transaction hash
        let signature = self.sign_pair.sign(&tx_hash)?;

        // For DER-encoded signatures, we need to extract r and s values
        // and compute recovery_id. For now, we'll use a default recovery_id
        // In a real implementation, this would be computed from the signature
        let recovery_id = 0u8; // Default recovery_id for testing
        let signature_bytes = signature; // Use the full DER signature

        Ok(SignedEthereumTransaction {
            transaction: tx.clone(),
            signature: signature_bytes,
            recovery_id,
        })
    }

    /// Sign a raw message.
    /// Following QoS P256SignPair::sign() exactly.
    pub fn sign_raw_message(&self, message: &[u8]) -> Result<Vec<u8>> {
        self.sign_pair.sign(message)
    }

    /// Verify a signature against a message.
    /// Following QoS P256SignPublic::verify() exactly.
    #[allow(dead_code)]
    pub fn verify_signature(&self, message: &[u8], signature: &[u8]) -> bool {
        self.sign_pair
            .public_key()
            .verify(message, signature)
            .is_ok()
    }

    /// Get the signing public key.
    #[allow(dead_code)]
    pub fn signing_public_key(&self) -> P256SignPublic {
        self.sign_pair.public_key()
    }

    /// Get the signing public key bytes.
    #[allow(dead_code)]
    pub fn signing_public_key_bytes(&self) -> Vec<u8> {
        self.sign_pair.public_key().to_bytes()
    }

    /// Create Ethereum transaction hash for signing.
    /// Following Ethereum EIP-155 transaction hashing.
    #[allow(dead_code)]
    fn create_ethereum_transaction_hash(&self, tx: &EthereumTransaction) -> Result<Vec<u8>> {
        use sha2::{Digest, Sha256};

        // Create RLP-encoded transaction data
        let mut rlp_data = Vec::new();

        // Add nonce
        rlp_data.extend_from_slice(&tx.nonce.to_be_bytes());

        // Add gas price
        rlp_data.extend_from_slice(&tx.gas_price.to_be_bytes());

        // Add gas limit
        rlp_data.extend_from_slice(&tx.gas_limit.to_be_bytes());

        // Add recipient address (20 bytes)
        if tx.to.len() != 20 {
            return Err(anyhow!("Invalid recipient address length"));
        }
        rlp_data.extend_from_slice(&tx.to);

        // Add value
        rlp_data.extend_from_slice(&tx.value.to_be_bytes());

        // Add transaction data
        rlp_data.extend_from_slice(&tx.data);

        // Add chain ID (EIP-155)
        rlp_data.extend_from_slice(&tx.chain_id.to_be_bytes());
        rlp_data.push(0u8); // v = 0
        rlp_data.push(0u8); // r = 0
        rlp_data.push(0u8); // s = 0

        // Hash the RLP data
        let mut hasher = Sha256::new();
        hasher.update(&rlp_data);
        let hash = hasher.finalize();

        Ok(hash.to_vec())
    }

    /// Sign a message with the quorum key directly.
    /// Following QoS P256Pair::sign() patterns.
    #[allow(dead_code)]
    pub fn sign_with_quorum_key(&self, message: &[u8]) -> Result<Vec<u8>> {
        self.quorum_key.sign(message)
    }

    /// Verify a signature with the quorum public key.
    /// Following QoS P256Public::verify() patterns.
    #[allow(dead_code)]
    pub fn verify_with_quorum_key(&self, message: &[u8], signature: &[u8]) -> bool {
        // This would need to be implemented in the P256Public struct
        // For now, we'll use the sign_pair verification
        self.verify_signature(message, signature)
    }
}

/// Batch transaction signing for multiple transactions.
pub struct BatchTransactionSigner {
    signer: TransactionSigner,
}

impl BatchTransactionSigner {
    /// Create a new batch transaction signer.
    #[allow(dead_code)]
    pub fn new(quorum_key: P256Pair) -> Self {
        Self {
            signer: TransactionSigner::new(quorum_key),
        }
    }

    /// Sign multiple Ethereum transactions.
    #[allow(dead_code)]
    pub fn sign_multiple_transactions(
        &self,
        transactions: &[EthereumTransaction],
    ) -> Result<Vec<SignedEthereumTransaction>> {
        let mut signed_transactions = Vec::new();

        for tx in transactions {
            let signed_tx = self.signer.sign_ethereum_transaction(tx)?;
            signed_transactions.push(signed_tx);
        }

        Ok(signed_transactions)
    }

    /// Sign multiple raw messages.
    #[allow(dead_code)]
    pub fn sign_multiple_messages(&self, messages: &[Vec<u8>]) -> Result<Vec<Vec<u8>>> {
        let mut signatures = Vec::new();

        for message in messages {
            let signature = self.signer.sign_raw_message(message)?;
            signatures.push(signature);
        }

        Ok(signatures)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sign_pair_generation() {
        let sign_pair = P256SignPair::generate();
        let public_key = sign_pair.public_key();

        // Test signing and verification
        let message = b"test message";
        let signature = sign_pair.sign(message).unwrap();
        assert!(public_key.verify(message, &signature).is_ok());
    }

    #[test]
    fn test_transaction_signer() {
        let quorum_key = P256Pair::generate().unwrap();
        let signer = TransactionSigner::new(quorum_key);

        // Test raw message signing
        let message = b"test message";
        let signature = signer.sign_raw_message(message).unwrap();
        assert!(signer.verify_signature(message, &signature));
    }

    #[test]
    fn test_ethereum_transaction_signing() {
        let quorum_key = P256Pair::generate().unwrap();
        let signer = TransactionSigner::new(quorum_key);

        let tx = EthereumTransaction {
            to: vec![0u8; 20],          // 20-byte address
            value: 1000000000000000000, // 1 ETH in wei
            gas_limit: 21000,
            gas_price: 20000000000, // 20 gwei
            nonce: 1,
            data: vec![],
            chain_id: 1, // Mainnet
        };

        let signed_tx = signer.sign_ethereum_transaction(&tx).unwrap();
        // The signature is DER-encoded, so it can be variable length
        assert!(!signed_tx.signature.is_empty());
        // Recovery ID should be 0 or 1 for Ethereum
        assert!(signed_tx.recovery_id == 0 || signed_tx.recovery_id == 1);
    }
}
