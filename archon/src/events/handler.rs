use crossterm::event::{Event, KeyCode, KeyEvent, KeyEventKind, KeyModifiers};

use super::Action;
use crate::ui::Screen;

/// Convert crossterm events to Actions
pub fn handle_event(event: Event, current_screen: &Screen) -> Action {
    match event {
        Event::Key(key_event) if key_event.kind == KeyEventKind::Press => {
            handle_key_event(key_event, current_screen)
        }
        _ => Action::None,
    }
}

fn handle_key_event(key: KeyEvent, current_screen: &Screen) -> Action {
    // Global key bindings (work on all screens)
    match (key.code, key.modifiers) {
        (KeyCode::Char('c'), KeyModifiers::CONTROL) | (KeyCode::Char('q'), KeyModifiers::NONE) => {
            return Action::Quit
        }
        (KeyCode::Char('?'), KeyModifiers::NONE) => return Action::NavigateTo(Screen::Help),
        (KeyCode::Esc, _) => return Action::NavigateBack,
        _ => {}
    }

    // Screen-specific key bindings
    match current_screen {
        Screen::Dashboard => handle_dashboard_keys(key),
        Screen::SitesList => handle_sites_list_keys(key),
        Screen::DomainsList => handle_domains_list_keys(key),
        Screen::NodesList => handle_nodes_list_keys(key),
        _ => Action::None,
    }
}

fn handle_dashboard_keys(key: KeyEvent) -> Action {
    match key.code {
        KeyCode::Char('1') | KeyCode::Char('d') => Action::NavigateTo(Screen::Dashboard),
        KeyCode::Char('2') | KeyCode::Char('s') => Action::NavigateTo(Screen::SitesList),
        KeyCode::Char('3') => Action::NavigateTo(Screen::DomainsList),
        KeyCode::Char('4') | KeyCode::Char('n') => Action::NavigateTo(Screen::NodesList),
        KeyCode::Char('r') => Action::LoadConfig, // Refresh
        _ => Action::None,
    }
}

fn handle_sites_list_keys(key: KeyEvent) -> Action {
    match key.code {
        KeyCode::Up | KeyCode::Char('k') => Action::SelectPrevious,
        KeyCode::Down | KeyCode::Char('j') => Action::SelectNext,
        KeyCode::Char('c') => Action::NavigateTo(Screen::SiteCreate),
        _ => Action::None,
    }
}

fn handle_domains_list_keys(key: KeyEvent) -> Action {
    match key.code {
        KeyCode::Up | KeyCode::Char('k') => Action::SelectPrevious,
        KeyCode::Down | KeyCode::Char('j') => Action::SelectNext,
        KeyCode::Char('c') => Action::NavigateTo(Screen::DomainCreate),
        _ => Action::None,
    }
}

fn handle_nodes_list_keys(key: KeyEvent) -> Action {
    match key.code {
        KeyCode::Up | KeyCode::Char('k') => Action::SelectPrevious,
        KeyCode::Down | KeyCode::Char('j') => Action::SelectNext,
        KeyCode::Char('c') => Action::NavigateTo(Screen::NodeCreate),
        _ => Action::None,
    }
}
