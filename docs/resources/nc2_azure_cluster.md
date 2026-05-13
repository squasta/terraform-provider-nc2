# nc2_azure_cluster (Resource)

Manages an NC2 cluster on Microsoft Azure.

> `access_policy` is **not** available on this resource (FR-010a).
>
> `desired_state` (hibernate / resume) is **not** available on this
> resource — the hibernate / resume lifecycle is AWS-only (FR-012).
> Setting `desired_state` in configuration is a schema error.

## Example Usage

```terraform
resource "nc2_azure_cluster" "demo" {
  organization_id     = nc2_organization.demo.id
  cloud_account_id    = nc2_cloud_account.azure.id
  name                = "demo-azure"
  region              = "eastus"
  host_access_ssh_key = "demo-key"
  license             = "aos"
  aos_version         = "6.7"
  software_tier       = "pro"

  capacity   = [{ host_type = "AN36P", number_of_hosts = "3" }]
  redundancy = { factor = "1" }
  network    = {
    mode              = "new"
    availability_zone = "1"
    vpc_cidr          = "10.0.0.0/16"
  }
}
```

## Schema

Same shape as `nc2_aws_cluster` minus `access_policy`. Azure-specific
network attributes (`virtual_network_id`, `delegated_subnet_id`)
live inside `network` as Map<String,String> entries.

### `capacity[*].host_type` constraint

NC2 on Azure only supports the following bare-metal SKUs as cluster
hosts:

| SKU | Description |
|---|---|
| `AN36P` | Nutanix bare-metal AN36P node (default for most workloads). |
| `AN64`  | Nutanix bare-metal AN64 node (memory- / capacity-optimised). |

Any other value (for example a generic Azure VM size such as
`Standard_D32s_v4`, or an AWS-style SKU such as `m5d.metal`) is
rejected at plan time with a per-element diagnostic that names both
the offending value and the allowed set. The check runs in
`ValidateConfig`, so no NC2 API call is made for invalid plans.

### `capacity[*].number_of_hosts` constraint

Across every cloud, NC2 only supports the following cluster sizes:

| `sum(capacity[*].number_of_hosts)` | Supported? |
|---|---|
| `1` | yes (single-host cluster) |
| `2` | **no** — a 2-host cluster has no quorum / metadata-redundancy story and is rejected by NC2 |
| `3` to `28` (inclusive) | yes (production cluster sizes) |
| `0` or `> 28` | no |

The check runs in the same `ValidateConfig` call as the host-type
check, sharing
`internal/resources/clustershared/ValidateCapacityHostCount`.
