// Copyright (C) 2018 Andrew Colin Kissa <andrew@datopdog.io>
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package protocol Golang Clamd client
Clamd - Golang clamd client
*/
package protocol

/*
Clamd Protocol

PING - Check the server's state. It should reply with "PONG".
VERSION - Print program and database versions.
RELOAD - Reload the virus databases.
SHUTDOWN - Perform a clean exit.
SCAN file/directory - Scan a file or a directory (recursively)
	with archive support enabled
CONTSCAN file/directory - Scan file or directory (recursively)
	with archive support enabled and don't stop the scanning
	when a virus is found.
MULTISCAN file/directory - Scan file in a standard way or scan
	directory (recursively) using multiple threads
INSTREAM - It is mandatory to prefix this command with n or z.
	Scan a stream of data. The stream is sent to clamd in chunks, after INSTREAM,
	on the same socket on which the command was sent. This avoids the overhead of
	establishing new TCP connections and problems with NAT. The format of the chunk is:
	'<length><data>' where <length> is the size of the following data in bytes expressed
	as a 4 byte unsigned integer in network byte order and <data> is the actual chunk.
	Streaming is terminated by sending a zero-length chunk. Note: do not exceed
	StreamMaxLength as defined in clamd.conf, otherwise clamd will reply with
	INSTREAM size limit exceeded and close the connection.
FILDES - It is mandatory to newline terminate this command, or prefix with n or z.
	This command only works on UNIX domain sockets. Scan a file descriptor.
	After issuing a FILDES command a subsequent rfc2292/bsd4.4 style packet
	(with at least one dummy character) is sent to clamd carrying the file
	descriptor to be scanned inside the ancillary data. Alternatively the
	file descriptor may be sent in the same packet, including the extra
	character.
STATS - It is mandatory to newline terminate this command, or prefix with n or z,
	it is recommended to only use the z prefix. Replies with statistics about the scan
	queue, contents of scan queue, and memory usage. The exact reply format is subject
	to change in future releases.
IDSESSION, END - It is mandatory to prefix this command with n or z, and all commands
	inside IDSESSION must be prefixed. Start/end a clamd session. Within a session multiple
	SCAN, INSTREAM, FILDES, VERSION, STATS commands can be sent on the same socket without
	opening new connections. Replies from clamd will be in the form '<id>: <response>' where
	<id> is the request number (in ascii, starting from 1) and <response> is the usual clamd
	reply. The reply lines have same delimiter as the corresponding command had. Clamd will
	process the commands asynchronously, and reply as soon as it has finished processing.
	Clamd requires clients to read all the replies it sent, before sending more commands to
	prevent send() deadlocks. The recommended way to implement a client that uses IDSESSION
	is with non-blocking sockets, and a select()/poll() loop: whenever send would block,
	sleep in select/poll until either you can write more data, or read more replies. Note
	that using non-blocking sockets without the select/poll loop and alternating recv()/send()
	doesn't comply with clamd's requirements. If clamd detects that a client has deadlocked,
	it will close the connection. Note that clamd may close an IDSESSION connection too if
	you don't follow the protocol's requirements. The client can use the PING command to
	keep the connection alive.
VERSIONCOMMANDS - It is mandatory to prefix this command with either n or z. It is
	recommended to use nVERSIONCOMMANDS. Print program and database versions, followed
	by "| COMMANDS:" and a space-delimited list of supported commands. Clamd <0.95 will
	recognize this as the VERSION command, and reply only with their version, without
	the commands list. This command can be used as an easy way to check for IDSESSION
	support for example.
*/

const (
	// Ping is the PING command
	Ping Command = iota + 1
	// Version is the VERSION command
	Version
	// Reload is the RELOAD command
	Reload
	// Shutdown is the SHUTDOWN command
	Shutdown
	// Scan is the SCAN command
	Scan
	// ContScan is the CONTSCAN command
	ContScan
	// MultiScan is the MULTISCAN command
	MultiScan
	// Instream is the INSTREAM command
	Instream
	// Fildes is the FILDES command
	Fildes
	// Stats is the STATS command
	Stats
	// IDSession is the IDSESSION command
	IDSession
	// EndSession is the END command
	EndSession
	// VersionCmds is the VERSIONCOMMANDS command
	VersionCmds
)

// A Command represents a Clamd Command
type Command int

func (c Command) String() (s string) {
	n := [...]string{
		"",
		"PING",
		"VERSION",
		"RELOAD",
		"SHUTDOWN",
		"SCAN",
		"CONTSCAN",
		"MULTISCAN",
		"INSTREAM",
		"FILDES",
		"STATS",
		"IDSESSION",
		"END",
		"VERSIONCOMMANDS",
	}
	if c < Ping || c > VersionCmds {
		s = ""
		return
	}
	s = n[c]
	return
}

// RequiresParam returns a bool to indicate if command takes a
// file or directory as a param
func (c Command) RequiresParam() (b bool) {
	switch c {
	case Scan, ContScan, MultiScan, Instream, Fildes:
		b = true
	}
	return
}
