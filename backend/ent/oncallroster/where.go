// Code generated by ent, DO NOT EDIT.

package oncallroster

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLTE(FieldID, id))
}

// ArchiveTime applies equality check predicate on the "archive_time" field. It's identical to ArchiveTimeEQ.
func ArchiveTime(v time.Time) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldArchiveTime, v))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldName, v))
}

// Slug applies equality check predicate on the "slug" field. It's identical to SlugEQ.
func Slug(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldSlug, v))
}

// ProviderID applies equality check predicate on the "provider_id" field. It's identical to ProviderIDEQ.
func ProviderID(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldProviderID, v))
}

// Timezone applies equality check predicate on the "timezone" field. It's identical to TimezoneEQ.
func Timezone(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldTimezone, v))
}

// ChatHandle applies equality check predicate on the "chat_handle" field. It's identical to ChatHandleEQ.
func ChatHandle(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldChatHandle, v))
}

// ChatChannelID applies equality check predicate on the "chat_channel_id" field. It's identical to ChatChannelIDEQ.
func ChatChannelID(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldChatChannelID, v))
}

// HandoverTemplateID applies equality check predicate on the "handover_template_id" field. It's identical to HandoverTemplateIDEQ.
func HandoverTemplateID(v uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldHandoverTemplateID, v))
}

// ArchiveTimeEQ applies the EQ predicate on the "archive_time" field.
func ArchiveTimeEQ(v time.Time) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldArchiveTime, v))
}

// ArchiveTimeNEQ applies the NEQ predicate on the "archive_time" field.
func ArchiveTimeNEQ(v time.Time) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNEQ(FieldArchiveTime, v))
}

// ArchiveTimeIn applies the In predicate on the "archive_time" field.
func ArchiveTimeIn(vs ...time.Time) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIn(FieldArchiveTime, vs...))
}

// ArchiveTimeNotIn applies the NotIn predicate on the "archive_time" field.
func ArchiveTimeNotIn(vs ...time.Time) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotIn(FieldArchiveTime, vs...))
}

// ArchiveTimeGT applies the GT predicate on the "archive_time" field.
func ArchiveTimeGT(v time.Time) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGT(FieldArchiveTime, v))
}

// ArchiveTimeGTE applies the GTE predicate on the "archive_time" field.
func ArchiveTimeGTE(v time.Time) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGTE(FieldArchiveTime, v))
}

// ArchiveTimeLT applies the LT predicate on the "archive_time" field.
func ArchiveTimeLT(v time.Time) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLT(FieldArchiveTime, v))
}

// ArchiveTimeLTE applies the LTE predicate on the "archive_time" field.
func ArchiveTimeLTE(v time.Time) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLTE(FieldArchiveTime, v))
}

// ArchiveTimeIsNil applies the IsNil predicate on the "archive_time" field.
func ArchiveTimeIsNil() predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIsNull(FieldArchiveTime))
}

// ArchiveTimeNotNil applies the NotNil predicate on the "archive_time" field.
func ArchiveTimeNotNil() predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotNull(FieldArchiveTime))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasSuffix(FieldName, v))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContainsFold(FieldName, v))
}

// SlugEQ applies the EQ predicate on the "slug" field.
func SlugEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldSlug, v))
}

// SlugNEQ applies the NEQ predicate on the "slug" field.
func SlugNEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNEQ(FieldSlug, v))
}

// SlugIn applies the In predicate on the "slug" field.
func SlugIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIn(FieldSlug, vs...))
}

// SlugNotIn applies the NotIn predicate on the "slug" field.
func SlugNotIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotIn(FieldSlug, vs...))
}

// SlugGT applies the GT predicate on the "slug" field.
func SlugGT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGT(FieldSlug, v))
}

// SlugGTE applies the GTE predicate on the "slug" field.
func SlugGTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGTE(FieldSlug, v))
}

// SlugLT applies the LT predicate on the "slug" field.
func SlugLT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLT(FieldSlug, v))
}

// SlugLTE applies the LTE predicate on the "slug" field.
func SlugLTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLTE(FieldSlug, v))
}

// SlugContains applies the Contains predicate on the "slug" field.
func SlugContains(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContains(FieldSlug, v))
}

// SlugHasPrefix applies the HasPrefix predicate on the "slug" field.
func SlugHasPrefix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasPrefix(FieldSlug, v))
}

// SlugHasSuffix applies the HasSuffix predicate on the "slug" field.
func SlugHasSuffix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasSuffix(FieldSlug, v))
}

// SlugEqualFold applies the EqualFold predicate on the "slug" field.
func SlugEqualFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEqualFold(FieldSlug, v))
}

// SlugContainsFold applies the ContainsFold predicate on the "slug" field.
func SlugContainsFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContainsFold(FieldSlug, v))
}

// ProviderIDEQ applies the EQ predicate on the "provider_id" field.
func ProviderIDEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldProviderID, v))
}

// ProviderIDNEQ applies the NEQ predicate on the "provider_id" field.
func ProviderIDNEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNEQ(FieldProviderID, v))
}

// ProviderIDIn applies the In predicate on the "provider_id" field.
func ProviderIDIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIn(FieldProviderID, vs...))
}

// ProviderIDNotIn applies the NotIn predicate on the "provider_id" field.
func ProviderIDNotIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotIn(FieldProviderID, vs...))
}

// ProviderIDGT applies the GT predicate on the "provider_id" field.
func ProviderIDGT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGT(FieldProviderID, v))
}

// ProviderIDGTE applies the GTE predicate on the "provider_id" field.
func ProviderIDGTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGTE(FieldProviderID, v))
}

// ProviderIDLT applies the LT predicate on the "provider_id" field.
func ProviderIDLT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLT(FieldProviderID, v))
}

// ProviderIDLTE applies the LTE predicate on the "provider_id" field.
func ProviderIDLTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLTE(FieldProviderID, v))
}

// ProviderIDContains applies the Contains predicate on the "provider_id" field.
func ProviderIDContains(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContains(FieldProviderID, v))
}

// ProviderIDHasPrefix applies the HasPrefix predicate on the "provider_id" field.
func ProviderIDHasPrefix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasPrefix(FieldProviderID, v))
}

// ProviderIDHasSuffix applies the HasSuffix predicate on the "provider_id" field.
func ProviderIDHasSuffix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasSuffix(FieldProviderID, v))
}

// ProviderIDEqualFold applies the EqualFold predicate on the "provider_id" field.
func ProviderIDEqualFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEqualFold(FieldProviderID, v))
}

// ProviderIDContainsFold applies the ContainsFold predicate on the "provider_id" field.
func ProviderIDContainsFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContainsFold(FieldProviderID, v))
}

// TimezoneEQ applies the EQ predicate on the "timezone" field.
func TimezoneEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldTimezone, v))
}

// TimezoneNEQ applies the NEQ predicate on the "timezone" field.
func TimezoneNEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNEQ(FieldTimezone, v))
}

// TimezoneIn applies the In predicate on the "timezone" field.
func TimezoneIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIn(FieldTimezone, vs...))
}

// TimezoneNotIn applies the NotIn predicate on the "timezone" field.
func TimezoneNotIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotIn(FieldTimezone, vs...))
}

// TimezoneGT applies the GT predicate on the "timezone" field.
func TimezoneGT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGT(FieldTimezone, v))
}

// TimezoneGTE applies the GTE predicate on the "timezone" field.
func TimezoneGTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGTE(FieldTimezone, v))
}

// TimezoneLT applies the LT predicate on the "timezone" field.
func TimezoneLT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLT(FieldTimezone, v))
}

// TimezoneLTE applies the LTE predicate on the "timezone" field.
func TimezoneLTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLTE(FieldTimezone, v))
}

// TimezoneContains applies the Contains predicate on the "timezone" field.
func TimezoneContains(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContains(FieldTimezone, v))
}

// TimezoneHasPrefix applies the HasPrefix predicate on the "timezone" field.
func TimezoneHasPrefix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasPrefix(FieldTimezone, v))
}

// TimezoneHasSuffix applies the HasSuffix predicate on the "timezone" field.
func TimezoneHasSuffix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasSuffix(FieldTimezone, v))
}

// TimezoneIsNil applies the IsNil predicate on the "timezone" field.
func TimezoneIsNil() predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIsNull(FieldTimezone))
}

// TimezoneNotNil applies the NotNil predicate on the "timezone" field.
func TimezoneNotNil() predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotNull(FieldTimezone))
}

// TimezoneEqualFold applies the EqualFold predicate on the "timezone" field.
func TimezoneEqualFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEqualFold(FieldTimezone, v))
}

// TimezoneContainsFold applies the ContainsFold predicate on the "timezone" field.
func TimezoneContainsFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContainsFold(FieldTimezone, v))
}

// ChatHandleEQ applies the EQ predicate on the "chat_handle" field.
func ChatHandleEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldChatHandle, v))
}

// ChatHandleNEQ applies the NEQ predicate on the "chat_handle" field.
func ChatHandleNEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNEQ(FieldChatHandle, v))
}

// ChatHandleIn applies the In predicate on the "chat_handle" field.
func ChatHandleIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIn(FieldChatHandle, vs...))
}

// ChatHandleNotIn applies the NotIn predicate on the "chat_handle" field.
func ChatHandleNotIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotIn(FieldChatHandle, vs...))
}

// ChatHandleGT applies the GT predicate on the "chat_handle" field.
func ChatHandleGT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGT(FieldChatHandle, v))
}

// ChatHandleGTE applies the GTE predicate on the "chat_handle" field.
func ChatHandleGTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGTE(FieldChatHandle, v))
}

// ChatHandleLT applies the LT predicate on the "chat_handle" field.
func ChatHandleLT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLT(FieldChatHandle, v))
}

// ChatHandleLTE applies the LTE predicate on the "chat_handle" field.
func ChatHandleLTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLTE(FieldChatHandle, v))
}

// ChatHandleContains applies the Contains predicate on the "chat_handle" field.
func ChatHandleContains(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContains(FieldChatHandle, v))
}

// ChatHandleHasPrefix applies the HasPrefix predicate on the "chat_handle" field.
func ChatHandleHasPrefix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasPrefix(FieldChatHandle, v))
}

// ChatHandleHasSuffix applies the HasSuffix predicate on the "chat_handle" field.
func ChatHandleHasSuffix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasSuffix(FieldChatHandle, v))
}

// ChatHandleIsNil applies the IsNil predicate on the "chat_handle" field.
func ChatHandleIsNil() predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIsNull(FieldChatHandle))
}

// ChatHandleNotNil applies the NotNil predicate on the "chat_handle" field.
func ChatHandleNotNil() predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotNull(FieldChatHandle))
}

// ChatHandleEqualFold applies the EqualFold predicate on the "chat_handle" field.
func ChatHandleEqualFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEqualFold(FieldChatHandle, v))
}

// ChatHandleContainsFold applies the ContainsFold predicate on the "chat_handle" field.
func ChatHandleContainsFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContainsFold(FieldChatHandle, v))
}

// ChatChannelIDEQ applies the EQ predicate on the "chat_channel_id" field.
func ChatChannelIDEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldChatChannelID, v))
}

// ChatChannelIDNEQ applies the NEQ predicate on the "chat_channel_id" field.
func ChatChannelIDNEQ(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNEQ(FieldChatChannelID, v))
}

// ChatChannelIDIn applies the In predicate on the "chat_channel_id" field.
func ChatChannelIDIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIn(FieldChatChannelID, vs...))
}

// ChatChannelIDNotIn applies the NotIn predicate on the "chat_channel_id" field.
func ChatChannelIDNotIn(vs ...string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotIn(FieldChatChannelID, vs...))
}

// ChatChannelIDGT applies the GT predicate on the "chat_channel_id" field.
func ChatChannelIDGT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGT(FieldChatChannelID, v))
}

// ChatChannelIDGTE applies the GTE predicate on the "chat_channel_id" field.
func ChatChannelIDGTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldGTE(FieldChatChannelID, v))
}

// ChatChannelIDLT applies the LT predicate on the "chat_channel_id" field.
func ChatChannelIDLT(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLT(FieldChatChannelID, v))
}

// ChatChannelIDLTE applies the LTE predicate on the "chat_channel_id" field.
func ChatChannelIDLTE(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldLTE(FieldChatChannelID, v))
}

// ChatChannelIDContains applies the Contains predicate on the "chat_channel_id" field.
func ChatChannelIDContains(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContains(FieldChatChannelID, v))
}

// ChatChannelIDHasPrefix applies the HasPrefix predicate on the "chat_channel_id" field.
func ChatChannelIDHasPrefix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasPrefix(FieldChatChannelID, v))
}

// ChatChannelIDHasSuffix applies the HasSuffix predicate on the "chat_channel_id" field.
func ChatChannelIDHasSuffix(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldHasSuffix(FieldChatChannelID, v))
}

// ChatChannelIDIsNil applies the IsNil predicate on the "chat_channel_id" field.
func ChatChannelIDIsNil() predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIsNull(FieldChatChannelID))
}

// ChatChannelIDNotNil applies the NotNil predicate on the "chat_channel_id" field.
func ChatChannelIDNotNil() predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotNull(FieldChatChannelID))
}

// ChatChannelIDEqualFold applies the EqualFold predicate on the "chat_channel_id" field.
func ChatChannelIDEqualFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEqualFold(FieldChatChannelID, v))
}

// ChatChannelIDContainsFold applies the ContainsFold predicate on the "chat_channel_id" field.
func ChatChannelIDContainsFold(v string) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldContainsFold(FieldChatChannelID, v))
}

// HandoverTemplateIDEQ applies the EQ predicate on the "handover_template_id" field.
func HandoverTemplateIDEQ(v uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldEQ(FieldHandoverTemplateID, v))
}

// HandoverTemplateIDNEQ applies the NEQ predicate on the "handover_template_id" field.
func HandoverTemplateIDNEQ(v uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNEQ(FieldHandoverTemplateID, v))
}

// HandoverTemplateIDIn applies the In predicate on the "handover_template_id" field.
func HandoverTemplateIDIn(vs ...uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIn(FieldHandoverTemplateID, vs...))
}

// HandoverTemplateIDNotIn applies the NotIn predicate on the "handover_template_id" field.
func HandoverTemplateIDNotIn(vs ...uuid.UUID) predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotIn(FieldHandoverTemplateID, vs...))
}

// HandoverTemplateIDIsNil applies the IsNil predicate on the "handover_template_id" field.
func HandoverTemplateIDIsNil() predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldIsNull(FieldHandoverTemplateID))
}

// HandoverTemplateIDNotNil applies the NotNil predicate on the "handover_template_id" field.
func HandoverTemplateIDNotNil() predicate.OncallRoster {
	return predicate.OncallRoster(sql.FieldNotNull(FieldHandoverTemplateID))
}

// HasSchedules applies the HasEdge predicate on the "schedules" edge.
func HasSchedules() predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, SchedulesTable, SchedulesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasSchedulesWith applies the HasEdge predicate on the "schedules" edge with a given conditions (other predicates).
func HasSchedulesWith(preds ...predicate.OncallSchedule) predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := newSchedulesStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasHandoverTemplate applies the HasEdge predicate on the "handover_template" edge.
func HasHandoverTemplate() predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, HandoverTemplateTable, HandoverTemplateColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasHandoverTemplateWith applies the HasEdge predicate on the "handover_template" edge with a given conditions (other predicates).
func HasHandoverTemplateWith(preds ...predicate.OncallHandoverTemplate) predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := newHandoverTemplateStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasAnnotations applies the HasEdge predicate on the "annotations" edge.
func HasAnnotations() predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, AnnotationsTable, AnnotationsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAnnotationsWith applies the HasEdge predicate on the "annotations" edge with a given conditions (other predicates).
func HasAnnotationsWith(preds ...predicate.OncallAnnotation) predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := newAnnotationsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasTeams applies the HasEdge predicate on the "teams" edge.
func HasTeams() predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, TeamsTable, TeamsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTeamsWith applies the HasEdge predicate on the "teams" edge with a given conditions (other predicates).
func HasTeamsWith(preds ...predicate.Team) predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := newTeamsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasShifts applies the HasEdge predicate on the "shifts" edge.
func HasShifts() predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, ShiftsTable, ShiftsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasShiftsWith applies the HasEdge predicate on the "shifts" edge with a given conditions (other predicates).
func HasShiftsWith(preds ...predicate.OncallUserShift) predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := newShiftsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasAlerts applies the HasEdge predicate on the "alerts" edge.
func HasAlerts() predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, AlertsTable, AlertsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAlertsWith applies the HasEdge predicate on the "alerts" edge with a given conditions (other predicates).
func HasAlertsWith(preds ...predicate.OncallAlert) predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := newAlertsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasUserWatchers applies the HasEdge predicate on the "user_watchers" edge.
func HasUserWatchers() predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, UserWatchersTable, UserWatchersPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasUserWatchersWith applies the HasEdge predicate on the "user_watchers" edge with a given conditions (other predicates).
func HasUserWatchersWith(preds ...predicate.User) predicate.OncallRoster {
	return predicate.OncallRoster(func(s *sql.Selector) {
		step := newUserWatchersStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.OncallRoster) predicate.OncallRoster {
	return predicate.OncallRoster(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.OncallRoster) predicate.OncallRoster {
	return predicate.OncallRoster(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.OncallRoster) predicate.OncallRoster {
	return predicate.OncallRoster(sql.NotPredicates(p))
}
