use floem::reactive::{RwSignal, SignalGet, SignalUpdate, SignalWith};
use uuid::Uuid;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum NotificationLevel {
    Success,
    Error,
    Warning,
    Info,
}

#[derive(Debug, Clone)]
pub struct Notification {
    pub id: Uuid,
    pub message: String,
    pub level: NotificationLevel,
}

impl Notification {
    pub fn success(message: impl Into<String>) -> Self {
        Self {
            id: Uuid::new_v4(),
            message: message.into(),
            level: NotificationLevel::Success,
        }
    }

    pub fn error(message: impl Into<String>) -> Self {
        Self {
            id: Uuid::new_v4(),
            message: message.into(),
            level: NotificationLevel::Error,
        }
    }

    pub fn warning(message: impl Into<String>) -> Self {
        Self {
            id: Uuid::new_v4(),
            message: message.into(),
            level: NotificationLevel::Warning,
        }
    }

    pub fn info(message: impl Into<String>) -> Self {
        Self {
            id: Uuid::new_v4(),
            message: message.into(),
            level: NotificationLevel::Info,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum AsyncOpStatus {
    Pending,
    Completed,
    Failed,
}

#[derive(Debug, Clone)]
pub struct AsyncOperation {
    pub id: Uuid,
    pub op_type: String,
    pub status: AsyncOpStatus,
    pub target: String,
}

/// Reactive notification state.
pub struct NotificationState {
    pub notifications: RwSignal<Vec<Notification>>,
    pub pending_operations: RwSignal<Vec<AsyncOperation>>,
}

impl NotificationState {
    pub fn new() -> Self {
        Self {
            notifications: RwSignal::new(Vec::new()),
            pending_operations: RwSignal::new(Vec::new()),
        }
    }

    pub fn push(&self, notification: Notification) {
        self.notifications.update(|n| n.push(notification));
    }

    pub fn dismiss(&self, id: Uuid) {
        self.notifications.update(|n| n.retain(|notif| notif.id != id));
    }

    pub fn clear(&self) {
        self.notifications.update(|n| n.clear());
    }

    pub fn has_notification(&self, message: &str) -> bool {
        self.notifications
            .with_untracked(|n| n.iter().any(|notif| notif.message == message))
    }

    pub fn add_operation(&self, op_type: &str, target: &str) -> Uuid {
        let id = Uuid::new_v4();
        self.pending_operations.update(|ops| {
            ops.push(AsyncOperation {
                id,
                op_type: op_type.to_string(),
                status: AsyncOpStatus::Pending,
                target: target.to_string(),
            });
        });
        id
    }

    pub fn complete_operation(&self, id: Uuid) {
        self.pending_operations.update(|ops| {
            if let Some(op) = ops.iter_mut().find(|o| o.id == id) {
                op.status = AsyncOpStatus::Completed;
            }
        });
    }

    pub fn fail_operation(&self, id: Uuid) {
        self.pending_operations.update(|ops| {
            if let Some(op) = ops.iter_mut().find(|o| o.id == id) {
                op.status = AsyncOpStatus::Failed;
            }
        });
    }

    pub fn remove_operation(&self, id: Uuid) {
        self.pending_operations
            .update(|ops| ops.retain(|o| o.id != id));
    }
}
