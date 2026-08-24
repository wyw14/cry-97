package api_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-97/internal/api"
	"github.com/wyw14/cry-97/internal/dosing"
	"github.com/wyw14/cry-97/internal/model"
)

type rejectedJournal struct{}

func (rejectedJournal) Append(context.Context, model.Event) (model.Event, error) {
	return model.Event{}, errors.New("no space left on device")
}

type commandRecorder struct {
	mu       sync.Mutex
	commands []model.DeviceCommand
}

func (r *commandRecorder) Publish(_ context.Context, command model.DeviceCommand) error {
	r.mu.Lock()
	r.commands = append(r.commands, command)
	r.mu.Unlock()
	return nil
}

func (r *commandRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.commands)
}

type dosingAdapter struct {
	service *dosing.Service
}

func (a dosingAdapter) ChangeDose(ctx context.Context, lineID model.LineID, batchID uuid.UUID, chemical string, rate float64) (model.DeviceCommand, error) {
	change, err := dosing.NewChange(lineID, batchID, chemical, rate, 4, time.Now())
	if err != nil {
		return model.DeviceCommand{}, err
	}
	return a.service.ChangeRate(ctx, change, time.Now())
}

func TestDoseChangeRequiresDurableBatch(t *testing.T) {
	recorder := &commandRecorder{}
	service, err := dosing.NewService(rejectedJournal{}, recorder, dosing.NewLedger())
	if err != nil {
		t.Fatal(err)
	}
	handler := api.NewDoseHandler(dosingAdapter{service: service})
	body := []byte(`{"line_id":"line-1","batch_id":"` + uuid.NewString() + `","chemical":"pac","rate":14.2}`)
	request := httptest.NewRequest(http.MethodPost, "/api/dosing", bytes.NewReader(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if recorder.count() != 0 {
		t.Fatalf("published %d device commands before durable batch", recorder.count())
	}
}
