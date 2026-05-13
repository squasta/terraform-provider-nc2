# `nc2_cloud_account` resource

A credentialed AWS / Azure / GCP account that NC2 uses to provision
infrastructure on behalf of an organization. Mapped to data-model.md
§2.

## ⚠️ Destroy semantics (FR-006)

NC2 has **no DELETE endpoint** for cloud accounts. Terraform `Destroy`
removes the resource from state only; the underlying NC2 cloud
account remains live until removed via the NC2 UI or another
channel. A plan-time warning is emitted when destroy is planned.

## Lifecycle

- **Create** issues `POST /organizations/{id}/cloud-accounts/{cloud_provider}`.
- **Read** issues `GET /cloud-accounts/{id}`. 404 → state removal
  (FR-009).
- **Update** routes the diff:
  - `credentials` change → `POST /cloud-accounts/{id}/update-credentials`.
  - `name` / `description` change → `PATCH /cloud-accounts/{id}`.
  - Both in the same diff → credentials first, then PATCH.
- **Delete** is state-only (see warning above).
- **ImportState** uses the cloud account UUID as the import key.

## Sensitive registry

The full `credentials` subtree is sensitive (auto-classified by the
FR-002 `credential` substring pattern; the parent attribute is also
schema-marked Sensitive). The per-resource registry adds explicit
entries for `secret_access_key`, `client_secret`, and
`service_account_json` for traceability and tools/sensitive-lint
drift detection.

## OpenAPI coverage

See `openapi_mapping.go`.
