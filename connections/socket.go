package connections

import (
	"crypto/tls"
	"github.com/massbitprotocol/turbo/utils"
	"net"
	"strconv"
	"time"
)

// Socket represents an in between interface between connection objects and network sockets
type Socket interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close(string) error
	Equals(s Socket) bool

	SetReadDeadline(t time.Time) error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Properties() (utils.TbSSLProperties, error)
}

const dialTimeout = 30 * time.Second

// TLS wraps a raw TLS network connection to implement the Socket interface
type TLS struct {
	*tls.Conn
}

// NewTLS dials and creates a new TLS connection
func NewTLS(ip string, port int, certs *utils.SSLCerts) (*TLS, error) {
	config, err := certs.LoadPrivateConfig()
	if err != nil {
		return nil, err
	}

	ipAddress := ip + ":" + strconv.Itoa(port)
	ipConn, err := net.DialTimeout("tcp", ipAddress, dialTimeout)
	if err != nil {
		return nil, err
	}

	tlsClient := tls.Client(ipConn, config)
	err = tlsClient.Handshake()
	if err != nil {
		return nil, err
	}

	return &TLS{Conn: tlsClient}, nil
}

// NewTLSFromConn creates a new TLS wrapper on an existing TLS connection
func NewTLSFromConn(conn *tls.Conn) *TLS {
	return &TLS{Conn: conn}
}

// Properties returns the SSL properties embedded in TLS certificates
func (t TLS) Properties() (utils.TbSSLProperties, error) {
	state := t.Conn.ConnectionState()
	var (
		err             error
		tbSSLExtensions utils.TbSSLProperties
	)
	for _, peerCertificate := range state.PeerCertificates {
		tbSSLExtensions, err = utils.ParseSSLCertificate(peerCertificate)
		if err == nil {
			break
		}
	}
	return tbSSLExtensions, err
}

// Close closes the underlying TLS connection
func (t TLS) Close(_ string) error {
	return t.Conn.Close()
}

// Equals compares two connection IDs
func (t TLS) Equals(s Socket) bool {
	return t == s
}
