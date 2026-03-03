use floem::peniko::Color;

// ---------------------------------------------------------------------------
// Color palette
// ---------------------------------------------------------------------------

pub const BG_PRIMARY: Color = Color::rgb8(24, 24, 27);       // zinc-900
pub const BG_SECONDARY: Color = Color::rgb8(39, 39, 42);     // zinc-800
pub const BG_ELEVATED: Color = Color::rgb8(52, 52, 56);      // zinc-700
pub const BG_HOVER: Color = Color::rgb8(63, 63, 70);         // zinc-600
pub const BG_CARD: Color = Color::rgb8(32, 32, 36);          // between primary and secondary

pub const TEXT_PRIMARY: Color = Color::rgb8(244, 244, 245);   // zinc-100
pub const TEXT_SECONDARY: Color = Color::rgb8(161, 161, 170); // zinc-400
pub const TEXT_MUTED: Color = Color::rgb8(113, 113, 122);     // zinc-500

pub const BORDER_DEFAULT: Color = Color::rgb8(63, 63, 70);    // zinc-600
pub const BORDER_MUTED: Color = Color::rgb8(52, 52, 56);      // zinc-700

pub const ACCENT_BLUE: Color = Color::rgb8(59, 130, 246);     // blue-500
pub const ACCENT_BLUE_HOVER: Color = Color::rgb8(96, 165, 250); // blue-400
pub const ACCENT_GREEN: Color = Color::rgb8(34, 197, 94);     // green-500
pub const ACCENT_RED: Color = Color::rgb8(239, 68, 68);       // red-500
pub const ACCENT_RED_MUTED: Color = Color::rgb8(153, 50, 50); // muted red for backgrounds
pub const ACCENT_YELLOW: Color = Color::rgb8(234, 179, 8);    // yellow-500
pub const ACCENT_ORANGE: Color = Color::rgb8(249, 115, 22);   // orange-500

// ---------------------------------------------------------------------------
// Status colors
// ---------------------------------------------------------------------------

pub fn status_color(status: &str) -> Color {
    match status {
        "running" | "online" => ACCENT_GREEN,
        "stopped" | "inactive" => TEXT_MUTED,
        "deploying" => ACCENT_BLUE,
        "failed" | "offline" => ACCENT_RED,
        "degraded" => ACCENT_ORANGE,
        _ => TEXT_SECONDARY,
    }
}

// ---------------------------------------------------------------------------
// Spacing constants
// ---------------------------------------------------------------------------

pub const SPACING_XS: f64 = 4.0;
pub const SPACING_SM: f64 = 8.0;
pub const SPACING_MD: f64 = 12.0;
pub const SPACING_LG: f64 = 16.0;
pub const SPACING_XL: f64 = 24.0;
pub const SPACING_2XL: f64 = 32.0;

pub const BORDER_RADIUS: f64 = 6.0;
pub const BORDER_RADIUS_SM: f64 = 4.0;
pub const BORDER_RADIUS_MD: f64 = 8.0;

pub const FONT_SIZE_SM: f32 = 12.0;
pub const FONT_SIZE_MD: f32 = 14.0;
pub const FONT_SIZE_LG: f32 = 16.0;
pub const FONT_SIZE_XL: f32 = 20.0;
pub const FONT_SIZE_TITLE: f32 = 24.0;

/// Max width for form content to prevent stretching on wide screens.
pub const FORM_MAX_WIDTH: f64 = 640.0;

/// Standard input height for consistency.
pub const INPUT_HEIGHT: f64 = 38.0;

/// Standard input padding.
pub const INPUT_PADDING: f64 = 10.0;

pub const SIDEBAR_WIDTH: f64 = 180.0;
pub const STATUS_BAR_HEIGHT: f64 = 32.0;
