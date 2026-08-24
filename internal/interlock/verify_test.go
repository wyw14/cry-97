package interlock_test

import (
	"sync"
	"testing"
	"time"

	"github.com/wyw14/cry-97/internal/interlock"
	"github.com/wyw14/cry-97/internal/model"
)

func TestBackwashAndDrainReserveOneRoute(t *testing.T) {
	for attempt := 0; attempt < 12; attempt++ {
		arbiter := interlock.NewArbiter()
		backwash, _ := interlock.NewRequest(model.LineID("line-2"), interlock.OperationBackwash, "BW2", []string{"line:line-2", "valve:shared-2"}, time.Now())
		drain, _ := interlock.NewRequest(model.LineID("line-2"), interlock.OperationDrain, "SD2", []string{"line:line-2", "valve:shared-2"}, time.Now())
		start := make(chan struct{})
		results := make(chan error, 2)
		var wait sync.WaitGroup
		for _, request := range []interlock.Request{backwash, drain} {
			request := request
			wait.Add(1)
			go func() {
				defer wait.Done()
				<-start
				_, err := arbiter.Reserve(request, time.Now())
				results <- err
			}()
		}
		close(start)
		wait.Wait()
		close(results)
		successes := 0
		for err := range results {
			if err == nil {
				successes++
			}
		}
		if successes != 1 || len(arbiter.Active()) != 1 {
			t.Fatalf("attempt %d admitted %d conflicting routes", attempt, successes)
		}
	}
}
