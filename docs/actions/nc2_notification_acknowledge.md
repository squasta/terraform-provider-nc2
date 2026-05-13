# nc2_notification_acknowledge (Action)

Acknowledges a notification by setting `acknowledged=true` via
`PATCH /notifications/{id}`.

## Example Usage

```terraform
action "nc2_notification_acknowledge" "ack" {
  config {
    notification_id = "..."
  }
}
```

## Schema

### Required

- `notification_id` (String) Target notification id.
