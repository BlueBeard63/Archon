use floem::prelude::*;

use crate::logic::{commands, AppState, Screen};
use crate::ui::components::button::primary_button;
use crate::ui::components::data_table::{empty_state, table_header};
use crate::ui::styles::*;

/// Sites list screen.
pub fn sites_list_screen(state: &'static AppState) -> impl IntoView {
    let sites = state.sites;
    let domains = state.domains;

    v_stack((
        // Header row
        h_stack((
            "Sites".style(|s| {
                s.font_size(FONT_SIZE_TITLE)
                    .color(TEXT_PRIMARY)
                    .font_bold()
            }),
            empty().style(|s| s.flex_grow(1.0)),
            primary_button("New Site", move || {
                state.navigation.navigate_to(Screen::SiteCreate);
            }),
        ))
        .style(|s| s.width_full().items_center().margin_bottom(SPACING_LG)),
        // Table header
        table_header(vec![
            ("Name", 200.0),
            ("Domain", 200.0),
            ("Image", 200.0),
            ("Status", 100.0),
        ]),
        // Sites list
        dyn_view(move || {
            let site_list = sites.get();
            let domain_list = domains.get();

            if site_list.is_empty() {
                return "No sites configured. Click 'New Site' to create one.".to_string();
            }

            site_list
                .iter()
                .map(|site| {
                    let domain_name = site
                        .domain_mappings
                        .first()
                        .and_then(|m| {
                            domain_list
                                .iter()
                                .find(|d| d.id == m.domain_id)
                                .map(|d| d.name.clone())
                        })
                        .or_else(|| {
                            domain_list
                                .iter()
                                .find(|d| d.id == site.domain_id)
                                .map(|d| d.name.clone())
                        })
                        .unwrap_or_else(|| "—".to_string());

                    format!(
                        "  {} | {} | {} | {}",
                        site.name, domain_name, site.docker_image, site.status
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
