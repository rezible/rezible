package db

import (
	"net/url"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type DocumentsServiceSuite struct{ test.Suite }

func TestDocumentsServiceSuite(t *testing.T) {
	suite.Run(t, &DocumentsServiceSuite{Suite: test.NewSuite()})
}

func (s *DocumentsServiceSuite) TestClosedReviewSessionIsReadOnly() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	owner, ownerErr := client.User.Query().First(ctx)
	s.Require().NoError(ownerErr)
	doc, docErr := client.Document.Create().SetContent([]byte{}).Save(ctx)
	s.Require().NoError(docErr)
	analysis, analysisErr := client.SystemAnalysis.Create().Save(ctx)
	s.Require().NoError(analysisErr)
	createIncident := client.Incident.Create().
		SetSlug(uuid.NewString()).
		SetTitle("Checkout outage")
	inc, incidentErr := createIncident.Save(ctx)
	s.Require().NoError(incidentErr)
	createRetro := client.Retrospective.Create().
		SetIncidentID(inc.ID).
		SetDocumentID(doc.ID).
		SetSystemAnalysisID(analysis.ID).
		SetState(retrospective.StateDraft)
	retro, retroErr := createRetro.Save(ctx)
	s.Require().NoError(retroErr)
	serverURL, urlErr := url.Parse("http://documents.test")
	s.Require().NoError(urlErr)
	key := paseto.NewV4SymmetricKey()
	service := &DocumentsService{db: tdb, sessionKey: key, documentsServerUrl: serverURL}
	before, beforeErr := service.GetUserDocumentAccess(ctx, doc.ID, owner.ID)
	s.Require().NoError(beforeErr)
	s.True(before.CanEdit)
	closeReview := client.Retrospective.UpdateOneID(retro.ID).
		SetState(retrospective.StateClosed).
		SetClosedAt(time.Now()).
		SetClosedByID(owner.ID)
	s.Require().NoError(closeReview.Exec(ctx))
	access, accessErr := service.GetUserDocumentAccess(ctx, doc.ID, owner.ID)
	s.Require().NoError(accessErr)
	s.True(access.CanView)
	s.False(access.CanEdit)
	s.False(access.CanManage)
	auth, authErr := service.CreateDocumentEditorSessionAuth(ctx, doc.ID, owner.ID)
	s.Require().NoError(authErr)
	token, parseErr := paseto.NewParser().ParseV4Local(key, auth.Token, nil)
	s.Require().NoError(parseErr)
	var canEdit bool
	s.Require().NoError(token.Get("can_edit", &canEdit))
	s.False(canEdit)
}
