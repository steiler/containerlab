package main

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/srl-labs/containerlab/runtime/cri"
	"github.com/srl-labs/containerlab/types"
)

func main() {

	c := cri.CRI{}
	err := c.Init()
	if err != nil {
		log.Fatalf("error: ", err)
	}

	ctx := context.Background()

	container := "2d554325df4d"
	status := c.GetContainerStatus(ctx, container)
	log.Infof("Container %s, Status: %s", container, status)

	image := "nginx:latest"
	err = c.PullImage(ctx, "alpine", types.PullPolicyAlways)
	if err != nil {
		log.Fatalf("error: ", err)
	}
	log.Infof("Image %q pulled successfull!", image)

	// create container
	cID, err := c.CreateContainer(ctx,
		&types.NodeConfig{
			ShortName: "myNode",
			LongName:  "myNode.long.name",
			Fqdn:      "myfqdn",
			Image:     "nginx:latest",
			Labels:    map[string]string{"MyLabel": "MyLabelValue"},
		},
	)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("CID: %s", cID)

	// start container
	_, err = c.StartContainer(ctx, cID, &types.NodeConfig{})
	if err != nil {
		log.Fatalf("error: ", err)
	} else {
		log.Infof("successfully started %q", container)
	}

	// List Conainers
	list, err := c.ListContainers(ctx, nil)
	if err != nil {
		log.Fatalf("error: ", err)
	} else {
		for _, e := range list {
			log.Info("ID: %s, Labels: %s", e.ID, e.Labels)
		}
	}

	os.Exit(0)

	// stop container
	err = c.StopContainer(ctx, "xdp-xdp-1")
	if err != nil {
		log.Fatalf("error: ", err)
	} else {
		log.Infof("successfully stopped %q", container)
	}

	// NSPATH
	info, err := c.GetNSPath(ctx, "xdp-xdp-1")
	if err != nil {
		log.Fatalf("error: ", err)
	} else {
		log.Infof("GetNSPath: %q -> %s", container, info)
	}
}
