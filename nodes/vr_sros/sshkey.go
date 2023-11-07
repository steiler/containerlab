package vr_sros

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// mapSSHPubKeys provides extracted key values based on key-algo for usage in vrSROS configuration
func (s *vrSROS) mapSSHPubKeys(sshKeyMapping map[string]*[]string) {
	// iterate through keys
	for _, k := range s.sshPubKeys {
		// find mapped slice for key type
		list, mappingFound := sshKeyMapping[k.Type()]
		if !mappingFound {
			log.Debugf("no mapping for key type %q found, ignoring key", k.Type())
		}
		// extract the fields
		// <keytype> <key> <comment>
		keyFields := strings.Fields(string(ssh.MarshalAuthorizedKey(k)))

		*list = append((*list), keyFields[1])
	}
}
