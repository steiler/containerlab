package links

import (
	"context"
	"fmt"
	"sync"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/srl-labs/containerlab/nodes/state"
	"github.com/vishvananda/netlink"
)

// LinkVEthRaw is the raw (string) representation of a veth link as defined in the topology file.
type LinkVEthRaw struct {
	LinkCommonParams `yaml:",inline"`
	Endpoints        []*EndpointRaw `yaml:"endpoints"`
}

func (r *LinkVEthRaw) MarshalYAML() (interface{}, error) {
	x := struct {
		Type        string `yaml:"type"`
		LinkVEthRaw `yaml:",inline"`
	}{
		Type:        string(LinkTypeVEth),
		LinkVEthRaw: *r,
	}
	return x, nil
}

// ToLinkConfig converts the raw link into a LinkConfig.
func (r *LinkVEthRaw) ToLinkConfig() *LinkBrief {
	lc := &LinkBrief{
		Endpoints: []string{},
		LinkCommonParams: LinkCommonParams{
			MTU:    r.MTU,
			Labels: r.Labels,
			Vars:   r.Vars,
		},
	}

	for _, e := range r.Endpoints {
		lc.Endpoints = append(lc.Endpoints, fmt.Sprintf("%s:%s", e.Node, e.Iface))
	}
	return lc
}

func (r *LinkVEthRaw) GetType() LinkType {
	return LinkTypeVEth
}

// Resolve resolves the raw veth link definition into a Link interface that is implemented
// by a concrete LinkVEth struct.
// Resolving a veth link resolves its endpoints.
func (r *LinkVEthRaw) Resolve(params *ResolveParams) (Link, error) {
	// create LinkVEth struct
	l := &LinkVEth{
		LinkCommonParams: r.LinkCommonParams,
		Endpoints:        make([]Endpoint, 0, 2),
	}

	// resolve raw endpoints (epr) to endpoints (ep)
	for _, epr := range r.Endpoints {
		ep, err := epr.Resolve(params, l)
		if err != nil {
			return nil, err
		}

		l.Endpoints = append(l.Endpoints, ep)
	}

	return l, nil
}

func vEthFromLinkConfig(lb *LinkBrief) (*LinkVEthRaw, error) {
	host, hostIf, node, nodeIf := extractHostNodeInterfaceData(lb, 0)

	result := &LinkVEthRaw{
		LinkCommonParams: LinkCommonParams{
			MTU:    lb.MTU,
			Labels: lb.Labels,
			Vars:   lb.Vars,
		},
		Endpoints: []*EndpointRaw{
			NewEndpointRaw(host, hostIf, ""),
			NewEndpointRaw(node, nodeIf, ""),
		},
	}
	return result, nil
}

type LinkVEth struct {
	// m mutex is used when deploying the link.
	m sync.Mutex `yaml:"-"`
	LinkCommonParams
	Endpoints []Endpoint

	deploymentState LinkDeploymentState
}

func (*LinkVEth) GetType() LinkType {
	return LinkTypeVEth
}

func (l *LinkVEth) Verify() {

}

func (l *LinkVEth) Deploy(ctx context.Context) error {
	l.m.Lock()
	defer l.m.Unlock()

	// since each node calls deploy on its links, we need to make sure that we only deploy
	// the link once, even if multiple nodes call deploy on the same link.
	if l.deploymentState == LinkDeploymentStateDeployed {
		return nil
	}

	for _, ep := range l.GetEndpoints() {
		if ep.GetNode().GetState() != state.Deployed {
			return nil
		}
	}

	// build the netlink.Veth struct for the link provisioning
	linkA := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: l.Endpoints[0].GetRandIfaceName(),
			MTU:  l.MTU,
			// Mac address is set later on
		},
		PeerName: l.Endpoints[1].GetRandIfaceName(),
		// PeerMac address is set later on
	}

	// add the link
	err := netlink.LinkAdd(linkA)
	if err != nil {
		return err
	}

	// retrieve the netlink.Link for the B / Peer side of the link
	linkB, err := netlink.LinkByName(l.Endpoints[1].GetRandIfaceName())
	if err != nil {
		return err
	}

	// for both ends of the link
	for idx, link := range []netlink.Link{linkA, linkB} {
		var adjustmentFunc func(ns.NetNS) error
		// if the endpoint is a bridge we also need to set the master of the interface to the bridge
		switch l.Endpoints[idx].GetNode().GetLinkEndpointType() {
		case LinkEndpointTypeBridge:
			// retrieve bridge name via node name
			bridgeName := l.Endpoints[idx].GetNode().GetShortName()
			// set the adjustmentFunc to the function that, besides the name, mac and up state
			// also sets the Master of the interface to the bridge
			adjustmentFunc = SetNameMACMasterAndUpInterface(link, l.Endpoints[idx], bridgeName)
		default:
			// use the simple function that renames the link in the container, sets the MAC
			// as well as its state to up
			adjustmentFunc = SetNameMACAndUpInterface(link, l.Endpoints[idx])
		}

		// if the node is a regular namespace node
		// add link to node, rename, set mac and Up
		err = l.Endpoints[idx].GetNode().AddNetlinkLinkToContainer(ctx, link, adjustmentFunc)
		if err != nil {
			return err
		}
	}

	// we need to update the Endpoint with system assigned data like
	// MAC addresses tht have not been assigned by the user
	for _, ep := range l.GetEndpoints() {
		// update the Endpoint with the assigned mac address
		err = ep.VerifyAndPopulateMacAddress()
		if err != nil {
			return err
		}
	}

	l.deploymentState = LinkDeploymentStateDeployed

	return nil
}

func (l *LinkVEth) Remove(_ context.Context) error {
	// TODO
	return nil
}

func (l *LinkVEth) GetEndpoints() []Endpoint {
	return l.Endpoints
}
