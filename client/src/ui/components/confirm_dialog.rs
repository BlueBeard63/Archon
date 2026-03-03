use std::rc::Rc;

use floem::prelude::*;
use floem::style::CursorStyle;

use crate::logic::AppState;
use crate::ui::components::button::secondary_button;
use crate::ui::styles::*;

/// A deletion confirmation dialog.
///
/// Shows a warning asking the user to type the entity name to confirm deletion.
pub fn delete_confirm_dialog(
    state: &'static AppState,
    on_confirm: impl Fn() + 'static,
    on_cancel: impl Fn() + 'static,
) -> impl IntoView {
    let target_name = state.form.deletion_target_name;
    let target_type = state.form.deletion_target_type;
    let confirm_input = state.form.deletion_confirm_input;
    let on_confirm = Rc::new(on_confirm);

    v_stack((
        // Warning icon + header
        dyn_container(
            move || (target_type.get(), target_name.get()),
            move |(entity_type, name)| {
                v_stack((
                    format!("Delete {} '{}'?", entity_type, name)
                        .style(|s| {
                            s.font_size(FONT_SIZE_XL)
                                .color(ACCENT_RED)
                                .font_bold()
                        }),
                    "This action cannot be undone."
                        .style(|s| {
                            s.font_size(FONT_SIZE_MD)
                                .color(TEXT_MUTED)
                                .margin_top(SPACING_SM)
                        }),
                ))
                .style(|s| s.margin_bottom(SPACING_XL))
                .into_any()
            },
        ),
        // Show what to type
        dyn_container(
            move || target_name.get(),
            move |name| {
                format!("Type '{}' to confirm:", name)
                    .style(|s| {
                        s.font_size(FONT_SIZE_MD)
                            .color(TEXT_SECONDARY)
                            .margin_bottom(SPACING_SM)
                    })
                    .into_any()
            },
        ),
        // Input field
        text_input(confirm_input).style(|s| {
            s.width_full()
                .max_width(420.0)
                .height(INPUT_HEIGHT)
                .padding_horiz(INPUT_PADDING)
                .padding_vert(SPACING_SM)
                .background(BG_ELEVATED)
                .color(TEXT_PRIMARY)
                .border(1.0)
                .border_color(BORDER_DEFAULT)
                .border_radius(BORDER_RADIUS)
                .font_size(FONT_SIZE_MD)
                .margin_bottom(SPACING_XL)
                .focus(|s| s.border_color(ACCENT_RED))
        }),
        // Buttons
        h_stack((
            secondary_button("Cancel", on_cancel),
            // Delete button - styled based on confirmation state
            dyn_container(
                move || {
                    let input = confirm_input.get();
                    let target = target_name.get();
                    !input.is_empty() && input == target
                },
                {
                    let on_confirm = on_confirm.clone();
                    move |confirmed| {
                        let label = "Delete".to_string();
                        if confirmed {
                            let cb = on_confirm.clone();
                            label
                                .style(|s| {
                                    s.padding_vert(SPACING_SM)
                                        .padding_horiz(SPACING_XL)
                                        .min_height(36.0)
                                        .background(ACCENT_RED)
                                        .color(TEXT_PRIMARY)
                                        .border_radius(BORDER_RADIUS)
                                        .font_size(FONT_SIZE_MD)
                                        .font_bold()
                                        .cursor(CursorStyle::Pointer)
                                        .hover(|s| s.background(Color::rgb8(220, 50, 50)))
                                })
                                .on_click_stop(move |_| cb())
                                .into_any()
                        } else {
                            label
                                .style(|s| {
                                    s.padding_vert(SPACING_SM)
                                        .padding_horiz(SPACING_XL)
                                        .min_height(36.0)
                                        .background(BG_ELEVATED)
                                        .color(TEXT_MUTED)
                                        .border_radius(BORDER_RADIUS)
                                        .font_size(FONT_SIZE_MD)
                                })
                                .into_any()
                        }
                    }
                },
            ),
        ))
        .style(|s| s.gap(SPACING_MD)),
    ))
    .style(|s| {
        s.width_full()
            .max_width(520.0)
            .padding(SPACING_2XL)
            .background(BG_SECONDARY)
            .border_radius(BORDER_RADIUS_MD)
            .border(1.0)
            .border_color(BORDER_DEFAULT)
    })
}
