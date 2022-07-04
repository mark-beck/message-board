use crate::config::Config;
use crate::crypto::JwtIssuer;
use crate::api::middleware;
use crate::image_service::ImageService;
use actix_web::web::Data;
use actix_web::{web, App, HttpServer};
use tracing_subscriber::util::SubscriberInitExt;
use std::sync::Arc;
use tracing::info;
use actix_web_httpauth::middleware::HttpAuthentication;
use actix_cors::Cors;
use tracing_subscriber::layer::SubscriberExt;
use opentelemetry::global;
use actix_web_opentelemetry::RequestTracing;
use tracing_actix_web::TracingLogger;

mod api;
mod config;
mod crypto;
mod mongo;
mod schema;
mod image_service;

#[actix_web::main]
async fn main() -> anyhow::Result<()> {

    global::set_text_map_propagator(opentelemetry_jaeger::Propagator::new());
    let tracer = opentelemetry_jaeger::new_pipeline()
        .with_service_name("auth-service")
        .with_agent_endpoint("tracing:6831")
        .install_simple()?;

    let opentelemetry = tracing_opentelemetry::layer().with_tracer(tracer);
    tracing_subscriber::registry()
        .with(opentelemetry)
        .try_init()?;

    let config = Config::from_env()?;

    let mongo = Arc::new(mongo::Mongo::from_config(config.clone()).await?);

    let jwt_issuer = Arc::new(JwtIssuer::new(config.clone()).await?);

    let image_service = ImageService::new(config.image_service.clone());

    info!("starting auth_service on port 8080");

    HttpServer::new(move || {
        App::new()
            .app_data(Data::new(mongo.clone()))
            .app_data(Data::new(jwt_issuer.clone()))
            .app_data(Data::new(image_service.clone()))
            .wrap(RequestTracing::new())
            .wrap(actix_web::middleware::NormalizePath::default())
            .wrap(Cors::permissive())
            .service(
                web::scope("/auth")
                    .route("/signin", web::post().to(api::Auth::sign_in))
                    .route("/signup", web::post().to(api::Auth::sign_up))
                    .route("/reissue", web::post().to(api::Auth::reissue))
                    .service(
                        web::scope("/admin")
                            .wrap(HttpAuthentication::bearer(middleware::validate_admin))
                            .route("/create_user", web::post().to(api::AdminApi::create_user))
                            .route("/list_users", web::get().to(api::AdminApi::list_users))
                            .route("/update_user/{id}", web::post().to(api::AdminApi::update_user))
                            .route(
                                "/delete_user/{id}",
                                web::delete().to(api::AdminApi::delete_user),
                            ),
                    )
                    .service(
                        web::scope("/user")
                            .wrap(HttpAuthentication::bearer(middleware::validate_user))
                            .route("/info", web::get().to(api::UserApi::info))
                            .route("/update", web::post().to(api::UserApi::update))
                            .route("/delete", web::delete().to(api::UserApi::delete))
                            .route("/get_batch", web::post().to(api::UserApi::get_batch))
                            .route("/{id}", web::get().to(api::UserApi::get))
                            .route("/email/{email}", web::get().to(api::UserApi::get_email)),
                    )
                    .route("/version", web::get().to(api::version))
            )
            
    })
    .bind(("0.0.0.0", 8080))?
    .run()
    .await
    .map_err(anyhow::Error::from)?;

    global::shutdown_tracer_provider();
    Ok(())
}
