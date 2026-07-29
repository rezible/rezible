import {
	createIncidentFieldMutation,
	createIncidentRoleMutation,
	createIncidentSeverityMutation,
	createIncidentTagMutation,
	createIncidentTypeMutation,
	getIncidentMetadataOptions,
	getIncidentMetadataQueryKey,
	getUserSessionOptions,
	type ErrorModel,
	type IncidentField,
	type IncidentRole,
	type IncidentSeverity,
	type IncidentTag,
	type IncidentType,
	updateIncidentFieldMutation,
	updateIncidentRoleMutation,
	updateIncidentSeverityMutation,
	updateIncidentTagMutation,
	updateIncidentTypeMutation,
	updateOrganizationPreferencesMutation,
} from "$lib/api";
import { useUserSessionState } from "$lib/user-session.svelte";
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";
import { SvelteMap } from "svelte/reactivity";

type MetadataKind = "severities" | "types" | "roles" | "tags" | "fields";

export class IncidentSettingsController {
	session = useUserSessionState();
	integrations = useIntegrationsController();
	private queryClient = useQueryClient();

	private metadataQuery = createQuery(() => getIncidentMetadataOptions({ query: { archived: true } }));
	private metadata = $derived(this.metadataQuery.data?.data);

	severities = $derived(this.metadata?.severities ?? []);
	types = $derived(this.metadata?.types ?? []);
	roles = $derived(this.metadata?.roles ?? []);
	tags = $derived(this.metadata?.tags ?? []);
	fields = $derived(this.metadata?.fields ?? []);

	loading = $derived(this.metadataQuery.isPending);
	error = $derived(this.metadataQuery.error as ErrorModel | null);
	saveError = $state<ErrorModel>();

	incidentIntegrationInstalled = $derived(
		this.integrations.installed.some((intg) =>
			intg.attributes.capabilities.includes("incident_management")
		)
	);
	canEdit = $derived(this.session.isAdmin && !this.incidentIntegrationInstalled);
	showEnableToggle = $derived(this.session.isAdmin && !this.incidentIntegrationInstalled);
	incidentManagementEnabled = $derived(
		Boolean(this.session.orgPreferences?.enableIncidentManagement) || this.incidentIntegrationInstalled
	);

	newSeverity = $state({ name: "", rank: 1, color: "#ef4444", description: "" });
	newType = $state("");
	newRole = $state({ name: "", required: false });
	newTag = $state("");
	newField = $state({ name: "", options: "" });
	fieldOptionTextEdits = new SvelteMap<string, string>();

	private async invalidateMetadata() {
		await this.queryClient.invalidateQueries({ queryKey: getIncidentMetadataQueryKey() });
	}

	private mutationOptions() {
		return {
			onSuccess: async () => {
				this.saveError = undefined;
				await this.invalidateMetadata();
			},
			onError: (err: ErrorModel) => {
				this.saveError = err;
			},
		};
	}

	private updateOrgPrefsMut = createMutation(() => ({
		...updateOrganizationPreferencesMutation(),
		onSuccess: async () => {
			this.saveError = undefined;
			await this.queryClient.invalidateQueries({ queryKey: getUserSessionOptions().queryKey });
			this.session.refetch();
		},
		onError: (err) => {
			this.saveError = err;
		},
	}));

	private createSeverityMut = createMutation(() => ({
		...createIncidentSeverityMutation(),
		...this.mutationOptions(),
	}));
	private updateSeverityMut = createMutation(() => ({
		...updateIncidentSeverityMutation(),
		...this.mutationOptions(),
	}));
	private createTypeMut = createMutation(() => ({
		...createIncidentTypeMutation(),
		...this.mutationOptions(),
	}));
	private updateTypeMut = createMutation(() => ({
		...updateIncidentTypeMutation(),
		...this.mutationOptions(),
	}));
	private createRoleMut = createMutation(() => ({
		...createIncidentRoleMutation(),
		...this.mutationOptions(),
	}));
	private updateRoleMut = createMutation(() => ({
		...updateIncidentRoleMutation(),
		...this.mutationOptions(),
	}));
	private createTagMut = createMutation(() => ({
		...createIncidentTagMutation(),
		...this.mutationOptions(),
	}));
	private updateTagMut = createMutation(() => ({
		...updateIncidentTagMutation(),
		...this.mutationOptions(),
	}));
	private createFieldMut = createMutation(() => ({
		...createIncidentFieldMutation(),
		...this.mutationOptions(),
	}));
	private updateFieldMut = createMutation(() => ({
		...updateIncidentFieldMutation(),
		...this.mutationOptions(),
	}));

	saving = $derived(
		this.updateOrgPrefsMut.isPending ||
			this.createSeverityMut.isPending ||
			this.updateSeverityMut.isPending ||
			this.createTypeMut.isPending ||
			this.updateTypeMut.isPending ||
			this.createRoleMut.isPending ||
			this.updateRoleMut.isPending ||
			this.createTagMut.isPending ||
			this.updateTagMut.isPending ||
			this.createFieldMut.isPending ||
			this.updateFieldMut.isPending
	);

	setIncidentManagementEnabled(enabled: boolean) {
		const orgId = this.session.org?.id;
		if (!orgId || !this.showEnableToggle) return;
		this.updateOrgPrefsMut.mutate({
			path: { id: orgId },
			body: { attributes: { enableIncidentManagement: enabled } },
		});
	}

	createSeverity() {
		if (!this.canEdit || !this.newSeverity.name.trim()) return;
		this.createSeverityMut.mutate({
			body: { attributes: { ...this.newSeverity, name: this.newSeverity.name.trim() } },
		});
		this.newSeverity = { name: "", rank: this.newSeverity.rank + 1, color: "#ef4444", description: "" };
	}

	updateSeverity(item: IncidentSeverity) {
		if (!this.canEdit) return;
		this.updateSeverityMut.mutate({
			path: { id: item.id },
			body: { attributes: { ...item.attributes } },
		});
	}

	createType() {
		if (!this.canEdit || !this.newType.trim()) return;
		this.createTypeMut.mutate({ body: { attributes: { name: this.newType.trim() } } });
		this.newType = "";
	}

	updateType(item: IncidentType) {
		if (!this.canEdit) return;
		this.updateTypeMut.mutate({
			path: { id: item.id },
			body: { attributes: { name: item.attributes.name, archived: item.attributes.archived } },
		});
	}

	createRole() {
		if (!this.canEdit || !this.newRole.name.trim()) return;
		this.createRoleMut.mutate({
			body: { attributes: { name: this.newRole.name.trim(), required: this.newRole.required } },
		});
		this.newRole = { name: "", required: false };
	}

	updateRole(item: IncidentRole) {
		if (!this.canEdit) return;
		this.updateRoleMut.mutate({
			path: { id: item.id },
			body: {
				attributes: {
					name: item.attributes.name,
					required: item.attributes.required,
					archived: item.attributes.archived,
				},
			},
		});
	}

	createTag() {
		if (!this.canEdit || !this.newTag.trim()) return;
		this.createTagMut.mutate({ body: { attributes: { value: this.newTag.trim() } } });
		this.newTag = "";
	}

	updateTag(item: IncidentTag) {
		if (!this.canEdit) return;
		this.updateTagMut.mutate({
			path: { id: item.id },
			body: { attributes: { value: item.attributes.value, archived: item.attributes.archived } },
		});
	}

	private fieldOptionsFromText(options: string) {
		return options
			.split(",")
			.map((value) => value.trim())
			.filter(Boolean);
	}

	fieldOptionsText(item: IncidentField) {
		return (
			this.fieldOptionTextEdits.get(item.id) ??
			item.attributes.options.map((opt) => opt.attributes.value).join(", ")
		);
	}

	setFieldOptionsText(id: string, value: string) {
		this.fieldOptionTextEdits.set(id, value);
	}

	createField() {
		const options = this.fieldOptionsFromText(this.newField.options);
		if (!this.canEdit || !this.newField.name.trim() || options.length === 0) return;
		this.createFieldMut.mutate({
			body: {
				attributes: {
					name: this.newField.name.trim(),
					required: false,
					options: options.map((value) => ({ fieldOptionType: "custom", value })),
				},
			},
		});
		this.newField = { name: "", options: "" };
	}

	updateField(item: IncidentField, optionsText: string) {
		if (!this.canEdit) return;
		const optionValues = this.fieldOptionsFromText(optionsText);
		this.updateFieldMut.mutate({
			path: { id: item.id },
			body: {
				attributes: {
					name: item.attributes.name,
					archived: item.attributes.archived,
					required: false,
					options: optionValues.map((value) => {
						const existing = item.attributes.options.find((opt) => opt.attributes.value === value);
						return {
							id: existing?.id,
							fieldOptionType: existing?.attributes.optionType ?? "custom",
							value,
							archived: false,
						};
					}),
				},
			},
		});
	}
}

const ctx = new Context<IncidentSettingsController>("IncidentSettingsController");
export const initIncidentSettingsController = () => ctx.set(new IncidentSettingsController());
