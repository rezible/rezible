package github

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
)

var (
	testReceivedAt        = time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	checkoutLinkingValues = projections.LinkingAttributes{"repository.full_name": "acme/checkout"}
)

type EventProcessorSuite struct {
	test.Suite
}

func TestEventProcessorSuite(t *testing.T) {
	suite.Run(t, &EventProcessorSuite{Suite: test.NewSuite()})
}

func (s *EventProcessorSuite) processTestEvent(source string, payload any) []byte {
	s.T().Helper()
	normalized := s.processTestEvents(source, payload)
	s.Require().Len(normalized, 1)
	return normalized[0].Attributes
}

func (s *EventProcessorSuite) processTestEvents(source string, payload any) ent.NormalizedEvents {
	s.T().Helper()
	encoded, encodeErr := json.Marshal(payload)
	s.Require().NoError(encodeErr)

	event := rez.ProviderEvent{
		Provider:            ProviderName,
		ProviderNamespace:   "installation-1",
		ProviderEventSource: source,
		ProviderEventRef:    "delivery-1",
		Attributes:          encoded,
		ReceivedAt:          testReceivedAt,
	}
	normalized, processErr := EventProcessor{}.ProcessProviderEvent(s.T().Context(), event)
	s.Require().NoError(processErr)
	return normalized
}

// checkoutPullRequestEvent is pull request 7 of Acme/Checkout, whose default branch is main.
func checkoutPullRequestEvent(action string, merged bool) github.PullRequestEvent {
	pullRequest := &github.PullRequest{
		Number:    github.Ptr(7),
		Title:     github.Ptr("Raise checkout timeout"),
		HTMLURL:   github.Ptr("https://github.com/Acme/Checkout/pull/7"),
		CreatedAt: &github.Timestamp{Time: testReceivedAt.Add(-time.Hour)},
		Merged:    github.Ptr(merged),
		Base:      &github.PullRequestBranch{Ref: github.Ptr("main")},
		User:      &github.User{Login: github.Ptr("octocat")},
	}
	if merged {
		pullRequest.MergedAt = &github.Timestamp{Time: testReceivedAt.Add(-time.Minute)}
		pullRequest.MergeCommitSHA = github.Ptr("4F1C0D9E8B7A6F5E4D3C2B1A09F8E7D6C5B4A392")
	}
	return github.PullRequestEvent{
		Action: github.Ptr(action),
		Repo: &github.Repository{
			ID:            github.Ptr(int64(42)),
			FullName:      github.Ptr("Acme/Checkout"),
			DefaultBranch: github.Ptr("main"),
		},
		PullRequest: pullRequest,
	}
}

func (s *EventProcessorSuite) TestPushObservesRepositoryByFullName() {
	push := github.PushEvent{
		Ref:   github.Ptr("refs/heads/main"),
		After: github.Ptr("abc123"),
		Repo: &github.PushEventRepository{
			ID:       github.Ptr(int64(42)),
			FullName: github.Ptr("Acme/Checkout"),
		},
	}

	encoded := s.processTestEvent(sourcePushEvent, push)

	var attrs projections.CodeChangeEventAttributes
	s.Require().NoError(json.Unmarshal(encoded, &attrs))
	s.Require().Equal("Acme/Checkout", attrs.Repository.DisplayName)
	s.Require().Equal(checkoutLinkingValues, attrs.Repository.LinkingAttributes)
}

func (s *EventProcessorSuite) TestPullRequestObservesRepositoryByFullName() {
	pullRequest := checkoutPullRequestEvent("closed", true)

	encoded := s.processTestEvent(sourcePullEvent, pullRequest)

	var attrs projections.CodeChangeEventAttributes
	s.Require().NoError(json.Unmarshal(encoded, &attrs))
	s.Require().Equal(checkoutLinkingValues, attrs.Repository.LinkingAttributes)
}

func (s *EventProcessorSuite) TestPullRequestMergeIsRecorded() {
	s.Run("a merge records the merge fields", func() {
		pullRequest := checkoutPullRequestEvent("closed", true)

		normalized := s.processTestEvents(sourcePullEvent, pullRequest)

		s.Require().Len(normalized, 1)
		event := normalized[0]
		s.Require().Equal("change:42:pr:7", event.ProviderResourceRef)
		s.Require().Equal(testReceivedAt.Add(-time.Minute), event.OccurredAt)

		var attrs projections.CodeChangeEventAttributes
		s.Require().NoError(json.Unmarshal(event.Attributes, &attrs))
		s.Require().Equal("Raise checkout timeout", attrs.DisplayName)
		expected := &projections.CodeChangeMerge{
			MergedAt:          testReceivedAt.Add(-time.Minute),
			MergeCommitSha:    "4f1c0d9e8b7a6f5e4d3c2b1a09f8e7d6c5b4a392",
			BaseRef:           "main",
			IntoDefaultBranch: true,
			Number:            7,
			URL:               "https://github.com/Acme/Checkout/pull/7",
			AuthorLogin:       "octocat",
		}
		s.Require().Equal(expected, attrs.Merge)
	})

	s.Run("a merge into another branch is not into the default branch", func() {
		pullRequest := checkoutPullRequestEvent("closed", true)
		pullRequest.PullRequest.Base.Ref = github.Ptr("release-1")

		encoded := s.processTestEvent(sourcePullEvent, pullRequest)

		var attrs projections.CodeChangeEventAttributes
		s.Require().NoError(json.Unmarshal(encoded, &attrs))
		s.Require().Equal("release-1", attrs.Merge.BaseRef)
		s.Require().False(attrs.Merge.IntoDefaultBranch)
	})

	notChanges := []struct {
		name   string
		action string
		merged bool
	}{
		{name: "opened", action: "opened", merged: false},
		{name: "edited after merging", action: "edited", merged: true},
		{name: "closed without merging", action: "closed", merged: false},
	}
	for _, tc := range notChanges {
		s.Run(tc.name+" is not a change", func() {
			pullRequest := checkoutPullRequestEvent(tc.action, tc.merged)

			normalized := s.processTestEvents(sourcePullEvent, pullRequest)

			s.Require().Empty(normalized)
		})
	}
}

func (s *EventProcessorSuite) TestRepositorySyncObservesRepositoryByFullName() {
	repository := githubRepositoryObservedPayload{
		ID:       42,
		FullName: "Acme/Checkout",
	}

	encoded := s.processTestEvent(sourceRepositories, repository)

	var attrs projections.CodeForgeEventAttributes
	s.Require().NoError(json.Unmarshal(encoded, &attrs))
	s.Require().Equal("Acme/Checkout", attrs.DisplayName)
	s.Require().Equal(checkoutLinkingValues, attrs.LinkingAttributes)
}
