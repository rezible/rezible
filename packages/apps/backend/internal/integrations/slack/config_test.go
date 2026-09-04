package slackintegration

func (s *SlackIntegrationSuite) TestInstallationTargetResourceRef() {
	s.Equal("team:T123", (InstallationIds{TeamId: "T123", EnterpriseId: "E123"}).InstallationTargetResourceRef())
	s.Equal("enterprise:E123", (InstallationIds{EnterpriseId: "E123"}).InstallationTargetResourceRef())
	s.Empty((InstallationIds{}).InstallationTargetResourceRef())
}

func (s *SlackIntegrationSuite) TestAppInstallationConfigUsesServiceNamespace() {
	credentials := &InstallationConfig{
		AccessToken: "xoxb-token",
		BotUserID:   "U123",
		Team:        &TeamInfo{Id: "T123", Name: "Rezible"},
	}
	config := appInstallationConfig{
		providerNamespace: "slack_agent",
		credentials:       credentials,
	}

	ref := config.InstallationTargetRef()
	s.Equal(ProviderName, ref.Provider)
	s.Equal("slack_agent", ref.ProviderNamespace)
	s.Equal("team:T123", ref.ResourceRef)
	encoded, encodeErr := config.Encode()
	s.Require().NoError(encodeErr)
	decoded, decodeErr := GetValidatedConfig(encoded)
	s.Require().NoError(decodeErr)
	s.Equal(credentials.Team.Id, decoded.Team.Id)
}

func (s *SlackIntegrationSuite) TestGetValidatedConfig() {
	valid := &InstallationConfig{
		AccessToken: "xoxb-token",
		BotUserID:   "U123",
		Team:        &TeamInfo{Id: "T123", Name: "Rezible"},
	}
	encoded, encodeErr := valid.Encode()
	s.Require().NoError(encodeErr)
	decoded, decodeErr := GetValidatedConfig(encoded)
	s.Require().NoError(decodeErr)
	s.Equal("T123", decoded.Team.Id)

	tests := map[string]*InstallationConfig{
		"missing access token": {
			BotUserID: "U123",
			Team:      &TeamInfo{Id: "T123"},
		},
		"missing bot user": {
			AccessToken: "xoxb-token",
			Team:        &TeamInfo{Id: "T123"},
		},
		"missing target": {
			AccessToken: "xoxb-token",
			BotUserID:   "U123",
		},
		"blank team ID": {
			AccessToken: "xoxb-token",
			BotUserID:   "U123",
			Team:        &TeamInfo{},
		},
		"blank enterprise ID": {
			AccessToken: "xoxb-token",
			BotUserID:   "U123",
			Enterprise:  &TeamInfo{},
		},
	}
	for name, cfg := range tests {
		s.Run(name, func() {
			encoded, encodeErr := cfg.Encode()
			s.Require().NoError(encodeErr)
			_, decodeErr := GetValidatedConfig(encoded)
			s.Require().Error(decodeErr)
		})
	}
}
