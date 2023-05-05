package types

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/srl-labs/containerlab/utils"
)

const (
	ansibleInventoryFileName  = "ansible-inventory.yml"
	topologyExportDatFileName = "topology-data.json"
	authzKeysFileName         = "authorized_keys"
	tlsDir                    = ".tls"
	caDir                     = "ca"
	graph                     = "graph"
	labDirPrefix              = "clab-"
	backupFileSuffix          = ".bak"
	backupFilePrefix          = "."
	CertFileSuffix            = ".pem"
	KeyFileSuffix             = ".key"
	CSRFileSuffix             = ".csr"
)

// clabTmpDir is the directory where clab stores temporary and/or downloaded files.
var clabTmpDir = filepath.Join(os.TempDir(), ".clab")

// TopoPaths creates all the required absolute paths and filenames for a topology.
// generally all these paths are deduced from two main paths. The topology file path and the lab dir path.
type TopoPaths struct {
	topoFile           string
	labDir             string
	topoName           string
	explicitCaCertFile string // if an external CA certificate is used the name to the Cert file is stored here
	explicitCaKeyFile  string // if an external CA certificate is used the name to the Key file is stored here
}

// NewTopoPaths constructs a new TopoPaths instance.
func NewTopoPaths(topologyFile string) (*TopoPaths, error) {
	t := &TopoPaths{}
	err := t.SetTopologyFilePath(topologyFile)
	if err != nil {
		return nil, err
	}

	return t, err
}

func NewCaTopoPaths(labDir string) (*TopoPaths, error) {
	return &TopoPaths{
		labDir: labDir,
	}, nil
}

// SetTopologyFilePath sets the topology file path.
func (t *TopoPaths) SetTopologyFilePath(topologyFile string) error {
	absTopoFile, err := filepath.Abs(topologyFile)
	if err != nil {
		return err
	}

	// make sure topo file exists
	_, err = os.Stat(absTopoFile)
	if err != nil {
		return err
	}

	t.topoFile = absTopoFile

	return nil
}

// SetLabDir sets the labDir foldername (no abs path, but the last element) usually the topology name.
func (t *TopoPaths) SetLabDir(topologyName string) (err error) {
	t.topoName = topologyName
	// if "CLAB_LABDIR_BASE" Env Var is set, use that dir as a base
	// for the labDir, otherwise use PWD.
	baseDir := os.Getenv("CLAB_LABDIR_BASE")
	if baseDir == "" {
		baseDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	// construct the path
	t.labDir = path.Join(baseDir, labDirPrefix+topologyName)
	return nil
}

// SetExternalCaFiles sets the filename for the cert and key files if externally generated should be used
func (t *TopoPaths) SetExternalCaFiles(certFile, keyFile string) error {
	if !utils.FileExists(certFile) {
		return fmt.Errorf("external CA cert file %s does not exist", certFile)
	}
	if !utils.FileExists(keyFile) {
		return fmt.Errorf("external CA key file %s does not exist", keyFile)
	}
	t.explicitCaCertFile = certFile
	t.explicitCaKeyFile = keyFile
	return nil
}

// TLSBaseDir returns the path of the TLS directory structure.
func (t *TopoPaths) TLSBaseDir() string {
	return path.Join(t.labDir, tlsDir)
}

// NodeTLSDir returns the directory that contains the certificat data for the given node.
func (t *TopoPaths) NodeTLSDir(nodename string) string {
	return path.Join(t.TLSBaseDir(), nodename)
}

// AuthorizedKeysFilename returns the path for the generated AuthorizedKeysFile.
func (t *TopoPaths) AuthorizedKeysFilename() string {
	return path.Join(t.labDir, authzKeysFileName)
}

// GraphDir returns the directory that takes the graphs.
func (t *TopoPaths) GraphDir() string {
	return path.Join(t.labDir, graph)
}

// GraphFilename returns the filename for a given graph file with the provided extension.
func (t *TopoPaths) GraphFilename(ext string) string {
	// add `.` to the extension if not provided
	if len(ext) > 0 && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	return path.Join(t.GraphDir(), t.TopologyFilenameWithoutExt()+ext)
}

// NodeDir returns the directory in the labDir for the provided node.
func (t *TopoPaths) NodeDir(nodeName string) string {
	return path.Join(t.labDir, nodeName)
}

// TopoExportFile returns the path for the topology-export file.
func (t *TopoPaths) TopoExportFile() string {
	return path.Join(t.labDir, topologyExportDatFileName)
}

// AnsibleInventoryFileAbsPath returns the absolute path to the ansible-inventory file.
func (t *TopoPaths) AnsibleInventoryFileAbsPath() string {
	return path.Join(t.labDir, ansibleInventoryFileName)
}

// TopologyFilenameAbsPath returns the absolute path to the topology file.
func (t *TopoPaths) TopologyFilenameAbsPath() string {
	return t.topoFile
}

// ClabTmpDir returns the path to the temporary directory where clab stores temporary and/or downloaded files.
func (*TopoPaths) ClabTmpDir() string {
	return clabTmpDir
}

// StartupConfigDownloadFileAbsPath returns the absolute path to the startup-config file
// when it is downloaded from a remote location to the clab temp directory.
func (t *TopoPaths) StartupConfigDownloadFileAbsPath(node, postfix string) string {
	return filepath.Join(t.ClabTmpDir(), fmt.Sprintf("%s-%s-%s", t.topoName, node, postfix))
}

// TopologyFilenameBase returns the full filename of the topology file
// without any additional paths.
func (t *TopoPaths) TopologyFilenameBase() string {
	return filepath.Base(t.topoFile)
}

// TopologyFilenameWithoutExt returns the topology file name without the file extension.
func (t *TopoPaths) TopologyFilenameWithoutExt() string {
	name := t.TopologyFilenameBase()

	return name[:len(name)-len(filepath.Ext(name))]
}

func (t *TopoPaths) TopologyFileIsSet() bool {
	if t == nil {
		return false
	}

	return t.topoFile != ""
}

// TopologyBakFileAbsPath returns the backup topology file name.
func (t *TopoPaths) TopologyBakFileAbsPath() string {
	return path.Join(t.TopologyFileDir(), backupFilePrefix+t.TopologyFilenameBase()+backupFileSuffix)
}

// TopologyFileDir returns the abs path to the topology file directory.
func (t *TopoPaths) TopologyFileDir() string {
	return filepath.Dir(t.topoFile)
}

// TopologyLabDir returns the lab directory.
func (t *TopoPaths) TopologyLabDir() string {
	return t.labDir
}

// NodeCertKeyAbsFilename returns the path to a key file for the given identifier.
func (t *TopoPaths) NodeCertKeyAbsFilename(nodeName string) string {
	return path.Join(t.NodeTLSDir(nodeName), nodeName+KeyFileSuffix)
}

// NodeCertAbsFilename returns the path to a cert file for the given identifier.
func (t *TopoPaths) NodeCertAbsFilename(nodeName string) string {
	return path.Join(t.NodeTLSDir(nodeName), nodeName+CertFileSuffix)
}

// NodeCertCSRAbsFilename returns the path to a csr file for the given identifier.
func (t *TopoPaths) NodeCertCSRAbsFilename(nodeName string) string {
	return path.Join(t.NodeTLSDir(nodeName), nodeName+CSRFileSuffix)
}

func (t *TopoPaths) CaCertFile() string {
	if t.explicitCaCertFile != "" {
		return t.explicitCaCertFile
	}
	return t.NodeCertAbsFilename(caDir)
}

func (t *TopoPaths) CaKeyFile() string {
	if t.explicitCaKeyFile != "" {
		return t.explicitCaKeyFile
	}
	return t.NodeCertKeyAbsFilename(caDir)
}

func (t *TopoPaths) CaCSRFile() string {
	return t.NodeCertCSRAbsFilename(caDir)
}
