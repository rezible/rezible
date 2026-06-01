package testkit

type options struct {
	configOverrides          map[string]any
	configAllowValidationErr bool

	skipSeedOrganization bool
	skipSeedUser         bool
}

type SuiteOption func(*options)

func WithSkipSeedOrganization() SuiteOption {
	return func(o *options) { o.skipSeedOrganization = true }
}

func WithSkipSeedUser() SuiteOption {
	return func(o *options) { o.skipSeedUser = true }
}

func WithConfigOverrides(overrides map[string]any) SuiteOption {
	return func(o *options) { o.configOverrides = overrides }
}

func WithAllowConfigValidationErrors() SuiteOption {
	return func(o *options) { o.configAllowValidationErr = true }
}
