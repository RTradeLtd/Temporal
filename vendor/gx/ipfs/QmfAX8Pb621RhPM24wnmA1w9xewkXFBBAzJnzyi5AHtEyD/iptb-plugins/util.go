package ipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	config "github.com/ipfs/go-ipfs-config"
	_ "github.com/libp2p/go-libp2p"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"

	"github.com/ipfs/iptb/testbed/interfaces"
)

const (
	attrID   = "id"
	attrPath = "path"

	metricBwIn  = "bw_in"
	metricBwOut = "bw_out"
)

func InitIpfs(l testbedi.Core) error {
	return nil
}

func GetAttr(l testbedi.Core, attr string) (string, error) {
	switch attr {
	case attrID:
		pcid, err := l.PeerID()
		if err != nil {
			return "", err
		}
		return pcid, nil
	case attrPath:
		return l.Dir(), nil
	default:
		return "", errors.New("unrecognized attribute: " + attr)
	}
}

func GetMetric(l testbedi.Core, metric string) (string, error) {
	switch metric {
	case metricBwIn:
		bw, err := GetBW(l)
		if err != nil {
			return "", err
		}
		return fmt.Sprint(bw.TotalIn), nil
	case metricBwOut:
		bw, err := GetBW(l)
		if err != nil {
			return "", err
		}
		return fmt.Sprint(bw.TotalOut), nil
	default:
		return "", errors.New("unrecognized metric: " + metric)
	}
}

func GetPeerID(l testbedi.Config) (*cid.Cid, error) {
	icfg, err := l.Config()
	if err != nil {
		return nil, err
	}

	lcfg, ok := icfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("Error: GetConfig() is not an ipfs config")
	}

	pcid, err := cid.Decode(lcfg.Identity.PeerID)
	if err != nil {
		return nil, err
	}

	return &pcid, nil
}

func GetMetricList() []string {
	return []string{metricBwIn, metricBwOut}
}

func GetMetricDesc(metric string) (string, error) {
	switch metric {
	case metricBwIn:
		return "node input bandwidth", nil
	case metricBwOut:
		return "node output bandwidth", nil
	default:
		return "", errors.New("unrecognized metric")
	}
}

func GetAttrList() []string {
	return []string{attrID, attrPath}
}

func GetAttrDesc(attr string) (string, error) {
	switch attr {
	case attrID:
		return "node ID", nil
	case attrPath:
		return "node IPFS_PATH", nil
	default:
		return "", errors.New("unrecognized attribute")
	}
}

func ReadLogs(l testbedi.Libp2p) (io.ReadCloser, error) {
	addrStr, err := l.APIAddr()
	if err != nil {
		return nil, err
	}

	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return nil, err
	}

	//TODO(tperson) ipv6
	ip, err := addr.ValueForProtocol(multiaddr.P_IP4)
	if err != nil {
		return nil, err
	}
	pt, err := addr.ValueForProtocol(multiaddr.P_TCP)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%s/api/v0/log/tail", ip, pt))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

type BW struct {
	TotalIn  int
	TotalOut int
}

func GetBW(l testbedi.Libp2p) (*BW, error) {
	addrStr, err := l.APIAddr()
	if err != nil {
		return nil, err
	}

	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return nil, err
	}

	//TODO(tperson) ipv6
	ip, err := addr.ValueForProtocol(multiaddr.P_IP4)
	if err != nil {
		return nil, err
	}
	pt, err := addr.ValueForProtocol(multiaddr.P_TCP)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%s/api/v0/stats/bw", ip, pt))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bw BW
	err = json.NewDecoder(resp.Body).Decode(&bw)
	if err != nil {
		return nil, err
	}

	return &bw, nil
}

func GetAPIAddrFromRepo(dir string) (string, error) {
	out, err := ioutil.ReadFile(filepath.Join(dir, "api"))
	return string(out), err
}

func SwarmAddrs(l testbedi.Core) ([]string, error) {
	pcid, err := l.PeerID()
	if err != nil {
		return nil, err
	}

	output, err := l.RunCmd(context.TODO(), nil, "ipfs", "swarm", "addrs", "local")
	if err != nil {
		return nil, err
	}

	bs, err := ioutil.ReadAll(output.Stdout())
	if err != nil {
		return nil, err
	}

	straddrs := strings.Split(string(bs), "\n")

	var maddrs []string
	for _, straddr := range straddrs {
		if !strings.Contains(straddr, pcid) {
			fstraddr := fmt.Sprintf("%s/ipfs/%s", straddr, pcid)
			maddrs = append(maddrs, fstraddr)
		} else {
			maddrs = append(maddrs, straddr)
		}
	}

	return maddrs, nil
}

func WaitOnAPI(l testbedi.Libp2p) error {
	for i := 0; i < 50; i++ {
		err := tryAPICheck(l)
		if err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * 400)
	}

	pcid, err := l.PeerID()
	if err != nil {
		return err
	}

	return fmt.Errorf("node %s failed to come online in given time period", pcid)
}

func tryAPICheck(l testbedi.Libp2p) error {
	addrStr, err := l.APIAddr()
	if err != nil {
		return err
	}

	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return err
	}

	//TODO(tperson) ipv6
	ip, err := addr.ValueForProtocol(multiaddr.P_IP4)
	if err != nil {
		return err
	}
	pt, err := addr.ValueForProtocol(multiaddr.P_TCP)
	if err != nil {
		return err
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%s/api/v0/id", ip, pt))
	if err != nil {
		return err
	}

	out := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return fmt.Errorf("liveness check failed: %s", err)
	}

	id, ok := out["ID"]
	if !ok {
		return fmt.Errorf("liveness check failed: ID field not present in output")
	}

	pcid, err := l.PeerID()
	if err != nil {
		return err
	}

	idstr, ok := id.(string)
	if !ok {
		return fmt.Errorf("liveness check failed: ID field is unexpected type")
	}

	if idstr != pcid {
		return fmt.Errorf("liveness check failed: unexpected peer at endpoint")
	}

	return nil
}
