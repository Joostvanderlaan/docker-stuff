// Copyright 2013 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"bytes"
	"github.com/dotcloud/docker"
	dcli "github.com/fsouza/go-dockerclient"
	"github.com/fsouza/go-dockerclient/testing"
	"github.com/globocom/config"
	"github.com/globocom/docker-cluster/cluster"
	"github.com/globocom/tsuru/app"
	"github.com/globocom/tsuru/db"
	"labix.org/v2/mgo/bson"
	"launchpad.net/gocheck"
)

type SchedulerSuite struct {
	storage *db.Storage
}

var _ = gocheck.Suite(&SchedulerSuite{})

func (s *SchedulerSuite) SetUpSuite(c *gocheck.C) {
	var err error
	config.Set("database:url", "127.0.0.1:27017")
	config.Set("database:name", "docker_scheduler_tests")
	config.Set("docker:repository-namespace", "tsuru")
	s.storage, err = db.Conn()
	c.Assert(err, gocheck.IsNil)
}

func (s *SchedulerSuite) TearDownSuite(c *gocheck.C) {
	s.storage.Apps().Database.DropDatabase()
	s.storage.Close()
}

func (s *SchedulerSuite) TestSchedulerSchedule(c *gocheck.C) {
	server0, err := testing.NewServer(nil)
	c.Assert(err, gocheck.IsNil)
	defer server0.Stop()
	server1, err := testing.NewServer(nil)
	c.Assert(err, gocheck.IsNil)
	defer server1.Stop()
	server2, err := testing.NewServer(nil)
	c.Assert(err, gocheck.IsNil)
	defer server2.Stop()
	var buf bytes.Buffer
	client, _ := dcli.NewClient(server0.URL())
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/python"}, &buf)
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/impius"}, &buf)
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/mirror"}, &buf)
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/dedication"}, &buf)
	client, _ = dcli.NewClient(server1.URL())
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/python"}, &buf)
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/impius"}, &buf)
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/mirror"}, &buf)
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/dedication"}, &buf)
	client, _ = dcli.NewClient(server2.URL())
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/python"}, &buf)
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/impius"}, &buf)
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/mirror"}, &buf)
	client.PullImage(dcli.PullImageOptions{Repository: "tsuru/dedication"}, &buf)
	a1 := app.App{Name: "impius", Teams: []string{"tsuruteam", "nodockerforme"}}
	a2 := app.App{Name: "mirror", Teams: []string{"tsuruteam"}}
	a3 := app.App{Name: "dedication", Teams: []string{"nodockerforme"}}
	err = s.storage.Apps().Insert(a1, a2, a3)
	c.Assert(err, gocheck.IsNil)
	defer s.storage.Apps().Remove(bson.M{"name": bson.M{"$in": []string{a1.Name, a2.Name, a3.Name}}})
	coll := s.storage.Collection(schedulerCollection)
	err = coll.Insert(
		node{ID: "server0", Address: server0.URL(), Team: "tsuruteam"},
		node{ID: "server1", Address: server1.URL(), Team: "tsuruteam"},
		node{ID: "server2", Address: server2.URL()},
	)
	c.Assert(err, gocheck.IsNil)
	defer coll.Remove(bson.M{"_id": bson.M{"$in": []string{"server0", "server1", "server2"}}})
	var scheduler segregatedScheduler
	config := docker.Config{Cmd: []string{"/usr/sbin/sshd", "-D"}, Image: "tsuru/impius"}
	node, _, err := scheduler.Schedule(&config)
	c.Assert(err, gocheck.IsNil)
	c.Check(node, gocheck.Equals, "server2")
	config = docker.Config{Cmd: []string{"/usr/sbin/sshd", "-D"}, Image: "tsuru/mirror"}
	node, _, err = scheduler.Schedule(&config)
	c.Assert(err, gocheck.IsNil)
	c.Check(node == "server0" || node == "server1", gocheck.Equals, true)
	config = docker.Config{Cmd: []string{"/usr/sbin/sshd", "-D"}, Image: "tsuru/python"}
	node, _, err = scheduler.Schedule(&config)
	c.Assert(err, gocheck.IsNil)
	c.Check(node, gocheck.Equals, "server2")
	config = docker.Config{Cmd: []string{"/usr/sbin/sshd", "-D"}, Image: "tsuru/dedication"}
	node, _, err = scheduler.Schedule(&config)
	c.Assert(err, gocheck.IsNil)
	c.Check(node, gocheck.Equals, "server2")
	config = docker.Config{Cmd: []string{"/usr/sbin/sshd", "-D"}}
	node, _, _ = scheduler.Schedule(&config)
	c.Check(node, gocheck.Equals, "server2")
}

func (s *SchedulerSuite) TestSchedulerNoFallback(c *gocheck.C) {
	app := app.App{Name: "bill", Teams: []string{"jean"}}
	err := s.storage.Apps().Insert(app)
	c.Assert(err, gocheck.IsNil)
	defer s.storage.Apps().Remove(bson.M{"name": app.Name})
	config := docker.Config{Cmd: []string{"/usr/sbin/sshd", "-D"}, Image: "tsuru/python"}
	var scheduler segregatedScheduler
	node, container, err := scheduler.Schedule(&config)
	c.Assert(node, gocheck.Equals, "")
	c.Assert(container, gocheck.IsNil)
	c.Assert(err, gocheck.Equals, errNoFallback)
}

func (s *SchedulerSuite) TestSchedulerNoNamespace(c *gocheck.C) {
	old, _ := config.Get("docker:repository-namespace")
	defer config.Set("docker:repository-namespace", old)
	config.Unset("docker:repository-namespace")
	config := docker.Config{Cmd: []string{"/usr/sbin/sshd", "-D"}, Image: "tsuru/python"}
	var scheduler segregatedScheduler
	node, container, err := scheduler.Schedule(&config)
	c.Assert(node, gocheck.Equals, "")
	c.Assert(container, gocheck.IsNil)
	c.Assert(err, gocheck.NotNil)
}

func (s *SchedulerSuite) TestSchedulerInvalidEndpoint(c *gocheck.C) {
	app := app.App{Name: "bill", Teams: []string{"jean"}}
	err := s.storage.Apps().Insert(app)
	c.Assert(err, gocheck.IsNil)
	defer s.storage.Apps().Remove(bson.M{"name": app.Name})
	coll := s.storage.Collection(schedulerCollection)
	err = coll.Insert(node{ID: "server0", Address: "", Team: "jean"})
	c.Assert(err, gocheck.IsNil)
	defer coll.Remove(bson.M{"_id": "server0"})
	config := docker.Config{Cmd: []string{"/usr/sbin/sshd", "-D"}, Image: "tsuru/bill"}
	var scheduler segregatedScheduler
	node, container, err := scheduler.Schedule(&config)
	c.Assert(node, gocheck.Equals, "server0")
	c.Assert(container, gocheck.IsNil)
	c.Assert(err, gocheck.NotNil)
}

func (s *SchedulerSuite) TestSchedulerNodes(c *gocheck.C) {
	coll := s.storage.Collection(schedulerCollection)
	err := coll.Insert(
		node{ID: "server0", Address: "http://localhost:8080", Team: "tsuru"},
		node{ID: "server1", Address: "http://localhost:8081", Team: "tsuru"},
		node{ID: "server2", Address: "http://localhost:8082", Team: "tsuru"},
	)
	c.Assert(err, gocheck.IsNil)
	defer coll.RemoveAll(bson.M{"_id": bson.M{"$in": []string{"server0", "server1", "server2"}}})
	expected := []cluster.Node{
		{ID: "server0", Address: "http://localhost:8080"},
		{ID: "server1", Address: "http://localhost:8081"},
		{ID: "server2", Address: "http://localhost:8082"},
	}
	var scheduler segregatedScheduler
	nodes, err := scheduler.Nodes()
	c.Assert(err, gocheck.IsNil)
	c.Assert(nodes, gocheck.DeepEquals, expected)
}
