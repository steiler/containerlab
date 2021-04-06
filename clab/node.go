package clab

import (
	"context"

	"github.com/docker/go-connections/nat"
)

type INode interface {
	InitNode(c *CLab, nodeCfg NodeConfig, user string, envs map[string]string) error
	PostDeploy(ctx context.Context, c *CLab, node *Node, lworkers uint) error
}

// Node is a struct that contains the information of a container element
type Node struct {
	ShortName string
	LongName  string
	Fqdn      string
	LabDir    string // LabDir is a directory related to the node, it contains config items and/or other persistent state
	Index     int
	Group     string
	Kind      string
	// path to config template file that is used for config generation
	Config       string
	ResConfig    string // path to config file that is actually mounted to the container and is a result of templation
	NodeType     string
	Position     string
	License      string
	Image        string
	Topology     string
	Sysctls      map[string]string
	User         string
	Entrypoint   string
	Cmd          string
	Env          map[string]string
	Binds        []string    // Bind mounts strings (src:dest:options)
	PortBindings nat.PortMap // PortBindings define the bindings between the container ports and host ports
	PortSet      nat.PortSet // PortSet define the ports that should be exposed on a container
	// container networking mode. if set to `host` the host networking will be used for this node, else bridged network
	NetworkMode          string
	MgmtNet              string // name of the docker network this node is connected to with its first interface
	MgmtIPv4Address      string
	MgmtIPv4PrefixLength int
	MgmtIPv6Address      string
	MgmtIPv6PrefixLength int
	MacAddress           string
	ContainerID          string
	TLSCert              string
	TLSKey               string
	TLSAnchor            string
	NSPath               string   // network namespace path for this node
	Publish              []string //list of ports to publish with mysocketctl
	// container labels
	Labels map[string]string
}
