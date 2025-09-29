# Use cargo-chef for proper Rust dependency caching
FROM docker.io/lukemathwalker/cargo-chef:latest AS chef
WORKDIR /app

# Prepare recipe for dependency caching
FROM chef AS planner
COPY . .
RUN cargo chef prepare --recipe-path recipe.json

# Development stage with cargo-chef caching
FROM chef AS development
COPY --from=planner /app/recipe.json recipe.json
RUN cargo chef cook --recipe-path recipe.json
RUN cargo install cargo-watch
COPY . .
EXPOSE 8080
CMD ["cargo", "watch", "-x", "run"]

# Production stage with cargo-chef caching
FROM chef AS production
COPY --from=planner /app/recipe.json recipe.json
RUN cargo chef cook --release --recipe-path recipe.json
COPY . .
RUN cargo build --release
EXPOSE 8080
CMD ["./target/release/backend"]
