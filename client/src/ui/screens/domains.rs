use floem::prelude::*;

use crate::logic::{AppState, Screen};
use crate::ui::components::button::primary_button;
use crate::ui::components::data_table::table_header;
use crate::ui::styles::*;

/// Domains list screen.
pub fn domains_list_screen(state: &'static AppState) -> impl IntoView {
    let domains = state.domains;

    v_stack((
        // Header row
        h_stack((
            "Domains".style(|s| {
                s.font_size(FONT_SIZE_TITLE)
                    .color(TEXT_PRIMARY)
                    .font_bold()
            }),
            empty().style(|s| s.flex_grow(1.0)),
            primary_button("New Domain", move || {
                state.navigation.navigate_to(Screen::DomainCreate);
            }),
        ))
        .style(|s| s.width_full().items_center().margin_bottom(SPACING_LG)),
        // Table header
        table_header(vec![
            ("Name", 250.0),
            ("Provider", 150.0),
            ("Records", 100.0),
            ("Traefik", 100.0),
        ]),
        // Domains list
        dyn_view(move || {
            let domain_list = domains.get();

            if domain_list.is_empty() {
                return "No domains configured. Click 'New Domain' to add one.".to_string();
            }

            domain_list
                .iter()
                .map(|d| {
                    format!(
                        "  {} | {} | {} records | {}",
                        d.name,
                        d.provider_name(),
                        d.dns_records.len(),
                        if d.traefik_enabled { "Yes" } else { "No" }
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
