package plugindockeripfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	config "github.com/ipfs/go-ipfs-config"
	serial "github.com/ipfs/go-ipfs-config/serialize"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	cnet "github.com/whyrusleeping/go-ctrlnet"

	"github.com/ipfs/iptb-plugins"
	"github.com/ipfs/iptb/testbed/interfaces"
	"github.com/ipfs/iptb/util"
)

var ErrTimeout = errors.New("timeout")

var PluginName = "dockeripfs"

const (
	attrIfName    = "ifname"
	attrContainer = "container"
)

type DockerIpfs struct {
	image       string
	id          string
	dir         string
	repobuilder string
	peerid      *cid.Cid
	apiaddr     multiaddr.Multiaddr
	swarmaddr   multiaddr.Multiaddr
	mdns        bool
}

func NewNode(dir string, attrs map[string]string) (testbedi.Core, error) {
	imagename := "ipfs/go-ipfs"
	mdns := false

	apiaddr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/5001")
	if err != nil {
		return nil, err
	}

	swarmaddr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4001")
	if err != nil {
		return nil, err
	}

	var repobuilder string

	if v, ok := attrs["image"]; ok {
		imagename = v
	}

	if v, ok := attrs["repobuilder"]; ok {
		repobuilder = v
	} else {
		ipfspath, err := exec.LookPath("ipfs")
		if err != nil {
			return nil, fmt.Errorf("No `repobuilder` provided, could not find ipfs in path")
		}

		repobuilder = ipfspath
	}

	if apiaddrstr, ok := attrs["apiaddr"]; ok {
		var err error
		apiaddr, err = multiaddr.NewMultiaddr(apiaddrstr)

		if err != nil {
			return nil, err
		}
	}

	if swarmaddrstr, ok := attrs["swarmaddr"]; ok {
		var err error
		swarmaddr, err = multiaddr.NewMultiaddr(swarmaddrstr)

		if err != nil {
			return nil, err
		}
	}

	if _, ok := attrs["mdns"]; ok {
		mdns = true
	}

	return &DockerIpfs{
		dir:         dir,
		image:       imagename,
		repobuilder: repobuilder,
		apiaddr:     apiaddr,
		swarmaddr:   swarmaddr,
		mdns:        mdns,
	}, nil
}

func GetAttrList() []string {
	return append(ipfs.GetAttrList(), attrIfName, attrContainer)
}

func GetAttrDesc(attr string) (string, error) {
	switch attr {
	case attrIfName:
		return "docker ifname", nil
	case attrContainer:
		return "docker container id", nil
	}

	return ipfs.GetAttrDesc(attr)
}

func GetMetricList() []string {
	return ipfs.GetMetricList()
}

func GetMetricDesc(attr string) (string, error) {
	return ipfs.GetMetricDesc(attr)
}

/// Core Interface

func (l *DockerIpfs) Init(ctx context.Context, args ...string) (testbedi.Output, error) {
	env, err := l.env()
	if err != nil {
		return nil, fmt.Errorf("error getting env: %s", err)
	}

	cmd := exec.CommandContext(ctx, l.repobuilder, "init")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, string(out))
	}

	icfg, err := l.Config()
	if err != nil {
		return nil, err
	}

	lcfg, ok := icfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("Error: Config() is not an ipfs config")
	}

	lcfg.Bootstrap = nil
	lcfg.Addresses.Swarm = []string{l.swarmaddr.String()}
	lcfg.Addresses.API = []string{l.apiaddr.String()}
	lcfg.Addresses.Gateway = []string{""}
	lcfg.Discovery.MDNS.Enabled = l.mdns

	err = l.WriteConfig(lcfg)
	if err != nil {
		return nil, err
	}

	return nil, err
}

func (l *DockerIpfs) Start(ctx context.Context, wait bool, args ...string) (testbedi.Output, error) {
	alive, err := l.isAlive()
	if err != nil {
		return nil, err
	}

	if alive {
		return nil, fmt.Errorf("node is already running")
	}

	fargs := []string{"run", "-d", "-v", l.dir + ":/data/ipfs", l.image}
	if len(args) > 0 {
		fargs = append(fargs, "daemon")
	}
	fargs = append(fargs, args...)
	cmd := exec.Command("docker", fargs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, string(out))
	}

	id := bytes.TrimSpace(out)
	l.id = string(id)

	idfile := filepath.Join(l.dir, "dockerid")
	err = ioutil.WriteFile(idfile, id, 0664)

	if err != nil {
		killErr := l.killContainer()
		if killErr != nil {
			return nil, combineErrors(err, killErr)
		}
		return nil, err
	}

	if wait {
		return nil, ipfs.WaitOnAPI(l)
	}

	return nil, nil
}

func (l *DockerIpfs) Stop(ctx context.Context) error {
	err := l.killContainer()
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(l.dir, "dockerid"))
}

func (l *DockerIpfs) RunCmd(ctx context.Context, stdin io.Reader, args ...string) (testbedi.Output, error) {
	id, err := l.getID()
	if err != nil {
		return nil, err
	}

	if stdin != nil {
		args = append([]string{"exec", "-i", id}, args...)
	} else {
		args = append([]string{"exec", id}, args...)
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = stdin

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	stderrbytes, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, err
	}

	stdoutbytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	exiterr := cmd.Wait()

	var exitcode = 0
	switch oerr := exiterr.(type) {
	case *exec.ExitError:
		if ctx.Err() == context.DeadlineExceeded {
			err = errors.Wrapf(oerr, "context deadline exceeded for command: %q", strings.Join(cmd.Args, " "))
		}

		exitcode = 1
	case nil:
		err = oerr
	}

	return iptbutil.NewOutput(args, stdoutbytes, stderrbytes, exitcode, err), nil
}

func (l *DockerIpfs) Connect(ctx context.Context, n testbedi.Core) error {
	swarmaddrs, err := n.SwarmAddrs()
	if err != nil {
		return err
	}

	var output testbedi.Output
	var addr string
	for _, swarmaddr := range swarmaddrs {
		if strings.HasPrefix(swarmaddr, "/ip4/") && !strings.HasPrefix(swarmaddr, "/ip4/127.0.0.1/") {
			addr = swarmaddr
			break
		}
	}
	if addr == "" {
		return fmt.Errorf("could not find valid swarm address for peer %s", n.String())
	}
	output, err = l.RunCmd(ctx, nil, "ipfs", "swarm", "connect", addr)

	if err != nil {
		return err
	}

	if output.ExitCode() != 0 {
		out, err := ioutil.ReadAll(output.Stderr())
		if err != nil {
			return err
		}

		return fmt.Errorf("%s", string(out))
	}

	return nil
}

func (l *DockerIpfs) Shell(ctx context.Context, nodes []testbedi.Core) error {
	id, err := l.getID()
	if err != nil {
		return err
	}

	nenvs := []string{}
	for i, n := range nodes {
		peerid, err := n.PeerID()

		if err != nil {
			return err
		}

		nenvs = append(nenvs, fmt.Sprintf("NODE%d=%s", i, peerid))
	}

	args := []string{"exec", "-it"}
	for _, e := range nenvs {
		args = append(args, "-e", e)
	}

	args = append(args, id, "/bin/sh")
	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func (l *DockerIpfs) String() string {
	pcid, err := l.PeerID()
	if err != nil {
		return fmt.Sprintf("%s", l.Type())
	}
	return fmt.Sprintf("%s", pcid[0:12])
}

func (l *DockerIpfs) APIAddr() (string, error) {
	return ipfs.GetAPIAddrFromRepo(l.dir)
}

func (l *DockerIpfs) SwarmAddrs() ([]string, error) {
	return ipfs.SwarmAddrs(l)
}

func (l *DockerIpfs) Dir() string {
	return l.dir
}

func (l *DockerIpfs) PeerID() (string, error) {
	if l.peerid != nil {
		return l.peerid.String(), nil
	}

	var err error
	l.peerid, err = ipfs.GetPeerID(l)

	if err != nil {
		return "", err
	}

	return l.peerid.String(), nil
}

// Metric Interface

func (l *DockerIpfs) GetMetricList() []string {
	return GetMetricList()
}

func (l *DockerIpfs) GetMetricDesc(attr string) (string, error) {
	return GetMetricDesc(attr)
}

func (l *DockerIpfs) Metric(metric string) (string, error) {
	return ipfs.GetMetric(l, metric)
}

func (l *DockerIpfs) Heartbeat() (map[string]string, error) {
	return nil, nil
}

func (l *DockerIpfs) Events() (io.ReadCloser, error) {
	return ipfs.ReadLogs(l)
}

func (l *DockerIpfs) Logs() (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

// Attribute Interface

func (l *DockerIpfs) GetAttrList() []string {
	return GetAttrList()
}

func (l *DockerIpfs) GetAttrDesc(attr string) (string, error) {
	return GetAttrDesc(attr)
}

func (l *DockerIpfs) Attr(attr string) (string, error) {
	switch attr {
	case attrIfName:
		return l.getInterfaceName()
	case attrContainer:
		return l.getID()
	}

	return ipfs.GetAttr(l, attr)
}

func (l *DockerIpfs) SetAttr(attr string, val string) error {
	switch attr {
	case "latency":
		return l.setLatency(val)
	case "bandwidth":
		return l.setBandwidth(val)
	case "jitter":
		return l.setJitter(val)
	case "loss":
		return l.setPacketLoss(val)
	default:
		return fmt.Errorf("no attribute named: %s", attr)
	}
}

func (l *DockerIpfs) StderrReader() (io.ReadCloser, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (l *DockerIpfs) StdoutReader() (io.ReadCloser, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (l *DockerIpfs) Config() (interface{}, error) {
	return serial.Load(filepath.Join(l.dir, "config"))
}

func (l *DockerIpfs) WriteConfig(cfg interface{}) error {
	return serial.WriteConfigFile(filepath.Join(l.dir, "config"), cfg)
}

func (l *DockerIpfs) Type() string {
	return "ipfs"
}

func (l *DockerIpfs) Deployment() string {
	return "docker"
}

func (l *DockerIpfs) getID() (string, error) {
	if len(l.id) != 0 {
		return l.id, nil
	}

	b, err := ioutil.ReadFile(filepath.Join(l.dir, "dockerid"))
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (l *DockerIpfs) isAlive() (bool, error) {
	return false, nil
}

func (l *DockerIpfs) env() ([]string, error) {
	envs := os.Environ()
	ipfspath := "IPFS_PATH=" + l.dir

	for i, e := range envs {
		if strings.HasPrefix(e, "IPFS_PATH=") {
			envs[i] = ipfspath
			return envs, nil
		}
	}
	return append(envs, ipfspath), nil
}

func (l *DockerIpfs) killContainer() error {
	id, err := l.getID()
	if err != nil {
		return err
	}
	out, err := exec.Command("docker", "kill", "--signal=INT", id).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(out))
	}
	return nil
}

func (l *DockerIpfs) getInterfaceName() (string, error) {
	out, err := l.RunCmd(context.TODO(), nil, "ip", "link")
	if err != nil {
		return "", err
	}

	stdout, err := ioutil.ReadAll(out.Stdout())
	if err != nil {
		return "", err
	}

	var cside string
	for _, l := range strings.Split(string(stdout), "\n") {
		if strings.Contains(l, "@if") {
			ifnum := strings.Split(strings.Split(l, " ")[1], "@")[1]
			cside = ifnum[2 : len(ifnum)-1]
			break
		}
	}

	if cside == "" {
		return "", fmt.Errorf("container-side interface not found")
	}

	localout, err := exec.Command("ip", "link").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, localout)
	}

	for _, l := range strings.Split(string(localout), "\n") {
		if strings.HasPrefix(l, cside+": ") {
			return strings.Split(strings.Fields(l)[1], "@")[0], nil
		}
	}

	return "", fmt.Errorf("could not determine interface")
}

func (l *DockerIpfs) setLatency(val string) error {
	dur, err := time.ParseDuration(val)
	if err != nil {
		return err
	}

	ifn, err := l.getInterfaceName()
	if err != nil {
		return err
	}

	settings := &cnet.LinkSettings{
		Latency: uint(dur.Nanoseconds() / 1000000),
	}

	return cnet.SetLink(ifn, settings)
}

func (l *DockerIpfs) setJitter(val string) error {
	dur, err := time.ParseDuration(val)
	if err != nil {
		return err
	}

	ifn, err := l.getInterfaceName()
	if err != nil {
		return err
	}

	settings := &cnet.LinkSettings{
		Jitter: uint(dur.Nanoseconds() / 1000000),
	}

	return cnet.SetLink(ifn, settings)
}

// set bandwidth (expects Mbps)
func (l *DockerIpfs) setBandwidth(val string) error {
	bw, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return err
	}

	ifn, err := l.getInterfaceName()
	if err != nil {
		return err
	}

	settings := &cnet.LinkSettings{
		Bandwidth: uint(bw * 1000000),
	}

	return cnet.SetLink(ifn, settings)
}

// set packet loss percentage (dropped / total)
func (l *DockerIpfs) setPacketLoss(val string) error {
	ratio, err := strconv.ParseUint(val, 10, 8)
	if err != nil {
		return err
	}

	ifn, err := l.getInterfaceName()
	if err != nil {
		return err
	}

	settings := &cnet.LinkSettings{
		PacketLoss: uint8(ratio),
	}

	return cnet.SetLink(ifn, settings)
}

func combineErrors(err1, err2 error) error {
	return fmt.Errorf("%v\nwhile handling the above error, the following error occurred:\n%v", err1, err2)
}
