package discharge_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-97/internal/discharge"
	"github.com/wyw14/cry-97/internal/lab"
	"github.com/wyw14/cry-97/internal/model"
)

func TestSampleRevisionRevokesPriorDischargePermit(t *testing.T) {
	lineID := model.LineID("line-1")
	valves, err := discharge.NewValveController([]model.LineID{lineID})
	if err != nil {
		t.Fatal(err)
	}
	permits, err := discharge.NewPermitBook(valves)
	if err != nil {
		t.Fatal(err)
	}
	results := lab.NewResults()
	sampleID, batchID := uuid.New(), uuid.New()
	qualified := model.LabResult{SampleID: sampleID, BatchID: batchID, LineID: lineID, Revision: 7, COD: 21, Ammonia: 2, Valid: true}
	if _, err := results.Apply(qualified); err != nil {
		t.Fatal(err)
	}
	if _, err := permits.ApplyResult(qualified, time.Now()); err != nil {
		t.Fatal(err)
	}
	corrected := qualified
	corrected.Revision = 8
	corrected.COD = 88
	if _, err := results.Apply(corrected); err != nil {
		t.Fatalf("newer correction was rejected: %v", err)
	}
	stored, ok := results.Get(sampleID)
	if !ok || stored.Revision != 8 || stored.Qualified() {
		t.Fatalf("result registry did not retain correction: %#v", stored)
	}
	permit, err := permits.ApplyResult(corrected, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if permit.Active || permit.SampleRevision != 8 {
		t.Fatalf("prior permit survived correction: %#v", permit)
	}
	valveState := valves.Snapshot()
	if len(valveState) != 1 || valveState[0].Open {
		t.Fatalf("discharge valve remained open: %#v", valveState)
	}
}
