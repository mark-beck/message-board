use super::config::Config;
use super::crypto;
use crate::schema::{Role, User, UserWithHash, UserWithPw};
use anyhow::{anyhow, Result};
use futures_util::stream::StreamExt;
use mongodb::options::Credential;
use mongodb::{bson::doc, options::ClientOptions, Client, Collection};
use tracing::{error, info, warn};

#[derive(Clone)]
pub struct Mongo {
    users: Collection<UserWithHash>,
}

impl Mongo {
    pub async fn from_config(config: Config) -> Result<Self> {

        let mut client_options = ClientOptions::parse(format!(
            "mongodb://{}:{}",
            config.db.address.clone(), config.db.port.clone()
        ))
        .await?;

        // Manually set an option
        client_options.app_name = Some("auth_server".to_string());
        let cred = Credential::builder()
            .username(config.db.user.clone())
            .password(config.db.password.clone())
            .build();
        client_options.credential = Some(cred);

        // Get a handle to the cluster
        let client = Client::with_options(client_options)?;
        // Ping the server to see if you can connect to the cluster
        client
            .database("admin")
            .run_command(doc! {"ping": 1}, None)
            .await?;
        tracing::info!("Mongo Connection sucessfull");

        let mongo = Mongo {
            users: client
                .database("auth_server")
                .collection::<UserWithHash>("users"),
        };

        match mongo
            .users
            .find_one(doc! {"name": &config.default_user.name.clone()}, None)
            .await?
        {
            Some(_user) => {
                if !mongo
                    .verify_user(&UserWithPw {
                        name: config.default_user.name.clone(),
                        password: config.default_user.pass.clone(),
                    })
                    .await
                {
                    warn!("default user password changed from config");
                }
            }
            None => {
                if config.default_user.create {
                    warn!("No default user, creating according to config");
                    mongo
                        .create_user(
                            User {
                                name: config.default_user.name.clone(),
                                password: config.default_user.pass.clone(),
                                email: "".into(),
                                roles: vec![Role::Admin, Role::Moderator, Role::User],
                            }
                            .into(),
                        )
                        .await?;
                } else {
                    error!("No default user");
                    panic!("No default user");
                }
            }
        }
        Ok(mongo)
    }

    pub async fn create_user(&self, user: UserWithHash) -> Result<()> {
        if self.get_user_from_name(&user.name).await.is_some() {
            return Err(anyhow!("User exists!"));
        }
        self.users.insert_one(&user, None).await?;
        info!("adding user {:?} with roles {:?}", user.name, user.roles);
        Ok(())
    }

    pub async fn verify_user(&self, user: &UserWithPw) -> bool {
        match self.get_user_from_name(&user.name).await {
            None => false,
            Some(userhashed) => crypto::verify(userhashed.hash.0, &user.password),
        }
    }

    pub async fn get_user_from_name(&self, name: &str) -> Option<UserWithHash> {
        match self.users.find_one(doc! {"name": name}, None).await {
            Ok(u) => u,
            Err(e) => {
                warn!("Error while searching for user: {:?}", e);
                None
            }
        }
    }

    pub async fn delete_user(&self, name: &str) -> Result<()> {
        match self.users.delete_one(doc! {"name": name}, None).await {
            Ok(_) => Ok(()),
            Err(e) => Err(anyhow::Error::new(e)),
        }
    }

    pub async fn get_all_users(&self) -> Result<Vec<UserWithHash>> {
        return match self.users.find(doc! {}, None).await {
            Ok(cursor) => Ok(cursor.filter_map(|v| async { v.ok() }).collect().await),
            Err(e) => {
                warn!("error while listing all users: {:?}", e);
                Err(e.into())
            }
        };
    }
}
