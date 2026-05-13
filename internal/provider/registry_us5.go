package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/action"

	close_tunnel "github.com/nutanix/terraform-provider-nc2/internal/actions/cluster_close_support_tunnel"
	condemn_host "github.com/nutanix/terraform-provider-nc2/internal/actions/cluster_condemn_host"
	extend_tunnel "github.com/nutanix/terraform-provider-nc2/internal/actions/cluster_extend_support_tunnel"
	open_tunnel "github.com/nutanix/terraform-provider-nc2/internal/actions/cluster_open_support_tunnel"
	scale_fgw "github.com/nutanix/terraform-provider-nc2/internal/actions/cluster_scale_flow_gateway"
	start_recovery "github.com/nutanix/terraform-provider-nc2/internal/actions/cluster_start_recovery"
	upgrade_fgw "github.com/nutanix/terraform-provider-nc2/internal/actions/cluster_upgrade_flow_gateway"
	notif_ack "github.com/nutanix/terraform-provider-nc2/internal/actions/notification_acknowledge"
)

//nolint:gochecknoinits // init is the documented per-story registration mechanism.
func init() {
	actionRegistry = append(actionRegistry,
		func() action.Action { return condemn_host.NewAction() },
		func() action.Action { return open_tunnel.NewAction() },
		func() action.Action { return extend_tunnel.NewAction() },
		func() action.Action { return close_tunnel.NewAction() },
		func() action.Action { return scale_fgw.NewAction() },
		func() action.Action { return upgrade_fgw.NewAction() },
		func() action.Action { return start_recovery.NewAction() },
		func() action.Action { return notif_ack.NewAction() },
	)
}
