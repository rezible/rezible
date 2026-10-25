package integrations

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

var errDisallowedAddress = errors.New("destination address not allowed")

var (
	awsMetadataV6  = netip.MustParseAddr("fd00:ec2::254")
	sharedAddress  = netip.MustParsePrefix("100.64.0.0/10")
	benchmarkRange = netip.MustParsePrefix("198.18.0.0/15")
)

func checkDialProviderHttpAddress(address string, allowPrivateNetworks bool) error {
	addrPort, parseErr := netip.ParseAddrPort(address)
	if parseErr != nil {
		return fmt.Errorf("parse dial address: %w", parseErr)
	}
	addr := addrPort.Addr().Unmap()
	if addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.WithZone("") == awsMetadataV6 {
		return fmt.Errorf("%w: %s", errDisallowedAddress, addr)
	}
	if allowPrivateNetworks {
		return nil
	}
	if addr.IsLoopback() || addr.IsPrivate() || sharedAddress.Contains(addr) || benchmarkRange.Contains(addr) || addr.IsMulticast() || addr.IsUnspecified() {
		return fmt.Errorf("%w: %s", errDisallowedAddress, addr)
	}
	return nil
}

type ProviderHttpClient = http.Client

// NewProviderHTTPClient returns a client that never connects to link-local addresses, including cloud metadata, and
// unless allowPrivateNetworks never to loopback, private, shared, benchmarking, multicast or unspecified
// addresses. Each connection checks the address it dials, so a DNS answer that changes between connections is
// checked again. Redirects are returned rather than followed.
func NewProviderHTTPClient(allowPrivateNetworks bool) *ProviderHttpClient {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control: func(_, address string, _ syscall.RawConn) error {
			return checkDialProviderHttpAddress(address, allowPrivateNetworks)
		},
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// A proxy would be dialed in place of the destination, hiding it from the check.
	transport.Proxy = nil
	transport.DialContext = dialer.DialContext
	return &http.Client{
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
