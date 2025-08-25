use axum::{extract::State, http::StatusCode, Json};
use log::{debug, error, info, warn};
use uuid::Uuid;

use crate::AppState;
use renclave_shared::*;

/// Health check endpoint
pub async fn health_check() -> StatusCode {
    debug!("🏥 Health check endpoint called");
    StatusCode::OK
}

/// Get service information
pub async fn get_info(
    State(state): State<AppState>,
) -> std::result::Result<Json<InfoResponse>, (StatusCode, Json<ErrorResponse>)> {
    info!("ℹ️  Service info requested");

    // Get network status
    let network_status = state.network_manager.get_status().await;
    let network_status_str = if network_status.connectivity.external {
        "connected"
    } else if network_status.connectivity.gateway {
        "limited"
    } else {
        "disconnected"
    }
    .to_string();

    // Try to get enclave info
    let enclave_id = match state.enclave_client.get_info().await {
        Ok(response) => match response.result {
            EnclaveResult::Info { enclave_id, .. } => enclave_id,
            _ => "unknown".to_string(),
        },
        Err(_) => "unavailable".to_string(),
    };

    let info = InfoResponse {
        version: env!("CARGO_PKG_VERSION").to_string(),
        service: "QEMU Host API Gateway".to_string(),
        enclave_id,
        capabilities: vec![
            "seed_generation".to_string(),
            "seed_validation".to_string(),
            "network_connectivity".to_string(),
            "enclave_communication".to_string(),
        ],
        network_status: network_status_str,
    };

    debug!("✅ Service info response prepared");
    Ok(Json(info))
}

/// Generate seed phrase
pub async fn generate_seed(
    State(state): State<AppState>,
    Json(request): Json<GenerateSeedRequest>,
) -> std::result::Result<Json<GenerateSeedResponse>, (StatusCode, Json<ErrorResponse>)> {
    let request_id = Uuid::new_v4().to_string();
    info!("🔑 Seed generation requested (ID: {})", request_id);

    // Validate request
    let strength = request.strength.unwrap_or(256);
    if ![128, 160, 192, 224, 256].contains(&strength) {
        warn!("❌ Invalid strength requested: {}", strength);
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Invalid strength. Must be 128, 160, 192, 224, or 256 bits".to_string(),
                code: 400,
                request_id: Some(request_id),
            }),
        ));
    }

    debug!(
        "📋 Request validated - strength: {}, passphrase: {}",
        strength,
        request.passphrase.is_some()
    );

    // Send request to enclave
    match state
        .enclave_client
        .generate_seed(strength, request.passphrase)
        .await
    {
        Ok(enclave_response) => match enclave_response.result {
            EnclaveResult::SeedGenerated {
                seed_phrase,
                entropy,
                strength,
                word_count,
            } => {
                info!("✅ Seed generation successful (ID: {})", request_id);
                Ok(Json(GenerateSeedResponse {
                    seed_phrase,
                    entropy,
                    strength,
                    word_count,
                }))
            }
            EnclaveResult::Error { message, code } => {
                error!("❌ Enclave error during seed generation: {}", message);
                Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse {
                        error: message,
                        code,
                        request_id: Some(request_id),
                    }),
                ))
            }
            _ => {
                error!("❌ Unexpected response type from enclave");
                Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse {
                        error: "Unexpected response from enclave".to_string(),
                        code: 500,
                        request_id: Some(request_id),
                    }),
                ))
            }
        },
        Err(e) => {
            error!("❌ Failed to communicate with enclave: {}", e);
            Err((
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    error: format!("Enclave communication failed: {}", e),
                    code: 503,
                    request_id: Some(request_id),
                }),
            ))
        }
    }
}

/// Validate seed phrase
pub async fn validate_seed(
    State(state): State<AppState>,
    Json(request): Json<ValidateSeedRequest>,
) -> std::result::Result<Json<ValidateSeedResponse>, (StatusCode, Json<ErrorResponse>)> {
    let request_id = Uuid::new_v4().to_string();
    info!("🔍 Seed validation requested (ID: {})", request_id);

    // Validate request
    if request.seed_phrase.trim().is_empty() {
        warn!("❌ Empty seed phrase provided");
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Seed phrase cannot be empty".to_string(),
                code: 400,
                request_id: Some(request_id),
            }),
        ));
    }

    debug!(
        "📋 Request validated - seed phrase length: {}",
        request.seed_phrase.len()
    );

    // Send request to enclave
    match state
        .enclave_client
        .validate_seed(request.seed_phrase)
        .await
    {
        Ok(enclave_response) => match enclave_response.result {
            EnclaveResult::SeedValidated { valid, word_count } => {
                info!(
                    "✅ Seed validation completed (ID: {}, valid: {})",
                    request_id, valid
                );
                Ok(Json(ValidateSeedResponse { valid, word_count }))
            }
            EnclaveResult::Error { message, code } => {
                error!("❌ Enclave error during seed validation: {}", message);
                Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse {
                        error: message,
                        code,
                        request_id: Some(request_id),
                    }),
                ))
            }
            _ => {
                error!("❌ Unexpected response type from enclave");
                Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse {
                        error: "Unexpected response from enclave".to_string(),
                        code: 500,
                        request_id: Some(request_id),
                    }),
                ))
            }
        },
        Err(e) => {
            error!("❌ Failed to communicate with enclave: {}", e);
            Err((
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    error: format!("Enclave communication failed: {}", e),
                    code: 503,
                    request_id: Some(request_id),
                }),
            ))
        }
    }
}

/// Get network status
pub async fn network_status(State(state): State<AppState>) -> Json<serde_json::Value> {
    debug!("🌐 Network status requested");

    let status = state.network_manager.get_status().await;

    let response = serde_json::json!({
        "tap_interface": status.tap_interface,
        "guest_ip": status.guest_ip,
        "gateway_ip": status.gateway_ip,
        "connectivity": {
            "loopback": status.connectivity.loopback,
            "gateway": status.connectivity.gateway,
            "external": status.connectivity.external,
            "dns": status.connectivity.dns,
        }
    });

    debug!("✅ Network status response prepared");
    Json(response)
}

/// Test network connectivity
pub async fn test_connectivity(State(state): State<AppState>) -> Json<serde_json::Value> {
    info!("🔍 Network connectivity test requested");

    let report = match state.connectivity_tester.run_comprehensive_test().await {
        Ok(report) => report,
        Err(e) => {
            error!("❌ Connectivity test failed: {}", e);
            return Json(serde_json::json!({
                "success": false,
                "error": format!("Connectivity test failed: {}", e)
            }));
        }
    };

    let response = serde_json::json!({
        "success": true,
        "gateway_ping": {
            "success": report.gateway_ping.success,
            "target": report.gateway_ping.target,
            "packets_sent": report.gateway_ping.packets_sent,
            "packets_received": report.gateway_ping.packets_received,
            "avg_time_ms": report.gateway_ping.avg_time_ms,
        },
        "external_ping": {
            "success": report.external_ping.success,
            "target": report.external_ping.target,
            "packets_sent": report.external_ping.packets_sent,
            "packets_received": report.external_ping.packets_received,
            "avg_time_ms": report.external_ping.avg_time_ms,
        },
        "dns_test": {
            "success": report.dns_test.success,
            "hostname": report.dns_test.hostname,
            "duration_ms": report.dns_test.duration.as_millis(),
        },
        "http_test": {
            "success": report.http_test.success,
            "url": report.http_test.url,
            "duration_ms": report.http_test.duration.as_millis(),
        },
        "total_duration_ms": report.total_duration.as_millis(),
    });

    info!("✅ Network connectivity test completed");
    Json(response)
}

/// Get enclave information
pub async fn enclave_info(
    State(state): State<AppState>,
) -> std::result::Result<Json<serde_json::Value>, (StatusCode, Json<ErrorResponse>)> {
    debug!("🔒 Enclave info requested");

    // Check enclave health first
    let is_healthy = state.enclave_client.health_check().await.unwrap_or(false);

    if !is_healthy {
        warn!("⚠️  Enclave is not healthy");
        return Err((
            StatusCode::SERVICE_UNAVAILABLE,
            Json(ErrorResponse {
                error: "Enclave is not available".to_string(),
                code: 503,
                request_id: None,
            }),
        ));
    }

    // Get enclave information
    match state.enclave_client.get_info().await {
        Ok(enclave_response) => match enclave_response.result {
            EnclaveResult::Info {
                version,
                enclave_id,
                capabilities,
            } => {
                let response = serde_json::json!({
                    "healthy": true,
                    "version": version,
                    "enclave_id": enclave_id,
                    "capabilities": capabilities,
                });

                debug!("✅ Enclave info response prepared");
                Ok(Json(response))
            }
            EnclaveResult::Error { message, code } => {
                error!("❌ Enclave error: {}", message);
                Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse {
                        error: message,
                        code,
                        request_id: None,
                    }),
                ))
            }
            _ => {
                error!("❌ Unexpected response type from enclave");
                Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse {
                        error: "Unexpected response from enclave".to_string(),
                        code: 500,
                        request_id: None,
                    }),
                ))
            }
        },
        Err(e) => {
            error!("❌ Failed to communicate with enclave: {}", e);
            Err((
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    error: format!("Enclave communication failed: {}", e),
                    code: 503,
                    request_id: None,
                }),
            ))
        }
    }
}

/// Derive key from seed phrase
pub async fn derive_key(
    State(state): State<AppState>,
    Json(request): Json<DeriveKeyRequest>,
) -> std::result::Result<Json<DeriveKeyResponse>, (StatusCode, Json<ErrorResponse>)> {
    let request_id = Uuid::new_v4().to_string();
    info!("🔑 Key derivation requested (ID: {})", request_id);

    // Validate request
    if request.seed_phrase.trim().is_empty() {
        warn!("❌ Empty seed phrase provided");
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Seed phrase cannot be empty".to_string(),
                code: 400,
                request_id: Some(request_id),
            }),
        ));
    }

    if request.path.trim().is_empty() {
        warn!("❌ Empty derivation path provided");
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Derivation path cannot be empty".to_string(),
                code: 400,
                request_id: Some(request_id),
            }),
        ));
    }

    if request.curve.trim().is_empty() {
        warn!("❌ Empty curve provided");
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Curve cannot be empty".to_string(),
                code: 400,
                request_id: Some(request_id),
            }),
        ));
    }

    debug!(
        "📋 Request validated - path: {}, curve: {}",
        request.path, request.curve
    );

    // Send request to enclave
    match state
        .enclave_client
        .derive_key(request.seed_phrase, request.path, request.curve)
        .await
    {
        Ok(enclave_response) => match enclave_response.result {
            EnclaveResult::KeyDerived {
                private_key,
                public_key,
                address,
                path,
                curve,
            } => {
                info!("✅ Key derivation successful (ID: {})", request_id);
                Ok(Json(DeriveKeyResponse {
                    private_key,
                    public_key,
                    address,
                    path,
                    curve,
                }))
            }
            EnclaveResult::Error { message, code } => {
                error!("❌ Enclave error during key derivation: {}", message);
                Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse {
                        error: message,
                        code,
                        request_id: Some(request_id),
                    }),
                ))
            }
            _ => {
                error!("❌ Unexpected response type from enclave");
                Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse {
                        error: "Unexpected response from enclave".to_string(),
                        code: 500,
                        request_id: Some(request_id),
                    }),
                ))
            }
        },
        Err(e) => {
            error!("❌ Failed to communicate with enclave: {}", e);
            Err((
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    error: format!("Enclave communication failed: {}", e),
                    code: 503,
                    request_id: Some(request_id),
                }),
            ))
        }
    }
}

/// Derive address from seed phrase
pub async fn derive_address(
    State(state): State<AppState>,
    Json(request): Json<DeriveAddressRequest>,
) -> std::result::Result<Json<DeriveAddressResponse>, (StatusCode, Json<ErrorResponse>)> {
    let request_id = Uuid::new_v4().to_string();
    info!("📍 Address derivation requested (ID: {})", request_id);

    // Validate request
    if request.seed_phrase.trim().is_empty() {
        warn!("❌ Empty seed phrase provided");
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Seed phrase cannot be empty".to_string(),
                code: 400,
                request_id: Some(request_id),
            }),
        ));
    }

    if request.path.trim().is_empty() {
        warn!("❌ Empty derivation path provided");
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Derivation path cannot be empty".to_string(),
                code: 400,
                request_id: Some(request_id),
            }),
        ));
    }

    if request.curve.trim().is_empty() {
        warn!("❌ Empty curve provided");
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Curve cannot be empty".to_string(),
                code: 400,
                request_id: Some(request_id),
            }),
        ));
    }

    debug!(
        "📋 Request validated - path: {}, curve: {}",
        request.path, request.curve
    );

    // Send request to enclave
    match state
        .enclave_client
        .derive_address(request.seed_phrase, request.path, request.curve)
        .await
    {
        Ok(enclave_response) => match enclave_response.result {
            EnclaveResult::AddressDerived {
                address,
                path,
                curve,
            } => {
                info!("✅ Address derivation successful (ID: {})", request_id);
                Ok(Json(DeriveAddressResponse {
                    address,
                    path,
                    curve,
                }))
            }
            EnclaveResult::Error { message, code } => {
                error!("❌ Enclave error during address derivation: {}", message);
                Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse {
                        error: message,
                        code,
                        request_id: Some(request_id),
                    }),
                ))
            }
            _ => {
                error!("❌ Unexpected response type from enclave");
                Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse {
                        error: "Unexpected response from enclave".to_string(),
                        code: 500,
                        request_id: Some(request_id),
                    }),
                ))
            }
        },
        Err(e) => {
            error!("❌ Failed to communicate with enclave: {}", e);
            Err((
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    error: format!("Enclave communication failed: {}", e),
                    code: 503,
                    request_id: Some(request_id),
                }),
            ))
        }
    }
}
