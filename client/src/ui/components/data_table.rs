use floem::peniko::Color;
use floem::prelude::*;
use floem::style::CursorStyle;
use floem::views::h_stack_from_iter;

use crate::ui::styles::*;

/// A single row in a data table.
pub fn table_row(
    columns: Vec<(String, f64)>,
    is_selected: bool,
    on_click: impl Fn() + 'static,
) -> impl IntoView {
    let cells: Vec<_> = columns
        .into_iter()
        .map(|(text, width)| {
            text.style(move |s| {
                s.width(width)
                    .padding_horiz(SPACING_SM)
                    .font_size(FONT_SIZE_MD)
                    .color(TEXT_PRIMARY)
                    .text_ellipsis()
            })
            .into_any()
        })
        .collect();

    h_stack_from_iter(cells)
        .style(move |s| {
            s.width_full()
                .padding_vert(SPACING_SM)
                .border_bottom(1.0)
                .border_color(BORDER_MUTED)
                .background(if is_selected {
                    BG_ELEVATED
                } else {
                    BG_PRIMARY
                })
                .cursor(CursorStyle::Pointer)
                .hover(|s| s.background(BG_HOVER))
        })
        .on_click_stop(move |_| on_click())
}

/// A table row with inline action buttons.
pub fn table_row_with_actions(
    columns: Vec<(String, f64)>,
    actions: Vec<(String, Color, Box<dyn Fn() + 'static>)>,
    on_click: impl Fn() + 'static,
) -> impl IntoView {
    let mut cells: Vec<_> = columns
        .into_iter()
        .map(|(text, width)| {
            text.style(move |s| {
                s.width(width)
                    .padding_horiz(SPACING_SM)
                    .font_size(FONT_SIZE_MD)
                    .color(TEXT_PRIMARY)
                    .text_ellipsis()
            })
            .into_any()
        })
        .collect();

    // Actions column
    let action_buttons: Vec<_> = actions
        .into_iter()
        .map(|(label, color, handler)| {
            label
                .style(move |s| {
                    s.padding_horiz(SPACING_SM)
                        .padding_vert(2.0)
                        .font_size(FONT_SIZE_SM)
                        .color(color)
                        .cursor(CursorStyle::Pointer)
                        .hover(|s| s.background(BG_HOVER).border_radius(BORDER_RADIUS_SM))
                })
                .on_click_stop(move |_| handler())
                .into_any()
        })
        .collect();

    cells.push(
        h_stack_from_iter(action_buttons)
            .style(|s| s.items_center().gap(SPACING_XS))
            .into_any(),
    );

    h_stack_from_iter(cells)
        .style(move |s| {
            s.width_full()
                .padding_vert(SPACING_SM)
                .border_bottom(1.0)
                .border_color(BORDER_MUTED)
                .background(BG_PRIMARY)
                .items_center()
                .cursor(CursorStyle::Pointer)
                .hover(|s| s.background(BG_HOVER))
        })
        .on_click_stop(move |_| on_click())
}

/// Table header row.
pub fn table_header(columns: Vec<(&str, f64)>) -> impl IntoView {
    let cells: Vec<_> = columns
        .into_iter()
        .map(|(text, width)| {
            let text = text.to_string();
            text.style(move |s| {
                s.width(width)
                    .padding_horiz(SPACING_SM)
                    .font_size(FONT_SIZE_SM)
                    .color(TEXT_MUTED)
                    .font_bold()
            })
            .into_any()
        })
        .collect();

    h_stack_from_iter(cells).style(|s| {
        s.width_full()
            .padding_vert(SPACING_SM)
            .border_bottom(1.0)
            .border_color(BORDER_DEFAULT)
            .background(BG_SECONDARY)
    })
}

/// Empty state message for when a list has no items.
pub fn empty_state(message: &str) -> impl IntoView {
    let message = message.to_string();
    message.style(|s| {
        s.width_full()
            .padding(SPACING_XL)
            .font_size(FONT_SIZE_MD)
            .color(TEXT_MUTED)
            .justify_center()
            .items_center()
    })
}
