// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package clab

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/srl-labs/containerlab/nodes"
	_ "github.com/srl-labs/containerlab/nodes/all"
	"github.com/srl-labs/containerlab/runtime"
	_ "github.com/srl-labs/containerlab/runtime/all"
	"github.com/srl-labs/containerlab/types"
)

type CLab struct {
	Config        *Config   `json:"config,omitempty"`
	TopoFile      *TopoFile `json:"topofile,omitempty"`
	m             *sync.RWMutex
	Nodes         map[string]nodes.Node               `json:"nodes,omitempty"`
	Links         map[int]*types.Link                 `json:"links,omitempty"`
	Runtimes      map[string]runtime.ContainerRuntime `json:"runtimes,omitempty"`
	globalRuntime string                              `json:"global-runtime,omitempty"`
	Dir           *Directory                          `json:"dir,omitempty"`

	timeout time.Duration `json:"timeout,omitempty"`
}

type Directory struct {
	Lab       string
	LabCA     string
	LabCARoot string
	LabGraph  string
}

type ClabOption func(c *CLab) error

func WithTimeout(dur time.Duration) ClabOption {
	return func(c *CLab) error {
		if dur <= 0 {
			return errors.New("zero or negative timeouts are not allowed")
		}
		c.timeout = dur
		return nil
	}
}

func WithRuntime(name string, rtconfig *runtime.RuntimeConfig) ClabOption {
	return func(c *CLab) error {
		// define runtime name.
		// order of preference: cli flag -> env var -> default value of docker
		envN := os.Getenv("CLAB_RUNTIME")
		log.Debugf("envN runtime var value is %v", envN)
		switch {
		case name != "":
		case envN != "":
			name = envN
		default:
			name = runtime.DockerRuntime
		}
		c.globalRuntime = name

		if rInit, ok := runtime.ContainerRuntimes[name]; ok {
			r := rInit()
			log.Debugf("Running runtime.Init with params %+v and %+v", rtconfig, c.Config.Mgmt)
			err := r.Init(
				runtime.WithConfig(rtconfig),
				runtime.WithMgmtNet(c.Config.Mgmt),
			)
			if err != nil {
				return fmt.Errorf("failed to init the container runtime: %v", err)
			}

			c.Runtimes[name] = r
			log.Debugf("initialized a runtime with params %+v", r)
			return nil
		}
		return fmt.Errorf("unknown container runtime %q", name)
	}
}

func WithKeepMgmtNet() ClabOption {
	return func(c *CLab) error {
		c.GlobalRuntime().WithKeepMgmtNet()
		return nil
	}
}

func WithTopoFile(file, varsFile string) ClabOption {
	return func(c *CLab) error {
		if file == "" {
			return fmt.Errorf("provide a path to the clab topology file")
		}
		if err := c.GetTopology(file, varsFile); err != nil {
			return fmt.Errorf("failed to read topology file: %v", err)
		}

		return c.initMgmtNetwork()
	}
}

// NewContainerLab function defines a new container lab.
func NewContainerLab(opts ...ClabOption) (*CLab, error) {
	c := &CLab{
		Config: &Config{
			Mgmt:     new(types.MgmtNet),
			Topology: types.NewTopology(),
		},
		TopoFile: new(TopoFile),
		m:        new(sync.RWMutex),
		Nodes:    make(map[string]nodes.Node),
		Links:    make(map[int]*types.Link),
		Runtimes: make(map[string]runtime.ContainerRuntime),
	}

	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	var err error
	if c.TopoFile.path != "" {
		err = c.parseTopology()
	}

	return c, err
}

// initMgmtNetwork sets management network config.
func (c *CLab) initMgmtNetwork() error {
	log.Debugf("method initMgmtNetwork was called mgmt params %+v", c.Config.Mgmt)
	if c.Config.Mgmt.Network == "" {
		c.Config.Mgmt.Network = dockerNetName
	}

	if c.Config.Mgmt.IPv4Subnet == "" && c.Config.Mgmt.IPv6Subnet == "" {
		c.Config.Mgmt.IPv4Subnet = dockerNetIPv4Addr
		c.Config.Mgmt.IPv6Subnet = dockerNetIPv6Addr
	}

	// by default external access is enabled if not set by a user
	if c.Config.Mgmt.ExternalAccess == nil {
		c.Config.Mgmt.ExternalAccess = new(bool)
		*c.Config.Mgmt.ExternalAccess = true
	}

	log.Debugf("New mgmt params are %+v", c.Config.Mgmt)

	return nil
}

func (c *CLab) GlobalRuntime() runtime.ContainerRuntime {
	return c.Runtimes[c.globalRuntime]
}

// CreateNodes schedules nodes creation and returns a waitgroup for all nodes.
// Nodes interdependencies are created in this function.
func (c *CLab) CreateNodes(ctx context.Context, maxWorkers uint,
	serialNodes map[string]struct{}, dm DependencyManager,
) (*sync.WaitGroup, error) {

	// nodes with static mgmt IP should be scheduled before the dynamic ones
	err := createStaticDynamicDependency(c.Nodes, dm)
	if err != nil {
		return nil, err
	}

	// create user-defined node dependencies done with `wait-for` node property
	err = createWaitForDependency(c.Nodes, dm)
	if err != nil {
		return nil, err
	}

	// make network namespace shared containers start in the right order
	createNamespaceSharingDependency(c.Nodes, dm)

	// Add possible additional dependencies here

	// make sure that there are no unresolvable dependencies, which would deadlock.
	err = dm.CheckAcyclicity()
	if err != nil {
		return nil, err
	}

	// start scheduling
	NodesWg := c.scheduleNodes(ctx, int(maxWorkers), c.Nodes, dm)

	return NodesWg, nil
}

// createNamespaceSharingDependency adds dependency between the containerlab nodes that share a common network namespace.
// When a node_a in the topology configured to be started in the netns of a node_b as such:
//
// node_a:
//   network-mode: container:node_b
//
// then node_a depends on node_b, and waits for node_b to be scheduled first.
func createNamespaceSharingDependency(nodeMap map[string]nodes.Node, dm DependencyManager) {
	for nodeName, n := range nodeMap {
		nodeConfig := n.Config()
		netModeArr := strings.SplitN(nodeConfig.NetworkMode, ":", 2)
		if netModeArr[0] != "container" {
			// we only care about nodes with shared netns network-mode ("container:<CONTAINERNAME>")
			continue
		}
		// the referenced container might be an external pre-existing or a container defined in the given clab topology.
		referencedNode := netModeArr[1]

		// if the container does not exist in the list of containers, it must be an external dependency
		// it can be ignored for internal processing so -> continue
		if _, exists := nodeMap[netModeArr[1]]; !exists {
			continue
		}

		// since the referenced container is clab-managed node, we create a dependency between the nodes
		dm.AddDependency(nodeName, types.NewWaitFor(referencedNode, types.WaitForCreated))
	}
}

// createStaticDynamicDependency creates the dependencies between the nodes such that all nodes with dynamic mgmt IP
// are dependent on the nodes with static mgmt IP. This results in nodes with static mgmt IP to be scheduled before dynamic ones.
func createStaticDynamicDependency(n map[string]nodes.Node, dm DependencyManager) error {
	staticIPNodes := make(map[string]nodes.Node)
	dynIPNodes := make(map[string]nodes.Node)

	// divide the nodes into static and dynamic mgmt IP nodes.
	for name, n := range n {
		if n.Config().MgmtIPv4Address != "" || n.Config().MgmtIPv6Address != "" {
			staticIPNodes[name] = n
			continue
		}
		dynIPNodes[name] = n
	}

	// go through all the dynamic ip nodes
	for dynNodeName := range dynIPNodes {
		// and add their wait group to the the static nodes, while increasing the waitgroup
		for staticNodeName := range staticIPNodes {
			err := dm.AddDependency(dynNodeName, types.NewWaitFor(staticNodeName, types.WaitForCreated))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// createWaitForDependency reflects the dependencies defined in the configuration via the wait-for field.
func createWaitForDependency(n map[string]nodes.Node, dm DependencyManager) error {
	for waiterNode, node := range n {
		// add node's waitFor nodes to the dependency manager
		for _, waitForNode := range node.Config().WaitFor {
			err := dm.AddDependency(waiterNode, waitForNode)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *CLab) scheduleNodes(ctx context.Context, maxWorkers int,
	scheduledNodes map[string]nodes.Node, dm DependencyManager,
) *sync.WaitGroup {
	concurrentChan := make(chan nodes.Node)

	workerFunc := func(i int, input chan nodes.Node, wg *sync.WaitGroup, dm DependencyManager) {
		defer wg.Done()
		for {
			select {
			case node, ok := <-input:
				if node == nil || !ok {
					log.Debugf("Worker %d terminating...", i)
					return
				}
				log.Debugf("Worker %d received node: %+v", i, node.Config())

				// Apply any startup delay
				delay := node.Config().StartupDelay
				if delay > 0 {
					log.Infof("node %q is being delayed for %d seconds", node.Config().ShortName, delay)
					time.Sleep(time.Duration(delay) * time.Second)
				}

				// PreDeploy
				err := node.PreDeploy(c.Config.Name, c.Dir.LabCA, c.Dir.LabCARoot)
				if err != nil {
					log.Errorf("failed pre-deploy phase for node %q: %v", node.Config().ShortName, err)
					continue
				}
				// Deploy
				err = node.Deploy(ctx)
				if err != nil {
					log.Errorf("failed deploy phase for node %q: %v", node.Config().ShortName, err)
					continue
				}

				// set deployment status of a node to created to indicate that it finished creating
				// this status is checked during link creation to only schedule link creation if both nodes are ready
				c.m.Lock()
				node.Config().DeploymentStatus = "created"
				c.m.Unlock()

				// signal to dependency manager that this node is done
				dm.SignalDone(node.Config().ShortName, types.WaitForCreated)
			case <-ctx.Done():
				return
			}
		}
	}

	numScheduledNodes := len(scheduledNodes)
	if numScheduledNodes < maxWorkers {
		maxWorkers = numScheduledNodes
	}
	wg := new(sync.WaitGroup)

	// start concurrent workers
	wg.Add(int(maxWorkers))
	// it's safe to not check if all nodes are serial because in that case
	// maxWorkers will be 0
	for i := 0; i < maxWorkers; i++ {
		go workerFunc(i, concurrentChan, wg, dm)
	}

	// Waitgroup used to protect the channel towards the workers of being closed to early
	workerFuncChWG := new(sync.WaitGroup)

	for _, n := range scheduledNodes {
		workerFuncChWG.Add(1)
		// start a func for all the containers, then will wait for their own waitgroups
		// to be set to zero by their depending containers, then enqueue to the creation channel
		go func(node nodes.Node, dm DependencyManager, workerChan chan<- nodes.Node, wfcwg *sync.WaitGroup) {
			// wait for all the nodes that node depends on
			err := dm.WaitForNodeDependencies(node.Config().ShortName, types.WaitForCreated)
			if err != nil {
				log.Error(err)
			}
			// wait for possible external dependencies
			c.WaitForExternalNodeDependencies(ctx, node.Config().ShortName)
			// when all nodes that this node depends on are created, push it into the channel
			workerChan <- node
			// indicate we are done, such that only when all of these functions are done, the workerChan is being closed
			wfcwg.Done()
		}(n, dm, concurrentChan, workerFuncChWG) // execute this function straight away
	}

	// Gate to make sure the channel is not closed before all the nodes made it though the channel
	workerFuncChWG.Wait()
	// close the channel and thereby terminate the workerFuncs
	close(concurrentChan)

	return wg
}

// WaitForExternalNodeDependencies makes nodes that have a reference to an external container network-namespace (network-mode: container:<NAME>)
// to wait until the referenced container is in started status.
// The wait time is 15 minutes by default.
func (c *CLab) WaitForExternalNodeDependencies(ctx context.Context, nodeName string) {
	if _, exists := c.Nodes[nodeName]; !exists {
		log.Errorf("unable to find referenced node %q", nodeName)
		return
	}
	nodeConfig := c.Nodes[nodeName].Config()
	netModeArr := strings.SplitN(nodeConfig.NetworkMode, ":", 2)
	if netModeArr[0] != "container" {
		// we only care about nodes with NetMode "container:<CONTAINERNAME>"
		return
	}
	// the referenced container might be an external pre-existing or a container created also by the given clab topology.
	contName := netModeArr[1]

	// if the container does not exist in the list of container, it must be an external dependency
	// it can be ignored for internal processing so -> continue
	if _, exists := c.Nodes[contName]; exists {
		return
	}

	// how long to wait for the external container to become running
	statusCheckTimeout := 15 * time.Minute
	// frequency to check for new container state
	statusCheckFrequency := time.Second

	// setup a ticker
	ticker := time.NewTicker(statusCheckFrequency)
	// timeout sets the specified timeout
	timeout := time.After(statusCheckTimeout)
	// startTime is used to calculate elapsed waiting time
	startTime := time.Now()

TIMEOUT_LOOP:
	for {
		select {
		case <-ticker.C:
			runtimeStatus := c.Runtimes[c.globalRuntime].GetContainerStatus(ctx, contName)

			// if the dependency container is running we are allowed to schedule the node
			if runtimeStatus == runtime.Running {
				log.Infof("node %q starts since external container %q is running now", nodeName, contName)
				break TIMEOUT_LOOP
			}

			// if not, log and retry
			log.Infof("node %q depends on external container %q, which is not running yet. Waited %s. Retrying...",
				nodeName, contName, time.Since(startTime).Truncate(time.Second))

		case <-timeout:
			log.Errorf("node %q waited %s for external dependency container %q to come up, which did not happen. Giving up now",
				nodeName, time.Since(startTime), contName)
			break TIMEOUT_LOOP
		}
	}
}

// CreateLinks creates links using the specified number of workers.
func (c *CLab) CreateLinks(ctx context.Context, workers uint) {
	wg := new(sync.WaitGroup)
	wg.Add(int(workers))
	linksChan := make(chan *types.Link)

	log.Debug("creating links...")
	// wire the links between the nodes based on cabling plan
	for i := uint(0); i < workers; i++ {
		go func(i uint) {
			defer wg.Done()
			for {
				select {
				case link := <-linksChan:
					if link == nil {
						log.Debugf("Link worker %d terminating...", i)
						return
					}
					log.Debugf("Link worker %d received link: %+v", i, link)
					if err := c.CreateVirtualWiring(link); err != nil {
						log.Error(err)
					}
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	// create a copy of links map to loop over
	// so that we can wait till all the nodes are ready before scheduling a link
	linksCopy := map[int]*types.Link{}
	for k, v := range c.Links {
		linksCopy[k] = v
	}
	for {
		if len(linksCopy) == 0 {
			break
		}
		for k, link := range linksCopy {
			c.m.Lock()
			if link.A.Node.DeploymentStatus == "created" &&
				link.B.Node.DeploymentStatus == "created" {
				linksChan <- link
				delete(linksCopy, k)
			}
			c.m.Unlock()
		}
	}

	// close channel to terminate the workers
	close(linksChan)
	// wait for all workers to finish
	wg.Wait()
}

func (c *CLab) DeleteNodes(ctx context.Context, workers uint, serialNodes map[string]struct{}) {
	wg := new(sync.WaitGroup)

	concurrentChan := make(chan nodes.Node)
	serialChan := make(chan nodes.Node)

	workerFunc := func(i uint, input chan nodes.Node, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case n := <-input:
				if n == nil {
					log.Debugf("Worker %d terminating...", i)
					return
				}
				err := n.Delete(ctx)
				if err != nil {
					log.Errorf("could not remove container %q: %v", n.Config().LongName, err)
				}
			case <-ctx.Done():
				return
			}
		}
	}

	// start concurrent workers
	wg.Add(int(workers))
	for i := uint(0); i < workers; i++ {
		go workerFunc(i, concurrentChan, wg)
	}

	// start the serial worker
	if len(serialNodes) > 0 {
		wg.Add(1)
		go workerFunc(workers, serialChan, wg)
	}

	// send nodes to workers
	for _, n := range c.Nodes {
		if _, ok := serialNodes[n.Config().LongName]; ok {
			serialChan <- n
			continue
		}
		concurrentChan <- n
	}

	// close channel to terminate the workers
	close(concurrentChan)
	close(serialChan)

	wg.Wait()
}

func (c *CLab) ListContainers(ctx context.Context, labels []*types.GenericFilter) ([]types.GenericContainer, error) {
	var containers []types.GenericContainer

	for _, r := range c.Runtimes {
		ctrs, err := r.ListContainers(ctx, labels)
		if err != nil {
			return containers, fmt.Errorf("could not list containers: %v", err)
		}
		containers = append(containers, ctrs...)
	}
	return containers, nil
}

func (c *CLab) GetNodeRuntime(contName string) (runtime.ContainerRuntime, error) {
	shortName, err := getShortName(c.Config.Name, c.Config.Prefix, contName)
	if err != nil {
		return nil, err
	}

	if node, ok := c.Nodes[shortName]; ok {
		return node.GetRuntime(), nil
	}

	return nil, fmt.Errorf("could not find a container matching name %q", contName)
}

// VethCleanup iterates over links found in clab topology to initiate removal of dangling veths in host networking namespace
// See https://github.com/srl-labs/containerlab/issues/842 for the reference.
func (c *CLab) VethCleanup(_ context.Context) error {
	for _, link := range c.Links {
		err := c.RemoveHostVeth(link)
		if err != nil {
			log.Infof("Error during veth cleanup: %v", err)
		}
	}
	return nil
}
