use std::collections::HashMap;

use floem::prelude::*;
use floem::reactive::{RwSignal, SignalGet, SignalUpdate};
use floem::views::v_stack_from_iter;

use crate::logic::{commands, AppState, Screen};
use crate::ui::components::button::{primary_button, secondary_button};
use crate::ui::components::confirm_dialog::delete_confirm_dialog;
use crate::ui::components::data_table::{empty_state, table_header, table_row_with_actions};
use crate::ui::components::dropdown::dropdown;
use crate::ui::components::form_field::{form_section, readonly_field, text_field, text_field_with_placeholder};
use crate::ui::components::form_layout::form_screen;
use crate::ui::components::multi_pair::{domain_mapping_editor, env_var_editor, volume_editor};
use crate::ui::styles::*;

/// Sites screen dispatcher.
pub fn sites_list_screen(state: &'static AppState) -> impl IntoView {
    dyn_container(
        move || state.navigation.current_screen.get(),
        move |screen| match screen {
            Screen::SiteCreate => site_create_view(state).into_any(),
            Screen::SiteEdit => site_edit_view(state).into_any(),
            Screen::SiteEnvVars => site_env_vars_view(state).into_any(),
            Screen::SiteDeleteConfirm => site_delete_view(state).into_any(),
            _ => sites_list_view(state).into_any(),
        },
    )
    .style(|s| s.width_full().height_full())
}

fn sites_list_view(state: &'static AppState) -> impl IntoView {
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
                state.form.reset();
                state.navigation.navigate_to(Screen::SiteCreate);
            }),
        ))
        .style(|s| s.width_full().items_center().margin_bottom(SPACING_LG)),
        // Table header
        table_header(vec![
            ("Name", 160.0),
            ("Domain", 160.0),
            ("Image", 160.0),
            ("Status", 80.0),
            ("Actions", 200.0),
        ]),
        // Sites list
        scroll(
            dyn_container(
                move || (sites.get(), domains.get()),
                move |(site_list, domain_list)| {
                    if site_list.is_empty() {
                        return empty_state("No sites configured. Click 'New Site' to create one.")
                            .into_any();
                    }

                    let rows: Vec<_> = site_list
                        .iter()
                        .map(|site| {
                            let site_id = site.id;
                            let site_name = site.name.clone();
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
                                .unwrap_or_else(|| "\u{2014}".to_string());

                            table_row_with_actions(
                                vec![
                                    (site.name.clone(), 160.0),
                                    (domain_name, 160.0),
                                    (
                                        if site.docker_image.is_empty() {
                                            "compose".to_string()
                                        } else {
                                            site.docker_image.clone()
                                        },
                                        160.0,
                                    ),
                                    (site.status.to_string(), 80.0),
                                ],
                                vec![
                                    (
                                        "Edit".to_string(),
                                        ACCENT_BLUE,
                                        Box::new(move || {
                                            state.selected_site_id.set(Some(site_id));
                                            state.form.reset();
                                            state.form.edit_form_initialized.set(false);
                                            state.navigation.navigate_to(Screen::SiteEdit);
                                        }),
                                    ),
                                    (
                                        "Deploy".to_string(),
                                        ACCENT_GREEN,
                                        Box::new(move || {
                                            commands::deploy_site(state, site_id);
                                        }),
                                    ),
                                    (
                                        "Stop".to_string(),
                                        ACCENT_YELLOW,
                                        Box::new(move || {
                                            commands::stop_site(state, site_id);
                                        }),
                                    ),
                                    (
                                        "Delete".to_string(),
                                        ACCENT_RED,
                                        Box::new({
                                            let name = site_name.clone();
                                            move || {
                                                state.selected_site_id.set(Some(site_id));
                                                state.form.begin_deletion(site_id, &name, "site");
                                                state
                                                    .navigation
                                                    .navigate_to(Screen::SiteDeleteConfirm);
                                            }
                                        }),
                                    ),
                                ],
                                move || {
                                    state.selected_site_id.set(Some(site_id));
                                    state.form.reset();
                                    state.form.edit_form_initialized.set(false);
                                    state.navigation.navigate_to(Screen::SiteEdit);
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

fn site_create_view(state: &'static AppState) -> impl IntoView {
    let site_type = state.form.site_type;

    // Build node options
    let nodes = state.nodes;
    let settings = state.settings;

    form_screen(
        "Create Site",
        state,
        v_stack((
            // Basic Information section
            form_section(
                "Basic Information",
                v_stack((
                    text_field("Site Name", state.form.site_name),
                    dropdown(
                        "Site Type",
                        state.form.site_type,
                        vec!["container".to_string(), "compose".to_string()],
                    ),
                    // Node dropdown
                    dyn_container(
                        move || nodes.get(),
                        move |node_list| {
                            let node_names: Vec<String> =
                                node_list.iter().map(|n| n.name.clone()).collect();
                            let node_name_signal = RwSignal::new(String::new());

                            v_stack((
                                dropdown("Node", node_name_signal, node_names.clone()),
                                dyn_container(
                                    move || node_name_signal.get(),
                                    move |selected_name: String| {
                                        if let Some(node) =
                                            node_list.iter().find(|n| n.name == selected_name)
                                        {
                                            state.form.site_node_id.set(node.id.to_string());
                                        }
                                        empty().into_any()
                                    },
                                ),
                            ))
                            .into_any()
                        },
                    ),
                    text_field_with_placeholder(
                        "SSL Email (optional)",
                        state.form.site_ssl_email,
                        "admin@example.com",
                    ),
                )),
            ),
            // Container/Compose settings section
            dyn_container(
                move || site_type.get(),
                move |st: String| {
                    if st == "compose" {
                        return form_section(
                            "Compose Settings",
                            text_field("Compose Content", state.form.site_compose_content),
                        )
                        .into_any();
                    }

                    form_section(
                        "Container Settings",
                        v_stack((
                            text_field("Docker Image", state.form.site_docker_image),
                            dyn_container(
                                move || settings.get(),
                                move |s| {
                                    let cred_names: Vec<String> = std::iter::once(String::new())
                                        .chain(
                                            s.docker_credentials.iter().map(|c| c.name.clone()),
                                        )
                                        .collect();
                                    let cred_name_signal = RwSignal::new(String::new());
                                    let creds = s.docker_credentials.clone();

                                    v_stack((
                                        dropdown(
                                            "Docker Credential (optional)",
                                            cred_name_signal,
                                            cred_names,
                                        ),
                                        dyn_container(
                                            move || cred_name_signal.get(),
                                            move |selected_name: String| {
                                                if let Some(cred) =
                                                    creds.iter().find(|c| c.name == selected_name)
                                                {
                                                    state
                                                        .form
                                                        .site_docker_credential_id
                                                        .set(cred.id.to_string());
                                                } else {
                                                    state
                                                        .form
                                                        .site_docker_credential_id
                                                        .set(String::new());
                                                }
                                                empty().into_any()
                                            },
                                        ),
                                    ))
                                    .into_any()
                                },
                            ),
                        )),
                    )
                    .into_any()
                },
            ),
            // Volume and domain mapping editors (already have card styling)
            volume_editor(state),
            domain_mapping_editor(state),
        )),
        move || {
            commands::submit_create_site(state);
        },
        "Create Site",
    )
}

fn site_edit_view(state: &'static AppState) -> impl IntoView {
    let site_type = state.form.site_type;
    let nodes = state.nodes;
    let settings = state.settings;
    let initialized = state.form.edit_form_initialized;

    // Initialize form on first render
    dyn_container(
        move || (initialized.get(), state.selected_site_id.get()),
        move |(is_init, site_id_opt)| {
            if !is_init {
                if let Some(site_id) = site_id_opt {
                    if let Some(site) = state.find_site(site_id) {
                        let domains = state.domains.get_untracked();
                        state.form.init_from_site(&site, &domains);
                        state.form.edit_form_initialized.set(true);
                    }
                }
            }

            form_screen(
                "Edit Site",
                state,
                v_stack((
                    // Basic Information section
                    form_section(
                        "Basic Information",
                        v_stack((
                            text_field("Site Name", state.form.site_name),
                            // Show site type as read-only
                            dyn_container(
                                move || site_type.get(),
                                move |st: String| {
                                    readonly_field("Site Type", &st).into_any()
                                },
                            ),
                            // Node dropdown
                            dyn_container(
                                move || (nodes.get(), state.form.site_node_id.get()),
                                move |(node_list, current_node_id): (Vec<_>, String)| {
                                    let current_name = node_list
                                        .iter()
                                        .find(|n| n.id.to_string() == current_node_id)
                                        .map(|n| n.name.clone())
                                        .unwrap_or_default();
                                    let node_name_signal = RwSignal::new(current_name);
                                    let node_names: Vec<String> =
                                        node_list.iter().map(|n| n.name.clone()).collect();

                                    v_stack((
                                        dropdown("Node", node_name_signal, node_names),
                                        dyn_container(
                                            move || node_name_signal.get(),
                                            move |selected_name: String| {
                                                if let Some(node) = node_list
                                                    .iter()
                                                    .find(|n| n.name == selected_name)
                                                {
                                                    state
                                                        .form
                                                        .site_node_id
                                                        .set(node.id.to_string());
                                                }
                                                empty().into_any()
                                            },
                                        ),
                                    ))
                                    .into_any()
                                },
                            ),
                            text_field_with_placeholder(
                                "SSL Email (optional)",
                                state.form.site_ssl_email,
                                "admin@example.com",
                            ),
                        )),
                    ),
                    // Container/Compose settings section
                    dyn_container(
                        move || site_type.get(),
                        move |st: String| {
                            if st == "compose" {
                                return form_section(
                                    "Compose Settings",
                                    text_field(
                                        "Compose Content",
                                        state.form.site_compose_content,
                                    ),
                                )
                                .into_any();
                            }
                            form_section(
                                "Container Settings",
                                v_stack((
                                    text_field("Docker Image", state.form.site_docker_image),
                                    dyn_container(
                                        move || settings.get(),
                                        move |s| {
                                            let cred_names: Vec<String> =
                                                std::iter::once(String::new())
                                                    .chain(
                                                        s.docker_credentials
                                                            .iter()
                                                            .map(|c| c.name.clone()),
                                                    )
                                                    .collect();
                                            let current_cred = state
                                                .form
                                                .site_docker_credential_id
                                                .get_untracked();
                                            let current_name = s
                                                .docker_credentials
                                                .iter()
                                                .find(|c| c.id.to_string() == current_cred)
                                                .map(|c| c.name.clone())
                                                .unwrap_or_default();
                                            let cred_name_signal = RwSignal::new(current_name);
                                            let creds = s.docker_credentials.clone();

                                            v_stack((
                                                dropdown(
                                                    "Docker Credential (optional)",
                                                    cred_name_signal,
                                                    cred_names,
                                                ),
                                                dyn_container(
                                                    move || cred_name_signal.get(),
                                                    move |selected_name: String| {
                                                        if let Some(cred) = creds
                                                            .iter()
                                                            .find(|c| c.name == selected_name)
                                                        {
                                                            state
                                                                .form
                                                                .site_docker_credential_id
                                                                .set(cred.id.to_string());
                                                        } else {
                                                            state
                                                                .form
                                                                .site_docker_credential_id
                                                                .set(String::new());
                                                        }
                                                        empty().into_any()
                                                    },
                                                ),
                                            ))
                                            .into_any()
                                        },
                                    ),
                                )),
                            )
                            .into_any()
                        },
                    ),
                    // Volume and domain mapping editors (already have card styling)
                    volume_editor(state),
                    domain_mapping_editor(state),
                    // Edit env vars button
                    secondary_button("Edit Environment Variables", move || {
                        state.navigation.navigate_to(Screen::SiteEnvVars);
                    }),
                )),
                move || {
                    commands::submit_update_site(state);
                },
                "Save Changes",
            )
            .into_any()
        },
    )
    .style(|s| s.width_full().height_full())
}

fn site_env_vars_view(state: &'static AppState) -> impl IntoView {
    form_screen(
        "Environment Variables",
        state,
        v_stack((
            "Edit environment variables for this site."
                .style(|s| {
                    s.font_size(FONT_SIZE_MD)
                        .color(TEXT_SECONDARY)
                        .margin_bottom(SPACING_LG)
                }),
            env_var_editor(state),
        )),
        move || {
            // Save env vars back to the site
            if let Some(site_id) = state.selected_site_id.get_untracked() {
                let pairs = state.form.env_var_pairs.get_untracked();
                let env_vars: HashMap<String, String> = pairs
                    .into_iter()
                    .filter(|p| !p.key.is_empty())
                    .map(|p| (p.key, p.value))
                    .collect();

                state.sites.update(|sites| {
                    if let Some(site) = sites.iter_mut().find(|s| s.id == site_id) {
                        site.environment_vars = env_vars;
                    }
                });

                // Save to disk
                if let Some(site) = state.find_site(site_id) {
                    let domain_name = state.domain_name_for_site(&site);
                    let loader = crate::config::FileConfigLoader::new();
                    let _ = loader.save_site(&site, &domain_name);
                }

                state
                    .notifications
                    .push(crate::logic::Notification::success("Environment variables saved"));
            }
            state.navigation.navigate_back();
        },
        "Save",
    )
}

fn site_delete_view(state: &'static AppState) -> impl IntoView {
    v_stack((
        // Header
        h_stack((
            secondary_button("Back", move || {
                state.form.reset_deletion();
                state.navigation.navigate_back();
            }),
            "Delete Site".style(|s| {
                s.font_size(FONT_SIZE_TITLE)
                    .color(TEXT_PRIMARY)
                    .font_bold()
                    .margin_left(SPACING_MD)
            }),
        ))
        .style(|s| s.width_full().items_center().margin_bottom(SPACING_XL)),
        // Confirm dialog
        delete_confirm_dialog(
            state,
            move || {
                commands::submit_delete(state);
            },
            move || {
                state.form.reset_deletion();
                state.navigation.navigate_back();
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
