package clab

type SonicNode struct {
	Node
}

func (node *SonicNode) InitNode(c *CLab, nodeCfg NodeConfig, user string, envs map[string]string) error {
	var err error

	node.Config, err = c.configInit(&nodeCfg, node.Kind)
	if err != nil {
		return err
	}
	node.Image = c.imageInitialization(&nodeCfg, node.Kind)
	node.Group = c.groupInitialization(&nodeCfg, node.Kind)
	node.Position = c.positionInitialization(&nodeCfg, node.Kind)
	node.User = user

	// rewrite entrypoint so sonic won't start supervisord before we attach veth interfaces
	node.Entrypoint = "/bin/bash"

	return err
}
