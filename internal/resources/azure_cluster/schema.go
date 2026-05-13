// Package azure_cluster implements the nc2_azure_cluster managed
// resource. Mirrors the AWS variant minus AWS-only fields. Setting
// `access_policy` on this resource is impossible — the attribute
// does not exist in the schema (FR-010a).
package azure_cluster //nolint:revive,staticcheck

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
)

// model is the framework view of the azure cluster state. It reuses
// the shared model verbatim — Azure-specific subfields are encoded
// as elements of the shared `network` Map<String,String>.
type model struct {
	clustershared.Model
}

func resourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages an NC2 cluster on Microsoft Azure. " +
			"`access_policy` is intentionally absent (AWS-only per FR-010a).",
		Attributes: clustershared.CommonAttributes(),
	}
}
