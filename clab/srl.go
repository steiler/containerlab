package clab

import (
	"crypto/rand"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
)

type mac struct {
	MAC string
}

type SRLNode struct {
	Node
}

func generateSRLTopologyFile(src, labDir string, index int) error {
	dst := path.Join(labDir, "topology.yml")
	tpl, err := template.ParseFiles(src)
	if err != nil {
		return err
	}

	// generate random bytes to use in the 2-3rd bytes of a base mac
	// this ensures that different srl nodes will have different macs for their ports
	buf := make([]byte, 2)
	_, err = rand.Read(buf)
	if err != nil {
		return err
	}
	m := fmt.Sprintf("02:%02x:%02x:00:00:00", buf[0], buf[1])

	mac := mac{
		MAC: m,
	}
	log.Debug(mac, dst)
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = tpl.Execute(f, mac); err != nil {
		return err
	}
	return nil
}

func (node *SRLNode) InitNode(c *CLab, nodeCfg NodeConfig, user string, envs map[string]string) error {
	var err error
	// initialize the global parameters with defaults, can be overwritten later
	node.Config, err = c.configInit(&nodeCfg, node.Kind)
	if err != nil {
		return err
	}

	lp, err := c.licenseInit(&nodeCfg, &node.Node)
	if err != nil {
		return err
	}

	if lp == "" {
		return fmt.Errorf("no license found for node '%s' of kind '%s'", node.ShortName, node.Kind)
	}

	node.License = lp

	node.Image = c.imageInitialization(&nodeCfg, node.Kind)
	node.Group = c.groupInitialization(&nodeCfg, node.Kind)
	node.NodeType = c.typeInit(&nodeCfg, node.Kind)
	node.Position = c.positionInitialization(&nodeCfg, node.Kind)

	if filename, found := srlTypes[node.NodeType]; found {
		node.Topology = baseConfigDir + filename
	} else {
		keys := make([]string, 0, len(srlTypes))
		for key := range srlTypes {
			keys = append(keys, key)
		}
		log.Fatalf("wrong node type. '%s' doesn't exist. should be any of %s", node.NodeType, strings.Join(keys, ", "))
	}

	// initialize specifc container information
	node.Cmd = "sudo sr_linux"

	kindEnv := map[string]string{"SRLINUX": "1"}
	node.Env = mergeStringMaps(kindEnv, envs)

	// if user was not initialized to a value, use root
	if user == "" {
		user = "0:0"
	}
	node.User = user

	node.Sysctls = make(map[string]string)
	node.Sysctls["net.ipv4.ip_forward"] = "0"
	node.Sysctls["net.ipv6.conf.all.disable_ipv6"] = "0"
	node.Sysctls["net.ipv6.conf.all.accept_dad"] = "0"
	node.Sysctls["net.ipv6.conf.default.accept_dad"] = "0"
	node.Sysctls["net.ipv6.conf.all.autoconf"] = "0"
	node.Sysctls["net.ipv6.conf.default.autoconf"] = "0"

	// we mount a fixed path node.Labdir/license.key as the license referenced in topo file will be copied to that path
	// in (c *cLab) CreateNodeDirStructure
	node.Binds = append(node.Binds, fmt.Sprint(filepath.Join(node.LabDir, "license.key"), ":/opt/srlinux/etc/license.key:ro"))

	// mount config directory
	cfgPath := filepath.Join(node.LabDir, "config")
	node.Binds = append(node.Binds, fmt.Sprint(cfgPath, ":/etc/opt/srlinux/:rw"))

	// mount srlinux.conf
	srlconfPath := filepath.Join(node.LabDir, "srlinux.conf")
	node.Binds = append(node.Binds, fmt.Sprint(srlconfPath, ":/home/admin/.srlinux.conf:rw"))

	// mount srlinux topology
	topoPath := filepath.Join(node.LabDir, "topology.yml")
	node.Binds = append(node.Binds, fmt.Sprint(topoPath, ":/tmp/topology.yml:ro"))

	return err
}

func init() {
	RegisterNodeType("SRL")
}
