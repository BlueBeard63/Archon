use floem::prelude::*;
use floem::reactive::RwSignal;
use floem::style::CursorStyle;
use floem::views::v_stack_from_iter;

use crate::logic::app_state::AppState;
use crate::ui::styles::*;

/// Shared input style for inline inputs in multi-pair rows.
fn pair_input_style(s: floem::style::Style) -> floem::style::Style {
    s.height(INPUT_HEIGHT)
        .padding_horiz(INPUT_PADDING)
        .padding_vert(SPACING_SM)
        .background(BG_ELEVATED)
        .color(TEXT_PRIMARY)
        .border(1.0)
        .border_color(BORDER_DEFAULT)
        .border_radius(BORDER_RADIUS)
        .font_size(FONT_SIZE_MD)
        .focus(|s| s.border_color(ACCENT_BLUE))
}

/// A styled remove button for pair rows.
fn remove_button(on_click: impl Fn() + 'static) -> impl IntoView {
    "Remove"
        .style(|s| {
            s.padding_vert(SPACING_XS)
                .padding_horiz(SPACING_SM)
                .color(ACCENT_RED)
                .font_size(FONT_SIZE_SM)
                .border_radius(BORDER_RADIUS_SM)
                .border(1.0)
                .border_color(Color::TRANSPARENT)
                .cursor(CursorStyle::Pointer)
                .hover(|s| {
                    s.background(ACCENT_RED_MUTED)
                        .color(TEXT_PRIMARY)
                        .border_color(ACCENT_RED)
                })
        })
        .on_click_stop(move |_| on_click())
}

/// A styled add button.
fn add_button(label: &str, on_click: impl Fn() + 'static) -> impl IntoView {
    let label = format!("+ {}", label);
    label
        .style(|s| {
            s.padding_vert(SPACING_SM)
                .padding_horiz(SPACING_LG)
                .color(ACCENT_BLUE)
                .font_size(FONT_SIZE_MD)
                .border_radius(BORDER_RADIUS)
                .border(1.0)
                .border_color(ACCENT_BLUE)
                .cursor(CursorStyle::Pointer)
                .margin_top(SPACING_MD)
                .hover(|s| s.background(BG_ELEVATED).color(ACCENT_BLUE_HOVER))
        })
        .on_click_stop(move |_| on_click())
}

/// Section label for multi-pair editors.
fn section_label(text: &str) -> impl IntoView {
    text.to_string().style(|s| {
        s.font_size(FONT_SIZE_LG)
            .color(TEXT_PRIMARY)
            .font_bold()
            .margin_bottom(SPACING_MD)
    })
}

/// Column header label for multi-pair rows.
fn column_header(text: &str, width: f64) -> impl IntoView {
    text.to_string().style(move |s| {
        s.width(width)
            .font_size(FONT_SIZE_SM)
            .color(TEXT_MUTED)
            .margin_bottom(SPACING_XS)
    })
}

/// Environment variable key-value pair editor.
pub fn env_var_editor(state: &'static AppState) -> impl IntoView {
    let pairs = state.form.env_var_pairs;

    v_stack((
        section_label("Environment Variables"),
        // Column headers
        h_stack((
            column_header("Key", 0.0),
            empty().style(|s| s.flex_grow(1.0)),
            column_header("Value", 0.0),
            empty().style(|s| s.flex_grow(1.0)),
            empty().style(|s| s.width(70.0)),
        ))
        .style(|s| s.width_full().padding_horiz(SPACING_XS)),
        dyn_container(
            move || pairs.get(),
            move |pair_list| {
                let rows: Vec<_> = pair_list
                    .iter()
                    .enumerate()
                    .map(|(idx, pair)| {
                        let key_signal = RwSignal::new(pair.key.clone());
                        let val_signal = RwSignal::new(pair.value.clone());

                        h_stack((
                            text_input(key_signal)
                                .placeholder("KEY")
                                .style(|s| pair_input_style(s).flex_grow(1.0))
                                .on_event_stop(
                                    floem::event::EventListener::FocusLost,
                                    move |_| {
                                        let v = key_signal.get_untracked();
                                        pairs.update(|p| {
                                            if let Some(pair) = p.get_mut(idx) {
                                                pair.key = v;
                                            }
                                        });
                                    },
                                ),
                            text_input(val_signal)
                                .placeholder("value")
                                .style(|s| {
                                    pair_input_style(s).flex_grow(1.0).margin_left(SPACING_SM)
                                })
                                .on_event_stop(
                                    floem::event::EventListener::FocusLost,
                                    move |_| {
                                        let v = val_signal.get_untracked();
                                        pairs.update(|p| {
                                            if let Some(pair) = p.get_mut(idx) {
                                                pair.value = v;
                                            }
                                        });
                                    },
                                ),
                            remove_button(move || {
                                state.form.remove_env_var(idx);
                            })
                            .style(|s| s.margin_left(SPACING_SM)),
                        ))
                        .style(|s| {
                            s.width_full()
                                .items_center()
                                .margin_bottom(SPACING_SM)
                        })
                        .into_any()
                    })
                    .collect();

                v_stack_from_iter(rows)
                    .style(|s| s.width_full())
                    .into_any()
            },
        ),
        add_button("Add Variable", move || {
            state.form.add_env_var();
        }),
    ))
    .style(|s| {
        s.width_full()
            .padding(SPACING_LG)
            .background(BG_CARD)
            .border_radius(BORDER_RADIUS_MD)
            .border(1.0)
            .border_color(BORDER_MUTED)
            .margin_bottom(SPACING_XL)
    })
}

/// Domain mapping editor: subdomain, domain dropdown, port.
pub fn domain_mapping_editor(state: &'static AppState) -> impl IntoView {
    let pairs = state.form.domain_mapping_pairs;
    let domains = state.domains;

    v_stack((
        section_label("Domain Mappings"),
        // Column headers
        h_stack((
            "Subdomain"
                .style(|s| s.width(140.0).font_size(FONT_SIZE_SM).color(TEXT_MUTED)),
            "Domain"
                .style(|s| {
                    s.width(180.0)
                        .font_size(FONT_SIZE_SM)
                        .color(TEXT_MUTED)
                        .margin_left(SPACING_SM)
                }),
            "Port"
                .style(|s| {
                    s.width(90.0)
                        .font_size(FONT_SIZE_SM)
                        .color(TEXT_MUTED)
                        .margin_left(SPACING_SM)
                }),
        ))
        .style(|s| s.width_full().margin_bottom(SPACING_XS)),
        dyn_container(
            move || (pairs.get(), domains.get()),
            move |(pair_list, domain_list)| {
                let rows: Vec<_> = pair_list
                    .iter()
                    .enumerate()
                    .map(|(idx, pair)| {
                        let sub_signal = RwSignal::new(pair.subdomain.clone());
                        let domain_signal = RwSignal::new(pair.domain_id.clone());
                        let port_signal = RwSignal::new(pair.port.clone());

                        let domain_list_clone = domain_list.clone();

                        h_stack((
                            // Subdomain input
                            text_input(sub_signal)
                                .placeholder("www")
                                .style(|s| pair_input_style(s).width(140.0))
                                .on_event_stop(
                                    floem::event::EventListener::FocusLost,
                                    move |_| {
                                        let v = sub_signal.get_untracked();
                                        pairs.update(|p| {
                                            if let Some(pair) = p.get_mut(idx) {
                                                pair.subdomain = v;
                                            }
                                        });
                                    },
                                ),
                            // Domain selector
                            dyn_container(
                                move || domain_signal.get(),
                                {
                                    let dl = domain_list_clone.clone();
                                    move |current_id: String| {
                                        let display = if current_id.is_empty() {
                                            "Select domain...".to_string()
                                        } else if let Ok(uuid) =
                                            uuid::Uuid::parse_str(&current_id)
                                        {
                                            dl.iter()
                                                .find(|d| d.id == uuid)
                                                .map(|d| d.name.clone())
                                                .unwrap_or_else(|| "Unknown".to_string())
                                        } else {
                                            current_id.clone()
                                        };
                                        let open = RwSignal::new(false);
                                        let dl2 = dl.clone();

                                        v_stack((
                                            display
                                                .clone()
                                                .style(move |s| {
                                                    s.width(180.0)
                                                        .height(INPUT_HEIGHT)
                                                        .padding_horiz(INPUT_PADDING)
                                                        .padding_vert(SPACING_SM)
                                                        .background(BG_ELEVATED)
                                                        .color(if display == "Select domain..." {
                                                            TEXT_MUTED
                                                        } else {
                                                            TEXT_PRIMARY
                                                        })
                                                        .border(1.0)
                                                        .border_color(BORDER_DEFAULT)
                                                        .border_radius(BORDER_RADIUS)
                                                        .font_size(FONT_SIZE_MD)
                                                        .cursor(CursorStyle::Pointer)
                                                        .margin_left(SPACING_SM)
                                                        .text_ellipsis()
                                                        .hover(|s| s.border_color(ACCENT_BLUE))
                                                })
                                                .on_click_stop(move |_| {
                                                    open.set(!open.get_untracked());
                                                }),
                                            dyn_container(move || open.get(), {
                                                let dl3 = dl2.clone();
                                                move |is_open: bool| {
                                                    if !is_open {
                                                        return empty().into_any();
                                                    }
                                                    let items: Vec<_> = dl3
                                                        .iter()
                                                        .map(|d| {
                                                            let d_name = d.name.clone();
                                                            let d_id = d.id.to_string();
                                                            d_name
                                                                .clone()
                                                                .style(|s| {
                                                                    s.width_full()
                                                                        .padding_horiz(
                                                                            INPUT_PADDING,
                                                                        )
                                                                        .padding_vert(SPACING_SM)
                                                                        .min_height(34.0)
                                                                        .background(BG_ELEVATED)
                                                                        .color(TEXT_PRIMARY)
                                                                        .font_size(FONT_SIZE_MD)
                                                                        .cursor(
                                                                            CursorStyle::Pointer,
                                                                        )
                                                                        .hover(|s| {
                                                                            s.background(BG_HOVER)
                                                                        })
                                                                })
                                                                .on_click_stop(move |_| {
                                                                    domain_signal
                                                                        .set(d_id.clone());
                                                                    pairs.update(|p| {
                                                                        if let Some(pair) =
                                                                            p.get_mut(idx)
                                                                        {
                                                                            pair.domain_id =
                                                                                domain_signal
                                                                                    .get_untracked(
                                                                                    );
                                                                            pair.domain_name =
                                                                                d_name.clone();
                                                                        }
                                                                    });
                                                                    open.set(false);
                                                                })
                                                                .into_any()
                                                        })
                                                        .collect();
                                                    v_stack_from_iter(items)
                                                        .style(|s| {
                                                            s.width(180.0)
                                                                .border(1.0)
                                                                .border_color(ACCENT_BLUE)
                                                                .border_radius(BORDER_RADIUS)
                                                                .background(BG_ELEVATED)
                                                                .margin_left(SPACING_SM)
                                                                .margin_top(SPACING_XS)
                                                        })
                                                        .into_any()
                                                }
                                            }),
                                        ))
                                        .into_any()
                                    }
                                },
                            ),
                            // Port input
                            text_input(port_signal)
                                .placeholder("80")
                                .style(|s| pair_input_style(s).width(90.0).margin_left(SPACING_SM))
                                .on_event_stop(
                                    floem::event::EventListener::FocusLost,
                                    move |_| {
                                        let v = port_signal.get_untracked();
                                        pairs.update(|p| {
                                            if let Some(pair) = p.get_mut(idx) {
                                                pair.port = v;
                                            }
                                        });
                                    },
                                ),
                            // Remove button
                            remove_button(move || {
                                state.form.remove_domain_mapping(idx);
                            })
                            .style(|s| s.margin_left(SPACING_SM)),
                        ))
                        .style(|s| s.width_full().items_center().margin_bottom(SPACING_SM))
                        .into_any()
                    })
                    .collect();

                v_stack_from_iter(rows)
                    .style(|s| s.width_full())
                    .into_any()
            },
        ),
        add_button("Add Mapping", move || {
            state.form.add_domain_mapping();
        }),
    ))
    .style(|s| {
        s.width_full()
            .padding(SPACING_LG)
            .background(BG_CARD)
            .border_radius(BORDER_RADIUS_MD)
            .border(1.0)
            .border_color(BORDER_MUTED)
            .margin_bottom(SPACING_XL)
    })
}

/// Volume bind mount editor: host_path, container_path.
pub fn volume_editor(state: &'static AppState) -> impl IntoView {
    let pairs = state.form.volume_pairs;

    v_stack((
        section_label("Volumes"),
        // Column headers
        h_stack((
            "Host Path"
                .style(|s| s.font_size(FONT_SIZE_SM).color(TEXT_MUTED)),
            empty().style(|s| s.flex_grow(1.0)),
            "Container Path"
                .style(|s| {
                    s.font_size(FONT_SIZE_SM)
                        .color(TEXT_MUTED)
                        .margin_left(SPACING_SM)
                }),
            empty().style(|s| s.flex_grow(1.0)),
            empty().style(|s| s.width(70.0)),
        ))
        .style(|s| s.width_full().margin_bottom(SPACING_XS)),
        dyn_container(
            move || pairs.get(),
            move |pair_list| {
                if pair_list.is_empty() {
                    return "No volumes configured."
                        .style(|s| {
                            s.font_size(FONT_SIZE_MD)
                                .color(TEXT_MUTED)
                                .padding_vert(SPACING_SM)
                        })
                        .into_any();
                }

                let rows: Vec<_> = pair_list
                    .iter()
                    .enumerate()
                    .map(|(idx, pair)| {
                        let host_signal = RwSignal::new(pair.host_path.clone());
                        let container_signal = RwSignal::new(pair.container_path.clone());

                        h_stack((
                            text_input(host_signal)
                                .placeholder("/host/path")
                                .style(|s| pair_input_style(s).flex_grow(1.0))
                                .on_event_stop(
                                    floem::event::EventListener::FocusLost,
                                    move |_| {
                                        let v = host_signal.get_untracked();
                                        pairs.update(|p| {
                                            if let Some(pair) = p.get_mut(idx) {
                                                pair.host_path = v;
                                            }
                                        });
                                    },
                                ),
                            text_input(container_signal)
                                .placeholder("/container/path")
                                .style(|s| {
                                    pair_input_style(s).flex_grow(1.0).margin_left(SPACING_SM)
                                })
                                .on_event_stop(
                                    floem::event::EventListener::FocusLost,
                                    move |_| {
                                        let v = container_signal.get_untracked();
                                        pairs.update(|p| {
                                            if let Some(pair) = p.get_mut(idx) {
                                                pair.container_path = v;
                                            }
                                        });
                                    },
                                ),
                            remove_button(move || {
                                state.form.remove_volume(idx);
                            })
                            .style(|s| s.margin_left(SPACING_SM)),
                        ))
                        .style(|s| {
                            s.width_full()
                                .items_center()
                                .margin_bottom(SPACING_SM)
                        })
                        .into_any()
                    })
                    .collect();

                v_stack_from_iter(rows)
                    .style(|s| s.width_full())
                    .into_any()
            },
        ),
        add_button("Add Volume", move || {
            state.form.add_volume();
        }),
    ))
    .style(|s| {
        s.width_full()
            .padding(SPACING_LG)
            .background(BG_CARD)
            .border_radius(BORDER_RADIUS_MD)
            .border(1.0)
            .border_color(BORDER_MUTED)
            .margin_bottom(SPACING_XL)
    })
}
