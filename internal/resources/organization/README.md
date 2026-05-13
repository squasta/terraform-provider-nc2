# `nc2_organization` resource

An NC2 organization — root of the resource hierarchy and unit of
billing/audit. Mapped to data-model.md §1.

## Lifecycle

`pending → active → terminating → terminated`

- **Create** issues `POST /organizations`. Synchronous (no task).
- **Read** issues `GET /organizations/{id}`. 404 → state removal
  (FR-009).
- **Update** issues `PATCH /organizations/{id}` for `name` and
  `description` changes. Other attributes are computed.
- **Delete** issues `PATCH /organizations/{id}/terminate`. The
  organization transitions through `terminating` → `terminated`;
  subsequent Read returns 404 → state is removed.
- **ImportState** uses the organization UUID as the import key.

## Sensitive registry

Empty — no FR-002 pattern matches and no per-resource sensitive
attributes.

## OpenAPI coverage

See `openapi_mapping.go`. PUT (`update`) is intentionally listed for
coverage purposes; the resource never PUTs at runtime per FR-010b.
