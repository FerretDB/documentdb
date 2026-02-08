/*-------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation.  All rights reserved.
 *
 * src/context/connection.rs
 *
 *-------------------------------------------------------------------------
 */

use std::{
    hash::{DefaultHasher, Hash, Hasher},
    sync::Arc,
    time::Instant,
};

use bson::RawDocumentBuf;
use openssl::ssl::SslRef;
use uuid::{Builder, Uuid};

use crate::{
    auth::AuthState,
    configuration::DynamicConfiguration,
    context::{Cursor, CursorStoreEntry, ServiceContext},
    error::{DocumentDBError, Result},
    postgres::Connection,
    telemetry::TelemetryProvider,
};

pub struct ConnectionContext {
    pub start_time: Instant,
    pub connection_id: Uuid,
    pub service_context: Arc<ServiceContext>,
    pub auth_state: AuthState,
    pub requires_response: bool,
    pub client_information: Option<RawDocumentBuf>,
    pub transaction: Option<(Vec<u8>, i64)>,
    pub telemetry_provider: Option<Box<dyn TelemetryProvider>>,
    pub ip_address: String,
    pub cipher_type: i32,
    pub ssl_protocol: String,
}

impl ConnectionContext {
    pub fn new(
        service_context: ServiceContext,
        telemetry_provider: Option<Box<dyn TelemetryProvider>>,
        ip_address: String,
        tls_config: Option<&SslRef>,
        connection_id: Uuid,
    ) -> Self {
        let tls_provider = service_context.tls_provider();

        let cipher_type = tls_config
            .map(|tls| tls_provider.ciphersuite_to_i32(tls.current_cipher()))
            .unwrap_or_default();

        let ssl_protocol = tls_config
            .map(|tls| tls.version_str().to_string())
            .unwrap_or_default();

        ConnectionContext {
            start_time: Instant::now(),
            connection_id,
            service_context: Arc::new(service_context),
            auth_state: AuthState::new(),
            requires_response: true,
            client_information: None,
            transaction: None,
            telemetry_provider,
            ip_address,
            cipher_type,
            ssl_protocol,
        }
    }

    pub async fn get_cursor(&self, id: i64, username: &str) -> Option<CursorStoreEntry> {
        // If there is a transaction, get the cursor to its store
        if let Some((session_id, _)) = self.transaction.as_ref() {
            let transaction_store = self.service_context.transaction_store();
            if let Some((_, transaction)) =
                transaction_store.transactions.read().await.get(session_id)
            {
                return transaction
                    .cursors
                    .get_cursor((id, username.to_string()))
                    .await;
            }
        }

        self.service_context.get_cursor(id, username).await
    }

    pub async fn add_cursor(
        &self,
        conn: Option<Arc<Connection>>,
        cursor: Cursor,
        username: &str,
        db: &str,
        collection: &str,
        session_id: Option<Vec<u8>>,
    ) {
        let key = (cursor.cursor_id, username.to_string());
        let value = CursorStoreEntry {
            conn,
            cursor,
            db: db.to_string(),
            collection: collection.to_string(),
            timestamp: Instant::now(),
            session_id,
        };

        // If there is a transaction, add the cursor to its store
        if let Some((session_id, _)) = self.transaction.as_ref() {
            let transaction_store = self.service_context.transaction_store();
            if let Some((_, transaction)) =
                transaction_store.transactions.read().await.get(session_id)
            {
                transaction.cursors.add_cursor(key, value).await;
                return;
            }
        }

        // Otherwise add it to the service context
        self.service_context.add_cursor(key, value).await
    }

    pub async fn allocate_data_pool(&self) -> Result<()> {
        let username = self.auth_state.username()?;
        let password = self
            .auth_state
            .password
            .as_ref()
            .ok_or(DocumentDBError::internal_error(
                "Password is missing on pg connection acquisition".to_string(),
            ))?;

        self.service_context
            .allocate_data_pool(username, password)
            .await
    }

    pub fn dynamic_configuration(&self) -> Arc<dyn DynamicConfiguration> {
        self.service_context.dynamic_configuration()
    }

    /// Generates a per-request activity ID by embedding the given `request_id`
    /// into the caller’s connection UUID and returning it as a hyphenated string.
    ///
    /// The function copies the current `connection_id` (a 16-byte UUID), overwrites
    /// bytes 12..16 (the final 4 bytes) with `request_id.to_be_bytes()`
    /// to preserve UUID version/variant bits, then returns the resulting UUID’s
    /// canonical (lowercase, hyphenated) string form.
    ///
    /// # Parameters
    /// - `request_id`: 32-bit identifier to embed (stored big-endian in bytes 12–15).
    ///
    /// # Returns
    /// A `String` containing the new UUID, e.g. `"550e8400-e29b-41d4-a716-446655440000"`.
    pub fn generate_request_activity_id(&mut self, request_id: i32) -> String {
        let mut activity_id_bytes = *self.connection_id.as_bytes();
        activity_id_bytes[12..].copy_from_slice(&request_id.to_be_bytes());
        Builder::from_bytes(activity_id_bytes)
            .into_uuid()
            .to_string()
    }

    /// Returns a positive 64-bit integer derived from the `connection_id` UUID by hashing it.
    ///
    /// This computes a `u64` hash of `self.connection_id` using `DefaultHasher`
    /// and then masks the result to ensure the returned value is non-negative
    /// when cast to `i64`.
    ///
    /// # Behavior
    /// - Uses `std::collections::hash_map::DefaultHasher`
    /// - The current mask `0x7FFF_FFFF` keeps only **31 bits** of the hash
    ///   keeping the value non-negative.
    pub fn get_connection_id_as_i64(&self) -> i64 {
        let mut hasher = DefaultHasher::new();
        self.connection_id.hash(&mut hasher);
        let finished_hash = hasher.finish();
        (finished_hash & 0x7FFF_FFFF) as i64
    }
}
