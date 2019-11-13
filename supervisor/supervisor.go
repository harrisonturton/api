// Supervisor takes a bunch of concurrent processes
// and makes sure they all start and stop at the same time.
// These processes must respect the Worker(Context, WaitGroup)
// interface – they stop when the context is cancelled, and call
// WaitGroup.Done when they finish.

package supervisor

import (
	"context"
	"sync"
)

// Worker is a goroutine that stops when the
// context is cancelled and respects the
// WaitGroup conventions – it calls WaitGroup.Done
// when it finishes.
type Worker interface {
	Run(context.Context, *sync.WaitGroup)
}

// Supervisor handles an array of goroutines,
// starting and stopping them all at once, safely.
type Supervisor struct {
	workers []Worker
	ctx     context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
}

// NewSupervisor will create a new supervisor instance
// that watches the workers passed to it.
func NewSupervisor(workers ...Worker) *Supervisor {
	wg := new(sync.WaitGroup)
	ctx, cancel := context.WithCancel(context.Background())
	return &Supervisor{workers, ctx, cancel, wg}
}

// Start will start all the workers
func (supervisor *Supervisor) Start() {
	for _, worker := range supervisor.workers {
		supervisor.wg.Add(1)
		go worker.Run(supervisor.ctx, supervisor.wg)
	}
}

// Run will start all the workers, and stop them when the context
// is cancelled. This fulfills the Worker interface, allowing
// supervisors to be nested!
func (supervisor *Supervisor) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	supervisor.Start()
	<-ctx.Done()
	supervisor.Stop()
}

// Stop will stop all the workers, and block
// until they finish.
func (supervisor *Supervisor) Stop() {
	supervisor.cancel()
	supervisor.wg.Wait()
}
