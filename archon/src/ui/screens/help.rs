use ratatui::{
    widgets::{Block, Borders, List, ListItem, Paragraph},
    Frame,
};

use crate::ui::Theme;

pub fn render(frame: &mut Frame, area: ratatui::layout::Rect) {
    let theme = Theme::new();

    let help_items = vec![
        "ARCHON - Web Server Manager",
        "",
        "Global Keybindings:",
        "  q / Ctrl+C    - Quit application",
        "  ?             - Show this help screen",
        "  Esc           - Go back to previous screen",
        "",
        "Navigation:",
        "  1 / d         - Dashboard",
        "  2 / s         - Sites list",
        "  3             - Domains list",
        "  4 / n         - Nodes list",
        "",
        "List Navigation:",
        "  ↑ / k         - Move up",
        "  ↓ / j         - Move down",
        "  c             - Create new item",
        "  e             - Edit selected item",
        "  d             - Delete selected item",
        "  Enter         - View details",
        "",
        "Form Editing:",
        "  Tab           - Next field",
        "  Shift+Tab     - Previous field",
        "  Enter         - Submit form",
        "  Esc           - Cancel",
        "",
        "Site Management:",
        "  r             - Redeploy site",
        "  s             - Stop site",
        "  l             - Refresh logs",
        "",
        "Node Management:",
        "  h             - Run health check",
        "  r             - Refresh stats",
        "",
        "DNS Management:",
        "  a             - Add DNS record",
        "  s             - Sync with provider (if not manual)",
        "",
        "Press Esc to return to previous screen",
    ];

    let list_items: Vec<ListItem> = help_items
        .iter()
        .map(|item| ListItem::new(*item))
        .collect();

    let list = List::new(list_items)
        .block(
            Block::default()
                .borders(Borders::ALL)
                .title("Help - Keybindings")
                .style(theme.title_style()),
        );

    frame.render_widget(list, area);
}
