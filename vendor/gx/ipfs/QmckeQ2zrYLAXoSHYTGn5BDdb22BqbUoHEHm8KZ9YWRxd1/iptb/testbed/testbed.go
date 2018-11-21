package testbed

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/ipfs/iptb/testbed/interfaces"
	"github.com/ipfs/iptb/util"
)

type Testbed interface {
	Name() string

	// Spec returns a spec for node n
	Spec(n int) (*NodeSpec, error)

	// Specs returns all specs
	Specs() ([]*NodeSpec, error)

	// Node returns node n, specified by spec n
	Node(n int) (testbedi.Core, error)

	// Node returns all nodes, specified by all specs
	Nodes() ([]testbedi.Core, error)

	/****************/
	/* Future Ideas */

	// Would be neat to have a TestBed Config interface
	// The node interface GetAttr and SetAttr should be a shortcut into this
	// Config() (map[interface{}]interface{}, error)

}

type BasicTestbed struct {
	dir   string
	specs []*NodeSpec
	nodes []testbedi.Core
}

func NewTestbed(dir string) BasicTestbed {
	return BasicTestbed{
		dir: dir,
	}
}

func (tb *BasicTestbed) Dir() string {
	return tb.dir
}

func (tb BasicTestbed) Name() string {
	return tb.dir
}

func AlreadyInitCheck(dir string, force bool) error {
	if _, err := os.Stat(filepath.Join(dir, "nodespec.json")); !os.IsNotExist(err) {
		if !force && !iptbutil.YesNoPrompt("testbed nodes already exist, overwrite? [y/n]") {
			return nil
		}

		return os.RemoveAll(dir)
	}

	return nil
}

func BuildSpecs(base string, count int, typ string, attrs map[string]string) ([]*NodeSpec, error) {
	var specs []*NodeSpec

	for i := 0; i < count; i++ {
		dir := path.Join(base, fmt.Sprint(i))

		if err := os.MkdirAll(dir, 0775); err != nil {
			return nil, err
		}

		spec := &NodeSpec{
			Type:  typ,
			Dir:   dir,
			Attrs: attrs,
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

func (tb *BasicTestbed) Spec(n int) (*NodeSpec, error) {
	specs, err := tb.Specs()

	if err != nil {
		return nil, err
	}

	if n >= len(specs) {
		return nil, fmt.Errorf("Spec index out of range")
	}

	return specs[n], err
}

func (tb *BasicTestbed) Specs() ([]*NodeSpec, error) {
	if tb.specs != nil {
		return tb.specs, nil
	}

	return tb.loadSpecs()
}

func (tb *BasicTestbed) Node(n int) (testbedi.Core, error) {
	nodes, err := tb.Nodes()

	if err != nil {
		return nil, err
	}

	if n >= len(nodes) {
		return nil, fmt.Errorf("Node index out of range")
	}

	return nodes[n], err
}

func (tb *BasicTestbed) Nodes() ([]testbedi.Core, error) {
	if tb.nodes != nil {
		return tb.nodes, nil
	}

	return tb.loadNodes()
}

func (tb *BasicTestbed) loadSpecs() ([]*NodeSpec, error) {
	specs, err := ReadNodeSpecs(tb.dir)
	if err != nil {
		return nil, err
	}

	return specs, nil
}

func (tb *BasicTestbed) loadNodes() ([]testbedi.Core, error) {
	specs, err := tb.Specs()
	if err != nil {
		return nil, err
	}

	return NodesFromSpecs(specs)
}

func NodesFromSpecs(specs []*NodeSpec) ([]testbedi.Core, error) {
	var out []testbedi.Core
	for _, s := range specs {
		nd, err := s.Load()
		if err != nil {
			return nil, err
		}
		out = append(out, nd)
	}
	return out, nil
}

func ReadNodeSpecs(dir string) ([]*NodeSpec, error) {
	data, err := ioutil.ReadFile(filepath.Join(dir, "nodespec.json"))
	if err != nil {
		return nil, err
	}

	var specs []*NodeSpec
	err = json.Unmarshal(data, &specs)
	if err != nil {
		return nil, err
	}

	return specs, nil
}

func WriteNodeSpecs(dir string, specs []*NodeSpec) error {
	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return err
	}

	fi, err := os.Create(filepath.Join(dir, "nodespec.json"))
	if err != nil {
		return err
	}

	defer fi.Close()
	return json.NewEncoder(fi).Encode(specs)
}
