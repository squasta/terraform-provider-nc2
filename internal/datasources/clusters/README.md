# `internal/datasources/clusters`

Implements `data.nc2_clusters` — sorted-by-`id` list of every NC2
cluster visible to the configured credentials.

## Public API

- `NewDataSource() datasource.DataSource`
- `var OperationMappings []oapi.Mapping` — covers
  `CPanelWeb.Api.ClusterController.index`.

## Behavior

`Read` calls `GET /clusters`; results are sorted ascending by `id`
to satisfy FR-015.
