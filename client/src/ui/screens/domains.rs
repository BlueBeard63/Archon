use floem::prelude::*;
use floem::reactive::SignalGet;
use floem::views::v_stack_from_iter;

use crate::logic::{commands, AppState, Screen};
use crate::ui::components::button::{primary_button, secondary_button};
use crate::ui::components::data_table::{empty_state, table_header, table_row_with_actions};
use crate::ui::components::dropdown::dropdown;
use crate::ui::components::form_field::{form_section, text_field};
use crate::ui::components::form_layout::form_screen;
use crate::ui::styles::*;

/// Domains screen dispatcher.
pub fn domains_list_screen(state: &'static AppState) -> impl IntoView {
    dyn_container(
        move || state.navigation.current_screen.get(),
        move |screen| match screen {
            Screen::DomainCreate => domain_create_view(state).into_any(),
            Screen::DomainEdit => domain_edit_view(state).into_any(),
            Screen::DomainDnsRecords => domain_dns_records_view(state).into_any(),
            _ => domains_list_view(state).into_any(),
        },
    )
    .style(|s| s.width_full().height_full())
}

fn domains_list_view(state: &'static AppState) -> impl IntoView {
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
                state.form.reset();
                state.navigation.navigate_to(Screen::DomainCreate);
            }),
        ))
        .style(|s| s.width_full().items_center().margin_bottom(SPACING_LG)),
        // Table header
        table_header(vec![
            ("Name", 200.0),
            ("Provider", 120.0),
            ("Records", 80.0),
            ("Actions", 200.0),
        ]),
        // Domains list
        scroll(
            dyn_container(
                move || domains.get(),
                move |domain_list| {
                    if domain_list.is_empty() {
                        return empty_state(
                            "No domains configured. Click 'New Domain' to add one.",
                        )
                        .into_any();
                    }

                    let rows: Vec<_> = domain_list
                        .iter()
                        .map(|domain| {
                            let domain_id = domain.id;
                            let domain_name = domain.name.clone();

                            table_row_with_actions(
                                vec![
                                    (domain.name.clone(), 200.0),
                                    (domain.provider_name().to_string(), 120.0),
                                    (format!("{}", domain.dns_records.len()), 80.0),
                                ],
                                vec![
                                    (
                                        "Edit".to_string(),
                                        ACCENT_BLUE,
                                        Box::new(move || {
                                            state.selected_domain_id.set(Some(domain_id));
                                            state.form.reset();
                                            state.form.edit_form_initialized.set(false);
                                            state.navigation.navigate_to(Screen::DomainEdit);
                                        }),
                                    ),
                                    (
                                        "DNS".to_string(),
                                        ACCENT_GREEN,
                                        Box::new(move || {
                                            state.selected_domain_id.set(Some(domain_id));
                                            state
                                                .navigation
                                                .navigate_to(Screen::DomainDnsRecords);
                                        }),
                                    ),
                                    (
                                        "Delete".to_string(),
                                        ACCENT_RED,
                                        Box::new({
                                            let name = domain_name.clone();
                                            move || {
                                                state.selected_domain_id.set(Some(domain_id));
                                                state
                                                    .form
                                                    .begin_deletion(domain_id, &name, "domain");
                                                state
                                                    .navigation
                                                    .navigate_to(Screen::DomainDnsRecords);
                                                // Navigate to a delete-like state - reuse DNS screen with deletion
                                                // Actually, use inline deletion
                                                commands::delete_domain(state, domain_id);
                                            }
                                        }),
                                    ),
                                ],
                                move || {
                                    state.selected_domain_id.set(Some(domain_id));
                                    state.form.reset();
                                    state.form.edit_form_initialized.set(false);
                                    state.navigation.navigate_to(Screen::DomainEdit);
                                },
                            )
                            .into_any()
                        })
                        .collect();

                    v_stack_from_iter(rows)
                        .style(|s| s.width_full())
                        .into_any()
                },
            )
            .style(|s| s.width_full()),
        )
        .style(|s| s.width_full().flex_grow(1.0).min_height(0.0)),
    ))
    .style(|s| {
        s.width_full()
            .height_full()
            .padding(SPACING_XL)
            .flex_col()
    })
}

fn domain_create_view(state: &'static AppState) -> impl IntoView {
    let provider = state.form.domain_provider;

    form_screen(
        "Create Domain",
        state,
        v_stack((
            form_section(
                "Domain Information",
                v_stack((
                    text_field("Domain Name", state.form.domain_name),
                    dropdown(
                        "DNS Provider",
                        state.form.domain_provider,
                        vec![
                            "manual".to_string(),
                            "cloudflare".to_string(),
                            "route53".to_string(),
                        ],
                    ),
                )),
            ),
            // Provider-specific credentials
            dyn_container(
                move || provider.get(),
                move |prov: String| match prov.as_str() {
                    "cloudflare" => form_section(
                        "Cloudflare Settings",
                        text_field("Zone ID", state.form.domain_zone_id),
                    )
                    .into_any(),
                    "route53" => form_section(
                        "Route53 Settings",
                        v_stack((
                            text_field("Zone ID", state.form.domain_zone_id),
                            text_field("Access Key", state.form.domain_access_key),
                            text_field("Secret Key", state.form.domain_secret_key),
                        )),
                    )
                    .into_any(),
                    _ => empty().into_any(),
                },
            ),
        )),
        move || {
            commands::submit_create_domain(state);
        },
        "Create Domain",
    )
}

fn domain_edit_view(state: &'static AppState) -> impl IntoView {
    let provider = state.form.domain_provider;
    let initialized = state.form.edit_form_initialized;

    dyn_container(
        move || (initialized.get(), state.selected_domain_id.get()),
        move |(is_init, domain_id_opt)| {
            if !is_init {
                if let Some(domain_id) = domain_id_opt {
                    if let Some(domain) = state.find_domain(domain_id) {
                        state.form.init_from_domain(&domain);
                        state.form.edit_form_initialized.set(true);
                    }
                }
            }

            form_screen(
                "Edit Domain",
                state,
                v_stack((
                    form_section(
                        "Domain Information",
                        v_stack((
                            text_field("Domain Name", state.form.domain_name),
                            dropdown(
                                "DNS Provider",
                                state.form.domain_provider,
                                vec![
                                    "manual".to_string(),
                                    "cloudflare".to_string(),
                                    "route53".to_string(),
                                ],
                            ),
                        )),
                    ),
                    dyn_container(
                        move || provider.get(),
                        move |prov: String| match prov.as_str() {
                            "cloudflare" => form_section(
                                "Cloudflare Settings",
                                text_field("Zone ID", state.form.domain_zone_id),
                            )
                            .into_any(),
                            "route53" => form_section(
                                "Route53 Settings",
                                v_stack((
                                    text_field("Zone ID", state.form.domain_zone_id),
                                    text_field("Access Key", state.form.domain_access_key),
                                    text_field("Secret Key", state.form.domain_secret_key),
                                )),
                            )
                            .into_any(),
                            _ => empty().into_any(),
                        },
                    ),
                )),
                move || {
                    commands::submit_update_domain(state);
                },
                "Save Changes",
            )
            .into_any()
        },
    )
    .style(|s| s.width_full().height_full())
}

fn domain_dns_records_view(state: &'static AppState) -> impl IntoView {
    let domains = state.domains;
    let selected_id = state.selected_domain_id;

    v_stack((
        // Header
        h_stack((
            secondary_button("Back", move || {
                state.navigation.navigate_back();
            }),
            "DNS Records".style(|s| {
                s.font_size(FONT_SIZE_TITLE)
                    .color(TEXT_PRIMARY)
                    .font_bold()
                    .margin_left(SPACING_MD)
            }),
        ))
        .style(|s| s.width_full().items_center().margin_bottom(SPACING_LG)),
        // Table
        table_header(vec![
            ("Type", 80.0),
            ("Name", 200.0),
            ("Value", 200.0),
            ("TTL", 80.0),
            ("Proxied", 80.0),
        ]),
        scroll(
            dyn_container(
                move || (domains.get(), selected_id.get()),
                move |(domain_list, sel_id)| {
                    let domain = sel_id
                        .and_then(|id| domain_list.iter().find(|d| d.id == id));

                    match domain {
                        None => empty_state("No domain selected.").into_any(),
                        Some(d) if d.dns_records.is_empty() => {
                            empty_state("No DNS records for this domain.").into_any()
                        }
                        Some(d) => {
                            let rows: Vec<_> = d
                                .dns_records
                                .iter()
                                .map(|record| {
                                    h_stack((
                                        record
                                            .record_type
                                            .to_string()
                                            .style(|s| {
                                                s.min_width(80.0)
                                                    .flex_basis(0.0)
                                                    .flex_grow(1.0)
                                                    .padding_horiz(SPACING_SM)
                                                    .font_size(FONT_SIZE_MD)
                                                    .color(TEXT_PRIMARY)
                                                    .justify_center()
                                            }),
                                        record.name.clone().style(|s| {
                                            s.min_width(200.0)
                                                .flex_basis(0.0)
                                                .flex_grow(2.0)
                                                .padding_horiz(SPACING_SM)
                                                .font_size(FONT_SIZE_MD)
                                                .color(TEXT_PRIMARY)
                                                .justify_center()
                                                .text_ellipsis()
                                        }),
                                        record.value.clone().style(|s| {
                                            s.min_width(200.0)
                                                .flex_basis(0.0)
                                                .flex_grow(2.0)
                                                .padding_horiz(SPACING_SM)
                                                .font_size(FONT_SIZE_MD)
                                                .color(TEXT_PRIMARY)
                                                .justify_center()
                                                .text_ellipsis()
                                        }),
                                        record.ttl.to_string().style(|s| {
                                            s.min_width(80.0)
                                                .flex_basis(0.0)
                                                .flex_grow(1.0)
                                                .padding_horiz(SPACING_SM)
                                                .font_size(FONT_SIZE_MD)
                                                .color(TEXT_PRIMARY)
                                                .justify_center()
                                        }),
                                        (if record.proxied { "Yes" } else { "No" })
                                            .style(|s| {
                                                s.min_width(80.0)
                                                    .flex_basis(0.0)
                                                    .flex_grow(1.0)
                                                    .padding_horiz(SPACING_SM)
                                                    .font_size(FONT_SIZE_MD)
                                                    .color(TEXT_PRIMARY)
                                                    .justify_center()
                                            }),
                                    ))
                                    .style(|s| {
                                        s.width_full()
                                            .padding_vert(SPACING_SM)
                                            .border_bottom(1.0)
                                            .border_color(BORDER_MUTED)
                                    })
                                    .into_any()
                                })
                                .collect();

                            v_stack_from_iter(rows)
                                .style(|s| s.width_full())
                                .into_any()
                        }
                    }
                },
            )
            .style(|s| s.width_full()),
        )
        .style(|s| s.width_full().flex_grow(1.0).min_height(0.0)),
    ))
    .style(|s| {
        s.width_full()
            .height_full()
            .padding(SPACING_XL)
            .flex_col()
    })
}
