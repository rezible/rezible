import { goto } from "$app/navigation";
import {
	createIncidentMutation,
	getIncidentMetadataOptions,
	getIncidentMetadataQueryKey,
	listIncidentsQueryKey,
	type ErrorModel,
	type IncidentField,
	type IncidentSeverity,
	type IncidentType,
} from "$lib/api";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import {
	CreateIncidentFormSchema,
	getEmptyCreateIncidentForm,
	type CreateIncidentFormState,
} from "./form";

const getNamedLabel = (items: IncidentSeverity[] | IncidentType[], id: string | undefined, fallback: string) => {
	return !id ? fallback : (items.find((item) => item.id === id)?.attributes.name ?? fallback);
}

export class IncidentCreateDialogController {
	private queryClient = useQueryClient();

	open = $state(false);
	error = $state<ErrorModel>();
	form = $state<CreateIncidentFormState>(getEmptyCreateIncidentForm());

	metadataQuery = createQuery(() => ({
		...getIncidentMetadataOptions(),
		enabled: this.open,
	}));
	private metadata = $derived(this.metadataQuery.data?.data);
	loading = $derived(this.metadataQuery.isLoading);

	severities = $derived((this.metadata?.severities ?? []).sort((a, b) => a.attributes.rank - b.attributes.rank));
	types = $derived(this.metadata?.types ?? []);
	tags = $derived(this.metadata?.tags ?? []);
	fields = $derived(this.metadata?.fields ?? []);

	constructor() {
		watch(() => this.metadata, md => {
			if (!md) return;
			// set default form
		});
	}

	parsedForm = $derived(
		this.open
			? CreateIncidentFormSchema.safeParse({
					title: this.form.title,
					summary: this.form.summary,
					severityId: this.form.severityId,
					typeId: this.form.typeId,
					tagIds: this.form.tagIds,
					fieldSelections: this.form.fieldSelections,
				})
			: null,
	);

	fieldErrors = $derived(
		this.parsedForm && !this.parsedForm.success ? this.parsedForm.error.flatten().fieldErrors : {},
	);

	createMut = createMutation(() => ({
		...createIncidentMutation(),
		onSuccess: async ({ data: incident }) => {
			this.error = undefined;
			await Promise.all([
				this.queryClient.invalidateQueries({ queryKey: listIncidentsQueryKey() }),
				this.queryClient.invalidateQueries({ queryKey: getIncidentMetadataQueryKey() }),
			]);
			this.setOpen(false);
			await goto(`/incidents/${incident.attributes.slug}`);
		},
		onError: (err) => {
			this.error = err as ErrorModel;
		},
	}));

	isPending = $derived(this.createMut.isPending);
	canSubmit = $derived(!!this.parsedForm?.success && !this.isPending && !this.metadataQuery.isLoading);

	openCreate = () => {
		this.error = undefined;
		this.open = true;
	};

	setOpen = (open: boolean) => {
		this.open = open;
		if (!open) this.reset();
	};

	reset = () => {
		this.form = getEmptyCreateIncidentForm();
		this.error = undefined;
	};

	toggleTag = (tagId: string) => {
		const tagIds = new Set(this.form.tagIds);
		if (tagIds.has(tagId)) {
			tagIds.delete(tagId);
		} else {
			tagIds.add(tagId);
		}
		this.form.tagIds = [...tagIds];
	};

	setFieldSelection = (fieldId: string, optionId: string) => {
		if (!optionId) {
			const { [fieldId]: _, ...rest } = this.form.fieldSelections;
			this.form.fieldSelections = rest;
			return;
		}

		this.form.fieldSelections = {
			...this.form.fieldSelections,
			[fieldId]: optionId,
		};
	};

	getFieldSelection = (fieldId: string) => this.form.fieldSelections[fieldId] ?? "";

	getFieldSelectionLabel = (field: IncidentField) => {
		const selectedId = this.getFieldSelection(field.id);
		if (!selectedId) return "Select an option";
		return (
			field.attributes.options.find((option) => option.id === selectedId)?.attributes.value ??
			"Select an option"
		);
	};

	hasTag = (tagId: string) => this.form.tagIds.includes(tagId);

	submit = () => {
		if (!this.parsedForm?.success) return;
		this.createMut.mutate({
			body: {
				attributes: this.parsedForm.data,
			},
		});
	};

	getSeverityLabel = (severityId?: string) => getNamedLabel(this.severities, severityId, "Select severity");
	getTypeLabel = (typeId?: string) => getNamedLabel(this.types, typeId, "Select type");
}

const ctx = new Context<IncidentCreateDialogController>("IncidentCreateDialogController");
export const initIncidentCreateDialogController = () => ctx.set(new IncidentCreateDialogController());
export const useIncidentCreateDialog = () => ctx.get();
