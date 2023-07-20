package types

import "fmt"

// LinkMACVLANRaw is the raw (string) representation of a macvlan link as defined in the topology file.
type LinkMACVLANRaw struct {
	LinkCommonParams `yaml:",inline"`
	HostInterface    string       `yaml:"host-interface"`
	Endpoint         *EndpointRaw `yaml:"endpoint"`
}

// ToLinkConfig converts the raw link into a LinkConfig.
func (r *LinkMACVLANRaw) ToLinkConfig() *LinkConfig {
	lc := &LinkConfig{
		Vars:      r.Vars,
		Labels:    r.Labels,
		MTU:       r.Mtu,
		Endpoints: make([]string, 2),
	}

	lc.Endpoints[0] = fmt.Sprintf("%s:%s", r.Endpoint.Node, r.Endpoint.Iface)
	lc.Endpoints[1] = fmt.Sprintf("%s:%s", "macvlan", r.HostInterface)

	return lc
}

func macVlanFromLinkConfig(lc LinkConfig, specialEPIndex int) (*LinkMACVLANRaw, error) {
	_, hostIf, node, nodeIf := extractHostNodeInterfaceData(lc, specialEPIndex)

	result := &LinkMACVLANRaw{
		LinkCommonParams: LinkCommonParams{
			Mtu:    lc.MTU,
			Labels: lc.Labels,
			Vars:   lc.Vars,
		},
		HostInterface: hostIf,
		Endpoint:      NewEndpointRaw(node, nodeIf, ""),
	}

	return result, nil
}

func (r *LinkMACVLANRaw) Resolve() (LinkInterf, error) {
	// TODO: needs implementation
	return nil, nil
}
