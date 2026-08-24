package sludge_test

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-97/internal/journal"
	"github.com/wyw14/cry-97/internal/model"
	"github.com/wyw14/cry-97/internal/sludge"
)

func TestOldSludgeCompensationCannotStopNewSequence(t *testing.T) {
	now := time.Date(2026, 8, 24, 8, 0, 0, 0, time.UTC)
	lineID := model.LineID("line-1")
	pumps, err := sludge.NewPumpController([]sludge.Pump{
		{ID: "pump-A", LineID: lineID, Kind: model.DevicePump},
		{ID: "pump-B", LineID: lineID, Kind: model.DevicePump},
	})
	if err != nil {
		t.Fatal(err)
	}
	queue := journal.NewCompensationQueue()
	book, err := sludge.NewSequenceBook(pumps, queue)
	if err != nil {
		t.Fatal(err)
	}
	old, err := book.Start(lineID, "pump-B", 118, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := book.ConfirmFlow(lineID, old.ID, old.Generation, 24, now); err != nil {
		t.Fatal(err)
	}
	if _, err := book.Fail(lineID, now); err != nil {
		t.Fatal(err)
	}
	current, err := book.Start(lineID, "pump-B", 119, 2, now.Add(200*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := book.ConfirmFlow(lineID, current.ID, current.Generation, 26, now.Add(300*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	_, runErr := queue.ExecuteDue(context.Background(), now.Add(2*time.Second), func(ctx context.Context, compensation journal.Compensation) error {
		return book.ExecuteCompensation(ctx, compensation, now.Add(2*time.Second))
	})
	if runErr == nil {
		t.Fatal("stale compensation unexpectedly controlled the new pump owner")
	}
	state := pumps.Snapshot()
	var pumpB sludge.Pump
	for _, pump := range state {
		if pump.ID == "pump-B" {
			pumpB = pump
		}
	}
	if !pumpB.Running || pumpB.OwnerSequence != current.ID || pumpB.OwnerGeneration != current.Generation {
		t.Fatalf("new sequence lost pump ownership: %#v", pumpB)
	}
}
