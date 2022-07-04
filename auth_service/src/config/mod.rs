use anyhow::Result;
use serde::Deserialize;
use std::fs;
use derivative::Derivative;

#[derive(Deserialize, Clone, Derivative)]
#[derivative(Debug)]
pub struct Config {
    pub db: DB,
    pub default_user: DefaultUser,
    #[derivative(Debug="ignore")]
    pub jwt_config: Jwt,
}

impl Config {
    pub fn parse(path: &str) -> Result<Self> {
        let config_string = fs::read_to_string(path)?;
        let mut config: Self = toml::from_str(&config_string)?;
        tracing::debug!("{:?}", config);
        if let Some(ref path) = config.jwt_config.path {
            let private = std::fs::read(&path.0)?;
            let public = std::fs::read(&path.1)?;
            config.jwt_config.secret = Some(Secret::KeyPair(private, public));
        }
        tracing::debug!("\n{:?}", config);
        Ok(config)
    }
}

#[derive(Deserialize, Clone, Derivative)]
#[derivative(Debug)]
pub struct DB {
    pub address: String,
    pub port: String,
    pub user: String,
    #[derivative(Debug="ignore")]
    pub password: String,
}

#[derive(Deserialize, Clone, Derivative)]
#[derivative(Debug)]
pub struct DefaultUser {
    pub name: String,
    #[derivative(Debug="ignore")]
    pub pass: String,
    pub create: bool,
}

#[derive(Deserialize, Clone)]
pub struct Jwt {
    pub secret: Option<Secret>,
    pub path: Option<(String, String)>,
}

#[derive(Deserialize, Clone)]
#[serde(untagged)]
pub enum Secret {
    Pass(String),
    KeyPair(Vec<u8>, Vec<u8>),
}
