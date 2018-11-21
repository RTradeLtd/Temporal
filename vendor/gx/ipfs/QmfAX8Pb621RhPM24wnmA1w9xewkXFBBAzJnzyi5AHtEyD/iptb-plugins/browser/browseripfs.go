package pluginbrowseripfs

import (
	"context"
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

	"github.com/ipfs/iptb-plugins"
	"github.com/ipfs/iptb/testbed/interfaces"
	"github.com/ipfs/iptb/util"

	"github.com/ipfs/go-cid"
	config "github.com/ipfs/go-ipfs-config"
	serial "github.com/ipfs/go-ipfs-config/serialize"
	"github.com/pkg/errors"
)

var errTimeout = errors.New("timeout")

var PluginName = "browseripfs"

type BrowserIpfs struct {
	dir         string
	peerid      *cid.Cid
	repobuilder string
	apiaddr     string
	swarmaddr   string
	source      string
}

func NewNode(dir string, attrs map[string]string) (testbedi.Core, error) {
	if _, err := exec.LookPath("ipfs"); err != nil {
		return nil, err
	}

	if _, err := exec.LookPath("node"); err != nil {
		return nil, err
	}

	apiaddr := "/ip4/127.0.0.1/tcp/0"
	swarmaddr := "/ip4/127.0.0.1/tcp/0"

	if apiaddrstr, ok := attrs["apiaddr"]; ok {
		apiaddr = apiaddrstr
	}

	if swarmaddrstr, ok := attrs["swarmaddr"]; ok {
		swarmaddr = swarmaddrstr
	}

	// repobuilder is the binary used to run `Init`, a browser ipfs node cannot
	// do this itself. The repo is largely ignore, but the configuration we want
	var repobuilder string
	if v, ok := attrs["repobuilder"]; ok {
		repobuilder = v
	} else {
		jsipfspath, err := exec.LookPath("jsipfs")
		if err != nil {
			return nil, fmt.Errorf("No `repobuilder` provided, could not find jsipfs in path")
		}

		repobuilder = jsipfspath
	}

	// source is any js (at the moment) script which can read the ipfs repo and expose an ipfs-api
	// An implementation can be found @ https://github.com/travisperson/js-ipfs-browser-server
	var source string
	if v, ok := attrs["source"]; ok {
		source = v
	} else {
		return nil, fmt.Errorf("No `source` provided")
	}

	return &BrowserIpfs{
		dir:         dir,
		apiaddr:     apiaddr,
		swarmaddr:   swarmaddr,
		repobuilder: repobuilder,
		source:      source,
	}, nil

}

/// TestbedNode Interface

func (l *BrowserIpfs) Init(ctx context.Context, agrs ...string) (testbedi.Output, error) {
	agrs = append([]string{l.repobuilder, "init"}, agrs...)
	output, oerr := l.RunCmd(ctx, nil, agrs...)
	if oerr != nil {
		return nil, oerr
	}

	icfg, err := l.Config()
	if err != nil {
		return nil, err
	}

	lcfg, ok := icfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("Error: Config() is not an ipfs config")
	}

	// jsipfs does not like this value being nil, so it needs to be set to an empty array
	lcfg.Bootstrap = []string{}
	lcfg.Addresses.Swarm = []string{l.swarmaddr}
	lcfg.Addresses.API = []string{l.apiaddr}
	lcfg.Addresses.Gateway = []string{""}
	lcfg.Discovery.MDNS.Enabled = false

	if err = l.WriteConfig(lcfg); err != nil {
		return nil, err
	}

	return output, oerr
}

func (l *BrowserIpfs) Start(ctx context.Context, wait bool, args ...string) (testbedi.Output, error) {
	var err error

	dir := l.dir
	cmd := exec.Command("node", l.source)
	cmd.Dir = dir

	cmd.Env, err = l.env()
	if err != nil {
		return nil, err
	}

	stdout, err := os.Create(filepath.Join(dir, "daemon.stdout"))
	if err != nil {
		return nil, err
	}

	stderr, err := os.Create(filepath.Join(dir, "daemon.stderr"))
	if err != nil {
		return nil, err
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	pid := cmd.Process.Pid

	if err = ioutil.WriteFile(filepath.Join(dir, "daemon.pid"), []byte(fmt.Sprint(pid)), 0666); err != nil {
		return nil, err
	}

	if wait {
		return nil, ipfs.WaitOnAPI(l)
	}

	return nil, nil
}

func (l *BrowserIpfs) Stop(ctx context.Context) error {
	pid, err := l.getPID()
	if err != nil {
		return fmt.Errorf("error killing daemon %s: %s", l.dir, err)
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("error killing daemon %s: %s", l.dir, err)
	}

	waitch := make(chan struct{}, 1)
	go func() {
		p.Wait()
		waitch <- struct{}{}
	}()

	defer func() {
		err := os.Remove(filepath.Join(l.dir, "daemon.pid"))
		if err != nil && !os.IsNotExist(err) {
			panic(fmt.Errorf("error removing pid file for daemon at %s: %s", l.dir, err))
		}
	}()

	if err := l.signalAndWait(p, waitch, syscall.SIGINT, 1*time.Second); err != errTimeout {
		return err
	}

	if err := l.signalAndWait(p, waitch, syscall.SIGTERM, 2*time.Second); err != errTimeout {
		return err
	}

	if err := l.signalAndWait(p, waitch, syscall.SIGQUIT, 5*time.Second); err != errTimeout {
		return err
	}

	if err := l.signalAndWait(p, waitch, syscall.SIGKILL, 5*time.Second); err != errTimeout {
		return err
	}

	return fmt.Errorf("Could not stop browseripfs node with pid %d", pid)
}

func (l *BrowserIpfs) RunCmd(ctx context.Context, stdin io.Reader, args ...string) (testbedi.Output, error) {
	env, err := l.env()
	if err != nil {
		return nil, fmt.Errorf("error getting env: %s", err)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdin = stdin

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
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

func (l *BrowserIpfs) Connect(ctx context.Context, tbn testbedi.Core) error {
	swarmaddrs, err := tbn.SwarmAddrs()
	if err != nil {
		return err
	}

	for _, addr := range swarmaddrs {
		output, err := l.RunCmd(ctx, nil, "ipfs", "swarm", "connect", addr)
		if err != nil {
			return err
		}

		if output.ExitCode() == 0 {
			return nil
		}
	}

	return fmt.Errorf("Could not connect using any address")
}

func (l *BrowserIpfs) Shell(ctx context.Context, nodes []testbedi.Core) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return fmt.Errorf("no shell found")
	}

	if len(os.Getenv("IPFS_PATH")) != 0 {
		// If the users shell sets IPFS_PATH, it will just be overridden by the shell again
		return fmt.Errorf("shell has IPFS_PATH set, please unset before trying to use iptb shell")
	}

	nenvs, err := l.env()
	if err != nil {
		return err
	}

	// TODO(tperson): It would be great if we could guarantee that the shell
	// is using the same binary. However, the users shell may prepend anything
	// we change in the PATH

	for i, n := range nodes {
		peerid, err := n.PeerID()
		if err != nil {
			return err
		}

		nenvs = append(nenvs, fmt.Sprintf("NODE%d=%s", i, peerid))
	}

	return syscall.Exec(shell, []string{shell}, nenvs)
}

func (l *BrowserIpfs) String() string {
	pcid, err := l.PeerID()
	if err != nil {
		return fmt.Sprintf("%s", l.Type())
	}
	return fmt.Sprintf("%s", pcid[0:12])
}

func (l *BrowserIpfs) APIAddr() (string, error) {
	return ipfs.GetAPIAddrFromRepo(l.dir)
}

func (l *BrowserIpfs) SwarmAddrs() ([]string, error) {
	return ipfs.SwarmAddrs(l)
}

func (l *BrowserIpfs) Dir() string {
	return l.dir
}

func (l *BrowserIpfs) PeerID() (string, error) {
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

func (l *BrowserIpfs) Config() (interface{}, error) {
	return serial.Load(filepath.Join(l.dir, "config"))
}

func (l *BrowserIpfs) WriteConfig(cfg interface{}) error {
	return serial.WriteConfigFile(filepath.Join(l.dir, "config"), cfg)
}

func (l *BrowserIpfs) Type() string {
	return PluginName
}

func (l *BrowserIpfs) signalAndWait(p *os.Process, waitch <-chan struct{}, signal os.Signal, t time.Duration) error {
	if err := p.Signal(signal); err != nil {
		return errors.Wrapf(err, "error killing daemon %s", l.dir)
	}

	select {
	case <-waitch:
		return nil
	case <-time.After(t):
		return errTimeout
	}
}

func (l *BrowserIpfs) getPID() (int, error) {
	b, err := ioutil.ReadFile(filepath.Join(l.dir, "daemon.pid"))
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(string(b))
}

func (l *BrowserIpfs) env() ([]string, error) {
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
