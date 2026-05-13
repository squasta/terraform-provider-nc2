# `nc2_cloud_account_region` resource

A cloud-provider region enabled on an NC2 cloud account. Mapped to
data-model.md §3.

## ⚠️ Destroy semantics (FR-007)

NC2 has **no DELETE endpoint** for cloud account regions.
`terraform destroy` removes the resource from state only; the region
remains enabled in NC2 until you disable it via the NC2 UI. A
plan-time warning is emitted.

## Lifecycle

- **Create** issues `POST /cloud-accounts/{cloud_account_id}/regions`.
- **Read** issues `GET /cloud-accounts/{cloud_account_id}/regions`
  and filters client-side for the matching id (NC2 has no
  `/regions/{id}` endpoint).
- **Update** is a no-op; `region` and `cloud_account_id` are
  `RequiresReplace`.
- **Delete** is state-only (see warning above).
- **ImportState** expects `<cloud_account_id>/<region_id>`.

## Sensitive registry

Empty.
