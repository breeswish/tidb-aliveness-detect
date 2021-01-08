package sqlclient

import (
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func serveInitialPacket(t *testing.T, packet []byte, closeImmediately bool) (*HandshakePacket, error) {
	listener, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	go func() {
		conn, err := listener.Accept()
		assert.Nil(t, err)
		_, _ = conn.Write(packet)
		if !closeImmediately {
			readBuf := make([]byte, 10)
			_, _ = io.ReadFull(conn, readBuf)
		}
		_ = conn.Close()
		_ = listener.Close()
	}()
	return TryConnect(listener.Addr().String())
}

// Assuming that port 1 is not used.
func TestTryConnectNotListeningPort(t *testing.T) {
	t.Parallel()

	p, err := TryConnect("127.0.0.1:1")
	assert.Nil(t, p)
	assert.NotNil(t, err)
}

// A server that never accept.
func TestTryConnectServerNoAccept(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)

	p, err := TryConnect(listener.Addr().String())
	assert.Nil(t, p)
	assert.NotNil(t, err)
	assert.Nil(t, listener.Close())
}

// A server that wait for incoming packet first.
func TestTryConnectServerWaitClientRequest(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	go func() {
		conn, err := listener.Accept()
		assert.Nil(t, err)

		readBuf := make([]byte, 10)
		_, _ = io.ReadFull(conn, readBuf)
		_ = conn.Close()

		_ = listener.Close()
	}()

	p, err := TryConnect(listener.Addr().String())
	assert.Nil(t, p)
	assert.NotNil(t, err)
}

// A server that simply reject all connections.
func TestTryConnectServerRejectAll(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", ":0")
	assert.Nil(t, err)
	go func() {
		conn, err := listener.Accept()
		assert.Nil(t, err)

		_ = conn.Close()
		_ = listener.Close()
	}()

	p, err := TryConnect(listener.Addr().String())
	assert.Nil(t, p)
	assert.NotNil(t, err)
}

// From https://dev.mysql.com/doc/internals/en/connection-phase-packets.html
var sampleInitialPacket1 = []byte{
	0x36, 0x00, 0x00, 0x00, 0x0a, 0x35, 0x2e, 0x35, 0x2e, 0x32, 0x2d, 0x6d, 0x32, 0x00, 0x0b, 0x00,
	0x00, 0x00, 0x64, 0x76, 0x48, 0x40, 0x49, 0x2d, 0x43, 0x4a, 0x00, 0xff, 0xf7, 0x08, 0x02, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2a, 0x34, 0x64,
	0x7c, 0x63, 0x5a, 0x77, 0x6b, 0x34, 0x5e, 0x5d, 0x3a, 0x00,
}

// From https://dev.mysql.com/doc/internals/en/connection-phase-packets.html
var sampleInitialPacket2 = []byte{
	0x50, 0x00, 0x00, 0x00, 0x0a, 0x35, 0x2e, 0x36, 0x2e, 0x34, 0x2d, 0x6d, 0x37, 0x2d, 0x6c, 0x6f,
	0x67, 0x00, 0x56, 0x0a, 0x00, 0x00, 0x52, 0x42, 0x33, 0x76, 0x7a, 0x26, 0x47, 0x72, 0x00, 0xff,
	0xff, 0x08, 0x02, 0x00, 0x0f, 0xc0, 0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x2b, 0x79, 0x44, 0x26, 0x2f, 0x5a, 0x5a, 0x33, 0x30, 0x35, 0x5a, 0x47, 0x00, 0x6d, 0x79,
	0x73, 0x71, 0x6c, 0x5f, 0x6e, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77,
	0x6f, 0x72, 0x64, 0x00,
}

// Captured by connecting to a real TiDB server
var sampleInitialPacket3 = []byte{
	0x6d, 0x00, 0x00, 0x00, 0x0a, 0x35, 0x2e, 0x37, 0x2e, 0x32, 0x35, 0x2d, 0x54, 0x69, 0x44, 0x42,
	0x2d, 0x76, 0x34, 0x2e, 0x30, 0x2e, 0x30, 0x2d, 0x62, 0x65, 0x74, 0x61, 0x2e, 0x32, 0x2d, 0x31,
	0x39, 0x34, 0x36, 0x2d, 0x67, 0x65, 0x61, 0x65, 0x36, 0x34, 0x65, 0x34, 0x30, 0x66, 0x00, 0x9d,
	0x00, 0x00, 0x00, 0x29, 0x36, 0x0a, 0x65, 0x2c, 0x63, 0x0f, 0x55, 0x00, 0x8f, 0xa6, 0x2e, 0x02,
	0x00, 0x1b, 0x00, 0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3e, 0x68,
	0x32, 0x0b, 0x02, 0x46, 0x49, 0x03, 0x73, 0x6e, 0x26, 0x72, 0x00, 0x6d, 0x79, 0x73, 0x71, 0x6c,
	0x5f, 0x6e, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64,
	0x00,
}

// Captured by connecting to a real MySQL server
var sampleInitialPacket4 = []byte{
	0x4a, 0x00, 0x00, 0x00, 0x0a, 0x38, 0x2e, 0x30, 0x2e, 0x31, 0x39, 0x00, 0x08, 0x00, 0x00, 0x00,
	0x34, 0x3f, 0x13, 0x76, 0x3e, 0x05, 0x5a, 0x7d, 0x00, 0xff, 0xff, 0xff, 0x02, 0x00, 0xff, 0xc7,
	0x15, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x31, 0x4f, 0x3a, 0x23, 0x69,
	0x63, 0x25, 0x57, 0x1e, 0x70, 0x5c, 0x7f, 0x00, 0x63, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x5f,
	0x73, 0x68, 0x61, 0x32, 0x5f, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x00,
}

// Valid initial handshake packet.
func TestTryConnectServerValidHandshake(t *testing.T) {
	t.Parallel()

	for i := 0; i <=1; i++ {
		closeImmediately := i == 1
		p, err := serveInitialPacket(t, sampleInitialPacket1, closeImmediately)
		assert.Nil(t, err)
		assert.Equal(t, "5.5.2-m2", p.ServerVersion)
		assert.Equal(t, uint32(0x0b), p.ConnectionIDU32)

		p, err = serveInitialPacket(t, sampleInitialPacket2, closeImmediately)
		assert.Nil(t, err)
		assert.Equal(t, "5.6.4-m7-log", p.ServerVersion)
		assert.Equal(t, uint32(0x0a56), p.ConnectionIDU32)

		p, err = serveInitialPacket(t, sampleInitialPacket3, closeImmediately)
		assert.Nil(t, err)
		assert.Equal(t, "5.7.25-TiDB-v4.0.0-beta.2-1946-geae64e40f", p.ServerVersion)
		assert.Equal(t, uint32(157), p.ConnectionIDU32)
		assert.Equal(t, uint32(0x1ba68f), p.CapabilityFlag)

		p, err = serveInitialPacket(t, sampleInitialPacket4, closeImmediately)
		assert.Nil(t, err)
		assert.Equal(t, "8.0.19", p.ServerVersion)
		assert.Equal(t, uint32(8), p.ConnectionIDU32)
		assert.Equal(t, uint32(0xc7ffffff), p.CapabilityFlag)
	}
}

// A server that send incorrect initial handshake packet.
func TestTryConnectServerInvalidHandshake1(t *testing.T) {
	t.Parallel()

	for i := 0; i <= 1; i++ {
		closeImmediately := i == 1

		p, err := serveInitialPacket(t, []byte{0x00, 0x01}, closeImmediately)
		assert.Nil(t, p)
		assert.NotNil(t, err)

		p, err = serveInitialPacket(t, []byte{0x01, 0x00, 0x00, 0x00, 0x01}, closeImmediately)
		assert.Nil(t, p)
		assert.NotNil(t, err)

		p, err = serveInitialPacket(t, []byte{0x00, 0x00, 0x00, 0x00}, closeImmediately)
		assert.Nil(t, p)
		assert.NotNil(t, err)

		// Incorrect sequence number
		p, err = serveInitialPacket(t, []byte{0x01, 0x00, 0x00, 0x01, 0x01}, closeImmediately)
		assert.Nil(t, p)
		assert.NotNil(t, err)

		// Incomplete packet
		p, err = serveInitialPacket(t, sampleInitialPacket3[:10], closeImmediately)
		assert.Nil(t, p)
		assert.NotNil(t, err)
	}
}

// A server that send incorrect initial handshake packet.
func TestTryConnectServerInvalidHandshake2(t *testing.T) {
	t.Parallel()

	packet := make([]byte, len(sampleInitialPacket3))

	// Note: we do not resist all packet tamper.
	for l := 50; l >= 4; l-- {
		copy(packet, sampleInitialPacket3)
		// Pretend to be a good packet, but a bad initial handshake.
		packet[0] = byte(l - 4)
		p, err := serveInitialPacket(t, packet[:l], true)
		assert.Nil(t, p)
		assert.NotNil(t, err)
	}
}