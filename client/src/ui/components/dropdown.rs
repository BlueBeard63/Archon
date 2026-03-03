use floem::prelude::*;
use floem::reactive::RwSignal;
use floem::style::CursorStyle;
use floem::views::v_stack_from_iter;

use crate::ui::styles::*;

/// A custom dropdown/select component.
pub fn dropdown(
    label: &str,
    selected: RwSignal<String>,
    options: Vec<String>,
) -> impl IntoView {
    let label = label.to_string();
    let open = RwSignal::new(false);
    let options_clone = options.clone();

    v_stack((
        // Label
        label.style(|s| {
            s.font_size(FONT_SIZE_MD)
                .color(TEXT_SECONDARY)
                .margin_bottom(SPACING_SM)
        }),
        // Dropdown button + options
        v_stack((
            // Button showing current value
            dyn_container(
                move || (selected.get(), open.get()),
                move |(_val, is_open)| {
                    let current = selected.get_untracked();
                    let indicator = if is_open { " \u{25B2}" } else { " \u{25BC}" };
                    let display_text = if current.is_empty() {
                        "Select...".to_string()
                    } else {
                        current
                    };
                    let display = format!("{}{}", display_text, indicator);

                    display
                        .style(move |s| {
                            s.width_full()
                                .height(INPUT_HEIGHT)
                                .padding_horiz(INPUT_PADDING)
                                .padding_vert(SPACING_SM)
                                .background(BG_ELEVATED)
                                .color(TEXT_PRIMARY)
                                .border(1.0)
                                .border_color(if is_open { ACCENT_BLUE } else { BORDER_DEFAULT })
                                .border_radius(BORDER_RADIUS)
                                .font_size(FONT_SIZE_MD)
                                .cursor(CursorStyle::Pointer)
                                .hover(|s| s.border_color(ACCENT_BLUE))
                        })
                        .on_click_stop(move |_| {
                            open.set(!open.get_untracked());
                        })
                        .into_any()
                },
            ),
            // Options list (shown when open)
            dyn_container(
                move || open.get(),
                {
                    let options_for_list = options_clone.clone();
                    move |is_open: bool| {
                        if !is_open {
                            return empty().into_any();
                        }

                        let items: Vec<_> = options_for_list
                            .iter()
                            .map(|opt| {
                                let opt_value = opt.clone();
                                let opt_display = opt.clone();

                                opt_display
                                    .style(move |s| {
                                        let is_selected = selected.get_untracked() == opt_value;
                                        s.width_full()
                                            .padding_horiz(INPUT_PADDING)
                                            .padding_vert(SPACING_SM)
                                            .min_height(34.0)
                                            .background(if is_selected {
                                                BG_HOVER
                                            } else {
                                                BG_ELEVATED
                                            })
                                            .color(if is_selected {
                                                TEXT_PRIMARY
                                            } else {
                                                TEXT_SECONDARY
                                            })
                                            .font_size(FONT_SIZE_MD)
                                            .cursor(CursorStyle::Pointer)
                                            .hover(|s| {
                                                s.background(BG_HOVER).color(TEXT_PRIMARY)
                                            })
                                    })
                                    .on_click_stop({
                                        let value = opt.clone();
                                        move |_| {
                                            selected.set(value.clone());
                                            open.set(false);
                                        }
                                    })
                                    .into_any()
                            })
                            .collect();

                        v_stack_from_iter(items)
                            .style(|s| {
                                s.width_full()
                                    .border(1.0)
                                    .border_color(ACCENT_BLUE)
                                    .border_radius(BORDER_RADIUS)
                                    .background(BG_ELEVATED)
                                    .margin_top(SPACING_XS)
                            })
                            .into_any()
                    }
                },
            ),
        )),
    ))
    .style(|s| s.width_full().margin_bottom(SPACING_LG))
}
