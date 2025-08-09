use actix_web::{web, HttpResponse, Result};
use serde_json::json;

pub async fn test_endpoint() -> Result<HttpResponse> {
    Ok(HttpResponse::Ok().json(json!({
        "message": "Test endpoint is working!",
        "status": "success",
        "timestamp": chrono::Utc::now().to_rfc3339()
    })))
}

pub fn configure_test_routes(cfg: &mut web::ServiceConfig) {
    cfg.route("/test", web::get().to(test_endpoint));
}
