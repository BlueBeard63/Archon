use floem::prelude::*;
use floem::reactive::{RwSignal, SignalGet};
use floem::views::v_stack_from_iter;

use crate::logic::{commands, AppState, Screen};
use crate::ui::components::button::primary_button;
use crate::ui::components::data_table::{empty_state, table_header, table_row_with_actions};
use crate::ui::components::form_field::{form_section, text_field, text_field_with_placeholder};
use crate::ui::components::form_layout::form_screen;
use crate::ui::styles::*;

/// Settings screen dispatcher.
pub fn settings_screen(state: &'static AppState) -> impl IntoView {
    dyn_container(
        move || state.navigation.current_screen.get(),
        move |screen| match screen {
            Screen::DockerCredentialCreate => docker_cred_create_view(state).into_any(),
            Screen::DockerCredentialEdit => docker_cred_edit_view(state).into_any(),
            _ => settings_main_view(state).into_any(),
        },
    )
    .style(|s| s.width_full().height_full())
}

fn settings_main_view(state: &'static AppState) -> impl IntoView {
    let settings = state.settings;
    let settings_initialized = RwSignal::new(false);

    // Initialize settings form on first render
    dyn_container(
        move || (settings.get(), settings_initialized.get()),
        move |(s, initialized)| {
            if !initialized {
                state.form.init_from_settings(&s);
                settings_initialized.set(true);
            }

            v_stack((
                "Settings".style(|s| {
                    s.font_size(FONT_SIZE_TITLE)
                        .color(TEXT_PRIMARY)
                        .font_bold()
                        .margin_bottom(SPACING_XL)
                }),
                scroll(
                    v_stack((
                        // API Keys section
                        form_section(
                            "API Keys",
                            v_stack((
                                text_field(
                                    "Cloudflare API Token",
                                    state.form.settings_cloudflare_token,
                                ),
                                text_field(
                                    "Route53 Access Key",
                                    state.form.settings_route53_access_key,
                                ),
                                text_field(
                                    "Route53 Secret Key",
                                    state.form.settings_route53_secret_key,
                                ),
                                primary_button("Save Settings", move || {
                                    commands::submit_save_settings(state);
                                }),
                            )),
                        ),
                        // Docker Credentials section
                        h_stack((
                            "Docker Credentials".style(|s| {
                                s.font_size(FONT_SIZE_LG)
                                    .color(TEXT_PRIMARY)
                                    .font_bold()
                            }),
                            empty().style(|s| s.flex_grow(1.0)),
                            primary_button("New Credential", move || {
                                state.form.reset();
                                state
                                    .navigation
                                    .navigate_to(Screen::DockerCredentialCreate);
                            }),
                        ))
                        .style(|s| {
                            s.width_full()
                                .items_center()
                                .margin_bottom(SPACING_MD)
                        }),
                        table_header(vec![
                            ("Name", 150.0),
                            ("Registry", 150.0),
                            ("Username", 150.0),
                            ("Actions", 120.0),
                        ]),
                        dyn_container(
                            move || settings.get(),
                            move |s| {
                                if s.docker_credentials.is_empty() {
                                    return empty_state(
                                        "No Docker credentials configured.",
                                    )
                                    .into_any();
                                }

                                let rows: Vec<_> = s
                                    .docker_credentials
                                    .iter()
                                    .map(|cred| {
                                        let cred_id = cred.id;

                                        table_row_with_actions(
                                            vec![
                                                (cred.name.clone(), 150.0),
                                                (cred.registry.clone(), 150.0),
                                                (cred.username.clone(), 150.0),
                                            ],
                                            vec![
                                                (
                                                    "Edit".to_string(),
                                                    ACCENT_BLUE,
                                                    Box::new(move || {
                                                        state
                                                            .selected_docker_credential_id
                                                            .set(Some(cred_id));
                                                        state.form.reset();
                                                        state
                                                            .form
                                                            .edit_form_initialized
                                                            .set(false);
                                                        state.navigation.navigate_to(
                                                            Screen::DockerCredentialEdit,
                                                        );
                                                    }),
                                                ),
                                                (
                                                    "Delete".to_string(),
                                                    ACCENT_RED,
                                                    Box::new(move || {
                                                        commands::submit_delete_docker_credential(
                                                            state, cred_id,
                                                        );
                                                    }),
                                                ),
                                            ],
                                            move || {
                                                state
                                                    .selected_docker_credential_id
                                                    .set(Some(cred_id));
                                                state.form.reset();
                                                state.form.edit_form_initialized.set(false);
                                                state.navigation.navigate_to(
                                                    Screen::DockerCredentialEdit,
                                                );
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
                    ))
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
            .into_any()
        },
    )
    .style(|s| s.width_full().height_full())
}

fn docker_cred_create_view(state: &'static AppState) -> impl IntoView {
    form_screen(
        "Create Docker Credential",
        state,
        form_section(
            "Credential Details",
            v_stack((
                text_field("Name", state.form.docker_cred_name),
                text_field_with_placeholder(
                    "Registry",
                    state.form.docker_cred_registry,
                    "docker.io",
                ),
                text_field("Username", state.form.docker_cred_username),
                text_field("Token / Password", state.form.docker_cred_token),
            )),
        ),
        move || {
            commands::submit_create_docker_credential(state);
        },
        "Create Credential",
    )
}

fn docker_cred_edit_view(state: &'static AppState) -> impl IntoView {
    let initialized = state.form.edit_form_initialized;

    dyn_container(
        move || (initialized.get(), state.selected_docker_credential_id.get()),
        move |(is_init, cred_id_opt)| {
            if !is_init {
                if let Some(cred_id) = cred_id_opt {
                    let settings = state.settings.get_untracked();
                    if let Some(cred) = settings.get_docker_credential_by_id(cred_id) {
                        state.form.init_from_docker_credential(cred);
                        state.form.edit_form_initialized.set(true);
                    }
                }
            }

            form_screen(
                "Edit Docker Credential",
                state,
                form_section(
                    "Credential Details",
                    v_stack((
                        text_field("Name", state.form.docker_cred_name),
                        text_field_with_placeholder(
                            "Registry",
                            state.form.docker_cred_registry,
                            "docker.io",
                        ),
                        text_field("Username", state.form.docker_cred_username),
                        text_field("Token / Password", state.form.docker_cred_token),
                    )),
                ),
                move || {
                    commands::submit_update_docker_credential(state);
                },
                "Save Changes",
            )
            .into_any()
        },
    )
    .style(|s| s.width_full().height_full())
}
