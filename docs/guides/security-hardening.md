# Security hardening

This guide enumerates the provider's security posture and the
operational checklist for hardened deployments.

## TLS verification (FR-003b, FR-003c)

The provider verifies every NC2 endpoint's TLS certificate against
the system trust store. There is no schema attribute to disable
verification, and a static-analysis test
(`internal/provider/tls_test.go::TestBuildTLSConfig_NeverSkipsVerify`)
fails the build if any code path constructs
`&tls.Config{InsecureSkipVerify: true}`. The integration test
`tests/integration/no_skip_verify_test.go` extends the same
guarantee across every Go file in the module.

To trust an internal CA (private NC2 sandbox, on-prem proxy), set
the `ca_bundle` provider attribute to a PEM file. The bundle is
**appended** to the system trust store, not replaced.

```terraform
provider "nc2" {
  ca_bundle = "/etc/ssl/internal-ca.pem"
}
```

Invalid PEM input produces a Configure-time error with the parse
offset.

## Sensitive-field redaction (FR-002, FR-002a, FR-002b)

Two layers of redaction protect sensitive values from logs, plan
output, and audit records:

1. **Pattern match** (FR-002): any attribute whose name contains
   `credential`, `password`, `secret`, `token`, `private_key`, or
   `api_key` (case-insensitive) is automatically redacted to
   `(sensitive)`.
2. **Per-resource registry** (FR-002b): each resource package
   contributes a `Sensitive() []string` returning a list of dotted
   attribute paths the redactor must replace explicitly. This
   covers fields whose names do not match a pattern but are
   semantically sensitive (e.g.
   `network.prism_element_access_policy.ip_addresses`).

The CI tool `tools/sensitive-lint` cross-checks the OpenAPI spec
against both layers and fails (in `--strict` mode) on drift.

## Audit logging (FR-020a..d)

Every NC2 API call emits one structured log record under the
`audit` `tflog` subsystem. Fields:

- `terraform_op` — e.g. `nc2_aws_cluster.Create`.
- `nc2_endpoint`, `http_method`, `http_status`.
- `nc2_task_id`, `nc2_error_code` (when present).
- `correlation_id` — UUIDv4, pinned per-request.
- `actor` — derived from JWT claims; the operator's identity.

To capture audit logs:

```bash
TF_LOG_PROVIDER_AUDIT=INFO TF_LOG_PATH=audit.log terraform apply
```

The audit subsystem is **separate** from `TF_LOG=DEBUG`; you can
ship audit records to your SIEM without enabling verbose framework
logs. All sensitive fields pass through `redact.Registry.Redact`
before emission (FR-020b).

## Release signing & supply-chain attestation (FR-032, FR-032a, FR-032b)

The release pipeline produces:

- **Provider binaries** for `linux/darwin/windows/freebsd ×
  amd64/arm64/386` via GoReleaser.
- **`SHA256SUMS`** hash file, GPG-signed (`SHA256SUMS.sig`) by the
  Nutanix release key.
- **Cosign keyless signatures** (`.sig` + `.pem`) for every
  artifact via the GitHub OIDC token.
- **SLSA Level 3 provenance** (`provenance.intoto.jsonl`) via
  `slsa-github-generator`.
- **Vulnerability report** (`vulnerability-report.json`) from
  `govulncheck` and `osv-scanner`. HIGH and CRITICAL CVEs hard-fail
  the workflow (FR-032a).

Verify a downloaded provider with:

```bash
cosign verify-blob \
  --certificate terraform-provider-nc2_<version>_<os>_<arch>.zip.pem \
  --signature   terraform-provider-nc2_<version>_<os>_<arch>.zip.sig \
  --certificate-identity-regexp 'https://github.com/nutanix/terraform-provider-nc2/.github/workflows/release.yml@.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  terraform-provider-nc2_<version>_<os>_<arch>.zip

slsa-verifier verify-artifact \
  --provenance-path provenance.intoto.jsonl \
  --source-uri github.com/nutanix/terraform-provider-nc2 \
  terraform-provider-nc2_<version>_<os>_<arch>.zip
```

## Operator checklist

- [ ] Restrict the credentials file (`chmod 600`).
- [ ] Source NC2 credentials from short-lived env vars in CI rather
      than committing them to the credentials file.
- [ ] Pin the provider version in `required_providers`.
- [ ] Enable `TF_LOG_PROVIDER_AUDIT=INFO` and ship audit logs to a
      retained store (CloudWatch, Loki, etc.).
- [ ] Verify cosign signature + SLSA provenance on every download.
- [ ] If an internal CA is involved, set `ca_bundle` rather than
      disabling verification (which is not possible).
