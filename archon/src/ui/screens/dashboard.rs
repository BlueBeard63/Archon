use ratatui::{
    layout::{Constraint, Direction, Layout},
    widgets::{Block, Borders, List, ListItem, Paragraph},
    Frame,
};

use crate::state::AppState;
use crate::ui::Theme;

pub fn render(frame: &mut Frame, area: ratatui::layout::Rect, state: &AppState) {
    let theme = Theme::new();

    // Split into 3 columns
    let chunks = Layout::default()
        .direction(Direction::Horizontal)
        .constraints([
            Constraint::Percentage(33),
            Constraint::Percentage(33),
            Constraint::Percentage(34),
        ])
        .split(area);

    // Left column: Sites summary
    render_sites_summary(frame, chunks[0], state, &theme);

    // Middle column: Nodes summary
    render_nodes_summary(frame, chunks[1], state, &theme);

    // Right column: Domains summary
    render_domains_summary(frame, chunks[2], state, &theme);
}

fn render_sites_summary(
    frame: &mut Frame,
    area: ratatui::layout::Rect,
    state: &AppState,
    theme: &Theme,
) {
    let running = state
        .sites
        .iter()
        .filter(|s| matches!(s.status, crate::models::site::SiteStatus::Running))
        .count();
    let failed = state
        .sites
        .iter()
        .filter(|s| matches!(s.status, crate::models::site::SiteStatus::Failed))
        .count();

    let items = vec![
        format!("Total Sites: {}", state.sites.len()),
        format!("Running: {}", running),
        format!("Failed: {}", failed),
        String::new(),
        "Recent:".to_string(),
    ];

    let mut list_items: Vec<ListItem> = items.iter().map(|i| ListItem::new(i.as_str())).collect();

    // Add last 5 sites
    for site in state.sites.iter().rev().take(5) {
        list_items.push(ListItem::new(format!("  • {}", site.name)));
    }

    let list = List::new(list_items)
        .block(Block::default().borders(Borders::ALL).title("Sites"));

    frame.render_widget(list, area);
}

fn render_nodes_summary(
    frame: &mut Frame,
    area: ratatui::layout::Rect,
    state: &AppState,
    theme: &Theme,
) {
    let online = state
        .nodes
        .iter()
        .filter(|n| matches!(n.status, crate::models::node::NodeStatus::Online))
        .count();
    let offline = state
        .nodes
        .iter()
        .filter(|n| matches!(n.status, crate::models::node::NodeStatus::Offline))
        .count();

    let items = vec![
        format!("Total Nodes: {}", state.nodes.len()),
        format!("Online: {}", online),
        format!("Offline: {}", offline),
        String::new(),
        "Nodes:".to_string(),
    ];

    let mut list_items: Vec<ListItem> = items.iter().map(|i| ListItem::new(i.as_str())).collect();

    for node in state.nodes.iter().take(5) {
        let status_icon = match node.status {
            crate::models::node::NodeStatus::Online => "🟢",
            crate::models::node::NodeStatus::Offline => "🔴",
            crate::models::node::NodeStatus::Degraded => "🟡",
            _ => "⚪",
        };
        list_items.push(ListItem::new(format!("  {} {}", status_icon, node.name)));
    }

    let list = List::new(list_items)
        .block(Block::default().borders(Borders::ALL).title("Nodes"));

    frame.render_widget(list, area);
}

fn render_domains_summary(
    frame: &mut Frame,
    area: ratatui::layout::Rect,
    state: &AppState,
    theme: &Theme,
) {
    let manual_dns = state
        .domains
        .iter()
        .filter(|d| d.is_manual_dns())
        .count();

    let items = vec![
        format!("Total Domains: {}", state.domains.len()),
        format!("Manual DNS: {}", manual_dns),
        format!("API Managed: {}", state.domains.len() - manual_dns),
        String::new(),
        "Domains:".to_string(),
    ];

    let mut list_items: Vec<ListItem> = items.iter().map(|i| ListItem::new(i.as_str())).collect();

    for domain in state.domains.iter().take(5) {
        list_items.push(ListItem::new(format!(
            "  • {} ({})",
            domain.name,
            domain.provider_name()
        )));
    }

    let list = List::new(list_items)
        .block(Block::default().borders(Borders::ALL).title("Domains"));

    frame.render_widget(list, area);
}
