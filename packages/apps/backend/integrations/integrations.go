package integrations

import (
	"fmt"
	"net/http"
	"sync"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

func GetSourceQueryCursor(cursors map[string]string, source string) (string, bool) {
	sc, ok := cursors[source]
	return sc, ok || len(cursors) == 0
}

type PackageRegistry struct {
	pkgsMu            sync.RWMutex
	nameMap           map[string]rez.IntegrationPackage
	availablePackages []rez.IntegrationPackage
}

var packageRegistry = NewPackageRegistry()

func NewPackageRegistry() *PackageRegistry {
	return &PackageRegistry{
		nameMap: make(map[string]rez.IntegrationPackage),
	}
}

func DefaultPackageRegistry() *PackageRegistry {
	return packageRegistry
}

func (r *PackageRegistry) RegisterPackage(pkg rez.IntegrationPackage) error {
	r.pkgsMu.Lock()
	defer r.pkgsMu.Unlock()

	available, configErr := pkg.IsAvailable()
	if !available {
		return configErr
	}
	r.availablePackages = append(r.availablePackages, pkg)
	r.nameMap[pkg.Name()] = pkg

	return nil
}

func (r *PackageRegistry) GetPackage(name string) (rez.IntegrationPackage, error) {
	p, valid := r.nameMap[name]
	if !valid {
		return nil, fmt.Errorf("unknown integration: %s", name)
	}
	return p, nil
}

func (r *PackageRegistry) GetAvailable() []rez.IntegrationPackage {
	return r.availablePackages
}

func (r *PackageRegistry) GetWebhookHandlers() map[string]http.Handler {
	type IntegrationWithWebhookHandler interface {
		WebhookHandler() http.Handler
	}
	whs := make(map[string]http.Handler)
	for _, pkg := range r.availablePackages {
		if whPkg, hasWebhook := pkg.(IntegrationWithWebhookHandler); hasWebhook {
			whs[pkg.Name()] = whPkg.WebhookHandler()
		}
	}
	return whs
}

func (r *PackageRegistry) GetProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	type IntegrationWithProviderEventQuerier interface {
		MakeProviderEventQuerier(*ent.Integration) (rez.ProviderEventQuerier, error)
	}
	pkg, valid := r.nameMap[intg.IntegrationName]
	if !valid {
		return nil, fmt.Errorf("unknown integration package: %s", intg.IntegrationName)
	}
	if querierPkg, ok := pkg.(IntegrationWithProviderEventQuerier); ok {
		return querierPkg.MakeProviderEventQuerier(intg)
	}
	return nil, nil
}
