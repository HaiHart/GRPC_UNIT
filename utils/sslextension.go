package utils

import (
	"crypto/x509"
	"github.com/massbitprotocol/turbo/types"
)

type TbSSLProperties struct {
	NodeType NodeType
	NodeID   types.NodeID
}

// ParseSSLCertificate extracts specific extension information from the SSL certificates
func ParseSSLCertificate(certificate *x509.Certificate) (TbSSLProperties, error) {
	panic("not implemented")
}
