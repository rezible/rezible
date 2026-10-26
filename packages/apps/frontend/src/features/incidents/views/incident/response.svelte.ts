import {
	getLocalTimeZone,
	now,
	parseDateTime,
	parseAbsolute,
	toCalendarDateTime,
	toZoned,
} from "@internationalized/date";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";
import { toast } from "svelte-sonner";

import { goto } from "$app/navigation";
import { resolve } from "$app/paths";
import {
	archiveIncidentMutation,
	createIncidentMilestoneMutation,
	createIncidentRoleAssignmentMutation,
	deleteIncidentRoleAssignmentMutation,
	errorDisplay,
	getIncidentMetadataOptions,
	getIncidentQueryKey,
	getSituationQueryKey,
	linkIncidentSituationMutation,
	listIncidentsQueryKey,
	unlinkIncidentSituationMutation,
	updateIncidentMilestoneMutation,
	updateIncidentMutation,
	type ApiError,
	type IncidentMilestone,
	type IncidentMilestoneAttributes,
	type Situation,
} from "$lib/api";
import { useUserSessionState } from "$lib/user-session.svelte";

import {
	detailsFormFromIncident,
	getEmptyDetailsForm,
	UpdateIncidentFormSchema,
	type IncidentDetailsFormState,
} from "$features/incidents/components/create-incident-dialog/form";
import { responseActions, type ResponseAction } from "$features/incidents/lib/response";

import { useIncidentView } from "./controller.svelte";

type MilestoneKind = IncidentMilestoneAttributes["kind"];

/** Kinds that can be recorded from the milestone dialog. Resolving uses its own confirmation. */
export const recordableMilestoneKinds: { kind: MilestoneKind; label: string }[] = [
	{ kind: "impact", label: "Impact" },
	{ kind: "detection", label: "Detection" },
	{ kind: "investigation", label: "Investigation" },
	{ kind: "mitigation", label: "Mitigation" },
];

export type MilestoneDialogState = {
	title: string;
	/** Editing an existing milestone, whose kind can't change. */
	milestoneId?: string;
	/** Whether the kind can be chosen. */
	chooseKind: boolean;
	kind: MilestoneKind;
	/** Local date and time, as a datetime-local input holds it, to the minute. */
	time: string;
	/** The time the dialog opened with. The time is sent only when the person changed it, so an unchanged
	 * new milestone is recorded now and an unchanged edit keeps its exact time. */
	initialTime: string;
	description: string;
	placeholder: string;
};

export type RoleHolder = { assignmentId: string; userId: string; name: string };

export type RoleGroup = {
	roleId: string;
	name: string;
	countLabel: string;
	full: boolean;
	fullReason: string;
	holders: RoleHolder[];
	archived: boolean;
};

const localNow = () => toCalendarDateTime(now(getLocalTimeZone())).toString().slice(0, 16);

const localTime = (iso: string) =>
	toCalendarDateTime(toZoned(parseAbsolute(iso, getLocalTimeZone()), getLocalTimeZone()))
		.toString()
		.slice(0, 16);

const isoFromLocal = (time: string) => toZoned(parseDateTime(time), getLocalTimeZone()).toAbsoluteString();

/** The incident page's writes: details, milestones, roles, linked situations and archiving. */
export class IncidentResponseController {
	private view = useIncidentView();
	private queryClient = useQueryClient();
	private session = useUserSessionState();

	enabled = $derived(!!this.session.orgPreferences?.enableIncidentManagement);
	isAdmin = $derived(this.session.isAdmin);

	private attributes = $derived(this.view.incident?.attributes);
	actions = $derived<ResponseAction[]>(
		this.attributes ? responseActions(this.attributes.responseState, this.enabled) : []
	);

	metadataQuery = createQuery(() => getIncidentMetadataOptions());
	private metadata = $derived(this.metadataQuery.data?.data);
	roleGroups = $derived(this.buildRoleGroups());
	hasRoles = $derived(this.roleGroups.length > 0);

	linkedSituationIds = $derived((this.attributes?.situations ?? []).map((situation) => situation.id));
	linkSituationOpen = $state(false);

	milestoneDialog = $state<MilestoneDialogState>();
	milestoneError = $state<ApiError>();

	detailsOpen = $state(false);
	detailsForm = $state<IncidentDetailsFormState>(getEmptyDetailsForm());
	detailsError = $state<ApiError>();
	incidentMetadata = $derived(this.metadata);
	private parsedDetails = $derived(UpdateIncidentFormSchema.safeParse(this.detailsForm));
	detailsTitleError = $derived(this.getDetailsTitleError());

	archiveOpen = $state(false);
	archiveError = $state<ApiError>();

	private updateMut = createMutation(() => ({
		...updateIncidentMutation(),
		onSuccess: async () => {
			this.detailsOpen = false;
			await this.refreshIncident();
		},
		onError: (err) => {
			this.detailsError = err as ApiError;
		},
	}));
	detailsPending = $derived(this.updateMut.isPending);

	private createMilestoneMut = createMutation(() => ({
		...createIncidentMilestoneMutation(),
		onSuccess: () => this.onMilestoneSaved(),
		onError: (err) => {
			this.milestoneError = err as ApiError;
		},
	}));
	private updateMilestoneMut = createMutation(() => ({
		...updateIncidentMilestoneMutation(),
		onSuccess: () => this.onMilestoneSaved(),
		onError: (err) => {
			this.milestoneError = err as ApiError;
		},
	}));
	milestonePending = $derived(this.createMilestoneMut.isPending || this.updateMilestoneMut.isPending);

	private assignMut = createMutation(() => ({
		...createIncidentRoleAssignmentMutation(),
		onSuccess: () => this.refreshIncident(),
		onError: (err) => this.showError(err as ApiError),
	}));
	private unassignMut = createMutation(() => ({
		...deleteIncidentRoleAssignmentMutation(),
		onSuccess: () => this.refreshIncident(),
		onError: (err) => this.showError(err as ApiError),
	}));

	private linkMut = createMutation(() => ({
		...linkIncidentSituationMutation(),
		onError: (err) => this.showError(err as ApiError),
	}));
	private unlinkMut = createMutation(() => ({
		...unlinkIncidentSituationMutation(),
		onError: (err) => this.showError(err as ApiError),
	}));

	private archiveMut = createMutation(() => ({
		...archiveIncidentMutation(),
		onSuccess: async () => {
			this.archiveOpen = false;
			await this.queryClient.invalidateQueries({ queryKey: listIncidentsQueryKey() });
			toast.success("Incident archived");
			await goto(resolve("/incidents"));
		},
		onError: (err) => {
			this.archiveError = err as ApiError;
		},
	}));
	archivePending = $derived(this.archiveMut.isPending);

	openResponseAction = (action: ResponseAction) => {
		switch (action) {
			case "mitigated":
				this.openMilestoneDialog({
					title: "Record mitigation",
					kind: "mitigation",
					placeholder: "What changed?",
				});
				break;
			case "mitigationFailed":
				this.openMilestoneDialog({
					title: "Mitigation failed",
					kind: "impact",
					placeholder: "What happened?",
				});
				break;
			case "resolve":
				this.openMilestoneDialog({ title: "Resolve incident?", kind: "resolution", placeholder: "" });
				break;
		}
	};

	openRecordMilestone = () => {
		this.openMilestoneDialog({
			title: "Record milestone",
			kind: "impact",
			placeholder: "",
			chooseKind: true,
		});
	};

	openEditMilestone = (milestone: IncidentMilestone) => {
		this.milestoneError = undefined;
		this.milestoneDialog = {
			title: "Edit milestone",
			milestoneId: milestone.id,
			chooseKind: false,
			kind: milestone.attributes.kind,
			time: localTime(milestone.attributes.timestamp),
			initialTime: localTime(milestone.attributes.timestamp),
			description: milestone.attributes.description,
			placeholder: "",
		};
	};

	/** Clicking the chosen kind again would clear it; a milestone always has a kind. */
	setMilestoneKind = (kind: string) => {
		const option = recordableMilestoneKinds.find((candidate) => candidate.kind === kind);
		if (this.milestoneDialog && option) {
			this.milestoneDialog.kind = option.kind;
		}
	};

	closeMilestoneDialog = () => {
		this.milestoneDialog = undefined;
		this.milestoneError = undefined;
	};

	saveMilestone = () => {
		const dialog = this.milestoneDialog;
		if (!dialog || !this.view.incidentId) {
			return;
		}
		this.milestoneError = undefined;
		let timestamp: string | undefined;
		if (dialog.time !== dialog.initialTime) {
			timestamp = isoFromLocal(dialog.time);
		}
		if (dialog.milestoneId) {
			this.updateMilestoneMut.mutate({
				path: { id: dialog.milestoneId },
				body: { attributes: { timestamp, description: dialog.description } },
			});
			return;
		}
		this.createMilestoneMut.mutate({
			path: { id: this.view.incidentId },
			body: { attributes: { kind: dialog.kind, timestamp, description: dialog.description } },
		});
	};

	openEditDetails = () => {
		if (!this.view.incident) {
			return;
		}
		this.detailsForm = detailsFormFromIncident(this.view.incident);
		this.detailsError = undefined;
		this.detailsOpen = true;
	};

	saveDetails = () => {
		if (!this.parsedDetails.success || !this.view.incidentId) {
			return;
		}
		this.detailsError = undefined;
		this.updateMut.mutate({
			path: { id: this.view.incidentId },
			body: { attributes: this.parsedDetails.data },
		});
	};

	assignRole = (roleId: string, userId: string) => {
		if (!this.view.incidentId) {
			return;
		}
		this.assignMut.mutate({
			path: { id: this.view.incidentId },
			body: { attributes: { roleId, userId } },
		});
	};

	unassignRole = (assignmentId: string) => {
		this.unassignMut.mutate({ path: { id: assignmentId } });
	};

	openLinkSituation = () => {
		this.linkSituationOpen = true;
	};

	linkSituation = (situation: Situation) => {
		this.linkSituationOpen = false;
		if (!this.view.incidentId) {
			return;
		}
		const variables = {
			path: { id: this.view.incidentId },
			body: { attributes: { situationId: situation.id } },
		};
		this.linkMut.mutate(variables, {
			onSuccess: async () => {
				await this.refreshSituation();
				if (situation.attributes.stage === "closed") {
					toast.success("Linked. The situation stays closed.");
				}
			},
		});
	};

	unlinkSituation = (situationId: string) => {
		if (!this.view.incidentId) {
			return;
		}
		const variables = { path: { id: this.view.incidentId }, query: { situationId } };
		this.unlinkMut.mutate(variables, {
			onSuccess: () => this.refreshSituation(),
		});
	};

	openArchive = () => {
		this.archiveError = undefined;
		this.archiveOpen = true;
	};

	archive = () => {
		if (!this.view.incidentId) {
			return;
		}
		this.archiveMut.mutate({ path: { id: this.view.incidentId } });
	};

	private openMilestoneDialog(options: {
		title: string;
		kind: MilestoneKind;
		placeholder: string;
		chooseKind?: boolean;
	}) {
		const openedAt = localNow();
		this.milestoneError = undefined;
		this.milestoneDialog = {
			title: options.title,
			chooseKind: options.chooseKind ?? false,
			kind: options.kind,
			time: openedAt,
			initialTime: openedAt,
			description: "",
			placeholder: options.placeholder,
		};
	}

	private async onMilestoneSaved() {
		this.closeMilestoneDialog();
		await this.refreshIncident();
	}

	private async refreshIncident() {
		await this.queryClient.invalidateQueries({
			queryKey: getIncidentQueryKey({ path: { id: this.view.ref } }),
		});
	}

	private async refreshSituation() {
		await Promise.all([
			this.refreshIncident(),
			// The ID-less key matches every cached situation: a situation's own page caches it under its URL
			// value, which this page does not know.
			this.queryClient.invalidateQueries({
				queryKey: getSituationQueryKey({} as Parameters<typeof getSituationQueryKey>[0]),
			}),
		]);
	}

	private showError(error: ApiError) {
		const display = errorDisplay(error);
		toast.error(display.title, { description: display.detail });
	}

	private getDetailsTitleError() {
		if (this.parsedDetails.success) {
			return undefined;
		}
		return this.parsedDetails.error.flatten().fieldErrors.title?.[0];
	}

	/**
	 * One group per role, in settings order, with who holds it and whether it is full. An archived role
	 * appears only while someone still holds it.
	 */
	private buildRoleGroups(): RoleGroup[] {
		const assignments = this.attributes?.roles ?? [];
		const groups: RoleGroup[] = [];
		for (const role of this.metadata?.roles ?? []) {
			const holders: RoleHolder[] = [];
			for (const assignment of assignments) {
				if (assignment.attributes.role.id !== role.id) {
					continue;
				}
				holders.push({
					assignmentId: assignment.id,
					userId: assignment.attributes.user.id,
					name: assignment.attributes.user.attributes?.name ?? "Unknown user",
				});
			}
			const archived = role.attributes.archived;
			if (archived && holders.length === 0) {
				continue;
			}
			const max = role.attributes.maxAssignees;
			const name = role.attributes.name;
			let countLabel = `${holders.length}`;
			let full = false;
			let fullReason = "";
			if (max !== undefined) {
				countLabel = `${holders.length}/${max}`;
				full = holders.length >= max;
				fullReason = `${name}: ${holders.length} of ${max} assigned. Remove someone to assign another.`;
			}
			groups.push({ roleId: role.id, name, countLabel, full, fullReason, holders, archived });
		}
		return groups;
	}
}

const ctx = new Context<IncidentResponseController>("IncidentResponseController");
export const initIncidentResponseController = () => ctx.set(new IncidentResponseController());
export const useIncidentResponse = () => ctx.get();
