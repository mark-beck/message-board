use crate::config::Config;
use crate::crypto::JwtIssuer;
use crate::mail::Mailer;
use crate::api::middleware;
use actix_web::web::Data;
use actix_web::{web, App, HttpServer};
use std::sync::Arc;
use tracing::{info, Level};
use tracing_actix_web::TracingLogger;
use tracing_subscriber::FmtSubscriber;
use actix_web_httpauth::middleware::HttpAuthentication;

mod api;
mod config;
mod crypto;
mod mail;
mod mongo;
mod schema;

#[actix_web::main]
async fn main() -> anyhow::Result<()> {
    let subscriber = FmtSubscriber::builder()
        .with_max_level(Level::TRACE)
        .finish();
    tracing::subscriber::set_global_default(subscriber).expect("setting default subscriber failed");

    let config = Config::parse("./auth_config.toml").expect("parsing failed");

    let mongo = Arc::new(mongo::Mongo::from_config(config.clone()).await?);

    let jwt_issuer = Arc::new(JwtIssuer::new(config.clone()).await?);

    let mailer = Arc::new(Mailer::with_config(config));

    info!("starting HTTP server");

    HttpServer::new(move || {
        App::new()
            .app_data(Data::new(mongo.clone()))
            .app_data(Data::new(jwt_issuer.clone()))
            .app_data(Data::new(mailer.clone()))
            .wrap(TracingLogger::default())
            .wrap(actix_web::middleware::NormalizePath::default())
            .service(
                web::scope("/auth")
                    .route("/signin", web::post().to(api::Auth::sign_in))
                    .route("/signup", web::post().to(api::Auth::sign_up))
                    .route("/reissue", web::post().to(api::Auth::reissue)),
            )
            .service(
                web::scope("/admin")
                    .wrap(HttpAuthentication::bearer(middleware::validate_admin))
                    .route("/create_user", web::post().to(api::Admin::create_user))
                    .route("/list_users", web::get().to(api::Admin::list_users))
                    .route(
                        "/delete_user/{name}",
                        web::delete().to(api::Admin::delete_user),
                    ),
            )
            .route("/version", web::get().to(api::version))
            .route("/mail/{address}/{subject}/{body}", web::get().to(api::mail))
    })
    .bind(("0.0.0.0", 8080))?
    .run()
    .await
    .map_err(Into::into)
}
