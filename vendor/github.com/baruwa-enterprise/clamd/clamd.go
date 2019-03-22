// Copyright (C) 2018 Andrew Colin Kissa <andrew@datopdog.io>
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package clamd Golang Clamd client
Clamd - Golang clamd client
*/
package clamd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/baruwa-enterprise/clamd/protocol"
)

const (
	defaultTimeout      = 15 * time.Second
	defaultSleep        = 1 * time.Second
	defaultCmdTimeout   = 1 * time.Minute
	defaultSock         = "/var/run/clamav/clamd.sock"
	invalidRespErr      = "Invalid server response: %s"
	unsupportedProtoErr = "Protocol: %s is not supported"
	unixSockErr         = "The unix socket: %s does not exist"
	statsErr            = "Stats not returned"
	versionErr          = "Version not returned"
	fldesErr            = "Fildes can not be called on a non unix connection"
	reloadResp          = "RELOADING"
	pingResp            = "PONG"
	versionCmdsResp     = "COMMANDS: "
	// ChunkSize the size for chunking INSTREAM files
	ChunkSize = 1024
)

var (
	responseRe = regexp.MustCompile(`^(?P<filename>[^:]+):\s+(?:(?P<signature>[^:]+)\s+)?(?P<status>(FOUND|OK|ERROR))$`)
)

// Response is the response from the server
type Response struct {
	Filename  string
	Signature string
	Status    string
	Raw       string
}

// A Client represents a Clamd client.
type Client struct {
	network     string
	address     string
	connTimeout time.Duration
	connRetries int
	connSleep   time.Duration
	cmdTimeout  time.Duration
}

// SetConnTimeout sets the connection timeout
func (c *Client) SetConnTimeout(t time.Duration) {
	c.connTimeout = t
}

// SetCmdTimeout sets the cmd timeout
func (c *Client) SetCmdTimeout(t time.Duration) {
	c.cmdTimeout = t
}

// SetConnRetries sets the number of times
// connection is retried
func (c *Client) SetConnRetries(s int) {
	if s < 0 {
		s = 0
	}
	c.connRetries = s
}

// SetConnSleep sets the connection retry sleep
// duration in seconds
func (c *Client) SetConnSleep(s time.Duration) {
	c.connSleep = s
}

// Ping sends a ping to the server
func (c *Client) Ping() (b bool, err error) {
	var r string
	if r, err = c.basicCmd(protocol.Ping); err != nil {
		return
	}

	if err = checkError(r); err != nil {
		return
	}

	b = r == pingResp

	return
}

// Version returns the server version
func (c *Client) Version() (v string, err error) {
	if v, err = c.basicCmd(protocol.Version); err != nil {
		return
	}

	if v == "" && err == nil {
		err = fmt.Errorf(versionErr)
		return
	}

	if err = checkError(v); err != nil {
		return
	}

	return
}

// Reload the server
func (c *Client) Reload() (b bool, err error) {
	var r string
	if r, err = c.basicCmd(protocol.Reload); err != nil {
		return
	}

	if err = checkError(r); err != nil {
		return
	}

	b = r == reloadResp

	return
}

// Shutdown stops the server
func (c *Client) Shutdown() (err error) {
	_, err = c.basicCmd(protocol.Shutdown)
	return
}

// Scan a file or directory
func (c *Client) Scan(p string) (r []*Response, err error) {
	r, err = c.fileCmd(protocol.Scan, p)
	return
}

// ScanReader scans an io.reader
func (c *Client) ScanReader(i io.Reader) (r []*Response, err error) {
	r, err = c.readerCmd(i)
	return
}

// ContScan a file or directory
func (c *Client) ContScan(p string) (r []*Response, err error) {
	r, err = c.fileCmd(protocol.ContScan, p)
	return
}

// MultiScan a file or directory
func (c *Client) MultiScan(p string) (r []*Response, err error) {
	r, err = c.fileCmd(protocol.MultiScan, p)
	return
}

// InStream scan a stream
func (c *Client) InStream(p string) (r []*Response, err error) {
	r, err = c.fileCmd(protocol.Instream, p)
	return
}

// Fildes scan a FD
func (c *Client) Fildes(p string) (r []*Response, err error) {
	r, err = c.fileCmd(protocol.Fildes, p)
	return
}

// Stats returns server stats
func (c *Client) Stats() (s string, err error) {
	if s, err = c.basicCmd(protocol.Stats); err != nil {
		return
	}

	if s == "" && err == nil {
		err = fmt.Errorf(statsErr)
		return
	}

	if err = checkError(s); err != nil {
		return
	}

	return
}

// IDSession starts a session
// func (c *Client) IDSession() {
// }

// End closes a session
// func (c *Client) End() {
// }

// VersionCmds returns the supported cmds
func (c *Client) VersionCmds() (r []string, err error) {
	var s string
	if s, err = c.basicCmd(protocol.VersionCmds); err != nil {
		return
	}

	p := strings.Split(s, versionCmdsResp)
	if len(p) != 2 {
		err = fmt.Errorf(invalidRespErr, s)
		return
	}
	s = p[1]
	r = strings.Split(s, " ")
	return
}

func (c *Client) dial() (conn net.Conn, err error) {
	d := &net.Dialer{
		Timeout: c.connTimeout,
	}

	for i := 0; i <= c.connRetries; i++ {
		conn, err = d.Dial(c.network, c.address)
		if e, ok := err.(net.Error); ok && e.Timeout() {
			time.Sleep(c.connSleep)
			continue
		}
		break
	}
	return
}

func (c *Client) basicCmd(cmd protocol.Command) (r string, err error) {
	var conn net.Conn
	var l []byte
	var b strings.Builder
	var tc *textproto.Conn

	conn, err = c.dial()
	if err != nil {
		return
	}

	tc = textproto.NewConn(conn)
	defer tc.Close()

	id := tc.Next()
	tc.StartRequest(id)
	conn.SetDeadline(time.Now().Add(c.cmdTimeout))
	fmt.Fprintf(tc.W, "n%s\n", cmd)
	tc.W.Flush()
	tc.EndRequest(id)

	tc.StartResponse(id)
	defer tc.EndResponse(id)

	if cmd == protocol.Shutdown {
		return
	}

	for {
		conn.SetDeadline(time.Now().Add(c.cmdTimeout))
		if l, err = tc.R.ReadBytes('\n'); err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		fmt.Fprintf(&b, "%s", l)
	}

	r = strings.TrimRight(b.String(), "\n")

	return
}

func (c *Client) fileCmd(cmd protocol.Command, p string) (r []*Response, err error) {
	var id uint
	var conn net.Conn
	var tc *textproto.Conn

	if cmd == protocol.Instream || cmd == protocol.Fildes {
		if _, err = os.Stat(p); os.IsNotExist(err) {
			return
		}
	}

	if cmd == protocol.Fildes && c.network != "unix" && c.network != "unixpacket" {
		err = fmt.Errorf(fldesErr)
		return
	}

	conn, err = c.dial()
	if err != nil {
		return
	}

	tc = textproto.NewConn(conn)
	defer tc.Close()

	id = tc.Next()
	tc.StartRequest(id)

	conn.SetDeadline(time.Now().Add(c.cmdTimeout))
	if cmd == protocol.Instream {
		if err = c.instreamScan(tc, conn, p); err != nil {
			tc.EndRequest(id)
			return
		}
	} else if cmd == protocol.Fildes {
		if err = c.fildesScan(tc, conn, p); err != nil {
			tc.EndRequest(id)
			return
		}
	} else {
		fmt.Fprintf(tc.W, "n%s %s\n", cmd, p)
	}
	tc.W.Flush()
	tc.EndRequest(id)

	tc.StartResponse(id)
	defer tc.EndResponse(id)

	r, err = c.processResponse(tc, conn)

	return
}

func (c *Client) readerCmd(i io.Reader) (r []*Response, err error) {
	var conn net.Conn
	var tc *textproto.Conn

	if conn, err = c.dial(); err != nil {
		return
	}

	tc = textproto.NewConn(conn)
	defer tc.Close()

	id := tc.Next()
	tc.StartRequest(id)

	if err = c.streamCmd(tc, protocol.Instream, i, conn); err != nil {
		tc.EndRequest(id)
		return
	}

	tc.EndRequest(id)

	tc.StartResponse(id)
	defer tc.EndResponse(id)

	r, err = c.processResponse(tc, conn)

	return
}

func (c *Client) streamCmd(tc *textproto.Conn, cmd protocol.Command, f io.Reader, conn net.Conn) (err error) {
	var n int

	fmt.Fprintf(tc.W, "n%s\n", cmd)
	b := make([]byte, 4)
	for {
		buf := make([]byte, ChunkSize)
		if n, err = f.Read(buf); err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}
		if n > 0 {
			conn.SetDeadline(time.Now().Add(c.cmdTimeout))
			binary.BigEndian.PutUint32(b, uint32(n))
			if _, err = tc.W.Write(b); err != nil {
				return
			}
			if _, err = tc.W.Write(buf[0:n]); err != nil {
				return
			}
			tc.W.Flush()
		}
	}
	if _, err = tc.W.Write([]byte{0, 0, 0, 0}); err != nil {
		return
	}
	tc.W.Flush()

	return
}

func (c *Client) processResponse(tc *textproto.Conn, conn net.Conn) (r []*Response, err error) {
	var lineb []byte

	for {
		conn.SetDeadline(time.Now().Add(c.cmdTimeout))
		rs := Response{}
		if lineb, err = tc.R.ReadBytes('\n'); err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		mb := responseRe.FindSubmatch(bytes.TrimRight(lineb, "\n"))
		if mb == nil {
			if bytes.HasSuffix(lineb, []byte("ERROR\n")) {
				err = fmt.Errorf("%s", bytes.TrimRight(lineb, " ERROR\n"))
			} else {
				err = fmt.Errorf(invalidRespErr, lineb)
			}
			break
		}

		rs.Filename = string(mb[1])
		rs.Signature = string(mb[2])
		rs.Status = string(mb[3])
		rs.Raw = string(mb[0])

		r = append(r, &rs)
	}

	return
}

func (c *Client) instreamScan(tc *textproto.Conn, conn net.Conn, p string) (err error) {
	var f *os.File

	if f, err = os.Open(p); err != nil {
		return
	}
	defer f.Close()

	if err = c.streamCmd(tc, protocol.Instream, f, conn); err != nil {
		return
	}

	return
}

func (c *Client) fildesScan(tc *textproto.Conn, conn net.Conn, p string) (err error) {
	var f *os.File
	var vf *os.File

	fmt.Fprintf(tc.W, "n%s\n", protocol.Fildes)
	tc.W.Flush()

	if f, err = os.Open(p); err != nil {
		return
	}
	defer f.Close()

	s := conn.(*net.UnixConn)
	if vf, err = s.File(); err != nil {
		return
	}
	sock := int(vf.Fd())
	defer vf.Close()

	fds := []int{int(f.Fd())}
	rights := syscall.UnixRights(fds...)
	if err = syscall.Sendmsg(sock, nil, rights, nil, 0); err != nil {
		return
	}

	return
}

// NewClient returns a new Clamd client.
func NewClient(network, address string) (c *Client, err error) {
	if network == "" && address == "" {
		network = "unix"
		address = defaultSock
	}

	if network != "unix" && network != "unixpacket" && network != "tcp" && network != "tcp4" && network != "tcp6" {
		err = fmt.Errorf(unsupportedProtoErr, network)
		return
	}

	if network == "unix" || network == "unixpacket" {
		if _, err = os.Stat(address); os.IsNotExist(err) {
			err = fmt.Errorf(unixSockErr, address)
			return
		}
	}

	c = &Client{
		network:     network,
		address:     address,
		connTimeout: defaultTimeout,
		connSleep:   defaultSleep,
		cmdTimeout:  defaultCmdTimeout,
	}
	return
}

func checkError(s string) (err error) {
	if strings.HasSuffix(s, "ERROR") {
		err = fmt.Errorf("%s", strings.TrimRight(s, " ERROR"))
		return
	}

	return
}
