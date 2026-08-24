package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/wyw14/cry-97/internal/api"
	"github.com/wyw14/cry-97/internal/model"
	"github.com/wyw14/cry-97/internal/process"
)

func TestMalformedLabResultCannotQualifyBatch(t *testing.T) {
	clock := func() time.Time { return time.Date(2026, 8, 24, 4, 0, 0, 0, time.UTC) }
	plant, err := process.NewPlant(t.TempDir(), []model.LineID{"line-1"}, clock)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := plant.StartBatch(t.Context(), model.LineID("line-1"), 600, "north")
	if err != nil {
		t.Fatal(err)
	}
	for _, stage := range []model.ProcessStage{model.StageEqualizing, model.StageBiological, model.StageSettling} {
		if _, err := plant.AdvanceBatch(t.Context(), model.LineID("line-1"), stage); err != nil {
			t.Fatal(err)
		}
	}
	server, err := api.NewServer(plant, fstest.MapFS{"process.html": {Data: []byte("ok")}})
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"sample_id":"7a3d1d6f-3559-4acf-b68e-d52a3fcc5d88","line_id":"line-1","batch_id":"` + batch.ID.String() + `","revision":4,"cod":18,"ammonia":`)
	request := httptest.NewRequest(http.MethodPost, "/api/lab/results", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	current, ok := plant.ActiveBatch(model.LineID("line-1"))
	if !ok || current.Stage != model.StageSettling || current.Qualified {
		t.Fatalf("malformed result changed batch: %#v", current)
	}
	if len(plant.Permits()) != 0 {
		t.Fatalf("malformed result created discharge permit: %#v", plant.Permits())
	}
}
