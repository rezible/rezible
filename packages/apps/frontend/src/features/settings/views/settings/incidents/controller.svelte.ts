import {
	createIncidentFieldMutation,
	createIncidentRoleMutation,
	createIncidentSeverityMutation,
	createIncidentTagMutation,
	createIncidentTypeMutation,
	getIncidentMetadataOptions,
	getIncidentMetadataQueryKey,
	getUserSessionOptions,
	type ApiError,
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
import { Context, watch } from "runed";
import { toast } from "svelte-sonner";
import { SettingsDraft, SettingsDraftList } from "$features/settings/lib/settings-draft.svelte";
import { SvelteMap } from "svelte/reactivity";

export class IncidentSettingsController {
	session = useUserSessionState();
	integrations = useIntegrationsController();
	private queryClient = useQueryClient();

	metadataQuery = createQuery(() => getIncidentMetadataOptions({ query: { archived: true } }));
	private metadata = $derived(this.metadataQuery.data?.data);

	severities = new SettingsDraftList<IncidentSeverity>();
	types = new SettingsDraftList<IncidentType>();
	roles = new SettingsDraftList<IncidentRole>();
	tags = new SettingsDraftList<IncidentTag>();
	fields = new SettingsDraftList<IncidentField>();

	loading = $derived(this.metadataQuery.isPending);
	error = $derived(this.metadataQuery.error as ApiError | null);
	saveError = $state<ApiError>();

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

	newSeverity = new SettingsDraft({ name: "", rank: 1, color: "#ef4444", description: "" });
	newType = new SettingsDraft("");
	newRole = new SettingsDraft({ name: "", required: false });
	newTag = new SettingsDraft("");
	newField = new SettingsDraft({ name: "", options: "" });
	fieldOptionTextEdits = new SvelteMap<string, string>();

	private async invalidateMetadata() {
		await this.queryClient.invalidateQueries({ queryKey: getIncidentMetadataQueryKey() });
	}

	constructor() {
		watch(
			() => this.metadata,
			(metadata) => {
				if (!metadata) return;
				this.severities.receive(metadata.severities);
				this.types.receive(metadata.types);
				this.roles.receive(metadata.roles);
				this.tags.receive(metadata.tags);
				this.fields.receive(metadata.fields);
			}
		);
	}

	dirty = $derived(
		this.severities.dirty ||
			this.types.dirty ||
			this.roles.dirty ||
			this.tags.dirty ||
			this.fields.dirty ||
			this.fieldOptionTextEdits.size > 0 ||
			this.newSeverity.dirty ||
			this.newType.dirty ||
			this.newRole.dirty ||
			this.newTag.dirty ||
			this.newField.dirty
	);

	private mutationOptions<T extends { id: string }>(
		forms: SettingsDraftList<T>,
		onSaved?: (item: T) => void
	) {
		return {
			onSuccess: async ({ data }: { data: T }) => {
				this.saveError = undefined;
				forms.accept(data);
				onSaved?.(data);
				toast.success("Settings saved.");
				await this.invalidateMetadata();
			},
			onError: (err: ApiError) => {
				this.saveError = err;
			},
		};
	}

	private updateOrgPrefsMut = createMutation(() => ({
		...updateOrganizationPreferencesMutation(),
		onSuccess: async () => {
			this.saveError = undefined;
			toast.success("Settings saved.");
			await this.queryClient.invalidateQueries({ queryKey: getUserSessionOptions().queryKey });
			this.session.refetch();
		},
		onError: (err) => {
			this.saveError = err;
		},
	}));

	private createSeverityMut = createMutation(() => ({
		...createIncidentSeverityMutation(),
		...this.mutationOptions(this.severities, () => {
			this.newSeverity.accept({
				name: "",
				rank: this.newSeverity.value.rank + 1,
				color: "#ef4444",
				description: "",
			});
		}),
	}));
	private updateSeverityMut = createMutation(() => ({
		...updateIncidentSeverityMutation(),
		...this.mutationOptions(this.severities),
	}));
	private createTypeMut = createMutation(() => ({
		...createIncidentTypeMutation(),
		...this.mutationOptions(this.types, () => {
			this.newType.accept("");
		}),
	}));
	private updateTypeMut = createMutation(() => ({
		...updateIncidentTypeMutation(),
		...this.mutationOptions(this.types),
	}));
	private createRoleMut = createMutation(() => ({
		...createIncidentRoleMutation(),
		...this.mutationOptions(this.roles, () => {
			this.newRole.accept({ name: "", required: false });
		}),
	}));
	private updateRoleMut = createMutation(() => ({
		...updateIncidentRoleMutation(),
		...this.mutationOptions(this.roles),
	}));
	private createTagMut = createMutation(() => ({
		...createIncidentTagMutation(),
		...this.mutationOptions(this.tags, () => {
			this.newTag.accept("");
		}),
	}));
	private updateTagMut = createMutation(() => ({
		...updateIncidentTagMutation(),
		...this.mutationOptions(this.tags),
	}));
	private createFieldMut = createMutation(() => ({
		...createIncidentFieldMutation(),
		...this.mutationOptions(this.fields, () => {
			this.newField.accept({ name: "", options: "" });
		}),
	}));
	private updateFieldMut = createMutation(() => ({
		...updateIncidentFieldMutation(),
		...this.mutationOptions(this.fields, (item) => {
			this.fieldOptionTextEdits.delete(item.id);
		}),
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
		if (this.saving || !this.canEdit || !this.newSeverity.value.name.trim()) return;
		this.createSeverityMut.mutate({
			body: { attributes: { ...this.newSeverity.value, name: this.newSeverity.value.name.trim() } },
		});
	}

	updateSeverity(item: IncidentSeverity) {
		if (this.saving || !this.canEdit) return;
		this.updateSeverityMut.mutate({
			path: { id: item.id },
			body: { attributes: { ...item.attributes } },
		});
	}

	createType() {
		if (this.saving || !this.canEdit || !this.newType.value.trim()) return;
		this.createTypeMut.mutate({ body: { attributes: { name: this.newType.value.trim() } } });
	}

	updateType(item: IncidentType) {
		if (this.saving || !this.canEdit) return;
		this.updateTypeMut.mutate({
			path: { id: item.id },
			body: { attributes: { name: item.attributes.name, archived: item.attributes.archived } },
		});
	}

	createRole() {
		if (this.saving || !this.canEdit || !this.newRole.value.name.trim()) return;
		this.createRoleMut.mutate({
			body: {
				attributes: { name: this.newRole.value.name.trim(), required: this.newRole.value.required },
			},
		});
	}

	updateRole(item: IncidentRole) {
		if (this.saving || !this.canEdit) return;
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
		if (this.saving || !this.canEdit || !this.newTag.value.trim()) return;
		this.createTagMut.mutate({ body: { attributes: { value: this.newTag.value.trim() } } });
	}

	updateTag(item: IncidentTag) {
		if (this.saving || !this.canEdit) return;
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
		const savedText = this.fields
			.get(id)
			.value.attributes.options.map((option) => option.attributes.value)
			.join(", ");
		if (value === savedText) this.fieldOptionTextEdits.delete(id);
		else this.fieldOptionTextEdits.set(id, value);
	}

	cancelField = (id: string) => {
		this.fields.get(id).cancel();
		this.fieldOptionTextEdits.delete(id);
	};

	createField() {
		const options = this.fieldOptionsFromText(this.newField.value.options);
		if (this.saving || !this.canEdit || !this.newField.value.name.trim() || options.length === 0) return;
		this.createFieldMut.mutate({
			body: {
				attributes: {
					name: this.newField.value.name.trim(),
					required: false,
					options: options.map((value) => ({ fieldOptionType: "custom", value })),
				},
			},
		});
	}

	updateField(item: IncidentField, optionsText: string) {
		if (this.saving || !this.canEdit) return;
		const optionValues = this.fieldOptionsFromText(optionsText);
		this.updateFieldMut.mutate({
			path: { id: item.id },
			body: {
				attributes: {
					name: item.attributes.name,
					archived: item.attributes.archived,
					required: false,
					options: optionValues.map((value) => {
						const existing = item.attributes.options.find(
							(opt) => opt.attributes.value === value
						);
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
