use std::sync::Arc;
use crate::config::{Jwt, WatchedConfig};
use crate::config::Secret::{KeyPair, Pass};
use crate::schema::{Role, UserClaims};
use anyhow::{anyhow, Result};
use jsonwebtoken::{decode, encode, Algorithm, DecodingKey, EncodingKey, Header, Validation};
use sodiumoxide::crypto::pwhash::argon2id13;
use tracing::trace;

#[derive(Clone)]
pub struct JwtIssuer {
    header: Header,
    encoding_key: EncodingKey,
    validation: Validation,
    config: Arc<WatchedConfig>,
}

impl JwtIssuer {
    pub async fn new(config: Arc<WatchedConfig>) -> Result<Self> {
        let r = match &config.0.read().await.jwt_config.secret {
            Some(Pass(p)) => Self {
                header: Header::new(Algorithm::HS256),
                encoding_key: EncodingKey::from_secret(p.as_bytes()),
                validation: Validation::default(),
                config: config.clone(),
            },
            Some(KeyPair(private, _public)) => Self {
                header: Header::new(Algorithm::ES256),
                encoding_key: EncodingKey::from_ec_pem(private)?,
                validation: Validation::new(Algorithm::ES256),
                config: config.clone(),
            },
            None => panic!("server kaputt")
        };
        Ok(r)
    }

    pub fn issue<T>(&self, user: T) -> Result<String>
    where
        T: Into<UserClaims>,
    {
        let claim: UserClaims = user.into();
        encode(&self.header, &claim, &self.encoding_key).map_err(|e| e.into())
    }

    // pub fn reissue(&self, jwt: &str) -> Result<String> {
    //     self.issue(self.decode(jwt).await?)
    // }

    pub async fn decode(&self, jwt: &str) -> Result<UserClaims> {
        let header = jsonwebtoken::decode_header(jwt)?;
        if header != self.header {
            return Err(anyhow!("header does not match"));
        }
        match &self.config.0.read().await.jwt_config.secret {
            Some(Pass(p)) => decode::<UserClaims>(
                jwt,
                &DecodingKey::from_secret(p.as_bytes()),
                &self.validation,
            ),
            Some(KeyPair(_private, public)) => {
                decode::<UserClaims>(jwt, &DecodingKey::from_ec_pem(public)?, &self.validation)
            }
            None => panic!("server kaputt")
        }
        .map(|ts| ts.claims)
        .map_err(|e| e.into())
    }

    pub async fn validate_level(&self, jwt: &str, level: Role) -> bool {
        trace!("validating level {:?}", level);
        if let Ok(claims) = self.decode(jwt).await {
            trace!("succesfully decoded");
            return claims.user.roles.contains(&level);
        }
        false
    }
}

pub fn hash(passwd: &str) -> argon2id13::HashedPassword {
    sodiumoxide::init().unwrap();
    argon2id13::pwhash(
        passwd.as_bytes(),
        argon2id13::OPSLIMIT_INTERACTIVE,
        argon2id13::MEMLIMIT_INTERACTIVE,
    )
    .unwrap()
}

pub fn verify(hash: [u8; 128], passwd: &str) -> bool {
    sodiumoxide::init().unwrap();
    match argon2id13::HashedPassword::from_slice(&hash) {
        Some(hp) => argon2id13::pwhash_verify(&hp, passwd.as_bytes()),
        _ => false,
    }
}
