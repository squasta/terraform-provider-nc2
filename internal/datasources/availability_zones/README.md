# `internal/datasources/availability_zones`

Implements `data.nc2_availability_zones` — sorted-by-`id` list of
availability zones inside a single cloud-account region.

## Public API

- `NewDataSource() datasource.DataSource`
- `var OperationMappings []oapi.Mapping` — covers
  `CPanelWeb.Api.AvailabilityZoneController.index`.
