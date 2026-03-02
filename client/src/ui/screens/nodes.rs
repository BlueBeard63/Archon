use floem::prelude::*;

use crate::logic::{commands, AppState, Screen};
use crate::ui::components::button::{primary_button, secondary_button};
use crate::ui::components::data_table::table_header;
use crate::ui::styles::*;

/// Nodes list screen.
pub fn nodes_list_screen(state: &'static AppState) -> impl IntoView {
    let nodes = state.nodes;

    v_stack((
        // Header row
        h_stack((
            "Nodes".style(|s| {
                s.font_size(FONT_SIZE_TITLE)
                    .color(TEXT_PRIMARY)
                    .font_bold()
            }),
            empty().style(|s| s.flex_grow(1.0)),
            secondary_button("Refresh All", move || {
                commands::health_check_all(state);
            }),
            primary_button("New Node", move || {
                state.navigation.navigate_to(Screen::NodeCreate);
            }),
        ))
        .style(|s| {
            s.width_full()
                .items_center()
                .gap(SPACING_SM)
                .margin_bottom(SPACING_LG)
        }),
        // Table header
        table_header(vec![
            ("Name", 180.0),
            ("Endpoint", 250.0),
            ("Proxy", 100.0),
            ("Status", 100.0),
        ]),
        // Nodes list
        dyn_view(move || {
            let node_list = nodes.get();

            if node_list.is_empty() {
                return "No nodes configured. Click 'New Node' to add one.".to_string();
            }

            node_list
                .iter()
                .map(|n| {
                    format!(
                        "  {} | {} | {} | {}",
                        n.name, n.api_endpoint, n.proxy_type, n.status
                    )
                })
                .collect::<Vec<_>>()
                .join("\n")
        })
        .style(|s| {
            s.width_full()
                .flex_grow(1.0)
                .padding(SPACING_SM)
                .color(TEXT_PRIMARY)
                .font_size(FONT_SIZE_MD)
        }),
    ))
    .style(|s| {
        s.width_full()
            .height_full()
            .padding(SPACING_XL)
            .flex_col()
    })
}
