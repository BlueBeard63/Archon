use anyhow::Result;
use directories::ProjectDirs;

mod api;
mod app;
mod config;
mod dns;
mod events;
mod models;
mod state;
mod ui;

use app::App;

#[tokio::main]
async fn main() -> Result<()> {
    // Install color-eyre for better error messages
    color_eyre::install().ok(); // Ignore error if already installed

    // Determine config path
    let config_path = if let Some(proj_dirs) = ProjectDirs::from("com", "archon", "archon") {
        proj_dirs.config_dir().join("config.toml")
    } else {
        std::env::current_dir()?.join("archon.toml")
    };

    // Create application
    let mut app = App::new(config_path)?;

    // Initialize terminal
    let terminal = ratatui::init();

    // Run application
    let result = app.run(terminal).await;

    // Restore terminal
    ratatui::restore();

    result
}