// Copyright 2013 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docker

import (
	"github.com/globocom/config"
	"github.com/globocom/tsuru/app"
	"github.com/globocom/tsuru/db"
	"labix.org/v2/mgo/bson"
	"launchpad.net/gocheck"
)

type FlattenSuite struct {
	apps []app.App
	conn *db.Storage
}

var _ = gocheck.Suite(&FlattenSuite{})

func (s *FlattenSuite) SetUpSuite(c *gocheck.C) {
	var err error
	config.Set("docker:repository-namespace", "tsuru")
	config.Set("database:url", "127.0.0.1:27017")
	config.Set("database:name", "docker_flatten")
	s.conn, err = db.Conn()
	units := []app.Unit{{Name: "4fa6e0f0c678"}, {Name: "e90e34656806"}}
	app1 := app.App{Name: "app1", Platform: "python", Deploys: 40, Units: units}
	err = s.conn.Apps().Insert(app1)
	c.Assert(err, gocheck.IsNil)
	app2 := app.App{Name: "app2", Platform: "python", Deploys: 20, Units: units}
	err = s.conn.Apps().Insert(app2)
	c.Assert(err, gocheck.IsNil)
	app3 := app.App{Name: "app3", Platform: "python", Deploys: 3, Units: units}
	err = s.conn.Apps().Insert(app3)
	c.Assert(err, gocheck.IsNil)
	app4 := app.App{Name: "app4", Platform: "python", Deploys: 19, Units: units}
	err = s.conn.Apps().Insert(app4)
	c.Assert(err, gocheck.IsNil)
	s.apps = append(s.apps, []app.App{app1, app2, app3, app4}...)
}

func (s *FlattenSuite) TearDownSuite(c *gocheck.C) {
	names := make([]string, len(s.apps))
	for i, a := range s.apps {
		names[i] = a.GetName()
	}
	_, err := s.conn.Apps().RemoveAll(bson.M{"name": bson.M{"$in": names}})
	c.Assert(err, gocheck.IsNil)
}

func (s *FlattenSuite) TestImagesToFlattenRetrievesOnlyUnitsWith20DeploysOrMore(c *gocheck.C) {
	images := imagesToFlatten()
	c.Assert(len(images), gocheck.Equals, 2)
	expected := []string{"tsuru/app1", "tsuru/app2"}
	c.Assert(images, gocheck.DeepEquals, expected)
}
