use crate::config::Config;
use crate::config::Secret::{KeyPair, Pass};
use crate::schema::{Role, UserClaims};
use anyhow::{anyhow, Result};
use jsonwebtoken::{decode, encode, Algorithm, DecodingKey, EncodingKey, Header, Validation};
use sodiumoxide::crypto::pwhash::argon2id13;
use tracing::trace;
use std::sync::Arc;
use crate::mongo::Mongo;

#[derive(Clone)]
pub struct JwtIssuer {
    header: Header,
    encoding_key: EncodingKey,
    validation: Validation,
    config: Config,
}

impl JwtIssuer {
    pub async fn new(config: Config) -> Result<Self> {
        let r = match &config.jwt_config.secret {
            Some(Pass(p)) => Self {
                header: Header::new(Algorithm::HS256),
                encoding_key: EncodingKey::from_secret(p.as_bytes()),
                validation: Validation::default(),
                config,
            },
            Some(KeyPair(private, _public)) => Self {
                header: Header::new(Algorithm::ES256),
                encoding_key: EncodingKey::from_ec_pem(private)?,
                validation: Validation::new(Algorithm::ES256),
                config,
            },
            None => panic!("server kaputt")
        };
        Ok(r)
    }

    #[tracing::instrument(level="trace", skip(self, user))]
    pub fn issue<T>(&self, user: T) -> Result<String>
    where
        T: Into<UserClaims>,
    {
        let claim: UserClaims = user.into();
        encode(&self.header, &claim, &self.encoding_key).map_err(Into::into)
    }

    // pub fn reissue(&self, jwt: &str) -> Result<String> {
    //     self.issue(self.decode(jwt).await?)
    // }

    #[tracing::instrument(level="trace", skip(self))]
    pub async fn decode(&self, jwt: &str) -> Result<UserClaims> {
        let header = jsonwebtoken::decode_header(jwt)?;
        if header != self.header {
            return Err(anyhow!("header does not match"));
        }
        match &self.config.jwt_config.secret {
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
        .map_err(Into::into)
    }

    #[tracing::instrument(level="trace", skip(self, mongo))]
    pub async fn validate_level(&self, mongo: &Arc<Mongo>, jwt: &str, level: Role) -> bool {
        trace!("validating level {:?}", level);
        if let Ok(claims) = self.decode(jwt).await {
            trace!("succesfully decoded");

            let user = match mongo.get_user_from_id(&claims.user_id).await {
                Ok(user) => user,
                Err(_) => return false,
            };


            return user.roles.contains(&level);
        }
        false
    }
}

#[tracing::instrument(level="trace", skip(passwd))]
pub fn hash(passwd: &str) -> argon2id13::HashedPassword {
    sodiumoxide::init().unwrap();
    argon2id13::pwhash(
        passwd.as_bytes(),
        argon2id13::OPSLIMIT_INTERACTIVE,
        argon2id13::MEMLIMIT_INTERACTIVE,
    )
    .unwrap()
}

#[tracing::instrument(level="trace", skip(hash, passwd))]
pub fn verify(hash: [u8; 128], passwd: &str) -> bool {
    sodiumoxide::init().unwrap();
    match argon2id13::HashedPassword::from_slice(&hash) {
        Some(hp) => argon2id13::pwhash_verify(&hp, passwd.as_bytes()),
        _ => false,
    }
}
