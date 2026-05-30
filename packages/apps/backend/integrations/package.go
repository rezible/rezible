package integrations

import (
	"fmt"
	"strings"

	fakeprovider "github.com/rezible/rezible/internal/integrations/fake"
	"github.com/rezible/rezible/internal/integrations/github"
	"github.com/rezible/rezible/internal/integrations/google"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/samber/do/v2"

	rez "github.com/rezible/rezible"
)

type integrationPackages []rez.IntegrationPackage

var Package = do.Package(
	fakeprovider.Package,
	slack.Package,
	github.Package,
	google.Package,
	do.Lazy(func(i do.Injector) (*PackageRegistry, error) {
		tel := do.MustInvoke[rez.TelemetryService](i)
		reg := &PackageRegistry{
			nameMap: make(map[string]rez.IntegrationPackage),
			logger:  tel.NewLogger(rez.LoggerOptions{PackageName: "integrations"}),
		}
		return reg, nil
	}),
	do.Lazy(func(i do.Injector) (integrationPackages, error) {
		var pkgs integrationPackages
		for _, svcDesc := range i.ListProvidedServices() {
			if !strings.Contains(svcDesc.Service, "internal/integrations") {
				continue
			}
			s := do.MustInvokeNamed[any](i, svcDesc.Service)
			if pkg, ok := s.(rez.IntegrationPackage); ok {
				pkgs = append(pkgs, pkg)
			}
		}
		return pkgs, nil
	}),
)

func RegisterIntegrations(i do.Injector) error {
	reg := do.MustInvoke[*PackageRegistry](i)
	for _, pkg := range do.MustInvoke[integrationPackages](i) {
		if regErr := reg.registerPackage(pkg); regErr != nil {
			return fmt.Errorf("failed to register integration package: %w", regErr)
		}
	}
	return nil
}
