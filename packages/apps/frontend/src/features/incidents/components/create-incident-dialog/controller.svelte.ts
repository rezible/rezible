import { goto } from "$app/navigation";
import { resolve } from "$app/paths";
import { page } from "$app/state";
import {
	createIncidentMutation,
	getIncidentMetadataOptions,
	getIncidentOptions,
	getSituationQueryKey,
	listIncidentsQueryKey,
	type ApiError,
	type Incident,
	type Situation,
} from "$lib/api";
import { useUserSessionState } from "$lib/user-session.svelte";
import { createMutation, createQueries, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";

import {
	CreateIncidentFormSchema,
	getEmptyCreateIncidentForm,
	type CreateIncidentFormState,
	type DeclareSituation,
} from "./form";

const DECLARE_PARAM = "declare";

export type ExistingIncident = { title: string; href: string };

export const declareSituationFrom = (situation: Situation): DeclareSituation => ({
	id: situation.id,
	title: situation.attributes.title,
	stage: situation.attributes.stage,
	linkedIncidentIds: situation.attributes.linkedIncidentIds,
});

/** The declare incident dialog, mounted once in the app shell. */
export class IncidentCreateDialogController {
	private queryClient = useQueryClient();
	private session = useUserSessionState();

	enabled = $derived(!!this.session.orgPreferences?.enableIncidentManagement);
	isAdmin = $derived(this.session.isAdmin);

	open = $state(false);
	error = $state<ApiError>();
	form = $state<CreateIncidentFormState>(getEmptyCreateIncidentForm());
	private submitted = $state(false);

	metadataQuery = createQuery(() => ({
		...getIncidentMetadataOptions(),
		enabled: this.open,
	}));
	metadata = $derived(this.metadataQuery.data?.data);
	metadataError = $derived(this.metadataQuery.error as ApiError | null);

	private parsedForm = $derived(this.open ? CreateIncidentFormSchema.safeParse(this.form) : null);
	titleError = $derived(this.getTitleError());

	// Incidents the chosen situations are already part of, to point out an unresolved one.
	private linkedIncidentIds = $derived([
		...new Set(this.form.situations.flatMap((situation) => situation.linkedIncidentIds)),
	]);
	private linkedIncidentQueries = createQueries(() => ({
		queries: this.linkedIncidentIds.map((id) => getIncidentOptions({ path: { id } })),
	}));

	createMut = createMutation(() => ({
		...createIncidentMutation(),
		onSuccess: async ({ data }) => {
			await this.queryClient.invalidateQueries({ queryKey: listIncidentsQueryKey() });
			if (this.form.situations.length > 0) {
				// The ID-less key matches every cached situation: a situation's own page caches it under its
				// URL value, which this dialog does not know.
				await this.queryClient.invalidateQueries({
					queryKey: getSituationQueryKey({} as Parameters<typeof getSituationQueryKey>[0]),
				});
			}
			this.open = false;
			this.reset();
			await goto(resolve("/incidents/[ref]/[[view=incidentView]]", { ref: data.reference }));
		},
		onError: (err) => {
			this.error = err as ApiError;
		},
	}));

	isPending = $derived(this.createMut.isPending);

	/** Opens the dialog, prefilled with the situation when one is given. */
	openCreate = (options?: { situation?: DeclareSituation }) => {
		this.reset();
		const situation = options?.situation;
		if (situation) {
			this.form.title = situation.title;
			this.form.situations = [situation];
		}
		this.open = true;
	};

	setOpen = (open: boolean) => {
		this.open = open;
		if (!open) {
			this.reset();
			this.clearDeclareParam();
		}
	};

	/** Removes the parameter that asked for the dialog, so it doesn't open again. */
	clearDeclareParam = async () => {
		if (!page.url.searchParams.has(DECLARE_PARAM)) {
			return;
		}
		const url = new URL(page.url);
		url.searchParams.delete(DECLARE_PARAM);
		await goto(url, { replaceState: true, noScroll: true, keepFocus: true });
	};

	addSituation = (situation: Situation) => {
		if (this.form.situations.some((chosen) => chosen.id === situation.id)) {
			return;
		}
		this.form.situations = [...this.form.situations, declareSituationFrom(situation)];
	};

	removeSituation = (situationId: string) => {
		this.form.situations = this.form.situations.filter((situation) => situation.id !== situationId);
	};

	/** The newest unresolved incident the situation is already part of, if any. Equal open times are ordered by ID. */
	existingIncidentFor(situation: DeclareSituation): ExistingIncident | undefined {
		let newest: Incident | undefined;
		for (const query of this.linkedIncidentQueries) {
			const incident = query.data?.data;
			if (!incident || incident.attributes.responseState === "resolved") {
				continue;
			}
			if (!situation.linkedIncidentIds.includes(incident.id)) {
				continue;
			}
			if (!newest || this.openedLater(incident, newest)) {
				newest = incident;
			}
		}
		if (!newest) {
			return undefined;
		}
		const href = resolve("/incidents/[ref]/[[view=incidentView]]", { ref: newest.attributes.reference });
		return { title: newest.attributes.title, href };
	}

	private openedLater(incident: Incident, than: Incident) {
		const byTime = Date.parse(incident.attributes.openedAt) - Date.parse(than.attributes.openedAt);
		if (byTime !== 0) {
			return byTime > 0;
		}
		return incident.id > than.id;
	}

	submit = () => {
		this.error = undefined;
		this.submitted = true;
		if (!this.parsedForm?.success) {
			return;
		}
		this.createMut.mutate({ body: { attributes: this.parsedForm.data } });
	};

	private reset() {
		this.form = getEmptyCreateIncidentForm();
		this.error = undefined;
		this.submitted = false;
	}

	private getTitleError() {
		if (!this.submitted || !this.parsedForm || this.parsedForm.success) {
			return undefined;
		}
		return this.parsedForm.error.flatten().fieldErrors.title?.[0];
	}
}

const ctx = new Context<IncidentCreateDialogController>("IncidentCreateDialogController");
export const initIncidentCreateDialogController = () => ctx.set(new IncidentCreateDialogController());
export const useIncidentCreateDialog = () => ctx.get();
