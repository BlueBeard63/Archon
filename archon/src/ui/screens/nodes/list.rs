use ratatui::{
    layout::{Constraint, Rect},
    widgets::{Block, Borders, Cell, Row, Table},
    Frame,
};

use crate::models::node::NodeStatus;
use crate::state::AppState;
use crate::ui::Theme;

pub fn render(frame: &mut Frame, area: Rect, state: &AppState) {
    let theme = Theme::new();

    if state.nodes.is_empty() {
        let block = Block::default()
            .borders(Borders::ALL)
            .title("Nodes - No nodes configured");

        frame.render_widget(block, area);
        return;
    }

    let selected_index = state.selection_state.nodes_list_index;

    let header = Row::new(vec![
        Cell::from("Name"),
        Cell::from("IP Address"),
        Cell::from("Status"),
        Cell::from("Containers"),
        Cell::from("API Endpoint"),
    ])
    .style(theme.title_style())
    .height(1);

    let rows: Vec<Row> = state
        .nodes
        .iter()
        .enumerate()
        .map(|(idx, node)| {
            let status_icon = match node.status {
                NodeStatus::Online => "🟢",
                NodeStatus::Offline => "🔴",
                NodeStatus::Degraded => "🟡",
                NodeStatus::Unknown => "⚪",
            };

            let containers = node
                .docker_info
                .as_ref()
                .map(|d| d.containers_running.to_string())
                .unwrap_or_else(|| "?".to_string());

            let row = Row::new(vec![
                Cell::from(node.name.clone()),
                Cell::from(node.ip_address.to_string()),
                Cell::from(format!("{} {}", status_icon, node.status)),
                Cell::from(containers),
                Cell::from(node.api_endpoint.clone()),
            ]);

            if idx == selected_index {
                row.style(theme.selected_style())
            } else {
                row
            }
        })
        .collect();

    let widths = [
        Constraint::Percentage(20),
        Constraint::Percentage(15),
        Constraint::Percentage(15),
        Constraint::Length(12),
        Constraint::Percentage(50),
    ];

    let table = Table::new(rows, widths)
        .header(header)
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title("Nodes - [c]reate [e]dit [d]elete [Enter]detail [h]ealth check"),
        )
        .highlight_style(theme.selected_style());

    frame.render_widget(table, area);
}
