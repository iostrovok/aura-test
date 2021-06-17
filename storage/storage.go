package storage

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

const (
	// CountBunches is a number of bunches.
	// It dependents on now server/node performance and count of session.
	CountBunches = 100 // TODO: move to the configuration parameters or make calculable by server configuration
)

type Storage struct {
	sync.RWMutex

	Bunches       map[uint32]*Bunch
	CountBunches  uint32
	ctx           context.Context
	countSessions *int32
}

// New is a simple constructor.
func New(ctx context.Context) *Storage {
	s := &Storage{
		ctx:           ctx,
		Bunches:       make(map[uint32]*Bunch, CountBunches),
		CountBunches:  CountBunches,
		countSessions: new(int32),
	}

	for i := uint32(0); i < s.CountBunches; i++ {
		s.Bunches[i] = newBunch(s.ctx)
	}

	return s
}

/*
 * Interface functions
 */

func (s *Storage) Create(ttl uint32) string {
	id := uuid.New()
	s.getBunches(id).create(id.String(), ttl)

	return id.String()
}

func (s *Storage) Extend(id string, ttl uint32) bool {
	u := uuid.MustParse(id)

	return s.getBunches(u).extend(id, ttl)
}

func (s *Storage) Destroy(id string) bool {
	u := uuid.MustParse(id)

	return s.getBunches(u).destroy(id)
}

// ListAllSessions returns list of all active sessions and remaining TTL from all bunches.
func (s *Storage) ListAllSessions() []byte {
	/*
		1) create channel for getting result
		2) Start reading data from each bunch.
		3) Collect date from bunches and return result.
	*/

	// run collecting data on all bunches
	resCh := make(chan []byte, len(s.Bunches)+1)
	wg := sync.WaitGroup{}
	wg.Add(int(s.CountBunches))
	for i := uint32(0); i < s.CountBunches; i++ {
		go func(b *Bunch) {
			b.list(resCh)
			wg.Done()
		}(s.Bunches[i])
	}

	// collect data from all bunches
	done := make(chan struct{})
	out := []byte("[")
	go func() {
		defer func() {
			done <- struct{}{}
		}()

		for data := range resCh {
			out = append(out, data...)
		}
	}()

	// wait for all bunches
	wg.Wait()
	close(resCh)

	// wait for collect all results is done.
	<-done

	// replace last "," to "]"
	if len(out) > 1 {
		out[len(out)-1] = ']'
	} else {
		// empty list
		out = []byte("[]")
	}

	// final result
	return out
}

/*
 * Internal functions
 */

func (s *Storage) getBunches(u uuid.UUID) *Bunch {
	s.Lock()
	defer s.Unlock()

	return s.Bunches[u.ID()%s.CountBunches]
}
