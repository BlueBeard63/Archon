use floem::prelude::*;
use floem::window::WindowConfig;
use floem::kurbo::Size;

use crate::logic::{AppState, Screen};
use crate::ui::components::status_bar::status_bar;
use crate::ui::components::tab_bar::tab_bar;
use crate::ui::screens::{
    dashboard::dashboard_screen,
    domains::domains_list_screen,
    help::help_screen,
    nodes::nodes_list_screen,
    settings::settings_screen,
    sites::sites_list_screen,
};
use crate::ui::styles::*;

/// Launch the Floem GUI application.
pub fn run() {
    // Leak AppState so it has 'static lifetime for reactive closures
    let state: &'static AppState = Box::leak(Box::new(AppState::new()));

    floem::Application::new()
        .window(
            move |_| app_view(state),
            Some(
                WindowConfig::default()
                    .size(Size::new(1200.0, 800.0))
                    .title("Archon"),
            ),
        )
        .run();
}

fn app_view(state: &'static AppState) -> impl IntoView {
    let current_screen = state.navigation.current_screen;

    v_stack((
        // Main content area: sidebar + screen
        h_stack((
            tab_bar(state),
            // Screen content
            dyn_container(
                move || current_screen.get(),
                move |screen| match screen {
                    Screen::Dashboard => dashboard_screen(state).into_any(),
                    Screen::SitesList
                    | Screen::SiteCreate
                    | Screen::SiteEdit
                    | Screen::SiteEnvVars
                    | Screen::SiteDeleteConfirm => sites_list_screen(state).into_any(),
                    Screen::DomainsList
                    | Screen::DomainCreate
                    | Screen::DomainEdit
                    | Screen::DomainDnsRecords => domains_list_screen(state).into_any(),
                    Screen::NodesList
                    | Screen::NodeCreate
                    | Screen::NodeEdit
                    | Screen::NodeConfig
                    | Screen::NodeConfigSave
                    | Screen::NodeQuickConfig => nodes_list_screen(state).into_any(),
                    Screen::Settings
                    | Screen::DockerCredentialsList
                    | Screen::DockerCredentialCreate
                    | Screen::DockerCredentialEdit => settings_screen(state).into_any(),
                    Screen::Help => help_screen().into_any(),
                },
            )
            .style(|s| s.flex_grow(1.0).height_full()),
        ))
        .style(|s| s.width_full().flex_grow(1.0)),
        // Status bar
        status_bar(state),
    ))
    .style(|s| {
        s.width_full()
            .height_full()
            .background(BG_PRIMARY)
            .flex_col()
    })
}
