package clab

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// TopoFile type is a struct which defines parameters of the topology file
type TopoFile struct {
	path     string // topo file path
	fullName string // file name with extension
	name     string // file name without extension
}

// GetTopology parses the topology file into c.Conf structure
// as well as populates the TopoFile structure with the topology file related information
func (c *CLab) GetTopology(topo string) error {
	yamlFile, err := ioutil.ReadFile(topo)
	if err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("Topology file contents:\n%s\n", yamlFile))

	err = yaml.UnmarshalStrict(yamlFile, c.Config)
	if err != nil {
		return err
	}

	path, _ := filepath.Abs(topo)
	if err != nil {
		return err
	}

	s := strings.Split(topo, "/")
	file := s[len(s)-1]
	filename := strings.Split(file, ".")
	c.TopoFile = &TopoFile{
		path:     path,
		fullName: file,
		name:     filename[0],
	}
	return nil
}

func fileExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !f.IsDir()
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherwise, copy the file contents from src to dst.
func copyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	return copyFileContents(src, dst)
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func createFile(file, content string) {
	var f *os.File
	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.WriteString(content + "\n"); err != nil {
		panic(err)
	}
}

// CreateDirectory creates a directory by a path with a mode/permission specified by perm.
// If directory exists, the function does not do anything.
func CreateDirectory(path string, perm os.FileMode) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, perm)
	}
}

// CreateNodeDirStructure create the directory structure and files for the lab nodes
func (c *CLab) CreateNodeDirStructure(node *Node) (err error) {
	c.m.RLock()
	defer c.m.RUnlock()

	// create node directory in the lab directory
	// skip creation of node directory for linux/bridge kinds
	// since they don't keep any state normally
	if node.Kind != "linux" && node.Kind != "bridge" {
		CreateDirectory(node.LabDir, 0777)
	}

	switch node.Kind {
	case "srl":
		log.Debugf("Creating directory structure for SRL container: %s", node.ShortName)
		var src string
		var dst string

		// copy license file to node specific directory in lab
		src = node.License
		dst = path.Join(node.LabDir, "license.key")
		if err = copyFile(src, dst); err != nil {
			return fmt.Errorf("CopyFile src %s -> dst %s failed %v", src, dst, err)
		}
		log.Debugf("CopyFile src %s -> dst %s succeeded", src, dst)

		// generate SRL topology file
		err = generateSRLTopologyFile(node.Topology, node.LabDir, node.Index)
		if err != nil {
			return err
		}

		// generate a config file if the destination does not exist
		// if the node has a `config:` statement, the file specified in that section
		// will be used as a template in nodeGenerateConfig()
		CreateDirectory(path.Join(node.LabDir, "config"), 0777)
		dst = path.Join(node.LabDir, "config", "config.json")
		if !fileExists(dst) {
			err = node.generateConfig(dst)
			if err != nil {
				log.Errorf("node=%s, failed to generate config: %v", node.ShortName, err)
			}
		} else {
			log.Debugf("Config File Exists for node %s", node.ShortName)
		}

		// copy env config to node specific directory in lab
		src = "/etc/containerlab/templates/srl/srl_env.conf"
		dst = node.LabDir + "/" + "srlinux.conf"
		err = copyFile(src, dst)
		if err != nil {
			return fmt.Errorf("CopyFile src %s -> dst %s failed %v", src, dst, err)
		}
		log.Debugf("CopyFile src %s -> dst %s succeeded\n", src, dst)

	case "linux":
	case "ceos":
		if err := c.createCEOSFiles(node); err != nil {
			return err
		}

	case "crpd":
		// create config and logs directory that will be bind mounted to crpd
		CreateDirectory(path.Join(node.LabDir, "config"), 0777)
		CreateDirectory(path.Join(node.LabDir, "log"), 0777)

		// copy crpd config from default template or user-provided conf file
		cfg := path.Join(node.LabDir, "/config/juniper.conf")
		if !fileExists(cfg) {
			err = node.generateConfig(cfg)
			if err != nil {
				log.Errorf("node=%s, failed to generate config: %v", node.ShortName, err)
			}
		} else {
			log.Debugf("Config file exists for node %s", node.ShortName)
		}
		// copy crpd sshd conf file to crpd node dir
		src := "/etc/containerlab/templates/crpd/sshd_config"
		dst := node.LabDir + "/config/sshd_config"
		err = copyFile(src, dst)
		if err != nil {
			return fmt.Errorf("file copy [src %s -> dst %s] failed %v", src, dst, err)
		}
		log.Debugf("CopyFile src %s -> dst %s succeeded\n", src, dst)

		if node.License != "" {
			// copy license file to node specific lab directory
			src = node.License
			dst = path.Join(node.LabDir, "/config/license.conf")
			if err = copyFile(src, dst); err != nil {
				return fmt.Errorf("file copy [src %s -> dst %s] failed %v", src, dst, err)
			}
			log.Debugf("CopyFile src %s -> dst %s succeeded", src, dst)
		}
	case "vr-sros":
		// create config directory that will be bind mounted to vrnetlab container at / path
		CreateDirectory(path.Join(node.LabDir, "tftpboot"), 0777)

		if node.License != "" {
			// copy license file to node specific lab directory
			src := node.License
			dst := path.Join(node.LabDir, "/tftpboot/license.txt")
			if err = copyFile(src, dst); err != nil {
				return fmt.Errorf("file copy [src %s -> dst %s] failed %v", src, dst, err)
			}
			log.Debugf("CopyFile src %s -> dst %s succeeded", src, dst)

			cfg := path.Join(node.LabDir, "tftpboot", "config.txt")
			if node.Config != "" {
				err = node.generateConfig(cfg)
				if err != nil {
					log.Errorf("node=%s, failed to generate config: %v", node.ShortName, err)
				}
			} else {
				log.Debugf("Config file exists for node %s", node.ShortName)
			}
		}
	case "bridge":
	default:
	}

	return nil
}

// GenerateConfig generates configuration for the nodes
func (node *Node) generateConfig(dst string) error {
	log.Debugf("generating config for node %s from file %s", node.ShortName, node.Config)
	funcMap := template.FuncMap{
		// INSERT TEMPLTE FUNCTION HERE
		//"extractSrlInterfaces": nil,
	}
	tpl, err := template.New(filepath.Base(node.Config)).Funcs(funcMap).ParseFiles(node.Config)
	if err != nil {
		return err
	}

	dstBytes := new(bytes.Buffer)
	err = tpl.Execute(dstBytes, node)
	if err != nil {
		return err
	}
	log.Debugf("node '%s' generated config: %s", node.ShortName, dstBytes.String())
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(dstBytes.Bytes())
	return err
}

func readFileContent(file string) ([]byte, error) {
	// check file exists
	if !fileExists(file) {
		return nil, fmt.Errorf("file %s does not exist", file)
	}

	// read and return file content
	b, err := ioutil.ReadFile(file)
	return b, err
}
