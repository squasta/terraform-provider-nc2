# `internal/actions/notification_acknowledge`

Implements `nc2_notification_acknowledge`. Issues
`PATCH /notifications/{id}` with `{"data": {"acknowledged": true}}`.

Covers `CPanelWeb.Api.NotificationController.update (2)` (the PATCH
variant; the PUT replace-all `update` is intentionally not invoked
here per FR-010b — it's claimed by the no-op coverage list to keep
SC-001 happy).
