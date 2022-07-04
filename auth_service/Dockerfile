FROM rust:1.61 as builder
WORKDIR /auth_service
COPY . .
RUN --mount=type=cache,target=/auth_service/target \
    --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/local/cargo/git \
    --mount=type=cache,target=/usr/local/rustup \
    set -eux; \
    rustup install stable; \
    cargo build --release; \
    objcopy --compress-debug-sections target/release/auth_service ./auth_service

FROM debian:stable

RUN set -eux; \
    export DEBIAN_FRONTEND=noninteractive; \
    apt update; \
    apt install --yes --no-install-recommends libssl-dev; \
    apt clean autoclean; \
    apt autoremove --yes; \
    rm -rf /var/lib/{apt,dpkg,cache,log}/; \
    echo "Installed base utils"

WORKDIR /server
COPY --from=builder /auth_service .

ENTRYPOINT ["/server/auth_service"]
