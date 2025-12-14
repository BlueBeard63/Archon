use ratatui::{
    layout::{Constraint, Rect},
    widgets::{Block, Borders, Cell, Row, Table},
    Frame,
};

use crate::state::AppState;
use crate::ui::Theme;

pub fn render(frame: &mut Frame, area: Rect, state: &AppState) {
    let theme = Theme::new();

    if state.domains.is_empty() {
        let block = Block::default()
            .borders(Borders::ALL)
            .title("Domains - No domains configured");

        frame.render_widget(block, area);
        return;
    }

    let selected_index = state.selection_state.domains_list_index;

    let header = Row::new(vec![
        Cell::from("Domain Name"),
        Cell::from("DNS Provider"),
        Cell::from("Records"),
        Cell::from("Sites"),
    ])
    .style(theme.title_style())
    .height(1);

    let rows: Vec<Row> = state
        .domains
        .iter()
        .enumerate()
        .map(|(idx, domain)| {
            let sites_count = state
                .sites
                .iter()
                .filter(|s| s.domain_id == domain.id)
                .count();

            let provider_display = if domain.is_manual_dns() {
                "⚠️  Manual"
            } else {
                domain.provider_name()
            };

            let row = Row::new(vec![
                Cell::from(domain.name.clone()),
                Cell::from(provider_display),
                Cell::from(domain.dns_records.len().to_string()),
                Cell::from(sites_count.to_string()),
            ]);

            if idx == selected_index {
                row.style(theme.selected_style())
            } else {
                row
            }
        })
        .collect();

    let widths = [
        Constraint::Percentage(40),
        Constraint::Percentage(25),
        Constraint::Percentage(15),
        Constraint::Percentage(20),
    ];

    let table = Table::new(rows, widths)
        .header(header)
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title("Domains - [c]reate [e]dit DNS [d]elete [s]ync"),
        )
        .highlight_style(theme.selected_style());

    frame.render_widget(table, area);
}
