package broccoli

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/dropbox/rust-brotli/c/go/brotli"
	"github.com/libp2p/go-libp2p-core/compression"
)

// ID is the protocol ID for noise
const ID = "/compression/broccoli"

var _ compression.CompressedTransport = &Transport{}

// Transport implements the interface sec.SecureTransport
// https://godoc.org/github.com/libp2p/go-libp2p-core/sec#SecureConn
type compConn struct {
	rlock sync.Mutex
	wlock sync.Mutex
	raw   net.Conn

	brotliCompress *brotli.MultiCompressionReader
	compInBuffer   *bytes.Buffer
	compOutBuffer  *bytes.Buffer

	r *brotli.DecompressionReader
}

var options = brotli.CompressionOptions{
	NumThreads: 1,
	Quality:    4,
	Appendable: true,
	Magic:      true,
}

// Transport defines a compression transport with a compression level.
type Transport struct {
	level int
}

// New Creates a new tranport with a specific compression level.
func New() *Transport {
	return &Transport{}
}

//NewConn upgrades a raw connection into a compressed connection.
func (t *Transport) NewConn(raw net.Conn, isServer bool) (compression.CompressedConn, error) {
	return &compConn{
		raw: raw,
	}, nil
}

// Write compression wrapper
func (c *compConn) Write(b []byte) (int, error) {
	fmt.Println("Write")
	c.wlock.Lock()
	defer c.wlock.Unlock()
	if c.brotliCompress == nil {
		c.compInBuffer = bytes.NewBuffer(nil)
		c.brotliCompress = brotli.NewMultiCompressionReader(c.compInBuffer, options)
	}
	n, err := c.compInBuffer.Write(b)
	if err != nil {
		return n, err
	}
	_, err = io.Copy(c.raw, c.brotliCompress)
	fmt.Println("Copied", n, err)

	return n, err
}

// Read compression wrapper
func (c *compConn) Read(b []byte) (int, error) {
	c.rlock.Lock()
	defer c.rlock.Unlock()
	if c.r == nil {
		// This _needs_ to be lazy as it reads a header.
		c.r = brotli.NewDecompressionReader(c.raw)
	}
	n, err := c.r.Read(b)
	if err != nil {
		if err != io.EOF {
			c.r.Close()
		}
	}
	fmt.Println("Read", n, err)
	return n, err
}

func (c *compConn) Close() error {
	c.wlock.Lock()
	defer c.wlock.Unlock()
	// return multierr.Combine(c.brotliCompress.Close(), c.raw.Close())
	return c.raw.Close()
}

func (c *compConn) LocalAddr() net.Addr {
	return c.raw.LocalAddr()
}

func (c *compConn) RemoteAddr() net.Addr {
	return c.raw.RemoteAddr()
}

func (c *compConn) SetDeadline(t time.Time) error {
	return c.raw.SetDeadline(t)
}

func (c *compConn) SetReadDeadline(t time.Time) error {
	return c.raw.SetReadDeadline(t)
}

func (c *compConn) SetWriteDeadline(t time.Time) error {
	return c.raw.SetWriteDeadline(t)
}
