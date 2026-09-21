<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import SettingsFormActions from "$features/settings/components/settings-form-actions/SettingsFormActions.svelte";
	import { Badge } from "$components/ui/badge";
	import * as Card from "$components/ui/card";
	import { Checkbox } from "$components/ui/checkbox";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { Switch } from "$components/ui/switch";
	import { Textarea } from "$components/ui/textarea";
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initIncidentSettingsController } from "./controller.svelte";

	const controller = initIncidentSettingsController();

	registerPageDescriptor(() => ({
		title: "Incident management",
		parents: [{ label: "Settings", path: resolve("/settings") }],
	}));
</script>

<div class="flex max-w-5xl flex-col gap-4">
	{#if controller.saveError}
		<InlineAlert error={controller.saveError} />
	{/if}

	<Card.Root>
		<Card.Header>
			<Card.Title>Incident Management</Card.Title>
			<Card.Action>
				{#if controller.incidentIntegrationInstalled}
					<Badge>Managed by integration</Badge>
				{:else if controller.incidentManagementEnabled}
					<Badge>Enabled</Badge>
				{:else}
					<Badge variant="outline">Disabled</Badge>
				{/if}
			</Card.Action>
		</Card.Header>
		<Card.Content>
			{#if controller.showEnableToggle}
				<div class="flex items-center justify-between gap-4">
					<Label for="incident-management-enabled">Enable incidents</Label>
					<Switch
						id="incident-management-enabled"
						checked={controller.incidentManagementEnabled}
						disabled={controller.saving}
						onCheckedChange={(checked) => controller.setIncidentManagementEnabled(checked)}
					/>
				</div>
			{:else}
				<div class="text-sm text-muted-foreground">
					{controller.incidentIntegrationInstalled
						? "Incident management is controlled by the installed incident integration."
						: "Incident settings are read-only."}
				</div>
			{/if}
		</Card.Content>
	</Card.Root>

	<LoadingQueryWrapper query={controller.metadataQuery}>
		{#snippet view()}
			{#if controller.incidentIntegrationInstalled || controller.incidentManagementEnabled}
				<fieldset disabled={controller.saving} class="flex min-w-0 flex-col gap-4">
					<Card.Root>
						<Card.Header><Card.Title>Severities</Card.Title></Card.Header>
						<Card.Content class="grid gap-3">
							{#if controller.canEdit}
								<div
									class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_5rem_6rem_auto] gap-2"
								>
									<Input
										aria-label="Name"
										bind:value={controller.newSeverity.value.name}
										placeholder="Name"
									/>
									<Input
										type="number"
										aria-label="Rank"
										bind:value={controller.newSeverity.value.rank}
									/>
									<Input
										type="color"
										aria-label="Color"
										bind:value={controller.newSeverity.value.color}
									/>
									<SettingsFormActions
										onSave={() => controller.createSeverity()}
										onCancel={controller.newSeverity.cancel}
										dirty={controller.newSeverity.dirty}
										pending={controller.saving}
										saveLabel="Add"
									/>
								</div>
								<Textarea
									aria-label="Description"
									bind:value={controller.newSeverity.value.description}
									placeholder="Description"
								/>
							{/if}
							{#each controller.severities.items as severity (severity.id)}
								<div class="grid gap-2 border-t pt-3">
									<div
										class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_5rem_6rem_auto_auto] items-center gap-2"
									>
										<Input
											aria-label="Name"
											bind:value={severity.attributes.name}
											disabled={!controller.canEdit}
										/>
										<Input
											type="number"
											aria-label="Rank"
											bind:value={severity.attributes.rank}
											disabled={!controller.canEdit}
										/>
										<Input
											type="color"
											aria-label="Color"
											bind:value={severity.attributes.color}
											disabled={!controller.canEdit}
										/>
										<label class="flex items-center gap-2 text-sm">
											<Checkbox
												bind:checked={severity.attributes.archived}
												disabled={!controller.canEdit}
											/>
											Archive
										</label>
										<SettingsFormActions
											onSave={() => controller.updateSeverity(severity)}
											onCancel={controller.severities.get(severity.id).cancel}
											dirty={controller.severities.get(severity.id).dirty}
											pending={controller.saving}
											disabled={!controller.canEdit}
										/>
									</div>
									<Textarea
										aria-label="Description"
										bind:value={severity.attributes.description}
										disabled={!controller.canEdit}
									/>
								</div>
							{/each}
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header><Card.Title>Types</Card.Title></Card.Header>
						<Card.Content class="grid gap-3">
							{#if controller.canEdit}
								<div class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_auto] gap-2">
									<Input
										aria-label="Name"
										bind:value={controller.newType.value}
										placeholder="Type name"
									/>
									<SettingsFormActions
										onSave={() => controller.createType()}
										onCancel={controller.newType.cancel}
										dirty={controller.newType.dirty}
										pending={controller.saving}
										saveLabel="Add"
									/>
								</div>
							{/if}
							{#each controller.types.items as type (type.id)}
								<div
									class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-2 border-t pt-3"
								>
									<Input
										aria-label="Name"
										bind:value={type.attributes.name}
										disabled={!controller.canEdit}
									/>
									<label class="flex items-center gap-2 text-sm">
										<Checkbox
											bind:checked={type.attributes.archived}
											disabled={!controller.canEdit}
										/>
										Archive
									</label>
									<SettingsFormActions
										onSave={() => controller.updateType(type)}
										onCancel={controller.types.get(type.id).cancel}
										dirty={controller.types.get(type.id).dirty}
										pending={controller.saving}
										disabled={!controller.canEdit}
									/>
								</div>
							{/each}
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header><Card.Title>Roles</Card.Title></Card.Header>
						<Card.Content class="grid gap-3">
							{#if controller.canEdit}
								<div
									class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-2"
								>
									<Input
										aria-label="Name"
										bind:value={controller.newRole.value.name}
										placeholder="Role name"
									/>
									<label class="flex items-center gap-2 text-sm">
										<Checkbox bind:checked={controller.newRole.value.required} />
										Required
									</label>
									<SettingsFormActions
										onSave={() => controller.createRole()}
										onCancel={controller.newRole.cancel}
										dirty={controller.newRole.dirty}
										pending={controller.saving}
										saveLabel="Add"
									/>
								</div>
							{/if}
							{#each controller.roles.items as role (role.id)}
								<div
									class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_auto_auto_auto] items-center gap-2 border-t pt-3"
								>
									<Input
										aria-label="Name"
										bind:value={role.attributes.name}
										disabled={!controller.canEdit}
									/>
									<label class="flex items-center gap-2 text-sm">
										<Checkbox
											bind:checked={role.attributes.required}
											disabled={!controller.canEdit}
										/>
										Required
									</label>
									<label class="flex items-center gap-2 text-sm">
										<Checkbox
											bind:checked={role.attributes.archived}
											disabled={!controller.canEdit}
										/>
										Archive
									</label>
									<SettingsFormActions
										onSave={() => controller.updateRole(role)}
										onCancel={controller.roles.get(role.id).cancel}
										dirty={controller.roles.get(role.id).dirty}
										pending={controller.saving}
										disabled={!controller.canEdit}
									/>
								</div>
							{/each}
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header><Card.Title>Tags</Card.Title></Card.Header>
						<Card.Content class="grid gap-3">
							{#if controller.canEdit}
								<div class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_auto] gap-2">
									<Input
										aria-label="Name"
										bind:value={controller.newTag.value}
										placeholder="Tag value"
									/>
									<SettingsFormActions
										onSave={() => controller.createTag()}
										onCancel={controller.newTag.cancel}
										dirty={controller.newTag.dirty}
										pending={controller.saving}
										saveLabel="Add"
									/>
								</div>
							{/if}
							{#each controller.tags.items as tag (tag.id)}
								<div
									class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-2 border-t pt-3"
								>
									<Input
										aria-label="Tag"
										bind:value={tag.attributes.value}
										disabled={!controller.canEdit}
									/>
									<label class="flex items-center gap-2 text-sm">
										<Checkbox
											bind:checked={tag.attributes.archived}
											disabled={!controller.canEdit}
										/>
										Archive
									</label>
									<SettingsFormActions
										onSave={() => controller.updateTag(tag)}
										onCancel={controller.tags.get(tag.id).cancel}
										dirty={controller.tags.get(tag.id).dirty}
										pending={controller.saving}
										disabled={!controller.canEdit}
									/>
								</div>
							{/each}
						</Card.Content>
					</Card.Root>

					<Card.Root>
						<Card.Header><Card.Title>Fields</Card.Title></Card.Header>
						<Card.Content class="grid gap-3">
							{#if controller.canEdit}
								<div
									class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)_auto] gap-2"
								>
									<Input
										aria-label="Name"
										bind:value={controller.newField.value.name}
										placeholder="Field name"
									/>
									<Input
										aria-label="Name"
										bind:value={controller.newField.value.options}
										placeholder="Options, comma separated"
									/>
									<SettingsFormActions
										onSave={() => controller.createField()}
										onCancel={controller.newField.cancel}
										dirty={controller.newField.dirty}
										pending={controller.saving}
										saveLabel="Add"
									/>
								</div>
							{/if}
							{#each controller.fields.items as field (field.id)}
								<div
									class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)_auto_auto] items-center gap-2 border-t pt-3"
								>
									<Input
										aria-label="Name"
										bind:value={field.attributes.name}
										disabled={!controller.canEdit}
									/>
									<Input
										aria-label="Options, comma separated"
										value={controller.fieldOptionsText(field)}
										disabled={!controller.canEdit}
										oninput={(event) =>
											controller.setFieldOptionsText(
												field.id,
												event.currentTarget.value
											)}
									/>
									<label class="flex items-center gap-2 text-sm">
										<Checkbox
											bind:checked={field.attributes.archived}
											disabled={!controller.canEdit}
										/>
										Archive
									</label>
									<SettingsFormActions
										onSave={() =>
											controller.updateField(field, controller.fieldOptionsText(field))}
										onCancel={() => controller.cancelField(field.id)}
										dirty={controller.fields.get(field.id).dirty ||
											controller.fieldOptionTextEdits.has(field.id)}
										pending={controller.saving}
										disabled={!controller.canEdit}
									/>
								</div>
							{/each}
						</Card.Content>
					</Card.Root>
				</fieldset>
			{/if}
		{/snippet}
	</LoadingQueryWrapper>
</div>
