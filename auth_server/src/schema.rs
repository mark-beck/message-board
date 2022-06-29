use crate::crypto;
use actix_web::web::Json;
use time::{Duration, OffsetDateTime};
use serde::{Deserialize, Serialize};
use sodiumoxide::crypto::pwhash::argon2id13::HashedPassword;
use std::ops::Add;
use uuid::Uuid;
use derivative::Derivative;

#[derive(Copy, Clone, Serialize, Deserialize, Debug, Eq, PartialEq)]
pub enum Role {
    User,
    Moderator,
    Admin,
}

#[derive(Serialize, Deserialize)]
pub enum Status {
    Ok,
    Failed,
}

// #[derive(Serialize, Deserialize)]
// pub struct Resp {
//     pub(crate) status: Status,
//     pub(crate) message: Option<String>,
//     pub(crate) jwt: Option<String>,
// }

#[derive(Serialize, Deserialize, Debug)]
pub struct TokenResponse {
    pub(crate) token: String,
    pub(crate) user: UserInfoFull,
}

#[derive(Serialize, Deserialize, Derivative)]
#[derivative(Debug)]
pub struct RegisteringUser {
    pub(crate) name: String,
    #[derivative(Debug="ignore")]
    pub(crate) password: String,
    pub(crate) email: String,
    #[derivative(Debug="ignore")]
    pub(crate) image: Option<String>,
}

impl RegisteringUser {
    pub fn add_roles(self, roles: Vec<Role>) -> User {
        User {
            id : Uuid::new_v4().to_string(),
            name: self.name,
            password: self.password,
            email: self.email,
            roles,
            image: self.image,
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct UserClaims {
    exp: i64,
    nbf: i64,
    sub: String,
    pub user_id: String,
}

impl From<UserWithHash> for UserClaims {
    fn from(uh: UserWithHash) -> Self {
        UserClaims {
            exp: OffsetDateTime::now_utc().add(Duration::hours(1)).unix_timestamp(),
            nbf: OffsetDateTime::now_utc().unix_timestamp(),
            sub: uh.name.clone(),
            user_id: uh.id,
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Derivative)]
#[derivative(Debug)]
pub struct UserWithHash {
    pub id: String,
    pub name: String,
    #[derivative(Debug="ignore")]
    pub hash: HashedPassword,
    pub email: String,
    pub roles: Vec<Role>,
    #[derivative(Debug="ignore")]
    pub image: Option<String>,
}

impl From<User> for UserWithHash {
    fn from(user: User) -> Self {
        Self {
            id: user.id,
            name: user.name,
            hash: crypto::hash(&user.password),
            email: user.email,
            roles: user.roles,
            image: user.image,
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Derivative)]
#[derivative(Debug)]
pub struct UserInfo {
    pub(crate) id: String,
    pub(crate) name: String,
    pub(crate) roles: Vec<Role>,
    #[derivative(Debug="ignore")]
    pub(crate) image: Option<String>,
}

impl From<UserWithHash> for UserInfo {
    fn from(uh: UserWithHash) -> Self {
        Self {
            id: uh.id,
            name: uh.name,
            roles: uh.roles,
            image: uh.image,
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Derivative)]
#[derivative(Debug)]
pub struct UserInfoFull {
    pub id: String,
    pub name: String,
    pub email: String,
    pub roles: Vec<Role>,
    #[derivative(Debug="ignore")]
    pub image: Option<String>,
}

impl From<UserWithHash> for UserInfoFull {
    fn from(uh: UserWithHash) -> Self {
        Self {
            id: uh.id,
            name: uh.name,
            email: uh.email,
            roles: uh.roles,
            image: uh.image,
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Derivative)]
#[derivative(Debug)]
pub struct User {
    pub id: String,
    pub name: String,
    #[derivative(Debug="ignore")]
    pub password: String,
    pub email: String,
    pub roles: Vec<Role>,
    #[derivative(Debug="ignore")]
    pub image: Option<String>,
}

// #[derive(Serialize, Deserialize)]
// pub struct UserWithPw {
//     pub name: String,
//     pub password: String,
// }

#[derive(Serialize, Deserialize, Derivative)]
#[derivative(Debug)]
pub struct LoginRequest {
    pub email: String,
    #[derivative(Debug="ignore")]
    pub password: String,
}

#[derive(Serialize, Deserialize, Derivative)]
#[derivative(Debug)]
pub struct UpdateRequestUser {
    pub name: Option<String>,
    pub email: Option<String>,
    #[derivative(Debug="ignore")]
    pub password: Option<String>,
    #[derivative(Debug="ignore")]
    pub image: Option<String>,
}

#[derive(Serialize, Deserialize, Derivative)]
#[derivative(Debug)]
pub struct UpdateRequestAdmin {
    pub name: Option<String>,
    pub email: Option<String>,
    #[derivative(Debug="ignore")]
    pub password: Option<String>,
    #[derivative(Debug="ignore")]
    pub image: Option<String>,
    pub roles: Option<Vec<Role>>,
}

#[derive(Serialize, Deserialize, Derivative)]
#[derivative(Debug)]
pub enum UpdateRequest {
    User(UpdateRequestUser),
    Admin(UpdateRequestAdmin),
}

impl From<Json<UpdateRequestUser>> for UpdateRequest {
    fn from(ur: Json<UpdateRequestUser>) -> Self {
        Self::User(ur.0)
    }
}

impl From<Json<UpdateRequestAdmin>> for UpdateRequest {
    fn from(ur: Json<UpdateRequestAdmin>) -> Self {
        Self::Admin(ur.0)
    }
}