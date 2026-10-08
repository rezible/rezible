import {
	createSystemAnalysisEntryMutation,
	updateSystemAnalysisEntryMutation,
	type ApiError,
	type SystemAnalysisEntry,
} from "$lib/api";
import { createMutation } from "@tanstack/svelte-query";
import { entryInput } from "./entry-input";
import { Context, watch } from "runed";
import {
	fromDate,
	getLocalTimeZone,
	now,
	parseAbsoluteToLocal,
	type ZonedDateTime,
} from "@internationalized/date";
import { useSystemAnalysisController } from "$components/system-analysis";
import { useIncidentView } from "$features/incidents/views/incident";

export class IncidentEventDialogController {
	analysis = useSystemAnalysisController();
	incident = useIncidentView();
	open = $state(false);
	canEdit = $derived(this.incident.documentAccess?.canEdit ?? false);
	editingEntry = $state.raw<SystemAnalysisEntry>();
	title = $state("");
	kind = $state<SystemAnalysisEntry["attributes"]["kind"]>("observation");
	body = $state("");
	timestamp = $state<ZonedDateTime>(now(getLocalTimeZone()));
	hasTimestamp = $state(true);
	error = $state<ApiError>();
	private onChanged: () => unknown;
	private returnFocus?: HTMLElement;
	createEntryMut = createMutation(() => createSystemAnalysisEntryMutation());
	updateEntryMut = createMutation(() => updateSystemAnalysisEntryMutation());
	loading = $derived(this.createEntryMut.isPending || this.updateEntryMut.isPending);

	constructor(onChanged: () => unknown) {
		this.onChanged = onChanged;
		watch(
			() => this.incident.incidentId,
			() => {
				this.open = false;
			}
		);
	}

	setCreating = (attributes?: { timestamp?: string }) => {
		if (!this.canEdit) return;
		this.editingEntry = undefined;
		this.title = "";
		this.kind = "observation";
		this.body = "";
		const time = attributes?.timestamp ?? this.incident.incident?.attributes.openedAt;
		this.timestamp = time ? parseAbsoluteToLocal(time) : now(getLocalTimeZone());
		this.hasTimestamp = true;
		this.show();
	};
	setEditing = (entry: SystemAnalysisEntry) => {
		if (!this.canEdit) return;
		this.editingEntry = entry;
		this.title = entry.attributes.title;
		this.kind = entry.attributes.kind;
		this.body = entry.attributes.body ?? "";
		this.hasTimestamp = !!entry.attributes.occurredAt;
		this.timestamp = entry.attributes.occurredAt
			? parseAbsoluteToLocal(entry.attributes.occurredAt)
			: fromDate(new Date(), getLocalTimeZone());
		this.show();
	};
	private show() {
		this.error = undefined;
		this.returnFocus = document.activeElement instanceof HTMLElement ? document.activeElement : undefined;
		this.open = true;
	}
	clear = () => {
		if (!this.loading) this.open = false;
	};
	restoreFocus = () => {
		this.returnFocus?.focus({ preventScroll: true });
	};

	confirm = async () => {
		if (!this.canEdit || this.loading || !this.title.trim() || !this.analysis.analysisId) return;
		this.error = undefined;
		const analysisId = this.analysis.analysisId;
		const attributes = entryInput(
			{
				title: this.title.trim(),
				kind: this.kind,
				body: this.body,
				occurredAt: this.hasTimestamp ? this.timestamp.toAbsoluteString() : null,
			},
			this.editingEntry
		);
		try {
			if (this.editingEntry) {
				await this.updateEntryMut.mutateAsync({
					path: { id: this.editingEntry.id },
					body: { attributes },
				});
			} else {
				await this.createEntryMut.mutateAsync({
					path: { id: this.analysis.analysisId },
					body: { attributes },
				});
			}
			if (analysisId === this.analysis.analysisId) {
				void this.onChanged();
				this.open = false;
			}
		} catch (error) {
			if (analysisId === this.analysis.analysisId) this.error = error as ApiError;
		}
	};
}

const ctx = new Context<IncidentEventDialogController>("IncidentEventDialogController");
export const initEventDialog = (onChanged: () => unknown) =>
	ctx.set(new IncidentEventDialogController(onChanged));
export const useEventDialog = () => ctx.get();
