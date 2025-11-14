package apiv1

import (
	"context"
	"regexp"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi/v1"
	"github.com/rs/zerolog/log"
)

func getRequestAuthSession(ctx context.Context, auth rez.AuthService) *rez.AuthSession {
	sess, sessErr := auth.GetAuthSession(ctx)
	if sessErr != nil {
		panic("getRequestAuthSession: " + sessErr.Error())
	}
	return sess
}

func requestUserId(ctx context.Context, auth rez.AuthService) uuid.UUID {
	return getRequestAuthSession(ctx, auth).UserId
}

var (
	uniqueErrFieldRe         = regexp.MustCompile("unique constraint \".*_(.*)_key\"")
	enumValidationErrFieldRe = regexp.MustCompile("invalid enum value for")

	commonConstraints = map[string]string{
		"name":  "Name already exists",
		"value": "Value already exists",
	}
)

func apiError(msg string, err error) error {
	if oapi.IsClientError(err) {
		return err
	}

	if ent.IsNotFound(err) {
		return oapi.ErrorNotFound("Not found")
	}

	if enumValidationErrFieldRe.MatchString(err.Error()) {
		return err
	}

	if ent.IsConstraintError(err) {
		match := uniqueErrFieldRe.FindStringSubmatch(err.Error())
		if match == nil || len(match) < 2 {
			return oapi.ErrorBadRequest("Constraint failed")
		}

		field := match[1]
		cstrMsg, found := commonConstraints[field]
		if found {
			detail := oapi.NewErrorDetail(cstrMsg, field, nil)
			return oapi.ErrorBadRequest("Constraint Error", detail)
		}
		return oapi.ErrorBadRequest("Value is not unique")
	}

	log.Error().Err(err).Msg(msg)
	return oapi.ErrorInternal(msg, err)
}
