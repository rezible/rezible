package integrations

import (
	"context"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	rez "github.com/rezible/rezible"

	"golang.org/x/oauth2"
)

type IntegrationWithOAuth2Flow interface {
	OAuth2Config() *oauth2.Config
	RetrieveInstallationTargetOptions(context.Context, *oauth2.Token) ([]rez.IntegrationInstallationTarget, error)
}

func (r *PackageRegistry) GetOAuthIntegration(name string) (IntegrationWithOAuth2Flow, error) {
	ip, ipErr := r.GetPackage(name)
	if ipErr != nil {
		return nil, fmt.Errorf("invalid integration %s: %w", name, ipErr)
	}
	oauth2Intg, ok := ip.(IntegrationWithOAuth2Flow)
	if !ok {
		return nil, fmt.Errorf("oauth2 flow not supported for integration %s", name)
	}
	if oauth2Intg.OAuth2Config() == nil {
		return nil, fmt.Errorf("empty integration oauth2 configuration")
	}
	return oauth2Intg, nil
}

func EncodeInstallationTargetOptions(options []rez.IntegrationInstallationTarget) ([]map[string]any, error) {
	result := make([]map[string]any, 0, len(options))
	for _, opt := range options {
		var optMap map[string]any
		if encErr := mapstructure.Decode(opt, &optMap); encErr != nil {
			return nil, fmt.Errorf("failed to encode to map: %w", encErr)
		}
		result = append(result, optMap)
	}
	return result, nil
}

func DecodeInstallationTargetOptions(opts []map[string]any) ([]rez.IntegrationInstallationTarget, error) {
	result := make([]rez.IntegrationInstallationTarget, 0, len(opts))
	for _, opt := range opts {
		var decoded rez.IntegrationInstallationTarget
		if decErr := mapstructure.Decode(opt, &decoded); decErr != nil {
			return nil, fmt.Errorf("failed to decode option from map: %w", decErr)
		}
		result = append(result, decoded)
	}
	return result, nil
}
