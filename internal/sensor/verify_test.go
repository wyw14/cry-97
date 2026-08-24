package sensor_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/wyw14/cry-97/internal/journal"
	"github.com/wyw14/cry-97/internal/model"
	"github.com/wyw14/cry-97/internal/sensor"
)

type sampleAppender struct {
	fail   bool
	stored int
}

func (a *sampleAppender) Append(_ context.Context, event model.Event) (model.Event, error) {
	if a.fail {
		a.fail = false
		return model.Event{}, errors.New("no space left on device")
	}
	a.stored++
	return event, nil
}

func TestRejectedSampleCanReplayAfterRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cursor.json")
	cursor, err := journal.NewCursorStore(path)
	if err != nil {
		t.Fatal(err)
	}
	appender := &sampleAppender{fail: true}
	receiver, err := sensor.NewReceiver(appender, cursor)
	if err != nil {
		t.Fatal(err)
	}
	sample := model.NewSample(model.LineID("line-1"), "online-1", model.SampleCOD, 1, 31, "mg/L", time.Now())
	if err := receiver.Receive(context.Background(), sample, time.Now()); err == nil {
		t.Fatal("failed sample append was reported as successful")
	}
	restartedCursor, err := journal.NewCursorStore(path)
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := sensor.NewReceiver(appender, restartedCursor)
	if err != nil {
		t.Fatal(err)
	}
	if err := restarted.Receive(context.Background(), sample, time.Now()); err != nil {
		t.Fatalf("replayed sample was rejected: %v", err)
	}
	if appender.stored != 1 || restarted.Current(model.LineID("line-1")) != 1 {
		t.Fatalf("replay was skipped: stored=%d cursor=%d", appender.stored, restarted.Current(model.LineID("line-1")))
	}
}
