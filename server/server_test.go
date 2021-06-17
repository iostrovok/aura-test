package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	. "github.com/iostrovok/check"

	"github.com/iostrovok/aura-test/response"
	"github.com/iostrovok/aura-test/storage"
)

type testSuite struct{}

var _ = Suite(&testSuite{})

func TestSuite(t *testing.T) { TestingT(t) }

// helper.
func checkAllInStorage(c *C, id, all string) {
	c.Assert(strings.Index(all, "["), Equals, 0)
	c.Assert(strings.Index(all, "{"), Equals, 1)
	c.Assert(strings.LastIndex(all, "]"), Equals, len(all)-1)
	c.Assert(strings.LastIndex(all, "}"), Equals, len(all)-2)
	c.Assert(strings.LastIndex(all, id) > 0, Equals, true)
}

// helper.
func checkRemoteEmptyAllInStorage(c *C, url string) {
	res, err := http.Get(url + "/sessions")
	c.Assert(err, IsNil)
	c.Assert(string(readResponse(c, res)), Equals, "[]")
}

// helper.
func checkRemoteAllInStorage(c *C, url, id string) {
	res, err := http.Get(url + "/sessions")
	c.Assert(err, IsNil)
	c.Assert(res.StatusCode, Equals, http.StatusOK)
	checkAllInStorage(c, id, string(readResponse(c, res)))
}

func responseParser(c *C, res *http.Response) *response.Response {
	defer res.Body.Close()

	body := readResponse(c, res)
	out := &response.Response{}
	c.Assert(json.Unmarshal(body, &out), IsNil)

	return out
}

func readResponse(c *C, res *http.Response) []byte {
	out, err := io.ReadAll(res.Body)
	res.Body.Close()
	c.Assert(err, IsNil)

	return out
}

func CreateRequest(c *C, address, ttl string) *http.Response {
	client := http.Client{}

	fmt.Printf("CreateRequestCreateRequestCreateRequestCreateRequest+>val=> %+v\n\n", client)

	form := url.Values{}
	if ttl != "" {
		form.Add("TTL", ttl)
	}

	req, err := http.NewRequest(http.MethodPost, address+"/sessions", strings.NewReader(form.Encode()))
	c.Assert(err, IsNil)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	c.Assert(err, IsNil)

	return res
}

func DestroyRequest(c *C, url, id string) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, url+"/sessions/"+id, nil)
	c.Assert(err, IsNil)
	resp, err := client.Do(req)
	c.Assert(err, IsNil)

	return resp
}

func ExtendRequest(c *C, url, id, ttl string) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url+"/sessions/"+id+"/"+ttl, nil)
	c.Assert(err, IsNil)
	resp, err := client.Do(req)
	c.Assert(err, IsNil)

	return resp
}

func ListRequest(c *C, url string) *http.Response {
	res, err := http.Get(url + "/sessions")
	c.Assert(err, IsNil)

	return res
}

func (s *testSuite) TestBuild(c *C) {
	ctx := context.Background()
	keeper := storage.New(ctx)

	ts := httptest.NewServer(http.HandlerFunc(initSessionsHandlers(keeper)))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	c.Assert(err, IsNil)
	c.Assert(string(readResponse(c, res)), Equals, "[]")
}

func (s *testSuite) TestCreate(c *C) {
	ctx := context.Background()
	keeper := storage.New(ctx)

	ts := httptest.NewServer(http.HandlerFunc(initSessionsHandlers(keeper)))
	defer ts.Close()

	data := responseParser(c, CreateRequest(c, ts.URL, ""))
	c.Assert(data.Error, Equals, "")
	c.Assert(len(data.ID), Equals, 36) // UUID string
}

func (s *testSuite) TestCreateDelete(c *C) {
	ctx := context.Background()
	keeper := storage.New(ctx)

	ts := httptest.NewServer(http.HandlerFunc(initSessionsHandlers(keeper)))
	defer ts.Close()

	data := responseParser(c, CreateRequest(c, ts.URL, ""))
	checkRemoteAllInStorage(c, ts.URL, data.ID)

	res := DestroyRequest(c, ts.URL, data.ID)
	c.Assert(res.StatusCode, Equals, http.StatusOK)
	checkRemoteEmptyAllInStorage(c, ts.URL)

	// no more data with such id
	res = DestroyRequest(c, ts.URL, data.ID)
	c.Assert(res.StatusCode, Equals, http.StatusNotFound)
	checkRemoteEmptyAllInStorage(c, ts.URL)
}

func (s *testSuite) TestDelete404(c *C) {
	ctx := context.Background()
	keeper := storage.New(ctx)

	ts := httptest.NewServer(http.HandlerFunc(initSessionsHandlers(keeper)))
	defer ts.Close()

	data := responseParser(c, CreateRequest(c, ts.URL, "1"))
	checkRemoteAllInStorage(c, ts.URL, data.ID)

	res := DestroyRequest(c, ts.URL, "790c72b9-0000-0000-0000-000000000000")
	c.Assert(res.StatusCode, Equals, http.StatusNotFound)
	checkRemoteAllInStorage(c, ts.URL, data.ID)
}

func (s *testSuite) TestExpired(c *C) {
	ctx := context.Background()
	keeper := storage.New(ctx)

	ts := httptest.NewServer(http.HandlerFunc(initSessionsHandlers(keeper)))
	defer ts.Close()

	data := responseParser(c, CreateRequest(c, ts.URL, "1"))
	checkRemoteAllInStorage(c, ts.URL, data.ID)

	time.Sleep(2 * time.Second)
	checkRemoteEmptyAllInStorage(c, ts.URL)
}

func (s *testSuite) TestExtend(c *C) {
	ctx := context.Background()
	keeper := storage.New(ctx)

	ts := httptest.NewServer(http.HandlerFunc(initSessionsHandlers(keeper)))
	defer ts.Close()

	data := responseParser(c, CreateRequest(c, ts.URL, "1"))
	checkRemoteAllInStorage(c, ts.URL, data.ID)

	rep := ExtendRequest(c, ts.URL, data.ID, "10")
	c.Assert(rep.StatusCode, Equals, http.StatusOK)

	time.Sleep(2 * time.Second)
	checkRemoteAllInStorage(c, ts.URL, data.ID)
}

func (s *testSuite) TestWrongMethod(c *C) {
	ctx := context.Background()
	keeper := storage.New(ctx)

	ts := httptest.NewServer(http.HandlerFunc(initSessionsHandlers(keeper)))
	defer ts.Close()

	client := http.Client{}

	req, err := http.NewRequest(http.MethodHead, ts.URL+"/sessions", nil)
	c.Assert(err, IsNil)

	res, err := client.Do(req)
	c.Assert(err, IsNil)
	res.Body.Close()
	c.Assert(res.StatusCode, Equals, http.StatusMethodNotAllowed)
}

func (s *testSuite) TestList(c *C) {
	ctx := context.Background()
	keeper := storage.New(ctx)

	ts := httptest.NewServer(http.HandlerFunc(initSessionsHandlers(keeper)))
	defer ts.Close()

	for i := 0; i < 1000; i++ {
		CreateRequest(c, ts.URL, "10")
	}

	res := ListRequest(c, ts.URL)
	c.Assert(res.StatusCode, Equals, http.StatusOK)
	out, err := io.ReadAll(res.Body)
	res.Body.Close()
	c.Assert(err, IsNil)

	data := make([]*response.List, 0)
	err = json.Unmarshal(out, &data)
	c.Assert(err, IsNil)
	c.Assert(len(data), Equals, 1000)
}
