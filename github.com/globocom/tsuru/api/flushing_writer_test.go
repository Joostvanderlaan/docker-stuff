// Copyright 2013 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"launchpad.net/gocheck"
	"net/http/httptest"
)

type FlushingSuite struct{}

var _ = gocheck.Suite(&FlushingSuite{})

func (s *FlushingSuite) TestFlushingWriter(c *gocheck.C) {
	recorder := httptest.NewRecorder()
	writer := flushingWriter{recorder, false}
	data := []byte("ble")
	_, err := writer.Write(data)
	c.Assert(err, gocheck.IsNil)
	c.Assert(recorder.Body.Bytes(), gocheck.DeepEquals, data)
	c.Assert(writer.wrote, gocheck.Equals, true)
}

func (s *FlushingSuite) TestFlushingWriterShouldReturnTheDataSize(c *gocheck.C) {
	recorder := httptest.NewRecorder()
	writer := flushingWriter{recorder, false}
	data := []byte("ble")
	n, err := writer.Write(data)
	c.Assert(err, gocheck.IsNil)
	c.Assert(n, gocheck.Equals, len(data))
}

func (s *FlushingSuite) TestFlushingWriterHeader(c *gocheck.C) {
	recorder := httptest.NewRecorder()
	writer := flushingWriter{recorder, false}
	writer.Header().Set("Content-Type", "application/xml")
	c.Assert(recorder.Header().Get("Content-Type"), gocheck.Equals, "application/xml")
}

func (s *FlushingSuite) TestFlushingWriterWriteHeader(c *gocheck.C) {
	recorder := httptest.NewRecorder()
	writer := flushingWriter{recorder, false}
	expectedCode := 333
	writer.WriteHeader(expectedCode)
	c.Assert(recorder.Code, gocheck.Equals, expectedCode)
	c.Assert(writer.wrote, gocheck.Equals, true)
}
