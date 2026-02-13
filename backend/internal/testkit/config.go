package testkit

import "time"

type ConfigLoader struct {
	DatabaseURL             string
	AllowTenantCreationFlag bool
	AllowUserCreationFlag   bool
	DataSyncModeFlag        bool
	DebugModeFlag           bool
	SingleTenantModeFlag    bool
	Values                  map[string]string
	Bools                   map[string]bool
	Durations               map[string]time.Duration
}

func (c *ConfigLoader) GetString(key string) string { return c.Values[key] }

func (c *ConfigLoader) GetStringOr(key string, orDefault string) string {
	if v, ok := c.Values[key]; ok {
		return v
	}
	return orDefault
}

func (c *ConfigLoader) GetStrings(string) []string { return nil }
func (c *ConfigLoader) GetBool(key string) bool    { return c.Bools[key] }

func (c *ConfigLoader) GetBoolOr(key string, orDefault bool) bool {
	if v, ok := c.Bools[key]; ok {
		return v
	}
	return orDefault
}

func (c *ConfigLoader) GetDuration(key string) time.Duration { return c.Durations[key] }

func (c *ConfigLoader) GetDurationOr(key string, orDefault time.Duration) time.Duration {
	if v, ok := c.Durations[key]; ok {
		return v
	}
	return orDefault
}

func (c *ConfigLoader) SingleTenantMode() bool { return c.SingleTenantModeFlag }
func (c *ConfigLoader) DebugMode() bool        { return c.DebugModeFlag }
func (c *ConfigLoader) DataSyncMode() bool     { return c.DataSyncModeFlag }
func (c *ConfigLoader) DatabaseUrl() string    { return c.DatabaseURL }

func (c *ConfigLoader) DocumentsServerAddress() string {
	return c.GetString("DOCUMENTS_SERVER_ADDRESS")
}

func (c *ConfigLoader) AppUrl() string { return c.GetStringOr("APP_URL", "http://localhost:3000") }
func (c *ConfigLoader) ApiRouteBase() string {
	return c.GetStringOr("API_ROUTE_BASE", "/api/v1")
}
func (c *ConfigLoader) AuthRouteBase() string {
	return c.GetStringOr("AUTH_ROUTE_BASE", "/api/auth")
}
func (c *ConfigLoader) AllowTenantCreation() bool { return c.AllowTenantCreationFlag }
func (c *ConfigLoader) AllowUserCreation() bool   { return c.AllowUserCreationFlag }
