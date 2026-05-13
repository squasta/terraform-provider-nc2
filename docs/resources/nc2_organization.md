# nc2_organization (Resource)

Manages an NC2 organization. Organizations are the root of the NC2
resource hierarchy and the unit of billing/audit.

## Example Usage

```terraform
resource "nc2_organization" "demo" {
  name        = "demo-org"
  description = "Demo organization managed by Terraform."
}
```

## Schema

### Required

- `name` (String) — 1–128 characters.

### Optional

- `description` (String) — ≤ 2048 characters.

### Read-only

- `id` (String) — Organization UUID.
- `state` (String) — `active` | `terminating` | `terminated`.
- `created_at`, `updated_at` (String) — RFC 3339 timestamps.

## Lifecycle

`pending → active → terminating → terminated`. Terraform `Destroy`
issues `PATCH /organizations/{id}/terminate`; the organization
transitions through `terminating` → `terminated` and a subsequent
Read returns 404, removing the resource from state (FR-009).

## Import

```shell
terraform import nc2_organization.demo <organization-uuid>
```
