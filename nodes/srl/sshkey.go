package srl

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

// catenateKeys catenates the ssh public keys
// and produces a string that can be used in the
// cli config command to set the ssh public keys
// for users.
func (n *srl) catenateKeys() string {
	var keys string

	for _, k := range n.sshPubKeys {
		// marshall the publickey in authorizedKeys format
		// and trim spaces (cause there will be a trailing newline)
		ks := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(*k)))
		// catenate all ssh keys into a single quoted string accepted in CLI
		keys += fmt.Sprintf("%q ", ks)
	}

	return keys
}
