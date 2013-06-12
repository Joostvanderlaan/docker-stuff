// Copyright 2013 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/globocom/config"
	"github.com/globocom/tsuru/auth"
	"github.com/globocom/tsuru/cmd"
	"launchpad.net/gnuflag"
)

type tokenCmd struct {
	fs     *gnuflag.FlagSet
	config string
}

func (c *tokenCmd) Run(context *cmd.Context, client *cmd.Client) error {
	err := config.ReadAndWatchConfigFile(c.config)
	if err != nil {
		return err
	}
	t, err := auth.CreateApplicationToken("tsr")
	if err != nil {
		return err
	}
	fmt.Fprintf(context.Stdout, t.Token)
	return nil
}

func (tokenCmd) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "token",
		Usage:   "token",
		Desc:    "Generates a tsuru token.",
		MinArgs: 0,
	}
}

func (c *tokenCmd) Flags() *gnuflag.FlagSet {
	if c.fs == nil {
		c.fs = gnuflag.NewFlagSet("token", gnuflag.ExitOnError)
		c.fs.StringVar(&c.config, "config", "/etc/tsuru/tsuru.conf", "tsr collector config file.")
		c.fs.StringVar(&c.config, "c", "/etc/tsuru/tsuru.conf", "tsr collector config file.")
	}
	return c.fs
}
