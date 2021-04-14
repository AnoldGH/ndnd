/* YaNFD - Yet another NDN Forwarding Daemon
 *
 * Copyright (C) 2020-2021 Eric Newberry.
 *
 * This file is licensed under the terms of the MIT License, as found in LICENSE.md.
 */

package face

import (
	"net"
	"strconv"

	"github.com/eric135/YaNFD/core"
	"github.com/eric135/YaNFD/face/impl"
	"github.com/eric135/YaNFD/ndn"
	"github.com/eric135/YaNFD/ndn/tlv"
)

// UnixStreamTransport is a Unix stream transport for communicating with local applications.
type UnixStreamTransport struct {
	conn *net.UnixConn
	transportBase
}

// MakeUnixStreamTransport creates a Unix stream transport.
func MakeUnixStreamTransport(remoteURI *ndn.URI, localURI *ndn.URI, conn net.Conn) (*UnixStreamTransport, error) {
	// Validate URIs
	if !remoteURI.IsCanonical() || remoteURI.Scheme() != "fd" || !localURI.IsCanonical() || localURI.Scheme() != "unix" {
		return nil, core.ErrNotCanonical
	}

	t := new(UnixStreamTransport)
	t.makeTransportBase(remoteURI, localURI, PersistencyPersistent, ndn.Local, ndn.PointToPoint, tlv.MaxNDNPacketSize)

	// Set connection
	t.conn = conn.(*net.UnixConn)

	t.changeState(ndn.Up)

	return t, nil
}

func (t *UnixStreamTransport) String() string {
	return "UnixStreamTransport, FaceID=" + strconv.FormatUint(t.faceID, 10) + ", RemoteURI=" + t.remoteURI.String() + ", LocalURI=" + t.localURI.String()
}

// SetPersistency changes the persistency of the face.
func (t *UnixStreamTransport) SetPersistency(persistency Persistency) bool {
	if persistency == t.persistency {
		return true
	}

	if persistency == PersistencyPersistent {
		t.persistency = persistency
		return true
	}

	return false
}

// GetSendQueueSize returns the current size of the send queue.
func (t *UnixStreamTransport) GetSendQueueSize() uint64 {
	rawConn, err := t.conn.SyscallConn()
	if err != nil {
		core.LogWarn(t, "Unable to get raw connection to get socket length: "+err.Error())
	}
	return impl.SyscallGetSocketSendQueueSize(rawConn)
}

func (t *UnixStreamTransport) sendFrame(frame []byte) {
	if len(frame) > t.MTU() {
		core.LogWarn(t, "Attempted to send frame larger than MTU - DROP")
		return
	}

	core.LogDebug(t, "Sending frame of size "+strconv.Itoa(len(frame)))
	_, err := t.conn.Write(frame)
	if err != nil {
		core.LogWarn(t, "Unable to send on socket - DROP and Face DOWN")
		t.changeState(ndn.Down)
	}

	t.nOutBytes += uint64(len(frame))
}

func (t *UnixStreamTransport) runReceive() {
	core.LogTrace(t, "Starting receive thread")
	recvBuf := make([]byte, tlv.MaxNDNPacketSize)
	for {
		core.LogTrace(t, "Reading from socket")
		readSize, err := t.conn.Read(recvBuf)
		if err != nil {
			if err.Error() == "EOF" {
				core.LogDebug(t, "EOF - Face DOWN")
			} else {
				core.LogWarn(t, "Unable to read from socket ("+err.Error()+") - DROP and Face DOWN")
			}
			t.changeState(ndn.Down)
			break
		}

		core.LogTrace(t, "Receive of size "+strconv.Itoa(readSize))
		t.nInBytes += uint64(readSize)

		// Send up to link service
		t.linkService.handleIncomingFrame(recvBuf)
	}
}

func (t *UnixStreamTransport) changeState(new ndn.State) {
	if t.state == new {
		return
	}

	core.LogInfo(t, "state: "+t.state.String()+" -> "+new.String())
	t.state = new

	if t.state != ndn.Up {
		core.LogInfo(t, "Closing Unix stream socket")
		t.hasQuit <- true
		t.conn.Close()

		// Stop link service
		t.linkService.tellTransportQuit()

		FaceTable.Remove(t.faceID)
	}
}
