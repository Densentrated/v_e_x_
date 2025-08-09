use actix_web::{middleware::Logger, App, HttpServer};

mod routes;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();

    println!("Starting Actix Web server on http://127.0.0.1:8080");

    HttpServer::new(|| {
        App::new()
            .wrap(Logger::default())
            .configure(routes::configure_test_routes)
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}
