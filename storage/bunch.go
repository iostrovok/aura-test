package storage

import (
	"bytes"
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/cornelk/hashmap"
)

const (
	numberCyclesForReloadTime = 200
	MaxAllowedExtendedTTL     = int64(300)
	cleanerDelay              = 2 * time.Second
)

type Bunch struct {
	sync.RWMutex
	ctx      context.Context
	sessions *hashmap.HashMap
}

func newBunch(ctx context.Context) *Bunch {
	bunch := &Bunch{
		ctx:      ctx,
		sessions: &hashmap.HashMap{},
	}

	// run cleaner
	go bunch.deleteExpired(ctx)

	return bunch
}

// create creates new session. Always success.
func (b *Bunch) create(uuid string, ttl uint32) {
	// non-blocking operation
	s := time.Now().Unix() + int64(ttl)
	b.sessions.Set(uuid, s)
}

func (b *Bunch) extend(uuid string, ttl uint32) bool {
	// blocking operation
	b.RLock()
	defer b.RUnlock()

	value, ok := b.sessions.Get(uuid)
	if ok && value != nil {
		if newSession, find := extendTimeSession(value.(int64), time.Now().Unix(), int64(ttl)); find {
			// session is not expired now
			b.sessions.Set(uuid, newSession)

			return true
		}
	}

	return false
}

func (b *Bunch) destroy(id string) bool {
	// blocking operation
	b.Lock()
	defer b.Unlock()

	if _, ok := b.sessions.Get(id); ok {
		b.sessions.Del(id)

		return true
	}

	return false
}

func (b *Bunch) list(resCh chan []byte) {
	select {
	case <-b.ctx.Done():
	case resCh <- b.allSession():
	}
}

func (b *Bunch) allSession() []byte {
	buffer := bytes.NewBuffer([]byte{})

	now := time.Now().Unix()
	counter := 0
	for i := range b.sessions.Iter() {
		if i.Value != nil {
			if ttl := i.Value.(int64) - now; ttl > 0 { // session is not expired now
				buffer.WriteString(`{"id":"` + i.Key.(string) + `","ttl":` + strconv.FormatInt(ttl, 10) + `},`)
			}
		}

		// it does >> 1000 cycles per second so it doesn't make sense get time each times.
		counter++
		if counter > numberCyclesForReloadTime {
			counter = 0
			now = time.Now().Unix()
		}
	}

	// send final data
	return buffer.Bytes()
}

func (b *Bunch) deleteExpired(ctx context.Context) {
	// it's deleting sessions which are already expired.
	// it's not real method.
	for {
		select {
		case <-ctx.Done():
			// game over
			return
		case <-time.After(cleanerDelay):
			for i := range b.sessions.Iter() {
				if i.Value != nil {
					if i.Value.(int64) <= time.Now().Unix() {
						b.sessions.Del(i.Key)
					}
				}
			}
		}
	}
}

func extendTimeSession(session, now, ttl int64) (int64, bool) {
	if session < now { // session is expired
		return 0, false
	}

	session += ttl
	if session-now < MaxAllowedExtendedTTL {
		return session, true
	}

	return now + MaxAllowedExtendedTTL, true
}
