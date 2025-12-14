use ratatui::{
    layout::{Constraint, Layout, Direction},
    widgets::{Block, Borders, Paragraph},
    Frame,
};

use crate::state::AppState;
use super::{Screen, Theme};

pub fn render(frame: &mut Frame, state: &AppState) {
    let theme = Theme::new();

    // Main layout: header, menu, content, status bar
    let chunks = Layout::default()
        .direction(Direction::Vertical)
        .constraints([
            Constraint::Length(1),  // Header
            Constraint::Length(1),  // Menu
            Constraint::Min(0),     // Main content
            Constraint::Length(1),  // Status bar
        ])
        .split(frame.area());

    // Render header
    let header = Paragraph::new(format!(
        "ARCHON - Web Server Manager                      [?] Help  [q] Quit"
    ))
    .style(theme.title_style());
    frame.render_widget(header, chunks[0]);

    // Render menu
    let menu = Paragraph::new(format!(
        " [1] Dashboard  [2] Sites  [3] Domains  [4] Nodes"
    ));
    frame.render_widget(menu, chunks[1]);

    // Render main content based on current screen
    render_main_content(frame, chunks[2], state);

    // Render status bar
    let status_text = if !state.notifications.is_empty() {
        if let Some(notification) = state.notifications.front() {
            notification.message.clone()
        } else {
            format!("Ready | Pending ops: {} | Sites: {} | Domains: {} | Nodes: {}",
                state.pending_operations.len(),
                state.sites.len(),
                state.domains.len(),
                state.nodes.len())
        }
    } else {
        format!("Ready | Pending ops: {} | Sites: {} | Domains: {} | Nodes: {}",
            state.pending_operations.len(),
            state.sites.len(),
            state.domains.len(),
            state.nodes.len())
    };

    let status = Paragraph::new(status_text);
    frame.render_widget(status, chunks[3]);
}

fn render_main_content(frame: &mut Frame, area: ratatui::layout::Rect, state: &AppState) {
    use super::screens;

    match &state.current_screen {
        Screen::Dashboard => screens::dashboard::render(frame, area, state),
        Screen::Help => screens::help::render(frame, area),

        // For screens not yet implemented, show placeholder
        _ => {
            let placeholder = Paragraph::new(format!(
                "Screen: {:?}\n\nThis screen is not yet implemented.\n\nPress Esc to go back or 1-4 to navigate.",
                state.current_screen
            ))
            .block(Block::default().borders(Borders::ALL).title("Coming Soon"));
            frame.render_widget(placeholder, area);
        }
    }
}
