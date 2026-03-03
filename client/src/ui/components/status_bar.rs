use floem::prelude::*;

use crate::logic::AppState;
use crate::ui::styles::*;

/// Bottom status bar showing current screen and latest notification.
pub fn status_bar(state: &'static AppState) -> impl IntoView {
    let current_screen = state.navigation.current_screen;
    let notifications = state.notifications.notifications;

    h_stack((
        // Screen name
        dyn_view(move || {
            let screen = current_screen.get();
            screen.to_string()
        })
        .style(|s| {
            s.font_size(FONT_SIZE_SM)
                .color(TEXT_SECONDARY)
                .padding_horiz(SPACING_LG)
        }),
        // Spacer
        empty().style(|s| s.flex_grow(1.0)),
        // Latest notification
        dyn_view(move || {
            let notifs = notifications.get();
            notifs
                .last()
                .map(|n| n.message.clone())
                .unwrap_or_default()
        })
        .style(|s| {
            s.font_size(FONT_SIZE_SM)
                .color(TEXT_MUTED)
                .padding_horiz(SPACING_LG)
                .text_ellipsis()
                .max_width_pct(60.0)
        }),
    ))
    .style(|s| {
        s.width_full()
            .height(STATUS_BAR_HEIGHT)
            .background(BG_SECONDARY)
            .border_top(1.0)
            .border_color(BORDER_DEFAULT)
            .items_center()
    })
}
