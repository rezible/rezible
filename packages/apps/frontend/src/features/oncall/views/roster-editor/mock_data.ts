import type { OncallRoster } from "$lib/api";

export type RosterEditorParticipantDraft = {
	userId: string;
	order: number;
};

export type RosterEditorScheduleDraft = {
	key: string;
	id?: string;
	name: string;
	timezone: string;
	description: string;
	participants: RosterEditorParticipantDraft[];
};

export type RosterEditorDraft = {
	name: string;
	slug: string;
	timezone: string;
	chatHandle: string;
	chatChannelId: string;
	handoverTemplateId: string;
	schedules: RosterEditorScheduleDraft[];
};

type MockRosterSchemaData = {
	timezone: string;
	chatHandle: string;
	chatChannelId: string;
	schedules: Record<string, { name: string }>;
};

const defaultMockSchemaData: MockRosterSchemaData = {
	timezone: "Australia/Perth",
	chatHandle: "@platform-oncall",
	chatChannelId: "slack:C07PLATFORM",
	schedules: {},
};

const mockRosterSchemaDataBySlug: Record<string, MockRosterSchemaData> = {
	"platform-primary": {
		timezone: "Australia/Perth",
		chatHandle: "@platform-oncall",
		chatChannelId: "slack:C07PLATFORM",
		schedules: {},
	},
	"payments-oncall": {
		timezone: "America/Los_Angeles",
		chatHandle: "@payments-oncall",
		chatChannelId: "slack:C02PAYMENTS",
		schedules: {},
	},
};

export const rosterEditorTimezoneOptions = [
	"Australia/Perth",
	"Australia/Sydney",
	"UTC",
	"America/Los_Angeles",
	"America/New_York",
	"Europe/London",
	"Asia/Singapore",
];

export const mockBackedRosterEditorFields = [
	{
		label: "Roster timezone",
		reason: "Present in the Ent schema but not returned by the current oncall roster read API.",
	},
	{
		label: "Roster chat handle and channel",
		reason: "Present in the Ent schema but not returned by the current oncall roster read API.",
	},
	{
		label: "Schedule name",
		reason: "Present in the Ent schema but not returned by the current oncall roster read API.",
	},
];

const blankToUndefined = (value: string) => {
	const trimmed = value.trim();
	return trimmed ? trimmed : undefined;
};

export const slugifyRosterName = (value: string) =>
	value
		.toLowerCase()
		.trim()
		.replace(/[^a-z0-9]+/g, "-")
		.replace(/^-+|-+$/g, "");

export const createEmptyScheduleDraft = (key: string, index = 0): RosterEditorScheduleDraft => ({
	key,
	name: index === 0 ? "Primary Rotation" : `Schedule ${index + 1}`,
	timezone: "Australia/Perth",
	description: "",
	participants: [],
});

export const createEmptyRosterEditorDraft = (): RosterEditorDraft => ({
	name: "",
	slug: "",
	timezone: defaultMockSchemaData.timezone,
	chatHandle: "",
	chatChannelId: "",
	handoverTemplateId: "",
	schedules: [createEmptyScheduleDraft("schedule-1")],
});

export const mapOncallRosterToEditorDraft = (roster: OncallRoster): RosterEditorDraft => {
	const mockData = mockRosterSchemaDataBySlug[roster.attributes.slug] ?? defaultMockSchemaData;

	return {
		name: roster.attributes.name,
		slug: roster.attributes.slug,
		timezone: mockData.timezone,
		chatHandle: mockData.chatHandle,
		chatChannelId: mockData.chatChannelId,
		handoverTemplateId: roster.attributes.handoverTemplateId,
		schedules: roster.attributes.schedules.map((schedule, index) => ({
			key: schedule.id,
			id: schedule.id,
			name: mockData.schedules[schedule.id]?.name ?? `Schedule ${index + 1}`,
			timezone: schedule.attributes.timezone || mockData.timezone,
			description: schedule.attributes.description,
			participants: schedule.attributes.participants
				.toSorted((a, b) => a.order - b.order)
				.map((participant) => ({
					userId: participant.user.id,
					order: participant.order,
				})),
		})),
	};
};

export const createRosterSchemaPreview = (draft: RosterEditorDraft) => ({
	name: draft.name,
	slug: draft.slug,
	timezone: blankToUndefined(draft.timezone),
	chat_handle: blankToUndefined(draft.chatHandle),
	chat_channel_id: blankToUndefined(draft.chatChannelId),
	handover_template_id: blankToUndefined(draft.handoverTemplateId),
});

export const createScheduleSchemaPreview = (draft: RosterEditorDraft, rosterId?: string) =>
	draft.schedules.map((schedule) => ({
		name: schedule.name,
		roster_id: rosterId ?? "<new-roster-id>",
		timezone: blankToUndefined(schedule.timezone),
		participants: schedule.participants.map((participant) => ({
			user_id: participant.userId,
			index: participant.order,
		})),
	}));
