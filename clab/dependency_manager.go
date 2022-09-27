package clab

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/srl-labs/containerlab/types"
)

type DependencyManager interface {
	// AddNode adds a node to the dependency manager.
	AddNode(name string)
	// AddDependency adds a dependency between depender and dependee.
	// The depender will effectively wait for the dependee to finish.
	AddDependency(depender string, wf *types.WaitFor) error
	// WaitForNodeDependencies is called by a node that is meant to be created.
	// This call will bock until all the nodes that this node depends on are created.
	WaitForNodeDependencies(nodeName string, phase types.WaitForPhase) error
	// SignalDone is called by a node that has finished the creation process.
	// internally the dependent nodes will be "notified" that an additional (if multiple exist) dependency is satisfied.
	SignalDone(nodeName string, phase types.WaitForPhase)
	// CheckAcyclicity checks if dependencies contain cycles.
	CheckAcyclicity() error
	// String returns a string representation of dependencies recorded with dependency manager.
	String() string
	// IsHealthCheckRequired returns true if dependencies exist for the given node to turn healthy.
	IsHealthCheckRequired(nodeName string) (bool, error)
}

type defaultDependencyManager struct {
	// map of wait group per node.
	// The scheduling of the nodes creation is dependent on their respective wait group.
	// Other nodes, that the specific node relies on will increment this wait group.
	nodeWaitGroup map[string]*sync.WaitGroup
	// Names of the nodes that depend on a given node are listed here.
	// On successful creation of the said node, all the depending nodes (dependers) wait groups will be decremented.
	nodeDependers map[string]map[types.WaitForPhase][]string
}

func NewDependencyManager() DependencyManager {
	return &defaultDependencyManager{
		nodeWaitGroup: map[string]*sync.WaitGroup{},
		nodeDependers: map[string]map[types.WaitForPhase][]string{},
	}
}

// AddNode adds a node to the dependency manager.
func (dm *defaultDependencyManager) AddNode(name string) {
	dm.nodeWaitGroup[name] = &sync.WaitGroup{}
	dm.nodeDependers[name] = map[types.WaitForPhase][]string{}

	// init the structs for all WaitForPhases
	for _, phaseName := range types.WaitForPhases {
		dm.nodeDependers[name][phaseName] = []string{}
	}
}

// AddDependency adds a dependency between depender and dependee.
// The depender will effectively wait for the dependee to finish.
func (dm *defaultDependencyManager) AddDependency(depender string, wf *types.WaitFor) error {
	// first check if the referenced nodes are known to the dm
	if _, exists := dm.nodeWaitGroup[depender]; !exists {
		return fmt.Errorf("node %q is not known to the dependency manager", depender)
	}
	if _, exists := dm.nodeDependers[wf.Node]; !exists {
		return fmt.Errorf("node %q is not known to the dependency manager", wf.Node)
	}

	// increase the WaitGroup by one for the depender
	dm.nodeWaitGroup[depender].Add(1)

	// add a depender node name for a given dependee
	dm.nodeDependers[wf.Node][wf.Phase] = append(dm.nodeDependers[wf.Node][wf.Phase], depender)
	return nil
}

// WaitForNodeDependencies is called by a node that is meant to be created.
// This call will bock until all the nodes that this node depends on are created.
func (dm *defaultDependencyManager) WaitForNodeDependencies(nodeName string, phase types.WaitForPhase) error {
	// first check if the referenced node is known to the dm
	if _, exists := dm.nodeWaitGroup[nodeName]; !exists {
		return fmt.Errorf("node %q is not known to the dependency manager", nodeName)
	}
	dm.nodeWaitGroup[nodeName].Wait()
	return nil
}

// SignalDone is called by a node that has finished the creation process.
// internally the dependent nodes will be "notified" that an additional (if multiple exist) dependency is satisfied.
func (dm *defaultDependencyManager) SignalDone(nodeName string, phase types.WaitForPhase) {
	// first check if the referenced node is known to the dm
	if _, exists := dm.nodeDependers[nodeName]; !exists {
		log.Errorf("tried to Signal Done for node %q but node is unknown to the DependencyManager", nodeName)
		return
	}
	for _, depender := range dm.nodeDependers[nodeName][phase] {
		dm.nodeWaitGroup[depender].Done()
	}
}

// getDependersSliceWithoutPhasesDependency return the dm.nodeDependers but removes the phases information
// basically from "map[string]map[types.WaitForPhase][]string" we return map[string][]string. The entries of
// the different phases are merged, because for the acyclicity check the phases do not matter.
func (dm *defaultDependencyManager) getDependersSliceWithoutPhasesDependency() map[string][]string {
	dependers := map[string][]string{}

	// we just need a plain list of dependencies, phases basically do not matter
	// iterate through the dependers per node
	for d, _ := range dm.nodeDependers {
		// iterate through their phases
		for _, phase := range types.WaitForPhases {
			// if the dependers entity does not have a string slice under index "d" create it
			if _, exists := dependers[d]; !exists {
				dependers[d] = []string{}
			}
			// basically merge all the phases dependencies into the dependers datastructure
			dependers[d] = dm.nodeDependers[d][phase]
		}
	}
	return dependers
}

// CheckAcyclicity checks if dependencies contain cycles.
func (dm *defaultDependencyManager) CheckAcyclicity() error {
	log.Debugf("Dependencies:\n%s", dm.String())

	if !isAcyclic(dm.getDependersSliceWithoutPhasesDependency(), 1) {
		return fmt.Errorf("cyclic dependencies found!\n%s", dm.String())
	}

	return nil
}

// String returns a string representation of dependencies recorded with dependency manager.
func (dm *defaultDependencyManager) String() string {
	// since dm.nodeDependers contains a map of dependee->[dependers] it is not
	// particularly suitable for displaying the dependency graph
	// this function reverses the order so that it becomes depender->[dependees]

	// map to record the dependencies in string based representation
	dependencies := map[string][]string{}

	// prepare dependencies table
	for name := range dm.nodeWaitGroup {
		dependencies[name] = []string{}
	}

	// build the dependency datastruct
	for dependee, dependers := range dm.nodeDependers {
		// iterate the phases
		for _, phase := range types.WaitForPhases {
			// iterate all dependers of the phases
			for _, depender := range dependers[phase] {
				// add them to the dependencies
				dependencies[depender] = append(dependencies[depender], dependee)
			}
		}
	}

	result := []string{}
	// print dependencies
	for nodename, deps := range dependencies {
		result = append(result, fmt.Sprintf("%s -> [ %s ]", nodename, strings.Join(deps, ", ")))
	}
	return strings.Join(result, "\n")
}

// isAcyclic checks the provided dependencies map for cycles.
// i indicates the check round. Must be set to 1.
func isAcyclic(nodeDependers map[string][]string, i int) bool {
	// no more nodes then the graph is acyclic
	if len(nodeDependers) == 0 {
		log.Debugf("node creation graph is successfully validated as being acyclic")

		return true
	}

	// debug output
	d := []string{}
	for dependee, dependers := range nodeDependers {
		d = append(d, fmt.Sprintf("%s <- [ %s ]", dependee, strings.Join(dependers, ", ")))
	}
	log.Debugf("- cycle check round %d - \n%s", i, strings.Join(d, "\n"))

	remainingNodeDependers := map[string][]string{}
	leafNodes := []string{}
	// mark a node as a remaining dependency if other nodes still depend on it,
	// otherwise add it to the leaf list for it to be removed in the next round of recursive check
	for dependee, dependers := range nodeDependers {
		if len(dependers) > 0 {
			remainingNodeDependers[dependee] = dependers
		} else {
			leafNodes = append(leafNodes, dependee)
		}
	}

	// if nodes remain but none of them is a leaf node, must by cyclic
	if len(leafNodes) == 0 {
		return false
	}

	// iterate over remaining nodes, to remove all leaf nodes from the dependencies, because in the next round of recursion,
	// these will no longer be there, they suffice the satisfy the acyclicity property
	for dependee, dependers := range remainingNodeDependers {
		// new array that keeps track of remaining dependencies
		newRemainingNodeDependers := []string{}
		// iterate over deleted nodes
		for _, dep := range dependers {
			keep := true
			// check if the actual dep is a leafNode and should therefore be removed
			for _, delnode := range leafNodes {
				// if it is a node that is meant to be deleted, stop here and make sure its not taken over to the new array
				if delnode == dep {
					keep = false
					break
				}
			}
			if keep {
				newRemainingNodeDependers = append(newRemainingNodeDependers, dep)
			}
		}
		// replace previous with the new, cleanup dependencies.
		remainingNodeDependers[dependee] = newRemainingNodeDependers
	}
	return isAcyclic(remainingNodeDependers, i+1)
}

// IsHealthCheckRequired returns true if dependencies exist for the given node to turn healthy.
func (dm *defaultDependencyManager) IsHealthCheckRequired(nodeName string) (bool, error) {
	if _, exists := dm.nodeDependers[nodeName]; !exists {
		return true, fmt.Errorf("node %q not found in DependencyManager", nodeName)
	}
	return len(dm.nodeDependers[nodeName][types.WaitForHealthy]) > 0, nil
}
