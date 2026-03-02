use floem::prelude::*;

use crate::ui::styles::*;

/// Help screen with keyboard shortcuts and usage info.
pub fn help_screen() -> impl IntoView {
    let help_text = "\
Archon - Site Deployment Manager

Navigation:
  Use the sidebar tabs to switch between sections.

Sites:
  Create and manage Docker container deployments.
  Each site maps to a domain and runs on a node.

Domains:
  Configure domains with DNS providers (Cloudflare, Route53, or Manual).
  DNS records are managed automatically during deployment.

Nodes:
  Nodes are servers running the Archon node agent.
  They handle Docker container management and reverse proxy setup.

Docker Credentials:
  Store registry credentials for private Docker images.
  Credentials are encrypted before being sent to nodes.

Settings:
  Configure global defaults, API tokens, and health check intervals.";

    v_stack((
        "Help".style(|s| {
            s.font_size(FONT_SIZE_TITLE)
                .color(TEXT_PRIMARY)
                .font_bold()
                .margin_bottom(SPACING_XL)
        }),
        help_text
            .to_string()
            .style(|s| {
                s.width_full()
                    .padding(SPACING_LG)
                    .background(BG_SECONDARY)
                    .border_radius(BORDER_RADIUS)
                    .color(TEXT_SECONDARY)
                    .font_size(FONT_SIZE_MD)
                    .line_height(1.6)
            }),
    ))
    .style(|s| {
        s.width_full()
            .height_full()
            .padding(SPACING_XL)
            .flex_col()
    })
}
