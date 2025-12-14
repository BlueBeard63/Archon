use uuid::Uuid;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Screen {
    Dashboard,
    SitesList,
    SiteCreate,
    SiteEdit(Uuid),
    SiteDetail(Uuid),
    DomainsList,
    DomainCreate,
    DomainDnsEditor(Uuid),
    NodesList,
    NodeCreate,
    NodeEdit(Uuid),
    NodeDetail(Uuid),
    Help,
}

impl Default for Screen {
    fn default() -> Self {
        Screen::Dashboard
    }
}
