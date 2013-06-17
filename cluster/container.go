// Copyright 2013 docker-cluster authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"github.com/dotcloud/docker"
	dcli "github.com/fsouza/go-dockerclient"
	"net/http"
	"sync"
)

// CreateContainer creates a container in one of the nodes.
//
// It returns the ID of the node where the container is running, and the
// container, or an error, in case of failures. It will always return the ID of
// the node, even in case of failures.
func (c *Cluster) CreateContainer(config *docker.Config) (string, *docker.Container, error) {
	node := c.next()
	container, err := node.CreateContainer(config)
	return node.id, container, err
}

// InspectContainer returns information about a container by its ID, getting
// the information from the right node.
func (c *Cluster) InspectContainer(id string) (*docker.Container, error) {
	container, err := c.runOnNodes(func(n node) (interface{}, error) {
		return n.InspectContainer(id)
	})
	if err != nil {
		return nil, err
	}
	return container.(*docker.Container), err
}

// KillContainer kills a container, returning an error in case of failure.
func (c *Cluster) KillContainer(id string) error {
	_, err := c.runOnNodes(func(n node) (interface{}, error) {
		return nil, n.KillContainer(id)
	})
	return err
}

// ListContainers returns a slice of all containers in the cluster matching the
// given criteria.
func (c *Cluster) ListContainers(opts dcli.ListContainersOptions) ([]docker.APIContainers, error) {
	c.mut.RLock()
	defer c.mut.RUnlock()
	var wg sync.WaitGroup
	result := make(chan []docker.APIContainers, len(c.nodes))
	errs := make(chan error, len(c.nodes))
	for _, n := range c.nodes {
		wg.Add(1)
		go func(n node) {
			defer wg.Done()
			if containers, err := n.ListContainers(opts); err != nil {
				errs <- err
			} else {
				result <- containers
			}
		}(n)
	}
	wg.Wait()
	var group []docker.APIContainers
	var err error
	for {
		select {
		case containers := <-result:
			group = append(group, containers...)
		case err = <-errs:
		default:
			return group, err
		}
	}
}

// RemoveContainer removes a container from the cluster.
func (c *Cluster) RemoveContainer(id string) error {
	_, err := c.runOnNodes(func(n node) (interface{}, error) {
		return nil, n.RemoveContainer(id)
	})
	return err
}

func (c *Cluster) StartContainer(id string) error {
	_, err := c.runOnNodes(func(n node) (interface{}, error) {
		return nil, n.StartContainer(id)
	})
	return err
}

// StopContainer stops a container, killing it after the given timeout, if it
// fails to stop nicely.
func (c *Cluster) StopContainer(id string, timeout uint) error {
	_, err := c.runOnNodes(func(n node) (interface{}, error) {
		return nil, n.StopContainer(id, timeout)
	})
	return err
}

// RestartContainer restarts a container, killing it after the given timeout,
// if it fails to stop nicely.
func (c *Cluster) RestartContainer(id string, timeout uint) error {
	_, err := c.runOnNodes(func(n node) (interface{}, error) {
		return nil, n.RestartContainer(id, timeout)
	})
	return err
}

// WaitContainer blocks until the given container stops, returning the exit
// code of the container command.
func (c *Cluster) WaitContainer(id string) (int, error) {
	exit, err := c.runOnNodes(func(n node) (interface{}, error) {
		return n.WaitContainer(id)
	})
	return exit.(int), err
}

// AttachToContainer attaches to a container, using the given options.
func (c *Cluster) AttachToContainer(opts dcli.AttachToContainerOptions) error {
	_, err := c.runOnNodes(func(n node) (interface{}, error) {
		return nil, n.AttachToContainer(opts)
	})
	return err
}

// CommitContainer commits a container and returns the image id.
func (c *Cluster) CommitContainer(opts dcli.CommitContainerOptions) (*docker.Image, error) {
	image, err := c.runOnNodes(func(n node) (interface{}, error) {
		return n.CommitContainer(opts)
	})
	return image.(*docker.Image), err
}

type containerFunc func(node) (interface{}, error)

func (c *Cluster) runOnNodes(fn containerFunc) (interface{}, error) {
	c.mut.RLock()
	defer c.mut.RUnlock()
	var wg sync.WaitGroup
	finish := make(chan int8, 1)
	errChan := make(chan error, len(c.nodes))
	result := make(chan interface{}, 1)
	for _, n := range c.nodes {
		wg.Add(1)
		go func(n node) {
			defer wg.Done()
			value, err := fn(n)
			if err == nil {
				result <- value
			} else if e, ok := err.(*dcli.Error); ok && e.Status == http.StatusNotFound {
				result <- nil
			} else if err != dcli.ErrNoSuchContainer {
				errChan <- err
			}
		}(n)
	}
	go func() {
		wg.Wait()
		close(finish)
	}()
	select {
	case value := <-result:
		return value, nil
	case err := <-errChan:
		return nil, err
	case <-finish:
		return nil, dcli.ErrNoSuchContainer
	}
}
