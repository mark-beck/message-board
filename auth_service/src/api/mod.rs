use crate::crypto::JwtIssuer;
use crate::image_service::ImageService;
use crate::mongo::Mongo;
use crate::schema::{
    LoginRequest, RegisteringUser, Role, TokenResponse, UpdateRequestAdmin, UpdateRequestUser,
    UserClaims, UserInfo, UserInfoFull,
};
use actix_web::error::{Error, Result};
use actix_web::http::StatusCode;
use actix_web::web::{Data, Json, ReqData};
use actix_web::{error, web, HttpRequest, HttpResponse, Responder};
use std::sync::Arc;

use crate::api::middleware::get_jwt;
use tracing::{debug, info, trace, warn};

pub mod middleware;

pub async fn version() -> impl Responder {
    trace!("version served");
    option_env!("CARGO_PKG_VERSION")
}

pub struct Auth;

impl Auth {
    #[tracing::instrument(level = "trace", skip(mongo, jwt_issuer))]
    pub(crate) async fn sign_in(
        mongo: Data<Arc<Mongo>>,
        jwt_issuer: Data<Arc<JwtIssuer>>,
        request: web::Json<LoginRequest>,
    ) -> Result<Json<TokenResponse>> {
        if mongo.verify_user(&request).await {
            let user_hashed = mongo
                .get_user_from_email(&request.email)
                .await
                .http_log_result("error finding user", StatusCode::UNAUTHORIZED)?;

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

    #[tracing::instrument(level = "trace", skip(mongo))]
    pub async fn sign_up(
        mongo: Data<Arc<Mongo>>,
        user_request: web::Json<RegisteringUser>,
        image_service: Data<ImageService>,
    ) -> Result<impl Responder> {
        trace!("register");
        let name = user_request.name.clone();
        info!("registering user {}", name);
        let user = user_request
            .0
            .into_user(&image_service, vec![Role::User])
            .await
            .http_result(StatusCode::INTERNAL_SERVER_ERROR)?;
        mongo
            .create_user(user.into())
            .await
            .http_result(StatusCode::INTERNAL_SERVER_ERROR)?;
        Ok(HttpResponse::Created())
    }

    #[tracing::instrument(level = "trace", skip(mongo, jwt_issuer))]
    pub async fn reissue(
        jwt_issuer: Data<Arc<JwtIssuer>>,
        mongo: Data<Arc<Mongo>>,
        req: HttpRequest,
    ) -> Result<Json<TokenResponse>> {
        let old_jwt = get_jwt(req.headers()).http_result(StatusCode::BAD_REQUEST)?;
        let claims = jwt_issuer
            .decode(old_jwt)
            .await
            .http_result(StatusCode::BAD_REQUEST)?;
        let new_jwt = jwt_issuer
            .issue(claims.clone())
            .http_result(StatusCode::BAD_REQUEST)?;

        let user = mongo
            .get_user_from_id(&claims.user_id)
            .await
            .http_result(StatusCode::BAD_REQUEST)?;

        Ok(Json(TokenResponse {
            token: new_jwt,
            user: user.into(),
        }))
    }
}

pub struct AdminApi;

impl AdminApi {
    #[tracing::instrument(level = "trace", skip(mongo))]
    pub async fn list_users(mongo: Data<Arc<Mongo>>) -> Result<Json<Vec<UserInfoFull>>> {
        trace!("list_users");
        return mongo
            .get_all_users()
            .await
            .map(|v| Json(v.into_iter().map(UserInfoFull::from).collect()))
            .http_result(StatusCode::INTERNAL_SERVER_ERROR);
    }

    #[tracing::instrument(level = "trace", skip(mongo))]
    pub async fn create_user(
        mongo: Data<Arc<Mongo>>,
        user: web::Json<RegisteringUser>,
        image_service: Data<ImageService>,
    ) -> Result<impl Responder> {
        info!("creating user {}", &user.0.name);
        mongo
            .create_user(
                user.0
                    .into_user(&image_service, vec![Role::User])
                    .await
                    .http_result(StatusCode::INTERNAL_SERVER_ERROR)?
                    .into(),
            )
            .await
            .http_result(StatusCode::CONFLICT)?;
        Ok(HttpResponse::Created())
    }

    #[tracing::instrument(level = "trace", skip(mongo, req))]
    pub async fn update_user(
        mongo: Data<Arc<Mongo>>,
        req: HttpRequest,
        update_request: Json<UpdateRequestAdmin>,
    ) -> Result<impl Responder> {
        let id = req
            .match_info()
            .get("id")
            .http_result(StatusCode::BAD_REQUEST)?;

        info!("updating user {}", id);
        mongo
            .update_user(id, &update_request.into())
            .await
            .http_result(StatusCode::INTERNAL_SERVER_ERROR)?;
        Ok(HttpResponse::Ok())
    }

    #[tracing::instrument(level = "trace", skip(mongo))]
    pub async fn delete_user(mongo: Data<Arc<Mongo>>, req: HttpRequest) -> Result<impl Responder> {
        let id = req
            .match_info()
            .get("id")
            .http_result(StatusCode::BAD_REQUEST)?;
        mongo
            .delete_user(id)
            .await
            .http_result(StatusCode::BAD_REQUEST)?;
        Ok(HttpResponse::Ok())
    }
}

pub struct UserApi;

impl UserApi {
    #[tracing::instrument(level = "trace", skip(mongo, claims))]
    pub async fn info(
        mongo: Data<Arc<Mongo>>,
        claims: ReqData<UserClaims>,
        req: HttpRequest,
    ) -> Result<Json<UserInfo>> {
        info!("user_id: {:?}", claims.user_id);

        info!("headers: {:?}", req.headers());

        let user = mongo
            .get_user_from_id(&claims.user_id)
            .await
            .http_result(StatusCode::BAD_REQUEST)?;

        Ok(Json(user.into()))
    }

    #[tracing::instrument(level = "trace", skip(mongo))]
    pub async fn get(mongo: Data<Arc<Mongo>>, req: HttpRequest) -> Result<Json<UserInfo>> {
        let id = req
            .match_info()
            .get("id")
            .http_result(StatusCode::BAD_REQUEST)?;
        let user = mongo
            .get_user_from_id(id)
            .await
            .http_result(StatusCode::BAD_REQUEST)?;
        Ok(Json(user.into()))
    }

    #[tracing::instrument(level = "trace", skip(mongo))]
    pub async fn get_email(mongo: Data<Arc<Mongo>>, req: HttpRequest) -> Result<Json<UserInfo>> {
        let email = req
            .match_info()
            .get("email")
            .http_result(StatusCode::BAD_REQUEST)?;
        let user = mongo
            .get_user_from_email(email)
            .await
            .http_result(StatusCode::BAD_REQUEST)?;
        Ok(Json(user.into()))
    }

    #[tracing::instrument(level = "trace", skip(mongo))]
    pub async fn get_batch(
        mongo: Data<Arc<Mongo>>,
        batch: web::Json<Vec<String>>,
    ) -> Result<Json<Vec<UserInfo>>> {
        let mut infos = Vec::new();
        for id in batch.0 {
            let user = mongo.get_user_from_id(&id).await;
            match user {
                Ok(user) => infos.push(user.into()),
                Err(e) => {
                    warn!("{}", e);
                }
            }
        }
        debug!("{:?}", infos);
        Ok(Json(infos))
    }

    #[tracing::instrument(level = "trace", skip(mongo))]
    pub async fn update(
        mongo: Data<Arc<Mongo>>,
        claims: ReqData<UserClaims>,
        update: web::Json<UpdateRequestUser>,
    ) -> Result<impl Responder> {
        info!("updating user {}", claims.user_id);
        mongo
            .update_user(&claims.user_id, &update.into())
            .await
            .http_result(StatusCode::BAD_REQUEST)?;
        Ok(HttpResponse::Ok())
    }

    #[tracing::instrument(level = "trace", skip(mongo))]
    pub async fn delete(
        mongo: Data<Arc<Mongo>>,
        claims: ReqData<UserClaims>,
    ) -> Result<impl Responder> {
        info!("deleting user {}", claims.user_id);
        mongo
            .delete_user(&claims.user_id)
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
        self.map_or_else(
            || {
                warn!("http_error of option");
                Err(error::InternalError::new("", status_code).into())
            },
            |val| Ok(val),
        )
    }
    fn http_log_result(
        self,
        message: &str,
        status_code: StatusCode,
    ) -> std::prelude::rust_2015::Result<T, Error> {
        self.map_or_else(
            || {
                warn!("http_error of Option, message: {}", message);
                Err(error::InternalError::new("", status_code).into())
            },
            |val| Ok(val),
        )
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
