use ratatui::{
    layout::{Constraint, Rect},
    style::{Color, Modifier, Style},
    widgets::{Block, Borders, Cell, Row, Table},
    Frame,
};

use crate::models::site::SiteStatus;
use crate::state::AppState;
use crate::ui::Theme;

pub fn render(frame: &mut Frame, area: Rect, state: &AppState) {
    let theme = Theme::new();

    if state.sites.is_empty() {
        let block = Block::default()
            .borders(Borders::ALL)
            .title("Sites - No sites configured");

        frame.render_widget(block, area);
        return;
    }

    // Prepare table rows
    let selected_index = state.selection_state.sites_list_index;

    let header = Row::new(vec![
        Cell::from("Name"),
        Cell::from("Domain"),
        Cell::from("Node"),
        Cell::from("Status"),
        Cell::from("SSL"),
        Cell::from("Port"),
    ])
    .style(theme.title_style())
    .height(1);

    let rows: Vec<Row> = state
        .sites
        .iter()
        .enumerate()
        .map(|(idx, site)| {
            let domain_name = state
                .get_domain(site.domain_id)
                .map(|d| d.name.as_str())
                .unwrap_or("Unknown");

            let node_name = state
                .get_node(site.node_id)
                .map(|n| n.name.as_str())
                .unwrap_or("Unknown");

            let status_icon = match site.status {
                SiteStatus::Running => "🟢",
                SiteStatus::Deploying => "🟡",
                SiteStatus::Failed => "❌",
                SiteStatus::Stopped => "🔴",
                SiteStatus::Inactive => "⚪",
            };

            let ssl_icon = if site.ssl_enabled { "✓" } else { "✗" };

            let row = Row::new(vec![
                Cell::from(site.name.clone()),
                Cell::from(domain_name.to_string()),
                Cell::from(node_name.to_string()),
                Cell::from(format!("{} {}", status_icon, site.status)),
                Cell::from(ssl_icon),
                Cell::from(site.port.to_string()),
            ]);

            if idx == selected_index {
                row.style(theme.selected_style())
            } else {
                row
            }
        })
        .collect();

    let widths = [
        Constraint::Percentage(25),
        Constraint::Percentage(20),
        Constraint::Percentage(20),
        Constraint::Percentage(15),
        Constraint::Length(5),
        Constraint::Length(8),
    ];

    let table = Table::new(rows, widths)
        .header(header)
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title("Sites - [c]reate [e]dit [d]elete [Enter]detail [r]edeploy"),
        )
        .highlight_style(theme.selected_style());

    frame.render_widget(table, area);
}
