use super::config::Config;
use super::crypto;
use crate::schema::{Role, User, UserWithHash, LoginRequest, UpdateRequest};
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
    #[tracing::instrument(level="trace")]
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
            .run_command(doc! {"ping": 1u32}, None)
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
                    .verify_user(&LoginRequest {
                        email: config.default_user.name.clone(),
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
                                id: uuid::Uuid::new_v4().to_string(),
                                name: config.default_user.name.clone(),
                                password: config.default_user.pass.clone(),
                                email: "".into(),
                                roles: vec![Role::Admin, Role::Moderator, Role::User],
                                image: None,
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

    #[tracing::instrument(level="trace", skip(self))]
    pub async fn create_user(&self, user: UserWithHash) -> Result<()> {
        if self.get_user_from_email(&user.name).await.is_ok() {
            return Err(anyhow!("User exists!"));
        }
        self.users.insert_one(&user, None).await?;
        info!("adding user {:?} with roles {:?}", user.name, user.roles);
        Ok(())
    }

    #[tracing::instrument(level="trace", skip(self, request))]
    pub async fn verify_user(&self, request: &LoginRequest) -> bool {
        match self.get_user_from_email(&request.email).await {
            Err(_) => false,
            Ok(user) => crypto::verify(user.hash.0, &request.password),
        }
    }

    #[tracing::instrument(level="trace", skip(self))]
    pub async fn get_user_from_email(&self, email: &str) -> Result<UserWithHash> {
        match self.users.find_one(doc! {"email": email}, None).await {
            Ok(Some(u)) => Ok(u),
            Ok(None) => Err(anyhow!("User not found")),
            Err(e) => {
                warn!("Error while searching for user: {:?}", e);
                Err(e.into())
            }
        }
    }

    #[tracing::instrument(level="trace", skip(self))]
    pub async fn get_user_from_id(&self, id: &str) -> Result<UserWithHash> {
        match self.users.find_one(doc! {"id": id}, None).await {
            Ok(Some(u)) => Ok(u),
            Ok(None) => Err(anyhow!("User not found")),
            Err(e) => {
                warn!("Error while searching for user: {:?}", e);
                Err(e.into())
            }
        }
    }

    #[tracing::instrument(level="trace", skip(self, update_request))]
    pub async fn update_user(&self, id: &str, update_request: &UpdateRequest) -> Result<()> {
        let mut user = self.get_user_from_id(id).await?;
        if let Some(password) = update_request.password.clone() {
            user.hash = crypto::hash(&password);
        }
        if let Some(email) = update_request.email.clone() {
            user.email = email;
        }
        if let Some(image) = update_request.image.clone() {
            user.image = Some(image);
        }
        if let Some(name) = update_request.name.clone() {
            user.name = name;
        }
        self.users.replace_one(doc! {"id": id}, &user, None).await?;
        Ok(())
    }

    #[tracing::instrument(level="trace", skip(self))]
    pub async fn delete_user(&self, id: &str) -> Result<()> {
        match self.users.delete_one(doc! {"id": id}, None).await {
            Ok(_) => Ok(()),
            Err(e) => Err(anyhow::Error::new(e)),
        }
    }

    #[tracing::instrument(level="trace", skip(self))]
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
