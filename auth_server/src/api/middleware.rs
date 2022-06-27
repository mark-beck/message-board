// use actix_service::Transform;
// use actix_web::dev::{Response, Service, ServiceRequest, ServiceResponse};
// use actix_web::error::Error;
// use actix_web::http::{HeaderMap, StatusCode};
// use actix_web::web::Data;
// use futures::future::Either;
// use tracing::{trace, warn};
//
// use std::future::{ready, Ready};
//
// use std::rc::Rc;
// use std::sync::Arc;
//
// use crate::crypto::JwtIssuer;
// use crate::schema::Role;
//

//
// pub struct TokenValidator<S> {
//     accept_role: Role,
//     service: Rc<S>,
// }
//
// impl<S> Service<ServiceRequest> for TokenValidator<S>
// where
//     S: Service<ServiceRequest, Response = ServiceResponse, Error = Error> + 'static,
// {
//     type Response = ServiceResponse;
//     type Error = Error;
//     // type Future = LocalBoxFuture<'static, Result<Self::Response, Self::Error>>;
//     type Future = LocalBoxFuture<'static, Result<ServiceResponse<B>, Error>>;
//
//     actix_service::forward_ready!(service);
//
//     fn call(&self, req: ServiceRequest) -> Self::Future {
//
//         let service = Rc::clone(&self.service);
//
//         async move {
//             let (req, credentials) = match Extract::<T>::new(req).await {
//                 Ok(req) => req,
//                 Err((err, req)) => {
//                     return Ok(req.error_response(err));
//                 }
//             };
//
//             // TODO: alter to remove ? operator; an error response is required for downstream
//             // middleware to do their thing (eg. cors adding headers)
//             let req = process_fn(req, credentials).await?;
//             // Ensure `borrow_mut()` and `.await` are on separate lines or else a panic occurs.
//             let fut = service.borrow_mut().call(req);
//             fut.await
//         }
//             .boxed_local()
//     }
//         Box::pin(async move {
//         // We only need to hook into the `start` for this middleware.
//         trace!("checking token");
//         if let Some(jwt_issuer) = req.app_data::<Data<Arc<JwtIssuer>>>() {
//             if let Some(jwt) = get_jwt(req.headers()) {
//                 Box::pin(async move {
//                     if jwt_issuer.validate_level(jwt, self.accept_role).await {
//                         trace!("token ok");
//                         return Either::Left(self.service.call(req));
//                     }
//                 });
//                 trace!("token not ok");
//                 return Either::Right(ready(Ok(
//                     req.into_response(Response::new(StatusCode::FORBIDDEN))
//                 )));
//             }
//             trace!("token not found");
//             return Either::Right(ready(Ok(
//                 req.into_response(Response::new(StatusCode::FORBIDDEN))
//             )));
//         }
//         warn!("no JwtIssuer");
//         Either::Right(ready(Ok(
//             req.into_response(Response::new(StatusCode::FORBIDDEN))
//         )))
//     }
// }
//
// pub struct TokenValidatorFactory {
//     accept_role: Role,
// }
//
// impl TokenValidatorFactory {
//     pub fn new(accept_role: Role) -> Self {
//         Self { accept_role }
//     }
// }
//
// impl<S> Transform<S, ServiceRequest> for TokenValidatorFactory
// where
//     S: Service<ServiceRequest, Response = ServiceResponse, Error = Error> + 'static,
// {
//     type Response = ServiceResponse;
//     type Error = Error;
//     type Transform = TokenValidator<S>;
//     type InitError = ();
//     type Future = Ready<Result<Self::Transform, Self::InitError>>;
//
//     fn new_transform(&self, service: S) -> Self::Future {
//         ready(Ok(TokenValidator {
//             accept_role: self.accept_role,
//             service: Rc::new(service),
//         }))
//     }
// }

use std::sync::Arc;
use actix_web::dev::ServiceRequest;
use actix_web::{error, HttpMessage};
use actix_web::error::Error;
use actix_web::http::{StatusCode, header::HeaderMap};
use actix_web::web::Data;
use actix_web_httpauth::extractors::bearer::BearerAuth;
use crate::crypto::JwtIssuer;
use crate::mongo::Mongo;
use crate::schema::Role;

pub async fn validate_admin(req: ServiceRequest, credentials: BearerAuth) -> Result<ServiceRequest, Error> {
    validate(req, credentials, Role::Admin).await
}

pub async fn validate_user(req: ServiceRequest, credentials: BearerAuth) -> Result<ServiceRequest, Error> {
    validate(req, credentials, Role::User).await
}

pub async fn validate(req: ServiceRequest, credentials: BearerAuth, role: Role) -> Result<ServiceRequest, Error> {
    let jwt_validator = req.app_data::<Data<Arc<JwtIssuer>>>().unwrap();
    let mongo = req.app_data::<Data<Arc<Mongo>>>().unwrap();

    if jwt_validator.validate_level(mongo, credentials.token(), role).await {
        let claims = jwt_validator.decode(credentials.token()).await.unwrap();
        req.extensions_mut().insert(claims);
        Ok(req)
    } else {
        Err(Error::from(error::InternalError::new("", StatusCode::UNAUTHORIZED)))
    }
}
pub fn get_jwt(headers: &HeaderMap) -> Option<&str> {
    headers
        .get("Authorization")?
        .to_str()
        .ok()
        .filter(|header| header[..7].contains("Bearer"))
        .map(|header| &header[7..])
}
