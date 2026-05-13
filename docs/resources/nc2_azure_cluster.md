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

  capacity   = [{ host_type = "Standard_D32s_v4", number_of_hosts = "3" }]
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
