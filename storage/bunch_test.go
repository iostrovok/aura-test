package storage

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	. "github.com/iostrovok/check"
)

type testSuite struct{}

var _ = Suite(&testSuite{})

func TestStorage(t *testing.T) { TestingT(t) }

// helper.
func checkAllInBunch(c *C, id, all string) {
	c.Assert(strings.Index(all, "{"), Equals, 0)
	c.Assert(strings.LastIndex(all, ","), Equals, len(all)-1)
	c.Assert(strings.LastIndex(all, "}"), Equals, len(all)-2)
	c.Assert(strings.LastIndex(all, id) > 0, Equals, true)
}

func (s *testSuite) TestCreate(c *C) {
	bunch := newBunch(context.Background())
	c.Assert(string(bunch.allSession()), Equals, "")

	id := uuid.New()
	bunch.create(id.String(), 30)

	all := string(bunch.allSession())
	c.Logf("all: %s\n", all)
	checkAllInBunch(c, id.String(), all)
}

func (s *testSuite) TestExpired(c *C) {
	bunch := newBunch(context.Background())
	c.Assert(string(bunch.allSession()), Equals, "")

	id := uuid.New()
	bunch.create(id.String(), 1)
	checkAllInBunch(c, id.String(), string(bunch.allSession()))

	time.Sleep(2 * time.Second)
	c.Assert(string(bunch.allSession()), Equals, "")
}

func (s *testSuite) TestDestroy(c *C) {
	bunch := newBunch(context.Background())
	c.Assert(string(bunch.allSession()), Equals, "")

	id := uuid.New()
	bunch.create(id.String(), 30)
	checkAllInBunch(c, id.String(), string(bunch.allSession()))

	c.Assert(bunch.destroy(id.String()), Equals, true)
	c.Assert(string(bunch.allSession()), Equals, "")

	c.Assert(bunch.destroy(id.String()), Equals, false)
	c.Assert(string(bunch.allSession()), Equals, "")
}

func (s *testSuite) TestDestroyMassive(c *C) {
	bunch := newBunch(context.Background())
	c.Assert(string(bunch.allSession()), Equals, "")

	ids := make([]string, 1000, 1000)
	for i := 0; i < 1000; i++ {
		id := uuid.New()
		bunch.create(id.String(), 30)
		ids[i] = id.String()
	}

	all := string(bunch.allSession())
	for i := 0; i < 1000; i++ {
		checkAllInBunch(c, ids[i], all)
	}

	for i := 0; i < 1000; i++ {
		c.Assert(bunch.destroy(ids[i]), Equals, true)
	}
	c.Assert(string(bunch.allSession()), Equals, "")
}

func (s *testSuite) TestExtend(c *C) {
	bunch := newBunch(context.Background())
	c.Assert(string(bunch.allSession()), Equals, "")

	id := uuid.New()
	bunch.create(id.String(), 1)
	c.Assert(bunch.extend(id.String(), 10), Equals, true)
	checkAllInBunch(c, id.String(), string(bunch.allSession()))

	time.Sleep(2 * time.Second)
	checkAllInBunch(c, id.String(), string(bunch.allSession()))
}

func (s *testSuite) TestExtendExpired(c *C) {
	bunch := newBunch(context.Background())
	c.Assert(string(bunch.allSession()), Equals, "")

	id := uuid.New()
	bunch.create(id.String(), 1)
	checkAllInBunch(c, id.String(), string(bunch.allSession()))

	time.Sleep(2 * time.Second)
	c.Assert(bunch.extend(id.String(), 10), Equals, false)
	c.Assert(string(bunch.allSession()), Equals, "")
}

func (s *testSuite) TestCreateManyRecords(c *C) {
	bunch := newBunch(context.Background())
	c.Assert(string(bunch.allSession()), Equals, "")

	for i := 0; i < 1000; i++ {
		bunch.create(uuid.New().String(), 10)
	}

	id := uuid.New()
	bunch.create(id.String(), 10)

	checkAllInBunch(c, id.String(), string(bunch.allSession()))
}

func (s *testSuite) TestStopExpired(c *C) {
	ctx, cxtFunc := context.WithCancel(context.Background())
	bunch := newBunch(ctx)
	c.Assert(string(bunch.allSession()), Equals, "")
	cxtFunc()
	c.Assert(string(bunch.allSession()), Equals, "")
}

func (s *testSuite) TestExtendTimeSession(c *C) {
	a, find := extendTimeSession(100, 200, 30)
	c.Assert(find, Equals, false)
	c.Assert(a, Equals, int64(0))

	a, find = extendTimeSession(200, 100, 30)
	c.Assert(find, Equals, true)
	c.Assert(a, Equals, int64(230))

	a, find = extendTimeSession(200, 100, 4000)
	c.Assert(find, Equals, true)
	c.Assert(a, Equals, int64(100+300))
}
