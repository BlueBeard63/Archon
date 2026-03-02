use floem::prelude::*;
use floem::reactive::RwSignal;

use crate::logic::AppState;
use crate::ui::styles::*;

/// Settings screen.
pub fn settings_screen(state: &'static AppState) -> impl IntoView {
    let settings = state.settings;

    v_stack((
        "Settings".style(|s| {
            s.font_size(FONT_SIZE_TITLE)
                .color(TEXT_PRIMARY)
                .font_bold()
                .margin_bottom(SPACING_XL)
        }),
        // Settings display
        dyn_view(move || {
            let s = settings.get();
            format!(
                "Auto Save: {}\n\
                 Health Check Interval: {}s\n\
                 Default DNS TTL: {}s\n\
                 Theme: {}\n\
                 Docker Credentials: {} configured\n\
                 Cloudflare Token: {}\n\
                 Route53 Keys: {}",
                if s.auto_save { "Enabled" } else { "Disabled" },
                s.health_check_interval_secs,
                s.default_dns_ttl,
                s.theme,
                s.docker_credentials.len(),
                if s.cloudflare_api_token.is_empty() {
                    "Not set"
                } else {
                    "Configured"
                },
                if s.route53_access_key.is_empty() {
                    "Not set"
                } else {
                    "Configured"
                },
            )
        })
        .style(|s| {
            s.width_full()
                .padding(SPACING_LG)
                .background(BG_SECONDARY)
                .border_radius(BORDER_RADIUS)
                .color(TEXT_PRIMARY)
                .font_size(FONT_SIZE_MD)
                .line_height(1.8)
        }),
    ))
    .style(|s| {
        s.width_full()
            .height_full()
            .padding(SPACING_XL)
            .flex_col()
    })
}
