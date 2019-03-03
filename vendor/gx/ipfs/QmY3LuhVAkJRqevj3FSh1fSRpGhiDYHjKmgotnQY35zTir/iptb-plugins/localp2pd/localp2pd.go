package pluginlocalp2pd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	client "gx/ipfs/QmRpsHkKwSXnbFRaQUhAY27WtnUAC2m8tAPfws9Lk72T4h/go-libp2p-daemon/p2pclient"
	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	peer "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"
	"gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/testbed/interfaces"
)

const (
	// PluginName is the libp2p daemon IPTB plugin name.
	PluginName = "localp2pd"
)

// type check
var _ testbedi.Attribute = (*LocalP2pd)(nil)
var _ testbedi.Libp2p = (*LocalP2pd)(nil)

// var _ testbedi.Config = (*LocalP2pd)(nil)
// var _ testbedi.Metric = (*LocalP2pd)(nil)

type connManagerConfig struct {
	lowWatermark  *int
	highWatermark *int
	gracePeriod   *int
}

// LocalP2pd wraps behaviors of the libp2p daemon.
type LocalP2pd struct {
	// config options
	command        string
	dir            string
	connManager    *connManagerConfig
	dhtMode        string
	bootstrap      bool
	bootstrapPeers string

	// process
	process *os.Process
	alive   bool
}

// NewNode creates a localp2pd iptb core node that runs the libp2p daemon on the
// local system using control sockets within the specified directory.
// Attributes:
// - controladdr:
func NewNode(dir string, attrs map[string]string) (testbedi.Core, error) {
	// defaults
	var (
		connManager    *connManagerConfig
		dhtMode        string
		found          bool
		err            error
		bootstrapPeers string
		command        string
		process        *os.Process
		bootstrap      = false
	)

	pidpath := filepath.Join(dir, "p2pd.pid")
	pidbytes, err := ioutil.ReadFile(pidpath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	pid, err := strconv.Atoi(string(pidbytes))
	if err != nil {
		os.Remove(pidpath)
	} else {
		process, err = os.FindProcess(pid)
		if err != nil {
			return nil, err
		}
	}

	if dhtMode, found = attrs["dhtmode"]; !found {
		dhtMode = "off"
	}

	if _, found = attrs["connmanager"]; found {
		connManager = &connManagerConfig{}
	}

	if lowmark, ok := attrs["connmanagerlowmark"]; ok {
		if connManager == nil {
			return nil, errors.New("conn manager low watermark provided without enabling conn manager")
		}
		var lowmarki int
		if lowmarki, err = strconv.Atoi(lowmark); err != nil {
			return nil, fmt.Errorf("parsing low watermark: %s", err)
		}
		connManager.lowWatermark = &lowmarki
	}

	if highmark, ok := attrs["connmanagerhighmark"]; ok {
		if connManager == nil {
			return nil, errors.New("conn manager high watermark provided without enabling conn manager")
		}
		var highmarki int
		if highmarki, err = strconv.Atoi(highmark); err != nil {
			return nil, fmt.Errorf("parsing low watermark: %s", err)
		}
		connManager.highWatermark = &highmarki
	}

	if graceperiod, ok := attrs["connmanagergraceperiod"]; ok {
		if connManager == nil {
			return nil, errors.New("conn manager grace period provided without enabling conn manager")
		}
		var graceperiodi int
		if graceperiodi, err = strconv.Atoi(graceperiod); err != nil {
			return nil, fmt.Errorf("parsing low watermark: %s", err)
		}
		connManager.gracePeriod = &graceperiodi
	}

	if _, found = attrs["bootstrap"]; found {
		bootstrap = true
	}

	if bootstrapPeers, found = attrs["bootstrapPeers"]; !found {
		bootstrapPeers = ""
	}

	if command, found = attrs["command"]; !found {
		command = "p2pd"
	}

	p2pd := &LocalP2pd{
		command:        command,
		dir:            dir,
		dhtMode:        dhtMode,
		connManager:    connManager,
		bootstrap:      bootstrap,
		bootstrapPeers: bootstrapPeers,
		process:        process,
	}
	return p2pd, nil
}

func (l *LocalP2pd) sockPath() string {
	return l.dir + "/p2pd.sock"
}

func (l *LocalP2pd) cmdArgs() []string {
	var args []string

	switch l.dhtMode {
	case "full":
		args = append(args, "-dht")
	case "client":
		args = append(args, "-dhtClient")
	}

	if l.bootstrap {
		args = append(args, "-b")
	}

	if l.bootstrapPeers != "" {
		args = append(args, "-bootstrapPeers", l.bootstrapPeers)
	}

	if l.connManager != nil {
		args = append(args, "-connManager")

		if l.connManager.gracePeriod != nil {
			args = append(args, "-connGrace", strconv.Itoa(*l.connManager.gracePeriod))
		}

		if l.connManager.highWatermark != nil {
			args = append(args, "-connHi", strconv.Itoa(*l.connManager.highWatermark))
		}

		if l.connManager.lowWatermark != nil {
			args = append(args, "-connLo", strconv.Itoa(*l.connManager.lowWatermark))
		}

	}
	args = append(args, "-sock", l.sockPath())

	return args
}

// PeerID returns the peer id
func (l *LocalP2pd) PeerID() (string, error) {
	client, err := l.newClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	peerID, _, err := client.Identify()
	if err != nil {
		return "", err
	}
	return peerID.Pretty(), nil
}

// APIAddr returns the multiaddr for the api
func (l *LocalP2pd) APIAddr() (string, error) {
	return l.sockPath(), nil
}

// SwarmAddrs returns the swarm addrs for the node
func (l *LocalP2pd) SwarmAddrs() ([]string, error) {
	client, err := l.newClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	_, addrs, err := client.Identify()
	if err != nil {
		return nil, err
	}
	addrstrs := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		addrstrs = append(addrstrs, addr.String())
	}
	return addrstrs, nil
}

// Init is a no-op
func (l *LocalP2pd) Init(ctx context.Context, args ...string) (testbedi.Output, error) {
	return nil, nil
}

// Start launches a libp2p daemon
func (l *LocalP2pd) Start(ctx context.Context, wait bool, args ...string) (testbedi.Output, error) {
	if l.alive {
		return nil, fmt.Errorf("libp2p daemon is already running")
	}

	// set up command
	cmdargs := append(l.cmdArgs(), args...)
	cmd := exec.Command(l.command, cmdargs...)
	cmd.Dir = l.dir

	stdout, err := os.Create(filepath.Join(l.dir, "p2pd.stdout"))
	if err != nil {
		return nil, err
	}

	stderr, err := os.Create(filepath.Join(l.dir, "p2pd.stderr"))
	if err != nil {
		return nil, err
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	pid := cmd.Process.Pid

	err = ioutil.WriteFile(filepath.Join(l.dir, "p2pd.pid"), []byte(fmt.Sprint(pid)), 0666)
	if err != nil {
		return nil, fmt.Errorf("writing libp2p daemon pid: %s", err)
	}

	if wait {
		for i := 0; i < 50; i++ {
			_, err := os.Stat(l.sockPath())
			if err != nil {
				time.Sleep(time.Millisecond * 400)
				continue
			}
			return nil, nil
		}
		return nil, fmt.Errorf("libp2p daemon with pid %d failed to come online ", pid)
	}

	return nil, nil
}

// Stop shuts down the daemon
func (l *LocalP2pd) Stop(ctx context.Context) error {
	// Stop a client if it exists
	proc := l.process
	if proc == nil {
		return nil
	}

	waitch := make(chan struct{}, 1)
	go func() {
		proc.Wait()
		waitch <- struct{}{}
	}()

	// cleanup
	defer func() {
		os.Remove(filepath.Join(l.dir, "p2pd.pid"))
		os.Remove(filepath.Join(l.dir, "p2pd.sock"))
	}()

	for i := 0; i < 2; i++ {
		if err := l.signalAndWait(waitch, syscall.SIGTERM, time.Second*5); err != errTimeout {
			return err
		}
	}

	if err := l.signalAndWait(waitch, syscall.SIGQUIT, time.Second*5); err != errTimeout {
		return err
	}

	if err := l.signalAndWait(waitch, syscall.SIGKILL, time.Second*5); err != errTimeout {
		return err
	}

	return nil
}

var errTimeout = fmt.Errorf("timed out waiting for process to exit")

func (l *LocalP2pd) signalAndWait(waitch chan struct{}, sig syscall.Signal, timeout time.Duration) error {
	if err := l.process.Signal(sig); err != nil {
		return err
	}

	select {
	case <-waitch:
		return nil
	case <-time.After(timeout):
		return errTimeout
	}
}

// RunCmd is a no-op for the libp2p daemon
func (l *LocalP2pd) RunCmd(ctx context.Context, stdin io.Reader, args ...string) (testbedi.Output, error) {
	return nil, nil
}

// Connect the node to another
func (l *LocalP2pd) Connect(ctx context.Context, n testbedi.Core) error {
	client, err := l.newClient()
	if err != nil {
		return err
	}
	defer client.Close()

	peerstr, err := n.PeerID()
	if err != nil {
		return err
	}
	peer, err := peer.IDB58Decode(peerstr)
	if err != nil {
		return err
	}

	var addrs []ma.Multiaddr
	addrstrs, err := n.SwarmAddrs()
	if err != nil {
		return err
	}
	for _, addrstr := range addrstrs {
		addr, err := ma.NewMultiaddr(addrstr)
		// log?
		if err != nil {
			continue
		}
		addrs = append(addrs, addr)
	}

	return client.Connect(peer, addrs)
}

func (l *LocalP2pd) newClient() (*client.Client, error) {
	ctrlAddr, err := ma.NewComponent("unix", l.sockPath())
	if err != nil {
		return nil, err
	}
	daemonAddr, err := ma.NewComponent("unix", filepath.Join(l.dir, "p2pclient.sock"))
	if err != nil {
		return nil, err
	}
	client, err := client.NewClient(ctrlAddr, daemonAddr)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Shell is a no-op for the libp2p daemon.
func (l *LocalP2pd) Shell(ctx context.Context, ns []testbedi.Core) error {
	return nil
}

// Dir returns the iptb directory assigned to the node
func (l *LocalP2pd) Dir() string {
	return l.dir
}

// Type returns a string that identifies the implementation
// Examples localipfs, dockeripfs, etc.
func (l *LocalP2pd) Type() string {
	return PluginName
}

func (l *LocalP2pd) String() string {
	pcid, err := l.PeerID()
	if err != nil {
		return fmt.Sprintf("%s", l.Type())
	}
	return fmt.Sprintf("%s{%s}", l.Type(), pcid[0:12])
}

var attrDesc = map[string]string{
	"id":             "the peer id (base58 encoded)",
	"addresses":      "comma-separated list of multiaddrs peer is listening on(as strings)",
	"controladdress": "the address the daemon is listening on",
}

func GetAttrList() []string {
	attrs := make([]string, 0, len(attrDesc))
	for attr := range attrDesc {
		attrs = append(attrs, attr)
	}
	return attrs
}

func GetAttrDesc(attr string) (string, error) {
	desc, ok := attrDesc[attr]
	if !ok {
		return "", fmt.Errorf("the libp2p daemon does not expose an attribute named \"%s\"", attr)
	}
	return desc, nil
}

func (l *LocalP2pd) GetAttrList() []string {
	return GetAttrList()
}

func (l *LocalP2pd) GetAttrDesc(attr string) (string, error) {
	return GetAttrDesc(attr)
}

func (l *LocalP2pd) peerAddresses() ([]string, error) {
	client, err := l.newClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	_, addrs, err := client.Identify()
	if err != nil {
		return nil, err
	}
	addrstrs := make([]string, len(addrs))
	for i, addr := range addrs {
		addrstrs[i] = addr.String()
	}
	return addrstrs, nil
}

func (l *LocalP2pd) Attr(attr string) (string, error) {
	switch attr {
	case "id":
		return l.PeerID()
	case "addresses":
		addrs, err := l.peerAddresses()
		if err != nil {
			return "", err
		}
		return strings.Join(addrs, ","), nil
	case "controladdress":
		return l.APIAddr()
	}
	return "", fmt.Errorf("the libp2p daemon does not expose an attribute named \"%s\"", attr)
}

func (l *LocalP2pd) SetAttr(string, string) error {
	return fmt.Errorf("the libp2p daemon does not expose any writeable attributes")
}
