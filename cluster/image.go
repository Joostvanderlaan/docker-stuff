// Copyright 2013 docker-cluster authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	dcli "github.com/fsouza/go-dockerclient"
	"io"
)

// RemoveImage removes an image from all nodes in the cluster, returning an
// error in case of failure.
func (c *Cluster) RemoveImage(name string) error {
	_, err := c.runOnNodes(func(n node) (interface{}, error) {
		return nil, n.RemoveImage(name)
	}, dcli.ErrNoSuchImage, false)
	return err
}

// PullImage pulls an image from a remote registry server, returning an error
// in case of failure.
func (c *Cluster) PullImage(opts dcli.PullImageOptions, w io.Writer) error {
	_, err := c.runOnNodes(func(n node) (interface{}, error) {
		return nil, n.PullImage(opts, w)
	}, dcli.ErrNoSuchImage, true)
	return err
}

// PushImage pushes an image to a remote registry server, returning an error in
// case of failure.
func (c *Cluster) PushImage(opts dcli.PushImageOptions, auth dcli.AuthConfiguration, w io.Writer) error {
	if node, err := c.getNodeForImage(opts.Name); err == nil {
		return node.PushImage(opts, auth, w)
	} else if err != errStorageDisabled {
		return err
	}
	_, err := c.runOnNodes(func(n node) (interface{}, error) {
		return nil, n.PushImage(opts, auth, w)
	}, dcli.ErrNoSuchImage, false)
	return err
}

func (c *Cluster) getNodeForImage(image string) (node, error) {
	return c.getNode(func(s Storage) (string, error) {
		return s.RetrieveImage(image)
	})
}

// ImportImage imports an image from a url or stdin
func (c *Cluster) ImportImage(opts dcli.ImportImageOptions, w io.Writer) error {
	_, err := c.runOnNodes(func(n node) (interface{}, error) {
		return nil, n.ImportImage(opts, w)
	}, dcli.ErrNoSuchImage, false)
	return err
}
