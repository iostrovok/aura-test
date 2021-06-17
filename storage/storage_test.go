package storage

import (
	"context"
	"strings"
	"time"

	. "github.com/iostrovok/check"
)

// helper.
func checkAllInStorage(c *C, id, all string) {
	c.Assert(strings.Index(all, "["), Equals, 0)
	c.Assert(strings.Index(all, "{"), Equals, 1)
	c.Assert(strings.LastIndex(all, "]"), Equals, len(all)-1)
	c.Assert(strings.LastIndex(all, "}"), Equals, len(all)-2)
	c.Assert(strings.LastIndex(all, id) > 0, Equals, true)
}

func (s *testSuite) TestStorageCreate(c *C) {
	storage := New(context.Background())
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")

	id := storage.Create(30)
	checkAllInStorage(c, id, string(storage.ListAllSessions()))
}

func (s *testSuite) TestStorageExpired(c *C) {
	storage := New(context.Background())
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")

	id := storage.Create(1)
	checkAllInStorage(c, id, string(storage.ListAllSessions()))

	time.Sleep(2 * time.Second)
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")
}

func (s *testSuite) TestStorageDestroy(c *C) {
	storage := New(context.Background())
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")

	id := storage.Create(30)
	checkAllInStorage(c, id, string(storage.ListAllSessions()))

	c.Assert(storage.Destroy(id), Equals, true)
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")

	c.Assert(storage.Destroy(id), Equals, false)
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")
}

func (s *testSuite) TestStorageDestroyMassive(c *C) {
	storage := New(context.Background())
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")

	ids := make([]string, 1000, 1000)
	for i := 0; i < 1000; i++ {
		ids[i] = storage.Create(30)
	}

	all := string(storage.ListAllSessions())
	for i := 0; i < 1000; i++ {
		checkAllInStorage(c, ids[i], all)
	}

	for i := 0; i < 1000; i++ {
		c.Assert(storage.Destroy(ids[i]), Equals, true)
	}
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")
	for i := 0; i < 1000; i++ {
		c.Assert(storage.Destroy(ids[i]), Equals, false)
	}
}

func (s *testSuite) TestStorageExtend(c *C) {
	storage := New(context.Background())
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")

	id := storage.Create(1)
	checkAllInStorage(c, id, string(storage.ListAllSessions()))

	c.Assert(storage.Extend(id, 10), Equals, true)
	checkAllInStorage(c, id, string(storage.ListAllSessions()))

	time.Sleep(2 * time.Second)
	checkAllInStorage(c, id, string(storage.ListAllSessions()))
}

func (s *testSuite) TestStorageExtendExpired(c *C) {
	storage := New(context.Background())
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")

	id := storage.Create(1)
	checkAllInStorage(c, id, string(storage.ListAllSessions()))

	time.Sleep(2 * time.Second)
	c.Assert(storage.Extend(id, 10), Equals, false)
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")
}

func (s *testSuite) TestStorageCreateManyRecords(c *C) {
	storage := New(context.Background())
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")

	for i := 0; i < 1000; i++ {
		storage.Create(10)
	}

	id := storage.Create(10)
	checkAllInStorage(c, id, string(storage.ListAllSessions()))
}

func (s *testSuite) TestStorageStopExpired(c *C) {
	ctx, cxtFunc := context.WithCancel(context.Background())
	storage := New(ctx)
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")
	cxtFunc()
	c.Assert(string(storage.ListAllSessions()), Equals, "[]")
}
