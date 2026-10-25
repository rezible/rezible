package integrations

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type ProvidersSuite struct {
	test.Suite
}

func TestProvidersSuite(t *testing.T) {
	suite.Run(t, &ProvidersSuite{Suite: test.NewSuite()})
}

func (s *ProvidersSuite) TestCheckDialProviderHttpAddress() {
	r := s.Require()
	for _, address := range []string{"127.0.0.1", "::1", "::ffff:127.0.0.1", "10.0.0.1", "100.64.0.1", "192.168.1.1"} {
		dialAddress := net.JoinHostPort(address, "443")
		r.ErrorIs(checkDialProviderHttpAddress(dialAddress, false), errDisallowedAddress, address)
		r.NoError(checkDialProviderHttpAddress(dialAddress, true), address)
	}
	for _, address := range []string{"169.254.169.254", "fe80::1"} {
		dialAddress := net.JoinHostPort(address, "80")
		r.ErrorIs(checkDialProviderHttpAddress(dialAddress, false), errDisallowedAddress, address)
		r.ErrorIs(checkDialProviderHttpAddress(dialAddress, true), errDisallowedAddress, address)
	}
}

func (s *ProvidersSuite) TestProviderHTTPClientChecksDialedAddressAndKeepsRedirects() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elsewhere", http.StatusFound)
	}))
	s.T().Cleanup(server.Close)

	r := s.Require()

	_, refusedErr := NewProviderHTTPClient(false).Get(server.URL)
	r.ErrorIs(refusedErr, errDisallowedAddress)

	resp, getErr := NewProviderHTTPClient(true).Get(server.URL)
	r.NoError(getErr)
	s.T().Cleanup(func() { _ = resp.Body.Close() })
	r.Equal(http.StatusFound, resp.StatusCode)
}
