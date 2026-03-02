use floem::reactive::{RwSignal, SignalGet, SignalUpdate, SignalWith};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Screen {
    Dashboard,
    SitesList,
    SiteCreate,
    SiteEdit,
    SiteEnvVars,
    SiteDeleteConfirm,
    DomainsList,
    DomainCreate,
    DomainEdit,
    DomainDnsRecords,
    NodesList,
    NodeCreate,
    NodeEdit,
    NodeConfig,
    NodeConfigSave,
    NodeQuickConfig,
    Settings,
    DockerCredentialsList,
    DockerCredentialCreate,
    DockerCredentialEdit,
    Help,
}

impl std::fmt::Display for Screen {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Dashboard => write!(f, "Dashboard"),
            Self::SitesList => write!(f, "Sites"),
            Self::SiteCreate => write!(f, "Create Site"),
            Self::SiteEdit => write!(f, "Edit Site"),
            Self::SiteEnvVars => write!(f, "Environment Variables"),
            Self::SiteDeleteConfirm => write!(f, "Delete Site"),
            Self::DomainsList => write!(f, "Domains"),
            Self::DomainCreate => write!(f, "Create Domain"),
            Self::DomainEdit => write!(f, "Edit Domain"),
            Self::DomainDnsRecords => write!(f, "DNS Records"),
            Self::NodesList => write!(f, "Nodes"),
            Self::NodeCreate => write!(f, "Create Node"),
            Self::NodeEdit => write!(f, "Edit Node"),
            Self::NodeConfig => write!(f, "Node Config"),
            Self::NodeConfigSave => write!(f, "Save Node Config"),
            Self::NodeQuickConfig => write!(f, "Quick Config"),
            Self::Settings => write!(f, "Settings"),
            Self::DockerCredentialsList => write!(f, "Docker Credentials"),
            Self::DockerCredentialCreate => write!(f, "Create Docker Credential"),
            Self::DockerCredentialEdit => write!(f, "Edit Docker Credential"),
            Self::Help => write!(f, "Help"),
        }
    }
}

/// Tab indices matching the tab bar order.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Tab {
    Dashboard = 0,
    Sites = 1,
    Domains = 2,
    Nodes = 3,
    Settings = 4,
    Help = 5,
}

impl Tab {
    pub fn from_index(idx: usize) -> Option<Self> {
        match idx {
            0 => Some(Self::Dashboard),
            1 => Some(Self::Sites),
            2 => Some(Self::Domains),
            3 => Some(Self::Nodes),
            4 => Some(Self::Settings),
            5 => Some(Self::Help),
            _ => None,
        }
    }

    pub fn label(&self) -> &'static str {
        match self {
            Self::Dashboard => "Dashboard",
            Self::Sites => "Sites",
            Self::Domains => "Domains",
            Self::Nodes => "Nodes",
            Self::Settings => "Settings",
            Self::Help => "Help",
        }
    }

    pub fn default_screen(&self) -> Screen {
        match self {
            Self::Dashboard => Screen::Dashboard,
            Self::Sites => Screen::SitesList,
            Self::Domains => Screen::DomainsList,
            Self::Nodes => Screen::NodesList,
            Self::Settings => Screen::Settings,
            Self::Help => Screen::Help,
        }
    }
}

pub const TAB_COUNT: usize = 6;
pub const TAB_LABELS: [&str; TAB_COUNT] = [
    "Dashboard",
    "Sites",
    "Domains",
    "Nodes",
    "Settings",
    "Help",
];

/// Reactive navigation state.
pub struct NavigationState {
    pub current_screen: RwSignal<Screen>,
    pub previous_screens: RwSignal<Vec<Screen>>,
    pub active_tab: RwSignal<usize>,
}

impl NavigationState {
    pub fn new() -> Self {
        Self {
            current_screen: RwSignal::new(Screen::Dashboard),
            previous_screens: RwSignal::new(Vec::new()),
            active_tab: RwSignal::new(0),
        }
    }

    /// Navigate to a new screen, pushing the current screen onto the back stack.
    pub fn navigate_to(&self, screen: Screen) {
        let current = self.current_screen.get_untracked();
        self.previous_screens.update(|stack| stack.push(current));
        self.current_screen.set(screen);
    }

    /// Navigate back to the previous screen. Returns false if there's nothing to go back to.
    pub fn navigate_back(&self) -> bool {
        let mut popped = None;
        self.previous_screens.update(|stack| {
            popped = stack.pop();
        });
        if let Some(screen) = popped {
            self.current_screen.set(screen);
            true
        } else {
            false
        }
    }

    /// Navigate to a tab's default screen, clearing the back stack.
    pub fn navigate_to_tab(&self, tab_index: usize) {
        if let Some(tab) = Tab::from_index(tab_index) {
            self.previous_screens.update(|stack| stack.clear());
            self.current_screen.set(tab.default_screen());
            self.active_tab.set(tab_index);
        }
    }

    pub fn can_go_back(&self) -> bool {
        self.previous_screens.with_untracked(|stack| !stack.is_empty())
    }
}
