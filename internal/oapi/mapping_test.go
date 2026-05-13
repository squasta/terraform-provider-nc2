package oapi

import "testing"

// TestRegistry_RegisterAndSnapshot covers the happy path: register
// a few mappings, snapshot them sorted, and confirm de-duplication.
func TestRegistry_RegisterAndSnapshot(t *testing.T) {
	t.Parallel()

	r := &Registry{}
	r.Register(
		Mapping{OperationID: "B.op", TerraformOp: "tf.b"},
		Mapping{OperationID: "A.op", TerraformOp: "tf.a"},
		Mapping{OperationID: "A.op", TerraformOp: "tf.a"}, // duplicate
		Mapping{OperationID: "", TerraformOp: "tf.x"},     // empty -> ignored
		Mapping{OperationID: "C.op", TerraformOp: ""},     // empty -> ignored
	)
	if got := r.Len(); got != 2 {
		t.Errorf("Len = %d; want 2", got)
	}
	snap := r.Snapshot()
	if len(snap) != 2 || snap[0].OperationID != "A.op" || snap[1].OperationID != "B.op" {
		t.Errorf("snapshot order/contents unexpected: %+v", snap)
	}
}

// TestRegistry_CoveredOperationIDs returns deduplicated, sorted ids.
func TestRegistry_CoveredOperationIDs(t *testing.T) {
	t.Parallel()

	r := &Registry{}
	r.Register(
		Mapping{OperationID: "X.op", TerraformOp: "tf.x1"},
		Mapping{OperationID: "X.op", TerraformOp: "tf.x2"}, // distinct mapping, same op
		Mapping{OperationID: "Y.op", TerraformOp: "tf.y"},
	)
	got := r.CoveredOperationIDs()
	if len(got) != 2 || got[0] != "X.op" || got[1] != "Y.op" {
		t.Errorf("CoveredOperationIDs = %v", got)
	}
}

// TestRegistry_Reset wipes the registry.
func TestRegistry_Reset(t *testing.T) {
	t.Parallel()

	r := &Registry{}
	r.Register(Mapping{OperationID: "X", TerraformOp: "tf.x"})
	r.Reset()
	if r.Len() != 0 {
		t.Errorf("after Reset, Len = %d; want 0", r.Len())
	}
}
