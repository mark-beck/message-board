FROM rust:1.56 as builder
WORKDIR /usr/src/auth_server
COPY . .
RUN cargo install --path .

FROM debian:buster-slim
RUN apt-get update && apt-get install -y libssl-dev
WORKDIR /server
COPY --from=builder /usr/local/cargo/bin/auth_server .

EXPOSE 8080

ENTRYPOINT ["/server/auth_server"]
