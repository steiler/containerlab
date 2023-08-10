package links

import (
	"context"
	"fmt"

	"github.com/vishvananda/netlink"
)

// LinkMacVlanRaw is the raw (string) representation of a macvlan link as defined in the topology file.
type LinkMacVlanRaw struct {
	LinkCommonParams `yaml:",inline"`
	HostInterface    string       `yaml:"host-interface"`
	Endpoint         *EndpointRaw `yaml:"endpoint"`
	Mode             string       `yaml:"mode"`
}

// ToLinkConfig converts the raw link into a LinkConfig.
func (r *LinkMacVlanRaw) ToLinkConfig() *LinkBrief {
	lc := &LinkBrief{
		Endpoints: make([]string, 2),
		LinkCommonParams: LinkCommonParams{
			MTU:    r.MTU,
			Labels: r.Labels,
			Vars:   r.Vars,
		},
	}

	lc.Endpoints[0] = fmt.Sprintf("%s:%s", r.Endpoint.Node, r.Endpoint.Iface)
	lc.Endpoints[1] = fmt.Sprintf("%s:%s", "macvlan", r.HostInterface)

	return lc
}

func (r *LinkMacVlanRaw) GetType() LinkType {
	return LinkTypeMacVLan
}

func macVlanLinkFromBrief(lb *LinkBrief, specialEPIndex int) (*LinkMacVlanRaw, error) {
	_, hostIf, node, nodeIf := extractHostNodeInterfaceData(lb, specialEPIndex)

	result := &LinkMacVlanRaw{
		LinkCommonParams: LinkCommonParams{
			MTU:    lb.MTU,
			Labels: lb.Labels,
			Vars:   lb.Vars,
		},
		HostInterface: hostIf,
		Endpoint:      NewEndpointRaw(node, nodeIf, ""),
	}

	return result, nil
}

func (r *LinkMacVlanRaw) Resolve(params *ResolveParams) (Link, error) {

	hostLinkNode := GetFakeHostLinkNode()
	ep := &EndpointMacVlan{
		EndpointGeneric: EndpointGeneric{
			IfaceName: r.HostInterface,
			Node:      hostLinkNode,
			// Link is being set after the link in created further down
		},
	}

	link := &LinkMacVlan{
		LinkCommonParams: r.LinkCommonParams,
		HostEndpoint:     ep,
	}
	ep.Link = link
	// parse the MacVlanMode
	mode, err := MacVlanModeParse(r.Mode)
	if err != nil {
		return nil, err
	}
	// set the mode in the link struct
	link.Mode = mode
	// resolve the endpoint
	link.NodeEndpoint, err = r.Endpoint.Resolve(params, link)
	if err != nil {
		return nil, err
	}
	return link, nil
}

type LinkMacVlan struct {
	LinkCommonParams
	HostEndpoint Endpoint
	NodeEndpoint Endpoint
	Mode         MacVlanMode
}

type MacVlanMode string

const (
	MacVlanModeBridge   = "bridge"
	MacVlanModeVepa     = "vepa"
	MacVlanModePassthru = "passthru"
	MacVlanModePrivate  = "private"
	MacVlanModeSource   = "source"
)

func MacVlanModeParse(s string) (MacVlanMode, error) {
	switch s {
	case MacVlanModeBridge:
		return MacVlanModeBridge, nil
	case MacVlanModeVepa:
		return MacVlanModeVepa, nil
	case MacVlanModePassthru:
		return MacVlanModePassthru, nil
	case MacVlanModePrivate:
		return MacVlanModePrivate, nil
	case MacVlanModeSource:
		return MacVlanModeSource, nil
	case "":
		return MacVlanModeBridge, nil
	}
	return "", fmt.Errorf("unknown MacVlanMode %q", s)
}

func (l *LinkMacVlan) GetType() LinkType {
	return LinkTypeMacVLan
}

func (l *LinkMacVlan) GetParentInterfaceMtu() (int, error) {
	hostLink, err := netlink.LinkByName(l.HostEndpoint.GetIfaceName())
	if err != nil {
		return 0, err
	}
	return hostLink.Attrs().MTU, nil
}

func (l *LinkMacVlan) Deploy(ctx context.Context) error {
	// lookup the parent host interface
	parentInterface, err := netlink.LinkByName(l.HostEndpoint.GetIfaceName())
	if err != nil {
		return err
	}

	// set MacVlanMode
	mode := netlink.MACVLAN_MODE_BRIDGE
	switch l.Mode {
	case MacVlanModeBridge:
		break
	case MacVlanModePassthru:
		mode = netlink.MACVLAN_MODE_PASSTHRU
	case MacVlanModeVepa:
		mode = netlink.MACVLAN_MODE_VEPA
	case MacVlanModePrivate:
		mode = netlink.MACVLAN_MODE_PRIVATE
	case MacVlanModeSource:
		mode = netlink.MACVLAN_MODE_SOURCE
	}

	// build Netlink Macvlan struct
	link := &netlink.Macvlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:        l.NodeEndpoint.GetRandIfaceName(),
			ParentIndex: parentInterface.Attrs().Index,
			MTU:         l.MTU,
		},
		Mode: mode,
	}
	// add the link in the Host NetNS
	err = netlink.LinkAdd(link)
	if err != nil {
		return err
	}

	// retrieve the Link by name
	mvInterface, err := netlink.LinkByName(l.NodeEndpoint.GetRandIfaceName())
	if err != nil {
		return fmt.Errorf("failed to lookup %q: %v", l.NodeEndpoint.GetRandIfaceName(), err)
	}

	// add the link to the Node Namespace
	err = l.NodeEndpoint.GetNode().AddNetlinkLinkToContainer(ctx, mvInterface, SetNameMACAndUpInterface(mvInterface, l.NodeEndpoint))
	if err != nil {
		return err
	}
	// update the HostEndpoint with the assigned mac address
	err = l.HostEndpoint.VerifyAndPopulateMacAddress()
	if err != nil {
		return err
	}

	// update the NodeEndpoint with the assigned mac address
	err = l.NodeEndpoint.VerifyAndPopulateMacAddress()
	if err != nil {
		return err
	}

	return err
}

func (l *LinkMacVlan) Remove(_ context.Context) error {
	// TODO
	return nil
}

func (l *LinkMacVlan) GetEndpoints() []Endpoint {
	return []Endpoint{
		l.NodeEndpoint,
		l.HostEndpoint,
	}
}
