package mysocketio

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/srl-labs/containerlab/nodes"
	"github.com/srl-labs/containerlab/types"
)

var Kindnames = []string{"mysocketio"}

// Register registers the node in the NodeRegistry.
func Register(r *nodes.NodeRegistry) {
	r.Register(Kindnames, func() nodes.Node {
		return new(mySocketIO)
	}, nil)
}

type mySocketIO struct {
	nodes.DefaultNode
}

func (s *mySocketIO) Init(cfg *types.NodeConfig, opts ...nodes.NodeOption) error {
	// Init DefaultNode
	s.DefaultNode = *nodes.NewDefaultNode(s)

	s.Cfg = cfg
	for _, o := range opts {
		o(s)
	}

	return nil
}

func (s *mySocketIO) PostDeploy(ctx context.Context, params *nodes.PostDeployParams) error {
	log.Debugf("Running postdeploy actions for mysocketio '%s' node", s.Cfg.ShortName)
	err := types.DisableTxOffload(s.Cfg)
	if err != nil {
		return fmt.Errorf("failed to disable tx checksum offload for mysocketio kind: %v", err)
	}

	log.Infof("Creating mysocketio tunnels...")
	err = createMysocketTunnels(ctx, s, params.Nodes)
	return err
}
