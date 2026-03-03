use floem::prelude::*;
use floem::views::v_stack_from_iter;

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
        // Recent sites section
        "Recent Sites".style(|s| {
            s.font_size(FONT_SIZE_LG)
                .color(TEXT_PRIMARY)
                .font_bold()
                .margin_top(SPACING_XL)
                .margin_bottom(SPACING_MD)
        }),
        dyn_container(
            move || sites.get(),
            move |site_list| {
                if site_list.is_empty() {
                    return "No sites configured yet"
                        .style(|s| {
                            s.width_full()
                                .padding(SPACING_XL)
                                .font_size(FONT_SIZE_MD)
                                .color(TEXT_MUTED)
                                .items_center()
                                .justify_center()
                                .background(BG_SECONDARY)
                                .border(1.0)
                                .border_color(BORDER_DEFAULT)
                                .border_radius(BORDER_RADIUS)
                        })
                        .into_any();
                }

                let row_count = site_list.len().min(5);
                let rows: Vec<_> = site_list
                    .iter()
                    .take(5)
                    .enumerate()
                    .map(|(i, site)| {
                        let status_str = site.status.to_string();
                        let color = status_color(&status_str);
                        let is_last = i == row_count - 1;

                        h_stack((
                            // Status dot
                            empty().style(move |s| {
                                s.width(8.0)
                                    .height(8.0)
                                    .border_radius(4.0)
                                    .background(color)
                                    .flex_shrink(0.0)
                            }),
                            // Site name
                            site.name.clone().style(|s| {
                                s.font_size(FONT_SIZE_MD)
                                    .color(TEXT_PRIMARY)
                                    .flex_grow(1.0)
                                    .text_ellipsis()
                            }),
                            // Site type badge
                            site.site_type.to_string().style(|s| {
                                s.font_size(FONT_SIZE_SM)
                                    .color(TEXT_MUTED)
                                    .padding_horiz(SPACING_SM)
                                    .padding_vert(2.0)
                                    .background(BG_ELEVATED)
                                    .border_radius(BORDER_RADIUS_SM)
                                    .flex_shrink(0.0)
                            }),
                            // Status label
                            status_str.style(move |s| {
                                s.font_size(FONT_SIZE_SM)
                                    .color(color)
                                    .min_width(60.0)
                                    .justify_end()
                                    .flex_shrink(0.0)
                            }),
                        ))
                        .style(move |s| {
                            let s = s
                                .width_full()
                                .items_center()
                                .gap(SPACING_MD)
                                .padding_horiz(SPACING_LG)
                                .padding_vert(SPACING_SM + 2.0);
                            if is_last {
                                s
                            } else {
                                s.border_bottom(1.0).border_color(BORDER_MUTED)
                            }
                        })
                        .into_any()
                    })
                    .collect();

                v_stack_from_iter(rows)
                    .style(|s| {
                        s.width_full()
                            .background(BG_SECONDARY)
                            .border(1.0)
                            .border_color(BORDER_DEFAULT)
                            .border_radius(BORDER_RADIUS)
                    })
                    .into_any()
            },
        )
        .style(|s| s.width_full()),
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
