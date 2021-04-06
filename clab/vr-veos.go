package clab

import (
	"context"
	"fmt"
)

type VrVeosNode struct {
	Node
}

func init() {
	RegisterNodeType("VrVeos", &SRLNode{})
}

func (node *VrVeosNode) InitNode(c *CLab, nodeCfg NodeConfig, user string, envs map[string]string) error {
	var err error

	node.Image = c.imageInitialization(&nodeCfg, node.Kind)
	node.Group = c.groupInitialization(&nodeCfg, node.Kind)
	node.Position = c.positionInitialization(&nodeCfg, node.Kind)
	node.User = user

	// env vars are used to set launch.py arguments in vrnetlab container
	defEnv := map[string]string{
		"CONNECTION_MODE":    vrDefConnMode,
		"USERNAME":           "admin",
		"PASSWORD":           "admin",
		"DOCKER_NET_V4_ADDR": c.Config.Mgmt.IPv4Subnet,
		"DOCKER_NET_V6_ADDR": c.Config.Mgmt.IPv6Subnet,
	}
	node.Env = mergeStringMaps(defEnv, envs)

	if node.Env["CONNECTION_MODE"] == "macvtap" {
		// mount dev dir to enable macvtap
		node.Binds = append(node.Binds, "/dev:/dev")
	}

	node.Cmd = fmt.Sprintf("--username %s --password %s --hostname %s --connection-mode %s --trace", node.Env["USERNAME"], node.Env["PASSWORD"], node.ShortName, node.Env["CONNECTION_MODE"])

	return err
}

func (_ *VrVeosNode) PostDeploy(ctx context.Context, c *CLab, node *Node, lworkers uint) error {
	return nil
}
