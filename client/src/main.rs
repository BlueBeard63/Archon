mod api;
mod config;
mod crypto;
mod dns;
mod logic;
mod models;
mod ui;

fn main() {
    tracing_subscriber::fmt::init();
    ui::app::run();
}
