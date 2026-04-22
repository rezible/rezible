import { getOncallRosterOptions, listUsersOptions, type OncallRoster, type User } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import {
	createEmptyRosterEditorDraft,
	createEmptyScheduleDraft,
	createRosterSchemaPreview,
	createScheduleSchemaPreview,
	mapOncallRosterToEditorDraft,
	mockBackedRosterEditorFields,
	rosterEditorTimezoneOptions,
	slugifyRosterName,
	type RosterEditorDraft,
} from "./mock_data";

const serializeDraft = (draft: RosterEditorDraft) =>
	JSON.stringify({
		...draft,
		schedules: draft.schedules.map((schedule) => ({
			...schedule,
			participants: schedule.participants.toSorted((a, b) => a.order - b.order),
		})),
	});

export class OncallRosterEditorViewController {
	rosterSlug = $state<string | undefined>();
	draft = $state<RosterEditorDraft>(createEmptyRosterEditorDraft());
	selectedScheduleKey = $state<string>("schedule-1");
	pendingParticipantUserId = $state<string>("");
	slugLocked = $state(false);
	private initialSnapshot = $state("");
	private nextScheduleNumber = $state(2);
	private queryBackedDraftLoaded = $state(false);

	constructor(slugFn: Getter<string | undefined>) {
		this.resetDraft();
		const syncSlug = (slug?: string) => {
			this.rosterSlug = slug || undefined;
			this.slugLocked = !!slug;
			this.queryBackedDraftLoaded = false;
			this.resetDraft();
		};

		syncSlug(slugFn());
		watch(slugFn, syncSlug);
		watch(
			() => this.roster,
			(roster) => {
				if (this.mode === "edit" && roster && !this.queryBackedDraftLoaded) {
					this.loadDraftFromRoster(roster);
				}
			}
		);
	}

	private rosterQuery = createQuery(() => ({
		...getOncallRosterOptions({ path: { id: this.rosterSlug ?? "" } }),
		enabled: !!this.rosterSlug,
	}));

	private usersQuery = createQuery(() =>
		listUsersOptions({
			query: {
				archived: false,
				limit: 200,
			},
		})
	);

	mode = $derived<"create" | "edit">(this.rosterSlug ? "edit" : "create");
	roster = $derived(this.rosterQuery.data?.data);
	users = $derived(this.usersQuery.data?.data ?? []);
	loading = $derived(this.mode === "edit" && this.rosterQuery.isLoading);
	rosterId = $derived(this.roster?.id);

	title = $derived(this.mode === "edit" ? `Edit ${this.draft.name || "Roster"}` : "Create Oncall Roster");
	subtitle = $derived(
		this.mode === "edit"
			? "Drafting an edit flow over the existing roster query shape."
			: "Drafting a create flow mapped to the backend oncall roster and schedule schemas."
	);
	submitLabel = $derived(this.mode === "edit" ? "Update API Pending" : "Create API Pending");
	hasUnsavedChanges = $derived(serializeDraft(this.draft) !== this.initialSnapshot);
	scheduleCountLabel = $derived(
		`${this.draft.schedules.length} schedule${this.draft.schedules.length === 1 ? "" : "s"}`
	);

	selectedSchedule = $derived(
		this.draft.schedules.find((schedule) => schedule.key === this.selectedScheduleKey) ??
			this.draft.schedules[0]
	);
	userMap = $derived.by(() => new Map(this.users.map((user) => [user.id, user])));
	availableUsersForSelectedSchedule = $derived.by(() => {
		const selectedSchedule = this.selectedSchedule;
		if (!selectedSchedule) return [];

		const assignedUserIds = new Set(
			selectedSchedule.participants.map((participant) => participant.userId)
		);
		return this.users.filter((user) => !assignedUserIds.has(user.id));
	});

	rosterSchemaPreview = $derived(createRosterSchemaPreview(this.draft));
	scheduleSchemaPreview = $derived(createScheduleSchemaPreview(this.draft, this.rosterId));

	timezoneOptions = rosterEditorTimezoneOptions;
	mockBackedFields = mockBackedRosterEditorFields;

	private resetDraft() {
		const nextDraft = createEmptyRosterEditorDraft();
		this.draft = nextDraft;
		this.selectedScheduleKey = nextDraft.schedules[0]?.key ?? "";
		this.pendingParticipantUserId = "";
		this.nextScheduleNumber = nextDraft.schedules.length + 1;
		this.initialSnapshot = serializeDraft(nextDraft);
	}

	private loadDraftFromRoster(roster: OncallRoster) {
		const nextDraft = mapOncallRosterToEditorDraft(roster);
		this.draft = nextDraft;
		this.selectedScheduleKey = nextDraft.schedules[0]?.key ?? "";
		this.pendingParticipantUserId = "";
		this.nextScheduleNumber = nextDraft.schedules.length + 1;
		this.initialSnapshot = serializeDraft(nextDraft);
		this.queryBackedDraftLoaded = true;
	}

	private updateDraft(updater: (draft: RosterEditorDraft) => void) {
		const nextDraft = structuredClone(this.draft);
		updater(nextDraft);
		this.draft = nextDraft;
		if (!nextDraft.schedules.some((schedule) => schedule.key === this.selectedScheduleKey)) {
			this.selectedScheduleKey = nextDraft.schedules[0]?.key ?? "";
		}
	}

	setRosterName(name: string) {
		this.updateDraft((draft) => {
			draft.name = name;
			if (!this.slugLocked) {
				draft.slug = slugifyRosterName(name);
			}
		});
	}

	setRosterSlug(slug: string) {
		this.slugLocked = true;
		this.updateDraft((draft) => {
			draft.slug = slugifyRosterName(slug);
		});
	}

	setRosterTimezone(timezone: string) {
		this.updateDraft((draft) => {
			draft.timezone = timezone;
		});
	}

	setRosterChatHandle(chatHandle: string) {
		this.updateDraft((draft) => {
			draft.chatHandle = chatHandle;
		});
	}

	setRosterChatChannelId(chatChannelId: string) {
		this.updateDraft((draft) => {
			draft.chatChannelId = chatChannelId;
		});
	}

	setRosterHandoverTemplateId(handoverTemplateId: string) {
		this.updateDraft((draft) => {
			draft.handoverTemplateId = handoverTemplateId;
		});
	}

	selectSchedule(key: string) {
		this.selectedScheduleKey = key;
		this.pendingParticipantUserId = "";
	}

	addSchedule() {
		const nextSchedule = createEmptyScheduleDraft(
			`schedule-${this.nextScheduleNumber}`,
			this.draft.schedules.length
		);
		nextSchedule.timezone = this.draft.timezone || nextSchedule.timezone;
		this.nextScheduleNumber += 1;
		this.updateDraft((draft) => {
			draft.schedules.push(nextSchedule);
		});
		this.selectedScheduleKey = nextSchedule.key;
	}

	removeSelectedSchedule() {
		if (this.draft.schedules.length <= 1 || !this.selectedSchedule) return;
		const selectedScheduleKey = this.selectedSchedule.key;
		this.updateDraft((draft) => {
			draft.schedules = draft.schedules.filter((schedule) => schedule.key !== selectedScheduleKey);
		});
	}

	setSelectedScheduleName(name: string) {
		if (!this.selectedSchedule) return;
		const selectedScheduleKey = this.selectedSchedule.key;
		this.updateDraft((draft) => {
			const schedule = draft.schedules.find((item) => item.key === selectedScheduleKey);
			if (!schedule) return;
			schedule.name = name;
		});
	}

	setSelectedScheduleTimezone(timezone: string) {
		if (!this.selectedSchedule) return;
		const selectedScheduleKey = this.selectedSchedule.key;
		this.updateDraft((draft) => {
			const schedule = draft.schedules.find((item) => item.key === selectedScheduleKey);
			if (!schedule) return;
			schedule.timezone = timezone;
		});
	}

	setSelectedScheduleDescription(description: string) {
		if (!this.selectedSchedule) return;
		const selectedScheduleKey = this.selectedSchedule.key;
		this.updateDraft((draft) => {
			const schedule = draft.schedules.find((item) => item.key === selectedScheduleKey);
			if (!schedule) return;
			schedule.description = description;
		});
	}

	addSelectedScheduleParticipant(userId: string) {
		if (!this.selectedSchedule || !userId) return;
		const selectedScheduleKey = this.selectedSchedule.key;
		this.updateDraft((draft) => {
			const schedule = draft.schedules.find((item) => item.key === selectedScheduleKey);
			if (!schedule || schedule.participants.some((participant) => participant.userId === userId))
				return;
			schedule.participants.push({
				userId,
				order: schedule.participants.length,
			});
		});
		this.pendingParticipantUserId = "";
	}

	removeSelectedScheduleParticipant(userId: string) {
		if (!this.selectedSchedule) return;
		const selectedScheduleKey = this.selectedSchedule.key;
		this.updateDraft((draft) => {
			const schedule = draft.schedules.find((item) => item.key === selectedScheduleKey);
			if (!schedule) return;
			schedule.participants = schedule.participants
				.filter((participant) => participant.userId !== userId)
				.map((participant, index) => ({
					...participant,
					order: index,
				}));
		});
	}

	moveSelectedScheduleParticipant(userId: string, direction: -1 | 1) {
		if (!this.selectedSchedule) return;
		const selectedScheduleKey = this.selectedSchedule.key;
		this.updateDraft((draft) => {
			const schedule = draft.schedules.find((item) => item.key === selectedScheduleKey);
			if (!schedule) return;

			const idx = schedule.participants.findIndex((participant) => participant.userId === userId);
			const nextIdx = idx + direction;
			if (idx < 0 || nextIdx < 0 || nextIdx >= schedule.participants.length) return;

			const [participant] = schedule.participants.splice(idx, 1);
			schedule.participants.splice(nextIdx, 0, participant);
			schedule.participants = schedule.participants.map((item, index) => ({
				...item,
				order: index,
			}));
		});
	}

	getUser(userId: string): User | undefined {
		return this.userMap.get(userId);
	}
}

const ctx = new Context<OncallRosterEditorViewController>("OncallRosterEditorViewController");
export const initOncallRosterEditorViewController = (slugFn: Getter<string | undefined>) =>
	ctx.set(new OncallRosterEditorViewController(slugFn));
export const useOncallRosterEditorViewController = () => ctx.get();
