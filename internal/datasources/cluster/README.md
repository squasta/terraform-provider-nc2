# `internal/datasources/cluster`

Implements `data.nc2_cluster` — single-cluster lookup by `id`.

## Public API

- `NewDataSource() datasource.DataSource`
- `var OperationMappings []oapi.Mapping` — covers
  `CPanelWeb.Api.ClusterController.show` (shared with the managed
  resource read).

## Behavior

`Read` calls `GET /clusters/{id}`. A 404 surfaces as a
`nc2_cluster not found` diagnostic.
