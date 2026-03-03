use floem::prelude::*;
use floem::reactive::SignalGet;
use floem::views::v_stack_from_iter;

use crate::logic::{commands, AppState, Screen};
use crate::ui::components::button::{primary_button, secondary_button};
use crate::ui::components::data_table::{empty_state, table_header, table_row_with_actions};
use crate::ui::components::dropdown::dropdown;
use crate::ui::components::form_field::{form_section, readonly_field, text_field, text_field_with_placeholder};
use crate::ui::components::form_layout::form_screen;
use crate::ui::styles::*;

/// Nodes screen dispatcher.
pub fn nodes_list_screen(state: &'static AppState) -> impl IntoView {
    dyn_container(
        move || state.navigation.current_screen.get(),
        move |screen| match screen {
            Screen::NodeCreate => node_create_view(state).into_any(),
            Screen::NodeEdit => node_edit_view(state).into_any(),
            Screen::NodeConfig => node_config_view(state).into_any(),
            Screen::NodeQuickConfig => node_quick_config_view(state).into_any(),
            _ => nodes_list_view(state).into_any(),
        },
    )
    .style(|s| s.width_full().height_full())
}

fn nodes_list_view(state: &'static AppState) -> impl IntoView {
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
                state.form.reset();
                // Generate API key for new node
                let api_key = uuid::Uuid::new_v4().to_string();
                state.form.node_api_key.set(api_key);
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
            ("Name", 150.0),
            ("Endpoint", 200.0),
            ("Proxy", 80.0),
            ("Status", 80.0),
            ("Actions", 220.0),
        ]),
        // Nodes list
        scroll(
            dyn_container(
                move || nodes.get(),
                move |node_list| {
                    if node_list.is_empty() {
                        return empty_state(
                            "No nodes configured. Click 'New Node' to add one.",
                        )
                        .into_any();
                    }

                    let rows: Vec<_> = node_list
                        .iter()
                        .map(|node| {
                            let node_id = node.id;

                            table_row_with_actions(
                                vec![
                                    (node.name.clone(), 150.0),
                                    (node.api_endpoint.clone(), 200.0),
                                    (node.proxy_type.to_string(), 80.0),
                                    (node.status.to_string(), 80.0),
                                ],
                                vec![
                                    (
                                        "Edit".to_string(),
                                        ACCENT_BLUE,
                                        Box::new(move || {
                                            state.selected_node_id.set(Some(node_id));
                                            state.form.reset();
                                            state.form.edit_form_initialized.set(false);
                                            state.navigation.navigate_to(Screen::NodeEdit);
                                        }),
                                    ),
                                    (
                                        "Config".to_string(),
                                        ACCENT_GREEN,
                                        Box::new(move || {
                                            state.selected_node_id.set(Some(node_id));
                                            state.navigation.navigate_to(Screen::NodeConfig);
                                        }),
                                    ),
                                    (
                                        "Health".to_string(),
                                        ACCENT_YELLOW,
                                        Box::new(move || {
                                            commands::health_check_node(state, node_id);
                                        }),
                                    ),
                                    (
                                        "Delete".to_string(),
                                        ACCENT_RED,
                                        Box::new(move || {
                                            commands::delete_node(state, node_id);
                                        }),
                                    ),
                                ],
                                move || {
                                    state.selected_node_id.set(Some(node_id));
                                    state.form.reset();
                                    state.form.edit_form_initialized.set(false);
                                    state.navigation.navigate_to(Screen::NodeEdit);
                                },
                            )
                            .into_any()
                        })
                        .collect();

                    v_stack_from_iter(rows)
                        .style(|s| s.width_full())
                        .into_any()
                },
            ),
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

fn node_create_view(state: &'static AppState) -> impl IntoView {
    form_screen(
        "Create Node",
        state,
        v_stack((
            form_section(
                "Node Details",
                v_stack((
                    text_field("Name", state.form.node_name),
                    text_field_with_placeholder(
                        "API Endpoint",
                        state.form.node_api_endpoint,
                        "https://node.example.com:8443",
                    ),
                    text_field_with_placeholder(
                        "IP Address (optional)",
                        state.form.node_ip_address,
                        "0.0.0.0",
                    ),
                    dropdown(
                        "Proxy Type",
                        state.form.node_proxy_type,
                        vec![
                            "traefik".to_string(),
                            "nginx".to_string(),
                            "apache".to_string(),
                        ],
                    ),
                )),
            ),
            form_section(
                "Authentication",
                dyn_container(
                    move || state.form.node_api_key.get(),
                    move |key: String| {
                        readonly_field("API Key (auto-generated)", &key).into_any()
                    },
                ),
            ),
        )),
        move || {
            commands::submit_create_node(state);
        },
        "Create Node",
    )
}

fn node_edit_view(state: &'static AppState) -> impl IntoView {
    let initialized = state.form.edit_form_initialized;

    dyn_container(
        move || (initialized.get(), state.selected_node_id.get()),
        move |(is_init, node_id_opt)| {
            if !is_init {
                if let Some(node_id) = node_id_opt {
                    if let Some(node) = state.find_node(node_id) {
                        state.form.init_from_node(&node);
                        state.form.edit_form_initialized.set(true);
                    }
                }
            }

            form_screen(
                "Edit Node",
                state,
                v_stack((
                    form_section(
                        "Node Details",
                        v_stack((
                            text_field("Name", state.form.node_name),
                            text_field_with_placeholder(
                                "API Endpoint",
                                state.form.node_api_endpoint,
                                "https://node.example.com:8443",
                            ),
                            text_field_with_placeholder(
                                "IP Address",
                                state.form.node_ip_address,
                                "0.0.0.0",
                            ),
                            dropdown(
                                "Proxy Type",
                                state.form.node_proxy_type,
                                vec![
                                    "traefik".to_string(),
                                    "nginx".to_string(),
                                    "apache".to_string(),
                                ],
                            ),
                        )),
                    ),
                    form_section(
                        "Authentication",
                        text_field("API Key", state.form.node_api_key),
                    ),
                )),
                move || {
                    commands::submit_update_node(state);
                },
                "Save Changes",
            )
            .into_any()
        },
    )
    .style(|s| s.width_full().height_full())
}

fn node_config_view(state: &'static AppState) -> impl IntoView {
    let nodes = state.nodes;
    let selected_id = state.selected_node_id;

    v_stack((
        // Header
        h_stack((
            secondary_button("Back", move || {
                state.navigation.navigate_back();
            }),
            "Node Configuration".style(|s| {
                s.font_size(FONT_SIZE_TITLE)
                    .color(TEXT_PRIMARY)
                    .font_bold()
                    .margin_left(SPACING_MD)
            }),
        ))
        .style(|s| s.width_full().items_center().margin_bottom(SPACING_LG)),
        // Config display
        scroll(
            dyn_container(
                move || (nodes.get(), selected_id.get()),
                move |(node_list, sel_id)| {
                    let node = sel_id.and_then(|id| node_list.iter().find(|n| n.id == id));

                    match node {
                        None => "No node selected.".style(|s| {
                            s.color(TEXT_MUTED).font_size(FONT_SIZE_MD)
                        }).into_any(),
                        Some(n) => {
                            // Generate a TOML-like config display
                            let config = format!(
                                "[node]\n\
                                 name = \"{}\"\n\
                                 api_endpoint = \"{}\"\n\
                                 api_key = \"{}\"\n\
                                 ip_address = \"{}\"\n\
                                 proxy_type = \"{}\"\n\
                                 \n\
                                 [status]\n\
                                 status = \"{}\"\n\
                                 last_health_check = \"{}\"",
                                n.name,
                                n.api_endpoint,
                                n.api_key,
                                n.ip_address,
                                n.proxy_type,
                                n.status,
                                n.last_health_check.format("%Y-%m-%d %H:%M:%S UTC"),
                            );

                            v_stack((
                                "Copy this configuration to your node's config file:"
                                    .style(|s| {
                                        s.font_size(FONT_SIZE_MD)
                                            .color(TEXT_SECONDARY)
                                            .margin_bottom(SPACING_MD)
                                    }),
                                config.style(|s| {
                                    s.width_full()
                                        .padding(SPACING_LG)
                                        .background(BG_SECONDARY)
                                        .color(TEXT_PRIMARY)
                                        .border_radius(BORDER_RADIUS)
                                        .border(1.0)
                                        .border_color(BORDER_DEFAULT)
                                        .font_size(FONT_SIZE_SM)
                                        .font_family("monospace".to_string())
                                        .line_height(1.6)
                                }),
                                // Quick config button
                                secondary_button("Quick Config Upload", move || {
                                    state.navigation.navigate_to(Screen::NodeQuickConfig);
                                }),
                            ))
                            .style(|s| s.width_full())
                            .into_any()
                        }
                    }
                },
            ),
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

fn node_quick_config_view(state: &'static AppState) -> impl IntoView {
    let quick_url = state.quick_config_url;
    let quick_expires = state.quick_config_expires_at;

    v_stack((
        // Header
        h_stack((
            secondary_button("Back", move || {
                state.navigation.navigate_back();
            }),
            "Quick Config".style(|s| {
                s.font_size(FONT_SIZE_TITLE)
                    .color(TEXT_PRIMARY)
                    .font_bold()
                    .margin_left(SPACING_MD)
            }),
        ))
        .style(|s| s.width_full().items_center().margin_bottom(SPACING_LG)),
        // Instructions
        "Quick Config allows you to upload the node configuration file via a temporary URL."
            .style(|s| {
                s.font_size(FONT_SIZE_MD)
                    .color(TEXT_SECONDARY)
                    .margin_bottom(SPACING_LG)
            }),
        // Upload button
        primary_button("Generate Config URL", move || {
            // In a real implementation, this would call an async command
            // For now, show a placeholder
            state
                .notifications
                .push(crate::logic::Notification::info(
                    "Quick config upload is not yet implemented",
                ));
        }),
        // Show URL if available
        dyn_container(
            move || (quick_url.get(), quick_expires.get()),
            move |(url, expires): (String, String)| {
                if url.is_empty() {
                    return empty().into_any();
                }
                v_stack((
                    "Config URL:".style(|s| {
                        s.font_size(FONT_SIZE_SM)
                            .color(TEXT_SECONDARY)
                            .margin_top(SPACING_LG)
                            .margin_bottom(SPACING_XS)
                    }),
                    url.clone().style(|s| {
                        s.width_full()
                            .padding(SPACING_SM)
                            .background(BG_SECONDARY)
                            .color(ACCENT_BLUE)
                            .border_radius(BORDER_RADIUS_SM)
                            .font_size(FONT_SIZE_MD)
                            .font_family("monospace".to_string())
                            .margin_bottom(SPACING_SM)
                    }),
                    format!("Expires: {}", expires).style(|s| {
                        s.font_size(FONT_SIZE_SM).color(TEXT_MUTED)
                    }),
                ))
                .into_any()
            },
        ),
    ))
    .style(|s| {
        s.width_full()
            .height_full()
            .padding(SPACING_XL)
            .flex_col()
    })
}
