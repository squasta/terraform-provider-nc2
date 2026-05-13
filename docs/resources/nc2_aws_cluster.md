# nc2_aws_cluster (Resource)

Manages an NC2 cluster on AWS. Async create / update / delete are
handled by polling NC2 task IDs (FR-008).

## Example Usage

```terraform
resource "nc2_aws_cluster" "demo" {
  organization_id     = nc2_organization.demo.id
  cloud_account_id    = nc2_cloud_account.aws.id
  name                = "demo-aws"
  region              = "us-east-1"
  host_access_ssh_key = "demo-key"
  license             = "aos"
  aos_version         = "6.7"
  software_tier       = "pro"

  capacity   = [{ host_type = "m5d.metal", number_of_hosts = "3" }]
  redundancy = { factor = "1" }
  network    = {
    mode              = "new"
    availability_zone = "us-east-1a"
    vpc_cidr          = "10.0.0.0/16"
  }
}
```

## Schema

### Required

- `organization_id`, `cloud_account_id`, `name`, `region`,
  `host_access_ssh_key`, `license`, `aos_version`, `software_tier`,
  `capacity` (List of Map), `redundancy` (Map), `network` (Map).

### Optional

- `use_case` (String, default `general`)
- `resource_tags` (Map of String) — in-place via update-resource-tags.
- `access_policy` (Map of String) — **AWS only**, FR-010a.
- `desired_state` (String) — `running` or `hibernated` (US4).

### Read-only

- `id`, `state`, `created_at`, `updated_at`.

## Update routing

Diffs are routed through `internal/resources/clustershared` in this
order: `update-license`, `update-ssh-key`, `update-capacity`,
`update-resource-tags`, `update-access-policy`, generic `PATCH`,
`hibernate` / `resume`.

## Import

```shell
terraform import nc2_aws_cluster.demo <cluster-uuid>
```
