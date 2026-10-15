package email

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProviderDomains protects that the provider domains can be looked up
// with the domain part of a normalized address, which is always lower case,
// so every entry must be a lower case bare domain name.
func TestProviderDomains(t *testing.T) {
	domains := ProviderDomains()
	require.NotEmpty(t, domains)

	for domain := range domains {
		assert.Equal(t, strings.ToLower(domain), domain, "domain must be lower case")
		assert.Equal(t, strings.TrimSpace(domain), domain, "domain must not contain surrounding whitespace")
		assert.NotContains(t, domain, "@", "must be a domain, not an address")
		assert.Contains(t, domain, ".", "must be a fully qualified domain")
	}

	addr, err := NormalizedAddress("Someone <Someone@GMail.com>")
	require.NoError(t, err)
	assert.Contains(t, domains, addr.DomainPart(), "lookup via a normalized address")
	assert.NotContains(t, domains, "domonda.com", "company domains are not providers")
}

func TestProviderDomains_ReturnsIndependentMaps(t *testing.T) {
	// Callers may modify the returned map without affecting other callers.
	first := ProviderDomains()
	delete(first, "gmail.com")
	first["example.com"] = struct{}{}

	second := ProviderDomains()
	assert.Contains(t, second, "gmail.com")
	assert.NotContains(t, second, "example.com")
}
