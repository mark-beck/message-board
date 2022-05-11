use crate::crypto::JwtIssuer;
use crate::mongo::Mongo;
use crate::schema::{RegisteringUser, Role, TokenResponse, User, UserInfo, UserWithPw};
use actix_web::error::{Error, Result};
use actix_web::http::StatusCode;
use actix_web::web::{Data, Json};
use actix_web::{error, web, HttpRequest, HttpResponse, Responder};
use std::sync::Arc;

use crate::api::middleware::get_jwt;
use crate::mail::Mailer;
use tracing::{info, trace, warn};

pub mod middleware;

pub async fn version() -> impl Responder {
    trace!("version served");
    option_env!("CARGO_PKG_VERSION")
}

pub async fn mail(mailer: Data<Arc<Mailer>>, req: HttpRequest) -> Result<impl Responder> {
    trace!("mail handler");
    let address = req
        .match_info()
        .get("address")
        .http_result(StatusCode::BAD_REQUEST)?;
    let subject = req
        .match_info()
        .get("subject")
        .http_result(StatusCode::BAD_REQUEST)?;
    let body = req
        .match_info()
        .get("body")
        .http_result(StatusCode::BAD_REQUEST)?;
    Ok(mailer
        .send_mail(address, subject, body).await
        .http_result(StatusCode::INTERNAL_SERVER_ERROR)?
        .message()
        .collect::<String>())
}

pub struct Auth;

impl Auth {
    pub(crate) async fn sign_in(
        mongo: Data<Arc<Mongo>>,
        jwt_issuer: Data<Arc<JwtIssuer>>,
        user: web::Json<UserWithPw>,
    ) -> Result<Json<TokenResponse>> {
        trace!("signin");
        if mongo.verify_user(&user.0).await {
            let user_hashed = mongo
                .get_user_from_name(&user.name)
                .await
                .http_result(StatusCode::UNAUTHORIZED)?;

            let jwt = jwt_issuer
                .issue(user_hashed.clone())
                .http_log_result("jwt error", StatusCode::INTERNAL_SERVER_ERROR)?;
            info!("giving out JWT to {}", user_hashed.name);
            Ok(Json(TokenResponse {
                token: jwt,
                user: user_hashed.into(),
            }))
        } else {
            Err(error::InternalError::new("", StatusCode::UNAUTHORIZED).into())
        }
    }

    pub async fn sign_up(
        mongo: Data<Arc<Mongo>>,
        user: web::Json<RegisteringUser>,
    ) -> Result<impl Responder> {
        trace!("register");
        let name = user.name.clone();
        info!("registering user {}", name);
        let user = user.0.add_roles(vec![Role::User]);
        mongo
            .create_user(user.into())
            .await
            .http_result(StatusCode::INTERNAL_SERVER_ERROR)?;
        Ok(HttpResponse::Created())
    }

    pub async fn reissue(
        jwt_issuer: Data<Arc<JwtIssuer>>,
        req: HttpRequest,
    ) -> Result<Json<TokenResponse>> {
        let old_jwt = get_jwt(req.headers()).http_result(StatusCode::BAD_REQUEST)?;
        let claims = jwt_issuer
            .decode(old_jwt).await
            .http_result(StatusCode::BAD_REQUEST)?;
        let new_jwt = jwt_issuer
            .issue(claims.clone())
            .http_result(StatusCode::BAD_REQUEST)?;

        Ok(Json(TokenResponse {
            token: new_jwt,
            user: claims.user,
        }))
    }
}

pub struct Admin;

impl Admin {
    pub async fn list_users(mongo: Data<Arc<Mongo>>) -> Result<Json<Vec<UserInfo>>> {
        trace!("list_users");
        return mongo
            .get_all_users()
            .await
            .map(|v| Json(v.into_iter().map(UserInfo::from).collect()))
            .http_result(StatusCode::INTERNAL_SERVER_ERROR);
    }

    pub async fn create_user(
        mongo: Data<Arc<Mongo>>,
        user: web::Json<User>,
    ) -> Result<impl Responder> {
        trace!("create_user");

        info!("creating user {}", &user.name);
        mongo
            .create_user(user.0.into())
            .await
            .http_result(StatusCode::CONFLICT)?;
        Ok(HttpResponse::Created())
    }

    pub async fn delete_user(mongo: Data<Arc<Mongo>>, req: HttpRequest) -> Result<impl Responder> {
        let name = req
            .match_info()
            .get("name")
            .http_result(StatusCode::BAD_REQUEST)?;
        mongo
            .delete_user(name)
            .await
            .http_result(StatusCode::BAD_REQUEST)?;
        Ok(HttpResponse::Ok())
    }
}

pub trait IntoHttpError<T> {
    fn http_result(self, status_code: StatusCode) -> core::result::Result<T, actix_web::Error>;
    fn http_log_result(
        self,
        message: &str,
        status_code: StatusCode,
    ) -> core::result::Result<T, actix_web::Error>;
}

impl<T> IntoHttpError<T> for Option<T> {
    fn http_result(self, status_code: StatusCode) -> std::prelude::rust_2015::Result<T, Error> {
        if let Some(val) = self {
            Ok(val)
        } else {
            warn!("http_error of Option");
            Err(error::InternalError::new("", status_code).into())
        }
        // match self {
        //     Some(val) => Ok(val),
        //     None => {
        //         warn!("http_error of Option");
        //         Err(error::InternalError::new("", status_code).into())
        //     }
        // }
    }
    fn http_log_result(
        self,
        message: &str,
        status_code: StatusCode,
    ) -> std::prelude::rust_2015::Result<T, Error> {
        if let Some(val) = self {
            Ok(val)
        } else {
            warn!("http_error of Option, message: {}", message);
            Err(error::InternalError::new("", status_code).into())
        }
        // match self {
        //     Some(val) => Ok(val),
        //     None => {
        //         warn!("http_error of Option, message: {}", message);
        //         Err(error::InternalError::new("", status_code).into())
        //     }
        // }
    }
}

impl<T> IntoHttpError<T> for anyhow::Result<T> {
    fn http_result(self, status_code: StatusCode) -> core::result::Result<T, actix_web::Error> {
        match self {
            Ok(val) => Ok(val),
            Err(err) => {
                warn!("http_error: {:?}", err);
                Err(error::InternalError::new(err.to_string(), status_code).into())
            }
        }
    }

    fn http_log_result(
        self,
        message: &str,
        status_code: StatusCode,
    ) -> core::result::Result<T, actix_web::Error> {
        match self {
            Ok(val) => Ok(val),
            Err(err) => {
                warn!("http_error: {:?}, message {}", err, message);
                Err(error::InternalError::new(err.to_string(), status_code).into())
            }
        }
    }
}
