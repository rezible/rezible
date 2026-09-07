package db

import (
	"testing"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/suite"
)

type AuthSessionServiceSuite struct {
	test.Suite
}

func TestAuthSessionServiceSuite(t *testing.T) {
	suite.Run(t, &AuthSessionServiceSuite{Suite: test.NewSuite()})
}

func (s *AuthSessionServiceSuite) TestCreatingSessionPreservesExistingSessions() {
	database := s.CreateTestDatabase()
	jobs := mocks.NewMockJobService(s.T())
	organizations, _ := NewOrganizationService(database, jobs)
	users, _ := NewUserService(database, organizations)
	sessions, _ := NewAuthSessionService(database, organizations, users)
	providerSession := &rez.UserAuthProviderSession{
		User:      ent.User{Email: "session-test@example.com", Name: "Session Test", AuthProviderID: "session-test-user"},
		Org:       ent.Organization{Name: "Session Test Org", AuthProviderID: "session-test-org"},
		ExpiresAt: time.Now().Add(time.Hour),
	}

	first, firstErr := sessions.CreateFromUserAuthResponse(s.T().Context(), providerSession)
	s.Require().NoError(firstErr)
	second, secondErr := sessions.CreateFromUserAuthResponse(s.T().Context(), providerSession)
	s.Require().NoError(secondErr)
	s.NotEqual(first.ID, second.ID)

	loaded, lookupErr := sessions.LookupSession(s.T().Context(), first.ID)
	s.Require().NoError(lookupErr)
	s.Equal(first.ID, loaded.ID)
}
