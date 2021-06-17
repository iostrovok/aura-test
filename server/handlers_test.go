package server

import (
	. "github.com/iostrovok/check"
)

func (s *testSuite) TestParseURL(c *C) {
	_, _, e := _parseURL("")
	c.Assert(e, NotNil)

	_, _, e = _parseURL("session")
	c.Assert(e, NotNil)

	_, _, e = _parseURL("/sessions/bla-bla-bla")
	c.Assert(e, NotNil)

	id, ttl, err := _parseURL("sessions")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, "")
	c.Assert(ttl, Equals, 30)

	id, ttl, err = _parseURL("sessions/")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, "")
	c.Assert(ttl, Equals, 30)

	id, ttl, err = _parseURL("/sessions")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, "")
	c.Assert(ttl, Equals, 30)

	testID := "c4d987da-8f47-49a0-8775-b28f39544e6c"

	id, ttl, err = _parseURL("/sessions/" + testID)
	c.Assert(err, IsNil)
	c.Assert(id, Equals, testID)
	c.Assert(ttl, Equals, 30)

	_, _, e = _parseURL("/sessions/" + testID + "/sadasdas")
	c.Assert(e, NotNil)

	_, _, e = _parseURL("/sessions/" + testID + "/999999999999999999999999999999")
	c.Assert(e, NotNil)

	id, ttl, err = _parseURL("/sessions/" + testID + "/5000")
	c.Assert(err, IsNil)
	c.Assert(id, Equals, testID)
	c.Assert(ttl, Equals, 300)

}
