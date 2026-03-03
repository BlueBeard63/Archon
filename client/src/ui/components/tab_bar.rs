use floem::prelude::*;
use floem::style::CursorStyle;
use floem::views::v_stack_from_iter;

use crate::logic::{AppState, TAB_LABELS};
use crate::ui::styles::*;

/// Vertical sidebar tab bar.
pub fn tab_bar(state: &'static AppState) -> impl IntoView {
    let active_tab = state.navigation.active_tab;

    let tabs: Vec<_> = TAB_LABELS
        .iter()
        .enumerate()
        .map(|(i, &label)| {
            label
                .style(move |s| {
                    let is_active = active_tab.get() == i;
                    s.width_full()
                        .padding_vert(SPACING_SM)
                        .padding_horiz(SPACING_LG)
                        .font_size(FONT_SIZE_MD)
                        .color(if is_active {
                            TEXT_PRIMARY
                        } else {
                            TEXT_SECONDARY
                        })
                        .background(if is_active {
                            BG_ELEVATED
                        } else {
                            BG_SECONDARY
                        })
                        .border_left(if is_active { 3.0 } else { 0.0 })
                        .border_color(ACCENT_BLUE)
                        .cursor(CursorStyle::Pointer)
                        .hover(|s| s.background(BG_HOVER).color(TEXT_PRIMARY))
                })
                .on_click_stop(move |_| {
                    state.navigation.navigate_to_tab(i);
                })
                .into_any()
        })
        .collect();

    v_stack_from_iter(tabs)
        .style(|s| {
            s.width(SIDEBAR_WIDTH)
                .height_full()
                .background(BG_SECONDARY)
                .border_right(1.0)
                .border_color(BORDER_DEFAULT)
                .padding_top(SPACING_LG)
                .gap(SPACING_XS)
                .flex_col()
        })
}
