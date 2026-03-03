use floem::prelude::*;
use floem::style::CursorStyle;

use crate::ui::styles::*;

/// Primary action button (blue).
pub fn primary_button(label: &str, on_click: impl Fn() + 'static) -> impl IntoView {
    let label = label.to_string();
    label
        .style(|s| {
            s.padding_vert(SPACING_SM)
                .padding_horiz(SPACING_XL)
                .min_height(36.0)
                .background(ACCENT_BLUE)
                .color(TEXT_PRIMARY)
                .border_radius(BORDER_RADIUS)
                .font_size(FONT_SIZE_MD)
                .font_bold()
                .cursor(CursorStyle::Pointer)
                .hover(|s| s.background(ACCENT_BLUE_HOVER))
                .items_center()
                .justify_center()
        })
        .on_click_stop(move |_| on_click())
}

/// Secondary/ghost button (outlined).
pub fn secondary_button(label: &str, on_click: impl Fn() + 'static) -> impl IntoView {
    let label = label.to_string();
    label
        .style(|s| {
            s.padding_vert(SPACING_SM)
                .padding_horiz(SPACING_XL)
                .min_height(36.0)
                .background(BG_SECONDARY)
                .color(TEXT_SECONDARY)
                .border(1.0)
                .border_color(BORDER_DEFAULT)
                .border_radius(BORDER_RADIUS)
                .font_size(FONT_SIZE_MD)
                .cursor(CursorStyle::Pointer)
                .hover(|s| s.background(BG_ELEVATED).color(TEXT_PRIMARY))
                .items_center()
                .justify_center()
        })
        .on_click_stop(move |_| on_click())
}

/// Danger button (red).
pub fn danger_button(label: &str, on_click: impl Fn() + 'static) -> impl IntoView {
    let label = label.to_string();
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
                .items_center()
                .justify_center()
        })
        .on_click_stop(move |_| on_click())
}

/// Small inline action button for use in table rows etc.
pub fn inline_button(
    label: &str,
    color: floem::peniko::Color,
    on_click: impl Fn() + 'static,
) -> impl IntoView {
    let label = label.to_string();
    label
        .style(move |s| {
            s.padding_vert(SPACING_XS)
                .padding_horiz(SPACING_SM)
                .color(color)
                .font_size(FONT_SIZE_SM)
                .border_radius(BORDER_RADIUS_SM)
                .cursor(CursorStyle::Pointer)
                .hover(|s| s.background(BG_HOVER))
        })
        .on_click_stop(move |_| on_click())
}
