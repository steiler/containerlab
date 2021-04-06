package clab

import (
	"context"
	"reflect"

	"github.com/docker/go-connections/nat"
)

var NodeTypeRegistry map[string]reflect.Type

func RegisterNodeType(name string, nodetype INode) {
	NodeTypeRegistry[name] = reflect.TypeOf(nodetype)
	println("Registered NodeType: ", name)
}

func getNodeFromName(name string) INode {
	v := reflect.New(NodeTypeRegistry[name]).Elem()
	// Maybe fill in fields here if necessary
	return v.Interface().(INode)
}

type INode interface {
	ShortName() string
	SetShortName(string)
	LongName() string
	SetLongName(string)
	Fqdn() string
	SetFqdn(string)
	LabDir() string
	SetLabDir(string)
	Index() int
	SetIndex(int)
	Group() string
	SetGroup(string)
	Kind() string
	SetKind(string)
	Config() string
	SetConfig(string)
	ResConfig() string
	SetResConfig(string)
	NodeType() string
	SetNodeType(string)
	Position() string
	SetPosition(string)
	License() string
	SetLicense(string)
	Image() string
	SetImage(string)
	Topology() string
	SetTopology(string)
	Sysctls() map[string]string
	SetSysctls(map[string]string)
	User() string
	SetUser(string)
	Entrypoint() string
	SetEntrypoint(string)
	Cmd() string
	SetCmd(string)
	Env() map[string]string
	SetEnv(map[string]string)
	Binds() []string
	SetBinds([]string)
	PortBindings() nat.PortMap
	SetPortBindings(nat.PortMap)
	PortSet() nat.PortSet
	SetPortSet(nat.PortSet)
	NetworkMode() string
	SetNetworkMode(string)
	MgmtNet() string
	SetMgmtNet(string)
	MgmtIPv4Address() string
	SetMgmtIPv4Address(string)
	MgmtIPv4PrefixLength() int
	SetMgmtIPv4PrefixLength(int)
	MgmtIPv6Address() string
	SetMgmtIPv6Address(string)
	MgmtIPv6PrefixLength() int
	SetMgmtIPv6PrefixLength(int)
	MacAddress() string
	SetMacAddress(string)
	ContainerID() string
	SetContainerID(string)
	TLSCert() string
	SetTLSCert(string)
	TLSKey() string
	SetTLSKey(string)
	TLSAnchor() string
	SetTLSAnchor(string)
	NSPath() string
	SetNSPath(string)
	Publish() []string
	SetPublish([]string)
	Labels() map[string]string
	SetLabels(map[string]string)

	InitNode(c *CLab, nodeCfg NodeConfig, user string, envs map[string]string) error
	PostDeploy(ctx context.Context, c *CLab, lworkers uint) error
}

// Node is a struct that contains the information of a container element
type Node struct {
	shortName string
	longName  string
	fqdn      string
	labDir    string // LabDir is a directory related to the node, it contains config items and/or other persistent state
	index     int
	group     string
	kind      string
	// path to config template file that is used for config generation
	config       string
	resConfig    string // path to config file that is actually mounted to the container and is a result of templation
	nodeType     string
	position     string
	license      string
	image        string
	topology     string
	sysctls      map[string]string
	user         string
	entrypoint   string
	cmd          string
	env          map[string]string
	binds        []string    // Bind mounts strings (src:dest:options)
	portBindings nat.PortMap // PortBindings define the bindings between the container ports and host ports
	portSet      nat.PortSet // PortSet define the ports that should be exposed on a container
	// container networking mode. if set to `host` the host networking will be used for this node, else bridged network
	networkMode          string
	mgmtNet              string // name of the docker network this node is connected to with its first interface
	mgmtIPv4Address      string
	mgmtIPv4PrefixLength int
	mgmtIPv6Address      string
	mgmtIPv6PrefixLength int
	macAddress           string
	containerID          string
	tlsCert              string
	tlsKey               string
	tlsAnchor            string
	nsPath               string   // network namespace path for this node
	publish              []string //list of ports to publish with mysocketctl
	// container labels
	labels map[string]string
}

func (n *Node) PostDeploy(ctx context.Context, c *CLab, lworkers uint) error {
	return nil
}

func (n *Node) Binds() []string {
	return n.binds
}

func (n *Node) SetBinds(binds []string) {
	n.binds = binds
}

func (n *Node) ShortName() string {
	return n.shortName
}
func (n *Node) SetShortName(shortname string) {
	n.shortName = shortname
}
func (n *Node) LongName() string {
	return n.longName
}
func (n *Node) SetLongName(longname string) {
	n.longName = longname
}
func (n *Node) Fqdn() string {
	return n.fqdn
}
func (n *Node) SetFqdn(fqdn string) {
	n.fqdn = fqdn
}
func (n *Node) LabDir() string {
	return n.labDir
}
func (n *Node) SetLabDir(labdir string) {
	n.labDir = labdir
}
func (n *Node) Index() int {
	return n.index
}
func (n *Node) SetIndex(index int) {
	n.index = index
}
func (n *Node) Group() string {
	return n.group
}
func (n *Node) SetGroup(group string) {
	n.group = group
}
func (n *Node) Kind() string {
	return n.kind
}
func (n *Node) SetKind(kind string) {
	n.kind = kind
}
func (n *Node) Config() string {
	return n.config
}
func (n *Node) SetConfig(config string) {
	n.config = config
}
func (n *Node) ResConfig() string {
	return n.resConfig
}
func (n *Node) SetResConfig(resConfig string) {
	n.resConfig = resConfig
}
