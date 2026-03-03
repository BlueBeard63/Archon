use floem::prelude::*;
use floem::reactive::RwSignal;

use crate::ui::styles::*;

/// A labeled text input field.
pub fn text_field(label_text: &str, value: RwSignal<String>) -> impl IntoView {
    let label_text = label_text.to_string();
    v_stack((
        label_text.style(|s| {
            s.font_size(FONT_SIZE_MD)
                .color(TEXT_SECONDARY)
                .margin_bottom(SPACING_SM)
        }),
        text_input(value).style(|s| {
            s.width_full()
                .height(INPUT_HEIGHT)
                .font_size(FONT_SIZE_MD)
        }),
    ))
    .style(|s| s.width_full().margin_bottom(SPACING_LG))
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
            s.font_size(FONT_SIZE_MD)
                .color(TEXT_SECONDARY)
                .margin_bottom(SPACING_SM)
        }),
        text_input(value)
            .placeholder(placeholder)
            .style(|s| {
                s.width_full()
                    .height(INPUT_HEIGHT)
                    .font_size(FONT_SIZE_MD)
            }),
    ))
    .style(|s| s.width_full().margin_bottom(SPACING_LG))
}

/// A section header to visually group related fields.
pub fn form_section(title: &str, content: impl IntoView + 'static) -> impl IntoView {
    let title = title.to_string();
    v_stack((
        title.style(|s| {
            s.font_size(FONT_SIZE_LG)
                .color(TEXT_PRIMARY)
                .font_bold()
                .margin_bottom(SPACING_MD)
        }),
        content,
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

/// A read-only display field (label + value in a muted box).
pub fn readonly_field(label_text: &str, value: &str) -> impl IntoView {
    let label_text = label_text.to_string();
    let value = value.to_string();
    v_stack((
        label_text.style(|s| {
            s.font_size(FONT_SIZE_MD)
                .color(TEXT_SECONDARY)
                .margin_bottom(SPACING_SM)
        }),
        value.style(|s| {
            s.width_full()
                .height(INPUT_HEIGHT)
                .padding_horiz(INPUT_PADDING)
                .items_center()
                .background(BG_SECONDARY)
                .color(TEXT_MUTED)
                .border(1.0)
                .border_color(BORDER_MUTED)
                .border_radius(BORDER_RADIUS)
                .font_size(FONT_SIZE_MD)
        }),
    ))
    .style(|s| s.width_full().margin_bottom(SPACING_LG))
}
