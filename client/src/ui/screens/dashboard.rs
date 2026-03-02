use floem::prelude::*;

use crate::logic::AppState;
use crate::ui::styles::*;

/// Dashboard screen with summary cards for sites, nodes, and domains.
pub fn dashboard_screen(state: &'static AppState) -> impl IntoView {
    let sites = state.sites;
    let nodes = state.nodes;
    let domains = state.domains;

    v_stack((
        // Title
        "Dashboard".style(|s| {
            s.font_size(FONT_SIZE_TITLE)
                .color(TEXT_PRIMARY)
                .font_bold()
                .margin_bottom(SPACING_XL)
        }),
        // Summary cards row
        h_stack((
            summary_card("Sites", move || {
                let all = sites.get();
                let running = all.iter().filter(|s| s.status.to_string() == "running").count();
                format!("{} total, {} running", all.len(), running)
            }),
            summary_card("Nodes", move || {
                let all = nodes.get();
                let online = all.iter().filter(|n| n.status.to_string() == "online").count();
                format!("{} total, {} online", all.len(), online)
            }),
            summary_card("Domains", move || {
                let all = domains.get();
                format!("{} configured", all.len())
            }),
        ))
        .style(|s| s.width_full().gap(SPACING_LG)),
        // Sites overview list
        "Recent Sites"
            .style(|s| {
                s.font_size(FONT_SIZE_LG)
                    .color(TEXT_PRIMARY)
                    .font_bold()
                    .margin_top(SPACING_XL)
                    .margin_bottom(SPACING_MD)
            }),
        dyn_view(move || {
            let site_list = sites.get();
            if site_list.is_empty() {
                return "No sites configured yet".to_string();
            }
            site_list
                .iter()
                .take(5)
                .map(|s| format!("  {} - {}", s.name, s.status))
                .collect::<Vec<_>>()
                .join("\n")
        })
        .style(|s| {
            s.width_full()
                .padding(SPACING_MD)
                .background(BG_SECONDARY)
                .border_radius(BORDER_RADIUS)
                .color(TEXT_SECONDARY)
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

fn summary_card(
    title: &str,
    value_fn: impl Fn() -> String + 'static,
) -> impl IntoView {
    let title = title.to_string();
    v_stack((
        title.style(|s| {
            s.font_size(FONT_SIZE_SM)
                .color(TEXT_MUTED)
                .margin_bottom(SPACING_XS)
        }),
        dyn_view(value_fn).style(|s| {
            s.font_size(FONT_SIZE_LG)
                .color(TEXT_PRIMARY)
                .font_bold()
        }),
    ))
    .style(|s| {
        s.flex_grow(1.0)
            .padding(SPACING_LG)
            .background(BG_SECONDARY)
            .border(1.0)
            .border_color(BORDER_DEFAULT)
            .border_radius(BORDER_RADIUS)
    })
}
