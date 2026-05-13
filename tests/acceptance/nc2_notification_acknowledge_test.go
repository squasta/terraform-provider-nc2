package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_US5_NotificationAcknowledge covers the
// nc2_notification_acknowledge action (FR-025).
func TestAcc_US5_NotificationAcknowledge(t *testing.T) {
	SkipIfNotAcc(t)
	envs := SkipIfMissing(t,
		"NC2_API_KEY", "NC2_KEY_ID", "NC2_ISSUER",
		"NC2_TEST_NOTIFICATION_ID",
	)
	notificationID := envs[3]

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: notificationAcknowledgeConfig(notificationID)},
		},
	})
}

func notificationAcknowledgeConfig(notificationID string) string {
	return fmt.Sprintf(`
action "nc2_notification_acknowledge" "ack" {
  config {
    notification_id = %q
  }
}
`, notificationID)
}
