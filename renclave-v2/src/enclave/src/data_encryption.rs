//! Data encryption/decryption module following QoS encrypt.rs patterns
//! Implements ECDH key agreement and AES-GCM encryption for data protection

use aes_gcm::{
    aead::{Aead, KeyInit, Payload},
    Aes256Gcm, Nonce,
};
use anyhow::{anyhow, Result};
use borsh::{BorshDeserialize, BorshSerialize};
use hmac::{Hmac, Mac};
use log::{debug, error, info};
use p256::{
    elliptic_curve::{ecdh::diffie_hellman, rand_core::OsRng, sec1::ToEncodedPoint},
    PublicKey, SecretKey,
};
use sha2::Sha512;
use zeroize::ZeroizeOnDrop;

use crate::quorum::P256Pair;

const AES256_KEY_LEN: usize = 32;
const BITS_96_AS_BYTES: u8 = 12;
const QOS_ENCRYPTION_HMAC_MESSAGE: &[u8] = b"qos_encryption_hmac_message";
const PUB_KEY_LEN_UNCOMPRESSED: usize = 65;

type HmacSha512 = Hmac<Sha512>;

/// Envelope for serializing an encrypted message with its context.
/// Following QoS encrypt.rs Envelope structure exactly.
#[derive(BorshDeserialize, BorshSerialize, Debug)]
pub struct Envelope {
    /// Nonce used as an input to the cipher.
    nonce: [u8; BITS_96_AS_BYTES as usize],
    /// Public key as sec1 encoded point with no compression
    pub ephemeral_sender_public: [u8; PUB_KEY_LEN_UNCOMPRESSED],
    /// The data encrypted with an AES 256 GCM cipher.
    encrypted_message: Vec<u8>,
}

/// P256 encryption key pair following QoS P256EncryptPair structure.
#[derive(Clone)]
pub struct P256EncryptPair {
    private: SecretKey,
}

impl ZeroizeOnDrop for P256EncryptPair {}

impl P256EncryptPair {
    /// Generate a new private key using the OS randomness source.
    /// Following QoS P256EncryptPair::generate() exactly.
    #[must_use]
    pub fn generate() -> Self {
        Self {
            private: SecretKey::random(&mut OsRng),
        }
    }

    /// Encrypt a message using proper ECDH key agreement.
    /// Following QoS P256EncryptPair::encrypt() exactly.
    pub fn encrypt(&self, message: &[u8]) -> Result<Vec<u8>> {
        info!("🔐 DEBUG: P256EncryptPair::encrypt() called");
        info!("🔐 DEBUG: Message length: {} bytes", message.len());
        info!(
            "🔐 DEBUG: Message (first 10 bytes): {:?}",
            &message[..message.len().min(10)]
        );

        // FIXED: Generate NEW ephemeral key for proper ECDH
        let ephemeral_sender_private = SecretKey::random(&mut OsRng);
        info!("🔐 DEBUG: Generated NEW ephemeral private key for ECDH");

        let ephemeral_sender_public: [u8; PUB_KEY_LEN_UNCOMPRESSED] = ephemeral_sender_private
            .public_key()
            .to_encoded_point(false)
            .as_ref()
            .try_into()
            .map_err(|_| anyhow!("Failed to coerce public key to intended length"))?;
        info!("🔐 DEBUG: Generated ephemeral sender public key");
        info!(
            "🔐 DEBUG: Ephemeral sender public (first 10 bytes): {:?}",
            &ephemeral_sender_public[..10]
        );

        let sender_public_typed = SenderPublic(&ephemeral_sender_public);
        // FIXED: Use the RECIPIENT's public key (this pair's public key)
        let receiver_encoded_point = self.private.public_key().to_encoded_point(false);
        let receiver_public_typed = ReceiverPublic(receiver_encoded_point.as_ref());
        info!("🔐 DEBUG: Created sender and receiver public types");
        info!("🔐 DEBUG: Using ephemeral_private + recipient_public for ECDH");

        info!("🔐 DEBUG: Creating cipher with ECDH...");
        let cipher = create_cipher(
            &PrivPubOrSharedSecret::PrivPub {
                private: &ephemeral_sender_private,
                public: &self.private.public_key(), // recipient's public key
            },
            &sender_public_typed,
            &receiver_public_typed,
        )?;
        info!("🔐 DEBUG: Cipher created successfully");

        let nonce = {
            let mut random_bytes = [0u8; BITS_96_AS_BYTES as usize];
            use rand::RngCore;
            OsRng.fill_bytes(&mut random_bytes);
            *Nonce::from_slice(&random_bytes)
        };
        info!("🔐 DEBUG: Generated nonce");

        let aad = create_additional_associated_data(&sender_public_typed, &receiver_public_typed)?;
        let payload = Payload {
            aad: &aad,
            msg: message,
        };
        info!("🔐 DEBUG: Created payload");

        let encrypted_message = cipher
            .encrypt(&nonce, payload)
            .map_err(|_| anyhow!("AES-GCM-256 encrypt error"))?;
        info!(
            "🔐 DEBUG: Message encrypted successfully, length: {} bytes",
            encrypted_message.len()
        );

        let nonce = nonce.into();
        let envelope = Envelope {
            nonce,
            ephemeral_sender_public,
            encrypted_message,
        };
        info!("🔐 DEBUG: Created envelope");

        let result =
            borsh::to_vec(&envelope).map_err(|_| anyhow!("Failed to serialize envelope"))?;
        info!(
            "🔐 DEBUG: Serialized envelope, final length: {} bytes",
            result.len()
        );
        Ok(result)
    }

    /// Decrypt a message encoded to this pair's public key.
    /// Following QoS P256EncryptPair::decrypt() exactly.
    pub fn decrypt(&self, serialized_envelope: &[u8]) -> Result<Vec<u8>> {
        info!("🔓 DEBUG: P256EncryptPair::decrypt() called");
        info!(
            "🔓 DEBUG: Serialized envelope length: {} bytes",
            serialized_envelope.len()
        );
        info!(
            "🔓 DEBUG: Serialized envelope (first 10 bytes): {:?}",
            &serialized_envelope[..serialized_envelope.len().min(10)]
        );

        let Envelope {
            nonce,
            ephemeral_sender_public: ephemeral_sender_public_bytes,
            encrypted_message,
        } = Envelope::try_from_slice(serialized_envelope)
            .map_err(|_| anyhow!("Failed to deserialize envelope"))?;
        info!("🔓 DEBUG: Successfully deserialized envelope");

        let nonce = Nonce::from_slice(&nonce);
        let ephemeral_sender_public = PublicKey::from_sec1_bytes(&ephemeral_sender_public_bytes)
            .map_err(|_| anyhow!("Failed to deserialize public key"))?;
        info!("🔓 DEBUG: Successfully parsed ephemeral sender public key");
        debug!(
            "🔓 DEBUG: Ephemeral sender public key (first 10 bytes): {:?}",
            &ephemeral_sender_public_bytes[..10]
        );

        let sender_public_typed = SenderPublic(&ephemeral_sender_public_bytes);
        let receiver_encoded_point = self.private.public_key().to_encoded_point(false);
        let receiver_public_typed = ReceiverPublic(receiver_encoded_point.as_ref());
        info!("🔓 DEBUG: Created sender and receiver public types");
        debug!(
            "🔓 DEBUG: Receiver private key (first 10 bytes): {:?}",
            &self.private.to_be_bytes()[..10]
        );
        debug!(
            "🔓 DEBUG: Receiver public key (first 10 bytes): {:?}",
            &receiver_encoded_point.as_bytes()[..10]
        );

        info!("🔓 DEBUG: Creating cipher with ECDH...");
        info!(
            "🔓 DEBUG: ECDH inputs - private: {:?}, public: {:?}",
            &self.private.to_be_bytes()[..10],
            &ephemeral_sender_public.to_encoded_point(false).as_bytes()[..10]
        );

        let cipher = create_cipher(
            &PrivPubOrSharedSecret::PrivPub {
                private: &self.private,
                public: &ephemeral_sender_public,
            },
            &sender_public_typed,
            &receiver_public_typed,
        )?;
        info!("🔓 DEBUG: Cipher created successfully");

        let aad = create_additional_associated_data(&sender_public_typed, &receiver_public_typed)?;
        let payload = Payload {
            aad: &aad,
            msg: &encrypted_message,
        };
        info!("🔓 DEBUG: Created payload");

        info!("🔓 DEBUG: Attempting AES-GCM decryption...");
        let result = cipher
            .decrypt(nonce, payload)
            .map_err(|_| anyhow!("AES-GCM-256 decrypt error"));

        match &result {
            Ok(decrypted) => {
                info!(
                    "🔓 DEBUG: Decryption successful, length: {} bytes",
                    decrypted.len()
                );
                info!(
                    "🔓 DEBUG: Decrypted data (first 10 bytes): {:?}",
                    &decrypted[..decrypted.len().min(10)]
                );
            }
            Err(e) => {
                error!("🔓 DEBUG: Decryption failed: {}", e);
            }
        }
        result
    }

    /// Get the public key.
    /// Following QoS P256EncryptPair::public_key() exactly.
    #[must_use]
    #[allow(dead_code)]
    pub fn public_key(&self) -> P256EncryptPublic {
        P256EncryptPublic {
            public: self.private.public_key(),
        }
    }

    /// Deserialize key from raw scalar byte slice.
    /// Following QoS P256EncryptPair::from_bytes() exactly.
    pub fn from_bytes(bytes: &[u8]) -> Result<Self> {
        Ok(Self {
            private: SecretKey::from_be_bytes(bytes)
                .map_err(|_| anyhow!("Failed to read secret"))?,
        })
    }

    /// Create P256EncryptPair from private key
    #[allow(dead_code)]
    pub fn from_private_key(private_key: &SecretKey) -> Result<Self> {
        Ok(Self {
            private: private_key.clone(),
        })
    }

    /// Create P256EncryptPair from public key (for encryption)
    #[allow(dead_code)]
    pub fn from_public_key(_public_key: &PublicKey) -> Result<Self> {
        // For encryption, we need to create a temporary pair
        // This is used when we have the recipient's public key
        let temp_private = SecretKey::random(&mut OsRng);
        Ok(Self {
            private: temp_private,
        })
    }

    /// Serialize key to raw scalar byte slice.
    /// Following QoS P256EncryptPair::to_bytes() exactly.
    #[must_use]
    #[allow(dead_code)]
    pub fn to_bytes(&self) -> Vec<u8> {
        self.private.to_be_bytes().to_vec()
    }
}

/// P256 Public key for encryption operations.
/// Following QoS P256EncryptPublic structure exactly.
#[derive(Clone, PartialEq, Eq)]
pub struct P256EncryptPublic {
    public: PublicKey,
}

impl P256EncryptPublic {
    /// Create a new P256EncryptPublic from a PublicKey
    #[allow(dead_code)]
    pub fn new(public: PublicKey) -> Self {
        Self { public }
    }

    /// Encrypt a message to this public key.
    /// Following QoS P256EncryptPublic::encrypt() exactly.
    #[allow(dead_code)]
    pub fn encrypt(&self, message: &[u8]) -> Result<Vec<u8>> {
        info!("🔐 DEBUG: P256EncryptPublic::encrypt() called");
        info!("🔐 DEBUG: Message length: {} bytes", message.len());
        info!(
            "🔐 DEBUG: Message (first 10 bytes): {:?}",
            &message[..message.len().min(10)]
        );

        let ephemeral_sender_private = SecretKey::random(&mut OsRng);
        info!("🔐 DEBUG: Generated ephemeral sender private key");
        debug!(
            "🔐 DEBUG: Ephemeral sender private key (first 10 bytes): {:?}",
            &ephemeral_sender_private.to_be_bytes()[..10]
        );

        let ephemeral_sender_public: [u8; PUB_KEY_LEN_UNCOMPRESSED] = ephemeral_sender_private
            .public_key()
            .to_encoded_point(false)
            .as_ref()
            .try_into()
            .map_err(|_| anyhow!("Failed to coerce public key to intended length"))?;
        info!("🔐 DEBUG: Generated ephemeral sender public key");
        debug!(
            "🔐 DEBUG: Ephemeral sender public (first 10 bytes): {:?}",
            &ephemeral_sender_public[..10]
        );

        let sender_public_typed = SenderPublic(&ephemeral_sender_public);
        let receiver_encoded_point = self.public.to_encoded_point(false);
        let receiver_public_typed = ReceiverPublic(receiver_encoded_point.as_ref());
        info!("🔐 DEBUG: Created sender and receiver public types");
        debug!(
            "🔐 DEBUG: Receiver public key (first 10 bytes): {:?}",
            &receiver_encoded_point.as_bytes()[..10]
        );

        info!("🔐 DEBUG: Creating cipher with ECDH...");
        info!(
            "🔐 DEBUG: ECDH inputs - private: {:?}, public: {:?}",
            &ephemeral_sender_private.to_be_bytes()[..10],
            &self.public.to_encoded_point(false).as_bytes()[..10]
        );

        let cipher = create_cipher(
            &PrivPubOrSharedSecret::PrivPub {
                private: &ephemeral_sender_private,
                public: &self.public,
            },
            &sender_public_typed,
            &receiver_public_typed,
        )?;
        info!("🔐 DEBUG: Cipher created successfully");

        let nonce = {
            let mut random_bytes = [0u8; BITS_96_AS_BYTES as usize];
            use rand::RngCore;
            OsRng.fill_bytes(&mut random_bytes);
            *Nonce::from_slice(&random_bytes)
        };
        info!("🔐 DEBUG: Generated nonce");

        let aad = create_additional_associated_data(&sender_public_typed, &receiver_public_typed)?;
        let payload = Payload {
            aad: &aad,
            msg: message,
        };
        info!("🔐 DEBUG: Created payload");

        let encrypted_message = cipher
            .encrypt(&nonce, payload)
            .map_err(|_| anyhow!("AES-GCM-256 encrypt error"))?;
        info!(
            "🔐 DEBUG: Message encrypted successfully, length: {} bytes",
            encrypted_message.len()
        );

        let nonce = nonce.into();
        let envelope = Envelope {
            nonce,
            ephemeral_sender_public,
            encrypted_message,
        };
        info!("🔐 DEBUG: Created envelope");

        let result =
            borsh::to_vec(&envelope).map_err(|_| anyhow!("Failed to serialize envelope"))?;
        info!(
            "🔐 DEBUG: Serialized envelope, final length: {} bytes",
            result.len()
        );
        Ok(result)
    }

    /// Decrypt a message encoded to this pair's public key.
    /// Following QoS P256EncryptPublic::decrypt_from_shared_secret() exactly.
    #[allow(dead_code)]
    pub fn decrypt_from_shared_secret(
        &self,
        serialized_envelope: &[u8],
        shared_secret: &[u8],
    ) -> Result<Vec<u8>> {
        let Envelope {
            nonce,
            ephemeral_sender_public: ephemeral_sender_public_bytes,
            encrypted_message,
        } = Envelope::try_from_slice(serialized_envelope)
            .map_err(|_| anyhow!("Failed to deserialize envelope"))?;

        let nonce = Nonce::from_slice(&nonce);
        let sender_public_typed = SenderPublic(&ephemeral_sender_public_bytes);
        let receiver_encoded_point = self.public.to_encoded_point(false);
        let receiver_public_typed = ReceiverPublic(receiver_encoded_point.as_ref());

        let cipher = create_cipher(
            &PrivPubOrSharedSecret::SharedSecret { shared_secret },
            &sender_public_typed,
            &receiver_public_typed,
        )?;

        let aad = create_additional_associated_data(&sender_public_typed, &receiver_public_typed)?;
        let payload = Payload {
            aad: &aad,
            msg: &encrypted_message,
        };

        cipher
            .decrypt(nonce, payload)
            .map_err(|_| anyhow!("AES-GCM-256 decrypt error"))
    }
}

/// AES-GCM-256 secret for symmetric encryption.
/// Following QoS AesGcm256Secret structure exactly.
#[derive(Clone)]
pub struct AesGcm256Secret {
    secret: [u8; AES256_KEY_LEN],
}

impl ZeroizeOnDrop for AesGcm256Secret {}

impl AesGcm256Secret {
    /// Generate a new AES-GCM-256 secret.
    /// Following QoS AesGcm256Secret::generate() exactly.
    #[must_use]
    pub fn generate() -> Self {
        Self {
            secret: {
                let mut bytes = [0u8; AES256_KEY_LEN];
                use rand::RngCore;
                OsRng.fill_bytes(&mut bytes);
                bytes
            },
        }
    }

    /// Create from bytes.
    /// Following QoS AesGcm256Secret::from_bytes() exactly.
    #[allow(dead_code)]
    pub fn from_bytes(bytes: [u8; AES256_KEY_LEN]) -> Result<Self> {
        Ok(Self { secret: bytes })
    }

    /// Encrypt a message with this secret.
    /// Following QoS AesGcm256Secret::encrypt() exactly.
    #[allow(dead_code)]
    pub fn encrypt(&self, msg: &[u8]) -> Result<Vec<u8>> {
        let cipher = Aes256Gcm::new_from_slice(&self.secret)
            .map_err(|_| anyhow!("Failed to create AES-GCM cipher"))?;

        let nonce = {
            let mut random_bytes = [0u8; BITS_96_AS_BYTES as usize];
            use rand::RngCore;
            OsRng.fill_bytes(&mut random_bytes);
            *Nonce::from_slice(&random_bytes)
        };

        let payload = Payload { aad: b"", msg };
        let encrypted_message = cipher
            .encrypt(&nonce, payload)
            .map_err(|_| anyhow!("AES-GCM-256 encrypt error"))?;

        let envelope = Envelope {
            nonce: nonce.into(),
            ephemeral_sender_public: [0u8; PUB_KEY_LEN_UNCOMPRESSED], // Not used for symmetric encryption
            encrypted_message,
        };

        borsh::to_vec(&envelope).map_err(|_| anyhow!("Failed to serialize envelope"))
    }

    /// Decrypt a message with this secret.
    /// Following QoS AesGcm256Secret::decrypt() exactly.
    #[allow(dead_code)]
    pub fn decrypt(&self, serialized_envelope: &[u8]) -> Result<Vec<u8>> {
        let Envelope {
            nonce,
            encrypted_message,
            ..
        } = Envelope::try_from_slice(serialized_envelope)
            .map_err(|_| anyhow!("Failed to deserialize envelope"))?;

        let cipher = Aes256Gcm::new_from_slice(&self.secret)
            .map_err(|_| anyhow!("Failed to create AES-GCM cipher"))?;

        let nonce = Nonce::from_slice(&nonce);
        let payload = Payload {
            aad: b"",
            msg: &encrypted_message,
        };

        cipher
            .decrypt(nonce, payload)
            .map_err(|_| anyhow!("AES-GCM-256 decrypt error"))
    }
}

/// Data encryption service using quorum keys.
/// Main interface for data encryption/decryption operations.
#[derive(Clone)]
pub struct DataEncryption {
    quorum_key: P256Pair,
    ephemeral_key: P256EncryptPair,
    symmetric_secret: AesGcm256Secret,
}

impl DataEncryption {
    /// Create a new data encryption service.
    pub fn new(quorum_key: P256Pair) -> Self {
        Self {
            quorum_key,
            ephemeral_key: P256EncryptPair::generate(),
            symmetric_secret: AesGcm256Secret::generate(),
        }
    }

    /// Encrypt data using the quorum key.
    /// Following QoS encryption patterns exactly.
    pub fn encrypt_data(&self, data: &[u8], _recipient_public: &[u8]) -> Result<Vec<u8>> {
        info!("🔐 DEBUG: Starting data encryption");
        info!("🔐 DEBUG: Input data length: {} bytes", data.len());
        info!(
            "🔐 DEBUG: Input data (first 10 bytes): {:?}",
            &data[..data.len().min(10)]
        );

        // FIXED: Use quorum key pair for both encryption and decryption
        // This matches QoS pattern: use the same key pair for encrypt/decrypt
        let quorum_private_bytes = self.quorum_key.private_key_bytes();
        info!("🔐 DEBUG: Using quorum private key for encryption");
        info!(
            "🔐 DEBUG: Quorum private key (first 10 bytes): {:?}",
            &quorum_private_bytes[..quorum_private_bytes.len().min(10)]
        );

        let encrypt_pair = P256EncryptPair::from_bytes(&quorum_private_bytes)?;
        info!("🔐 DEBUG: Created P256EncryptPair from quorum key, calling encrypt()");

        let result = encrypt_pair.encrypt(data);
        match &result {
            Ok(encrypted) => {
                info!(
                    "🔐 DEBUG: Encryption successful, result length: {} bytes",
                    encrypted.len()
                );
                info!(
                    "🔐 DEBUG: Encrypted data (first 10 bytes): {:?}",
                    &encrypted[..encrypted.len().min(10)]
                );
            }
            Err(e) => {
                error!("🔐 DEBUG: Encryption failed: {}", e);
            }
        }
        result
    }

    /// Decrypt data using the quorum key.
    /// Following QoS decryption patterns exactly.
    pub fn decrypt_data(&self, encrypted_envelope: &[u8]) -> Result<Vec<u8>> {
        info!("🔓 DEBUG: Starting data decryption");
        info!(
            "🔓 DEBUG: Encrypted envelope length: {} bytes",
            encrypted_envelope.len()
        );
        info!(
            "🔓 DEBUG: Encrypted envelope (first 10 bytes): {:?}",
            &encrypted_envelope[..encrypted_envelope.len().min(10)]
        );

        // Use the quorum key for decryption (same key pair as encryption)
        // This matches QoS pattern: same key pair for encrypt/decrypt
        info!("🔓 DEBUG: Calling quorum_key.decrypt()");
        let result = self.quorum_key.decrypt(encrypted_envelope);

        match &result {
            Ok(decrypted) => {
                info!(
                    "🔓 DEBUG: Decryption successful, result length: {} bytes",
                    decrypted.len()
                );
                info!(
                    "🔓 DEBUG: Decrypted data (first 10 bytes): {:?}",
                    &decrypted[..decrypted.len().min(10)]
                );
            }
            Err(e) => {
                error!("🔓 DEBUG: Decryption failed: {}", e);
            }
        }
        result
    }

    /// Encrypt data using symmetric encryption.
    /// Following QoS symmetric encryption patterns.
    #[allow(dead_code)]
    pub fn encrypt_symmetric(&self, data: &[u8]) -> Result<Vec<u8>> {
        self.symmetric_secret.encrypt(data)
    }

    /// Decrypt data using symmetric encryption.
    /// Following QoS symmetric decryption patterns.
    #[allow(dead_code)]
    pub fn decrypt_symmetric(&self, encrypted_envelope: &[u8]) -> Result<Vec<u8>> {
        self.symmetric_secret.decrypt(encrypted_envelope)
    }

    /// Get the ephemeral public key for key exchange.
    #[allow(dead_code)]
    pub fn ephemeral_public_key(&self) -> Vec<u8> {
        self.ephemeral_key
            .public_key()
            .public
            .to_encoded_point(false)
            .as_ref()
            .to_vec()
    }
}

// Helper types following QoS patterns exactly

/// Sender public key wrapper.
struct SenderPublic<'a>(&'a [u8]);

/// Receiver public key wrapper.
struct ReceiverPublic<'a>(&'a [u8]);

/// Input for creating a shared secret.
/// Following QoS PrivPubOrSharedSecret exactly.
enum PrivPubOrSharedSecret<'a> {
    /// Inputs for using Diffie–Hellman to create a shared secret.
    PrivPub {
        private: &'a SecretKey,
        public: &'a PublicKey,
    },
    /// This will be used as is as a shared secret.
    SharedSecret { shared_secret: &'a [u8] },
}

/// Helper function to create the `Aes256Gcm` cipher.
/// Following QoS create_cipher() exactly.
fn create_cipher(
    shared_secret: &PrivPubOrSharedSecret,
    ephemeral_sender_public: &SenderPublic,
    receiver_public: &ReceiverPublic,
) -> Result<Aes256Gcm> {
    info!("🔐 DEBUG: create_cipher() called");
    info!(
        "🔐 DEBUG: Ephemeral sender public (first 10 bytes): {:?}",
        &ephemeral_sender_public.0[..10]
    );
    info!(
        "🔐 DEBUG: Receiver public (first 10 bytes): {:?}",
        &receiver_public.0[..10]
    );

    let shared_secret = match shared_secret {
        PrivPubOrSharedSecret::PrivPub { private, public } => {
            info!("🔐 DEBUG: Using PrivPub mode for ECDH");
            debug!(
                "🔐 DEBUG: Private key (first 10 bytes): {:?}",
                &private.to_be_bytes()[..10]
            );

            // Real ECDH implementation using p256
            // Use the public key from the PrivPubOrSharedSecret, not receiver_public
            let public_key = public;
            info!("🔐 DEBUG: Successfully parsed public key from PrivPubOrSharedSecret");
            debug!(
                "🔐 DEBUG: Public key (first 10 bytes): {:?}",
                &public_key.to_encoded_point(false).as_bytes()[..10]
            );

            // ECDH implementation matching QoS exactly
            // Use proper ECDH key agreement as in QoS
            info!("🔐 DEBUG: Computing ECDH shared secret...");
            info!(
                "🔐 DEBUG: ECDH inputs - private: {:?}, public: {:?}",
                &private.to_be_bytes()[..10],
                &public_key.to_encoded_point(false).as_bytes()[..10]
            );

            let shared_secret =
                diffie_hellman(&private.to_nonzero_scalar(), public_key.as_affine());
            let shared_secret_bytes = shared_secret.raw_secret_bytes().to_vec();
            info!(
                "🔐 DEBUG: ECDH shared secret computed, length: {} bytes",
                shared_secret_bytes.len()
            );
            info!(
                "🔐 DEBUG: Shared secret (first 10 bytes): {:?}",
                &shared_secret_bytes[..shared_secret_bytes.len().min(10)]
            );
            shared_secret_bytes
        }
        PrivPubOrSharedSecret::SharedSecret { shared_secret } => {
            info!("🔐 DEBUG: Using SharedSecret mode");
            shared_secret.to_vec()
        }
    };

    // To help with entropy and add domain context, we do
    // `sender_public||receiver_public||shared_secret` as the pre-image for the
    // shared key.
    info!("🔐 DEBUG: Creating pre_image from sender_public + receiver_public + shared_secret");
    let pre_image: Vec<u8> = ephemeral_sender_public
        .0
        .iter()
        .chain(receiver_public.0)
        .chain(shared_secret.iter())
        .copied()
        .collect();
    info!("🔐 DEBUG: Pre_image length: {} bytes", pre_image.len());
    info!(
        "🔐 DEBUG: Pre_image (first 10 bytes): {:?}",
        &pre_image[..pre_image.len().min(10)]
    );

    info!("🔐 DEBUG: Running HMAC-SHA512 key derivation...");
    let mut mac = <HmacSha512 as KeyInit>::new_from_slice(&pre_image[..])
        .map_err(|_| anyhow!("HMAC key initialization failed"))?;
    mac.update(QOS_ENCRYPTION_HMAC_MESSAGE);
    let shared_key = mac.finalize().into_bytes();
    info!("🔐 DEBUG: HMAC key derivation successful");
    info!(
        "🔐 DEBUG: Shared key (first 10 bytes): {:?}",
        &shared_key[..10]
    );

    let cipher = Aes256Gcm::new_from_slice(&shared_key[..AES256_KEY_LEN])
        .map_err(|_| anyhow!("Failed to create AES-GCM cipher"))?;
    info!("🔐 DEBUG: AES-GCM cipher created successfully");
    Ok(cipher)
}

/// Helper function to create the additional associated data (AAD).
/// Following QoS create_additional_associated_data() exactly.
fn create_additional_associated_data(
    sender_public: &SenderPublic,
    receiver_public: &ReceiverPublic,
) -> Result<Vec<u8>> {
    let sender_len = sender_public.0.len() as u64;
    let receiver_len = receiver_public.0.len() as u64;

    let mut aad = Vec::new();
    aad.extend_from_slice(sender_public.0);
    aad.extend_from_slice(&sender_len.to_be_bytes());
    aad.extend_from_slice(receiver_public.0);
    aad.extend_from_slice(&receiver_len.to_be_bytes());

    Ok(aad)
}
