# Contracts

This directory defines the **Terraform-facing interface contracts** the provider must implement. Each file is a JSON document describing one HCL block (the provider config, a managed resource, a data source, or an action) at the level of detail required to lock the user-visible surface before implementation.

These are not Go code, not OpenAPI specs, and not Terraform schema code — they are an intermediate, language-agnostic representation that:

1. Pins the **attribute names, types, modes, plan modifiers, sensitivity, and validators** the implementation must produce.
2. Pins the **mapping** from each CRUD verb (or action invocation) to one or more NC2 OpenAPI `operationId` values, so that the `tools/coverage-check` CI gate (FR-021, FR-021a) has an authoritative cross-reference.
3. Drive the **contract tests** (FR-022) that assert the Go implementation matches the declared shape (attribute names, types, sensitivity, validators all line up).

## File naming

```text
contracts/
├── README.md                              # this file
├── provider.json                          # full provider config schema
├── resources/<resource_type_name>.json    # one per managed resource
├── datasources/<data_source_type_name>.json
└── actions/<action_type_name>.json
```

## File format (informal JSON Schema)

Every contract file conforms to this structure:

```jsonc
{
  "kind": "provider" | "resource" | "data_source" | "action",
  "type_name": "nc2_<thing>",
  "description": "<human-readable purpose>",
  "schema_version": 0,
  "attributes": {
    "<attribute_path>": {
      "type": "<framework type expression>",
      "mode": "Required" | "Optional" | "Optional+Default" | "Optional+Computed" | "Computed",
      "default": "<value, only if Optional+Default>",
      "sensitive": true | false,
      "description": "<human-readable purpose>",
      "validators": ["<rule>", "..."],
      "plan_modifiers": ["RequiresReplace" | "UseStateForUnknown" | "..."],
      "nc2_mapping": {
        "request": "<dotted json path in request body, or 'path:{param}'>",
        "response": "<dotted json path in response body>"
      }
    }
  },
  "lifecycle": {
    "Create": { "operationIds": ["<openapi opId>", "..."] },
    "Read":   { "operationIds": ["<openapi opId>", "..."] },
    "Update": { "operationIds": ["<openapi opId>", "..."], "field_routing": { "<attribute>": "<openapi opId>" } },
    "Delete": { "operationIds": ["<openapi opId>", "..."] | [] }
  },
  "sensitive_registry": ["<attribute path>", "..."]
}
```

`kind: data_source` files omit `Create`, `Update`, `Delete` from `lifecycle`. `kind: action` files replace `lifecycle` with a single `Invoke` block.

## Coverage in this directory

This directory contains **canonical** contracts — one example per kind — to lock the convention. The full set (6 resources + 18 data sources + 8 actions = 32 contract files) is produced during implementation, driven by `/speckit.tasks`. `tools/coverage-check` validates the complete set against `openapi/openapi.json` on every build.

Files present in this directory:

| File | Kind | Purpose |
|---|---|---|
| `provider.json` | provider | Full provider config schema (auth, base URL, `ca_bundle`, profile, timeouts) — locks the FR-001..FR-003c shape |
| `resources/nc2_organization.json` | resource | Canonical managed resource — smallest meaningful resource; locks the basic CRUD-with-terminate pattern |
| `resources/nc2_aws_cluster.json` | resource | Most complex managed resource — locks the multi-endpoint Update field routing per FR-010 and the AWS-only `access_policy` per FR-010a |
| `datasources/nc2_cloud_account_regions.json` | data_source | Canonical data source — locks the scoping (cloud account → region list) and deterministic ordering per FR-015 |
| `actions/nc2_cluster_condemn_host.json` | action | Canonical action — locks the action shape (target identifier in, task result out) per FR-016/FR-017 |
