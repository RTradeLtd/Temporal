package testbed

import (
	"fmt"
	"plugin"

	"github.com/ipfs/iptb/testbed/interfaces"
)

// NodeSpec represents a node's specification
type NodeSpec struct {
	Type  string
	Dir   string
	Attrs map[string]string
}

// IptbPlugin contains exported symbols from loaded plugins
type IptbPlugin struct {
	From        string
	NewNode     testbedi.NewNodeFunc
	GetAttrList testbedi.GetAttrListFunc
	GetAttrDesc testbedi.GetAttrDescFunc
	PluginName  string
	BuiltIn     bool
}

var plugins map[string]IptbPlugin

func init() {
	plugins = make(map[string]IptbPlugin)
}

// GetPlugin returns a plugin registered with RegisterPlugin
func GetPlugin(name string) (IptbPlugin, bool) {
	plg, ok := plugins[name]
	return plg, ok
}

// RegisterPlugin registers a plugin, the `force` flag can be passed to
// override any plugin registered under the same IptbPlugin.PluginName
func RegisterPlugin(plg IptbPlugin, force bool) (bool, error) {
	overloaded := false

	if pl, exists := plugins[plg.PluginName]; exists && !force {
		if pl.BuiltIn {
			overloaded = true
		} else {
			return false, fmt.Errorf("plugin %s already loaded from %s", pl.PluginName, pl.From)
		}
	}

	plugins[plg.PluginName] = plg

	return overloaded, nil

}

// LoadPlugin loads a plugin from `path`
func LoadPlugin(path string) (*IptbPlugin, error) {
	return loadPlugin(path)
}

// LoadPluginCore loads core symbols from a golang plugin into an IptbPlugin
func loadPluginCore(pl *plugin.Plugin, plg *IptbPlugin) error {
	NewNodeSym, err := pl.Lookup("NewNode")
	if err != nil {
		return err
	}

	NewNode, ok := NewNodeSym.(*testbedi.NewNodeFunc)
	if !ok {
		return fmt.Errorf("Error: could not cast `NewNode` of %s", pl)
	}

	PluginNameSym, err := pl.Lookup("PluginName")
	if err != nil {
		return err
	}

	PluginName, ok := PluginNameSym.(*string)
	if !ok {
		return fmt.Errorf("Error: could not cast `PluginName` of %s", pl)
	}

	plg.PluginName = *PluginName
	plg.NewNode = *NewNode

	return nil
}

// LoadPluginCore loads attr symbols from a golang plugin into an IptbPlugin
func loadPluginAttr(pl *plugin.Plugin, plg *IptbPlugin) (bool, error) {
	GetAttrListSym, err := pl.Lookup("GetAttrList")
	if err != nil {
		return false, err
	}

	GetAttrList, ok := GetAttrListSym.(*testbedi.GetAttrListFunc)
	if !ok {
		return true, fmt.Errorf("Error: could not cast `GetAttrList` of %s", pl)
	}

	GetAttrDescSym, err := pl.Lookup("GetAttrDesc")
	if err != nil {
		return false, err
	}

	GetAttrDesc, ok := GetAttrDescSym.(*testbedi.GetAttrDescFunc)
	if !ok {
		return true, fmt.Errorf("Error: could not cast `GetAttrDesc` of %s", pl)
	}

	plg.GetAttrList = *GetAttrList
	plg.GetAttrDesc = *GetAttrDesc

	return true, nil
}

func loadPlugin(path string) (*IptbPlugin, error) {
	pl, err := plugin.Open(path)

	if err != nil {
		return nil, err
	}

	var plg IptbPlugin

	if err := loadPluginCore(pl, &plg); err != nil {
		return nil, err
	}

	if ok, err := loadPluginAttr(pl, &plg); ok && err != nil {
		return nil, err
	}

	return &plg, nil
}

// Load uses plugins registered with RegisterPlugin to construct a Core node
// from the NodeSpec
func (ns *NodeSpec) Load() (testbedi.Core, error) {
	pluginName := ns.Type

	if plg, ok := plugins[pluginName]; ok {
		return plg.NewNode(ns.Dir, ns.Attrs)
	}

	return nil, fmt.Errorf("Could not find plugin %s", pluginName)
}

// SetAttr sets an attribute on the NodeSpec
func (ns *NodeSpec) SetAttr(attr string, val string) {
	ns.Attrs[attr] = val
}

// GetAttr gets an attribute from the NodeSpec
func (ns *NodeSpec) GetAttr(attr string) (string, error) {
	if v, ok := ns.Attrs[attr]; ok {
		return v, nil
	}

	return "", fmt.Errorf("Attr not set")
}
