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

func TestBlanketAlarmDoesNotBlockStateRead(t *testing.T) {
	plant, err := process.NewPlant(t.TempDir(), []model.LineID{"line-2"}, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	serverAPI, err := api.NewServer(plant, fstest.MapFS{"process.html": {Data: []byte("ok")}})
	if err != nil {
		t.Fatal(err)
	}
	updateDone := make(chan error, 1)
	go func() {
		request := httptest.NewRequest(http.MethodPost, "/api/settling/line-2/blanket", bytes.NewBufferString(`{"level":3.2}`))
		response := httptest.NewRecorder()
		serverAPI.Handler().ServeHTTP(response, request)
		var err error
		if response.Code != http.StatusAccepted {
			err = &statusError{code: response.Code}
		}
		updateDone <- err
	}()
	time.Sleep(20 * time.Millisecond)
	readDone := make(chan error, 1)
	go func() {
		request := httptest.NewRequest(http.MethodGet, "/api/process-lines/line-2", nil)
		response := httptest.NewRecorder()
		serverAPI.Handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			readDone <- &statusError{code: response.Code}
			return
		}
		readDone <- nil
	}()
	select {
	case readErr := <-readDone:
		if readErr != nil {
			t.Fatalf("state query failed after blanket alarm: %v", readErr)
		}
	case <-time.After(350 * time.Millisecond):
		t.Fatal("state query blocked after blanket alarm")
	}
	select {
	case updateErr := <-updateDone:
		if updateErr != nil {
			t.Fatalf("blanket update did not finish: %v", updateErr)
		}
	case <-time.After(350 * time.Millisecond):
		t.Fatal("blanket update did not finish")
	}
	if len(plant.Alarms()) != 1 {
		t.Fatalf("alarm was not published: %#v", plant.Alarms())
	}
}

type statusError struct{ code int }

func (e *statusError) Error() string { return http.StatusText(e.code) }
