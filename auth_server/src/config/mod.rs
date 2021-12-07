#![allow(dead_code)]

use anyhow::Result;
use notify::{Watcher, RecursiveMode, watcher, DebouncedEvent};
use serde_derive::Deserialize;
use std::fs;
use std::path::Path;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;

pub struct WatchedConfig(pub(crate) Arc<RwLock<Config>>);

impl WatchedConfig {
    pub fn new(path: &str) -> Result<Self> {
        let config = Config::parse(path)?;

        // let mut hotwatch = Hotwatch::new().expect("hotwatch failed to initialize!");
        // if let Some(path) = &config.jwt_config.path {
        //     hotwatch.watch(&path.0, |event: Event| {
        //         tracing::info!("private key changed: {:?}", event)
        //     }).expect("failed to watch file!");
        //     hotwatch.watch(&path.1, |event: Event| {
        //         tracing::info!("public key changed: {:?}", event)
        //     }).expect("failed to watch file!");
        // }
        let keypath = config.jwt_config.path.clone();
        let rwlock = Arc::new(RwLock::new(config));
        let rwlock2 = rwlock.clone();
        let pathstr = path.to_string();
        tokio::spawn(async move {
            let (atx, mut arx) = tokio::sync::mpsc::unbounded_channel();
            let pathstr2 = pathstr.clone();
            std::thread::spawn(move || {
                let (tx, rx) = std::sync::mpsc::channel::<DebouncedEvent>();

                // Create a watcher object, delivering debounced events.
                // The notification back-end is selected based on the platform.
                let mut watcher = watcher(tx, Duration::from_secs(10)).unwrap();

                // Add a path to be watched. All files and directories at that path and
                // below will be monitored for changes.
                watcher.watch(pathstr2, RecursiveMode::Recursive).unwrap();
                if let Some(p) = keypath {
                    watcher.watch(p.0, RecursiveMode::Recursive).unwrap();
                    watcher.watch(p.1, RecursiveMode::Recursive).unwrap();
                }


                loop {
                    atx.send(rx.recv());
                }
            });

            loop {
                match arx.recv().await.unwrap() {
                    Ok(event) => {
                        tracing::info!("config file changed");
                        let mut conf = rwlock2.write().await;
                        if let Ok(newconf) = Config::parse(&pathstr) {
                            *conf = newconf;
                        }
                    },
                    Err(e) => tracing::debug!("filewatcher error: {}", e),
                }
            }

        });


        Ok(WatchedConfig(rwlock))
    }
}

#[derive(Deserialize, Debug, Clone)]
pub struct Config {
    pub mail: Mail,
    pub db: DB,
    pub default_user: DefaultUser,
    pub jwt_config: Jwt,
}

impl Config {
    pub fn parse(path: &str) -> Result<Self> {
        let config_string = fs::read_to_string(path)?;
        let mut config: Config = toml::from_str(&config_string)?;
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

#[derive(Deserialize, Debug, Clone)]
pub struct Mail {
    pub server: String,
    pub sender: String,
}

#[derive(Deserialize, Debug, Clone)]
pub struct DB {
    pub address: String,
    pub port: String,
    pub user: String,
    pub password: String,
}

#[derive(Deserialize, Debug, Clone)]
pub struct DefaultUser {
    pub name: String,
    pub pass: String,
    pub create: bool,
}

#[derive(Deserialize, Debug, Clone)]
pub struct Jwt {
    pub secret: Option<Secret>,
    pub path: Option<(String, String)>,
}

#[derive(Deserialize, Debug, Clone)]
#[serde(untagged)]
pub enum Secret {
    Pass(String),
    KeyPair(Vec<u8>, Vec<u8>),
}
