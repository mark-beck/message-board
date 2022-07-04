use anyhow::Result;
use derivative::Derivative;
use serde::Deserialize;
use std::env;
use std::fs;

#[derive(Deserialize, Clone, Derivative)]
#[derivative(Debug)]
pub struct Config {
    pub image_service: ImageServiceConfig,
    pub db: DB,
    pub default_user: DefaultUser,
    #[derivative(Debug = "ignore")]
    pub jwt_config: JwtSecret,
}

impl Config {
    #[tracing::instrument(level = "trace")]
    pub fn from_env() -> Result<Self> {
        let config = Config {
            image_service: ImageServiceConfig {
                url: env::var("IMAGE_SERVICE_URL")?,
            },
            db: DB {
                address: std::env::var("DB_ADDRESS")?,
                port: std::env::var("DB_PORT")?,
                user: std::env::var("DB_USER")?,
                password: std::env::var("DB_PASSWORD")?,
            },
            default_user: DefaultUser {
                name: std::env::var("DEFAULT_USER")?,
                pass: std::env::var("DEFAULT_PASSWORD")?,
                create: (std::env::var("CREATE_DEFAULT_USER").is_ok()),
            },
            jwt_config: if let Ok(secret) = std::env::var("JWT_SECRET") {
                JwtSecret::Pass(secret)
            } else {
                JwtSecret::KeyPair {
                    private: fs::read(env::var("JWT_PRIVATE_PATH")?)?,
                    public: fs::read(env::var("JWT_PUBLIC_PATH")?)?,
                }
            },
        };

        tracing::debug!("{:?}", config);
        Ok(config)
    }
}

#[derive(Deserialize, Clone, Derivative)]
#[derivative(Debug)]
pub struct DB {
    pub address: String,
    pub port: String,
    pub user: String,
    #[derivative(Debug = "ignore")]
    pub password: String,
}

#[derive(Deserialize, Clone, Debug)]
pub struct ImageServiceConfig {
    pub url: String,
}

#[derive(Deserialize, Clone, Derivative)]
#[derivative(Debug)]
pub struct DefaultUser {
    pub name: String,
    #[derivative(Debug = "ignore")]
    pub pass: String,
    pub create: bool,
}

#[derive(Deserialize, Clone)]
#[serde(untagged)]
pub enum JwtSecret {
    Pass(String),
    KeyPair { private: Vec<u8>, public: Vec<u8> },
}
