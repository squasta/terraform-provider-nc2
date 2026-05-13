# nc2_gcp_cluster (Resource)

Manages an NC2 cluster on Google Cloud Platform.

> `access_policy` is **not** available on this resource (FR-010a).
>
> `desired_state` (hibernate / resume) is **not** available on this
> resource — the hibernate / resume lifecycle is AWS-only (FR-012).
> Setting `desired_state` in configuration is a schema error.

## Example Usage

```terraform
resource "nc2_gcp_cluster" "demo" {
  organization_id     = nc2_organization.demo.id
  cloud_account_id    = nc2_cloud_account.gcp.id
  name                = "demo-gcp"
  region              = "europe-west1"
  host_access_ssh_key = "demo-key"
  license             = "aos"
  aos_version         = "6.7"
  software_tier       = "pro"

  capacity   = [{ host_type = "n2-standard-32", number_of_hosts = "3" }]
  redundancy = { factor = "1" }
  network    = {
    mode              = "new"
    availability_zone = "europe-west1-b"
    project_id        = "my-gcp-project"
  }
}
```

## Schema

Same shape as `nc2_aws_cluster` minus `access_policy`. GCP-specific
network attributes (`project_id`, `vpc_name`) live inside `network`
as Map<String,String> entries.
