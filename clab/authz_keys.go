// Copyright 2021 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package clab

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/srl-labs/containerlab/utils"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	pubKeysGlob = "~/.ssh/*.pub"
	// authorized keys file path on a clab host that is used to create the clabAuthzKeys file.
	authzKeysFPath = "~/.ssh/authorized_keys"
)

// CreateAuthzKeysFile creats the authorized_keys file in the lab directory
// if any files ~/.ssh/*.pub found.
func (c *CLab) CreateAuthzKeysFile() error {
	b := new(bytes.Buffer)

	// get keys registered with ssh-agent
	keys, err := RetrieveSSHPubKeys()
	if err != nil {
		log.Debug(err)
	}

	for _, k := range keys {
		x := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(*k)))
		addKeyToBuffer(b, x)
	}

	clabAuthzKeysFPath := c.TopoPaths.AuthorizedKeysFilename()
	if err := utils.CreateFile(clabAuthzKeysFPath, b.String()); err != nil {
		return err
	}

	// ensure authz_keys will have the permissions allowing it to be read by anyone
	return os.Chmod(clabAuthzKeysFPath, 0644) // skipcq: GSC-G302
}

// RetrieveSSHPubKeysFromFiles retrieves public keys from the ~/.ssh/*.authorized_keys
// and ~/.ssh/*.pub files.
func RetrieveSSHPubKeysFromFiles() ([]*ssh.PublicKey, error) {
	var keys []*ssh.PublicKey
	p := utils.ResolvePath(pubKeysGlob, "")

	all, err := filepath.Glob(p)
	if err != nil {
		return nil, fmt.Errorf("failed globbing the path %s", p)
	}

	f := utils.ResolvePath(authzKeysFPath, "")

	if utils.FileExists(f) {
		log.Debugf("%s found, adding the public keys it contains", f)
		all = append(all, f)
	}

	// iterate through all files with key material
	for _, fn := range all {
		rb, err := os.ReadFile(fn)
		if err != nil {
			return nil, fmt.Errorf("failed reading the file %s: %v", fn, err)
		}

		pubKey, _, _, _, err := ssh.ParseAuthorizedKey(rb)
		if err != nil {
			return nil, err
		}

		keys = append(keys, &pubKey)
	}

	return keys, nil
}

// RetrieveSSHPubKeys retrieves the PubKeys from the different sources
// SSHAgent as well as all home dir based /.ssh/*.pub files.
func RetrieveSSHPubKeys() ([]*ssh.PublicKey, error) {
	keys := make([]*ssh.PublicKey, 0)

	fkeys, err := RetrieveSSHPubKeysFromFiles()
	if err != nil {
		return nil, err
	}

	agentKeys, err := RetrieveSSHAgentKeys()
	if err != nil {
		return nil, err
	}

	keysM := map[string]*ssh.PublicKey{}
	for _, k := range append(fkeys, agentKeys...) {
		keysM[string(ssh.MarshalAuthorizedKey(*k))] = k
	}

	for _, k := range keysM {
		keys = append(keys, k)
	}

	return keys, nil
}

// addKeyToBuffer adds a key to the buffer if the key is not already present.
func addKeyToBuffer(b *bytes.Buffer, key string) {
	if !strings.Contains(b.String(), key) {
		b.WriteString(key + "\n")
	}
}

// RetrieveSSHAgentKeys retrieves public keys registered with the ssh-agent.
func RetrieveSSHAgentKeys() ([]*ssh.PublicKey, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if len(socket) == 0 {
		return nil, fmt.Errorf("SSH_AUTH_SOCK not set, skipping pubkey fetching")
	}
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSH_AUTH_SOCK: %w", err)
	}

	agentClient := agent.NewClient(conn)
	keys, err := agentClient.List()
	if err != nil {
		return nil, fmt.Errorf("error listing agent's pub keys %w", err)
	}

	log.Debugf("extracted %d keys from ssh-agent", len(keys))

	var pubKeys []*ssh.PublicKey

	for _, key := range keys {
		pkey, err := ssh.ParsePublicKey(key.Blob)
		if err != nil {
			return nil, err
		}
		pubKeys = append(pubKeys, &pkey)
	}

	return pubKeys, nil
}
