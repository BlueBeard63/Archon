use floem::prelude::*;
use floem::reactive::RwSignal;

use crate::ui::styles::*;

/// A labeled text input field.
pub fn text_field(label_text: &str, value: RwSignal<String>) -> impl IntoView {
    let label_text = label_text.to_string();
    v_stack((
        label_text.style(|s| {
            s.font_size(FONT_SIZE_SM)
                .color(TEXT_SECONDARY)
                .margin_bottom(SPACING_XS)
        }),
        text_input(value).style(|s| {
            s.width_full()
                .padding(SPACING_SM)
                .background(BG_ELEVATED)
                .color(TEXT_PRIMARY)
                .border(1.0)
                .border_color(BORDER_DEFAULT)
                .border_radius(BORDER_RADIUS_SM)
                .font_size(FONT_SIZE_MD)
                .focus(|s| s.border_color(ACCENT_BLUE))
        }),
    ))
    .style(|s| s.width_full().margin_bottom(SPACING_MD))
}

/// A labeled text input with placeholder.
pub fn text_field_with_placeholder(
    label_text: &str,
    value: RwSignal<String>,
    placeholder: &str,
) -> impl IntoView {
    let label_text = label_text.to_string();
    let placeholder = placeholder.to_string();
    v_stack((
        label_text.style(|s| {
            s.font_size(FONT_SIZE_SM)
                .color(TEXT_SECONDARY)
                .margin_bottom(SPACING_XS)
        }),
        text_input(value)
            .placeholder(placeholder)
            .style(|s| {
                s.width_full()
                    .padding(SPACING_SM)
                    .background(BG_ELEVATED)
                    .color(TEXT_PRIMARY)
                    .border(1.0)
                    .border_color(BORDER_DEFAULT)
                    .border_radius(BORDER_RADIUS_SM)
                    .font_size(FONT_SIZE_MD)
                    .focus(|s| s.border_color(ACCENT_BLUE))
            }),
    ))
    .style(|s| s.width_full().margin_bottom(SPACING_MD))
}
