use crate::crypto;
use chrono::{Duration, Utc};
use serde_derive::{Deserialize, Serialize};
use sodiumoxide::crypto::pwhash::argon2id13::HashedPassword;
use std::ops::Add;

#[allow(unused)]
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

#[derive(Serialize, Deserialize)]
pub struct Resp {
    pub(crate) status: Status,
    pub(crate) message: Option<String>,
    pub(crate) jwt: Option<String>,
}

#[derive(Serialize, Deserialize)]
pub struct TokenResponse {
    pub(crate) token: String,
    pub(crate) user: UserInfo,
}

#[derive(Serialize, Deserialize)]
pub struct RegisteringUser {
    pub(crate) name: String,
    pub(crate) password: String,
    pub(crate) email: String,
}

impl RegisteringUser {
    pub fn add_roles(self, roles: Vec<Role>) -> User {
        User {
            name: self.name,
            password: self.password,
            email: self.email,
            roles,
        }
    }
}

#[derive(Serialize, Deserialize, Clone)]
pub struct UserClaims {
    exp: i64,
    nbf: i64,
    sub: String,
    pub(crate) user: UserInfo,
}

impl From<UserWithHash> for UserClaims {
    fn from(uh: UserWithHash) -> Self {
        UserClaims {
            exp: Utc::now().add(Duration::hours(1)).timestamp(),
            nbf: Utc::now().timestamp(),
            sub: uh.name.clone(),
            user: uh.into(),
        }
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct UserWithHash {
    pub name: String,
    pub hash: HashedPassword,
    pub email: String,
    pub roles: Vec<Role>,
}

impl From<User> for UserWithHash {
    fn from(user: User) -> Self {
        Self {
            name: user.name,
            hash: crypto::hash(&user.password),
            email: user.email,
            roles: user.roles,
        }
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct UserInfo {
    pub name: String,
    pub email: String,
    pub roles: Vec<Role>,
}

impl From<UserWithHash> for UserInfo {
    fn from(uh: UserWithHash) -> Self {
        Self {
            name: uh.name,
            email: uh.email,
            roles: uh.roles,
        }
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct User {
    pub name: String,
    pub password: String,
    pub email: String,
    pub roles: Vec<Role>,
}

#[derive(Serialize, Deserialize)]
pub struct UserWithPw {
    pub name: String,
    pub password: String,
}
