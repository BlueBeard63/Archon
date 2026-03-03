use floem::prelude::*;

use crate::logic::AppState;
use crate::ui::components::button::{primary_button, secondary_button};
use crate::ui::styles::*;

/// A standard form screen with back button, title, scrollable content, and submit/cancel buttons.
pub fn form_screen(
    title: &str,
    state: &'static AppState,
    content: impl IntoView + 'static,
    on_submit: impl Fn() + 'static,
    submit_label: &str,
) -> impl IntoView {
    let title = title.to_string();
    let submit_label = submit_label.to_string();

    v_stack((
        // Top: Back button + title
        h_stack((
            secondary_button("Back", move || {
                state.form.reset();
                state.navigation.navigate_back();
            }),
            title.style(|s| {
                s.font_size(FONT_SIZE_TITLE)
                    .color(TEXT_PRIMARY)
                    .font_bold()
                    .margin_left(SPACING_LG)
            }),
        ))
        .style(|s| {
            s.width_full()
                .items_center()
                .margin_bottom(SPACING_XL)
                .padding_bottom(SPACING_LG)
                .border_bottom(1.0)
                .border_color(BORDER_MUTED)
        }),
        // Middle: scrollable content, constrained width, centered
        scroll(
            v_stack((
                v_stack((content,)).style(|s| {
                    s.width_full()
                        .max_width(FORM_MAX_WIDTH)
                        .padding_vert(SPACING_SM)
                }),
            ))
            .style(|s| s.width_full().items_center()),
        )
        .style(|s| s.width_full().flex_grow(1.0).min_height(0.0)),
        // Bottom: Cancel + Submit with separator
        h_stack((
            secondary_button("Cancel", move || {
                state.form.reset();
                state.navigation.navigate_back();
            }),
            primary_button(&submit_label, on_submit),
        ))
        .style(|s| {
            s.width_full()
                .gap(SPACING_MD)
                .margin_top(SPACING_LG)
                .padding_top(SPACING_LG)
                .border_top(1.0)
                .border_color(BORDER_MUTED)
                .justify_end()
        }),
    ))
    .style(|s| {
        s.width_full()
            .height_full()
            .padding(SPACING_XL)
            .flex_col()
    })
}
