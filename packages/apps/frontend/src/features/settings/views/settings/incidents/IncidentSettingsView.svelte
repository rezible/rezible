<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import SettingsFormActions from "$features/settings/components/settings-form-actions/SettingsFormActions.svelte";
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import RiPencilLine from "remixicon-svelte/icons/pencil-line";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import { Checkbox } from "$components/ui/checkbox";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { Switch } from "$components/ui/switch";
	import { Textarea } from "$components/ui/textarea";
	import { resolve } from "$app/paths";
	import type { IncidentSeverity } from "$lib/api";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initIncidentSettingsController } from "./controller.svelte";

	type SeverityFields = { name: string; rank: number; color: string; description: string };

	const controller = initIncidentSettingsController();

	registerPageDescriptor(() => ({
		title: "Incident management",
		parents: [
			{ label: "Settings", path: resolve("/settings") },
			{ label: "Organization", path: resolve("/settings/organization") },
		],
	}));
</script>

<div class="flex max-w-5xl flex-col gap-4">
	{#if controller.saveError}
		<InlineAlert error={controller.saveError} />
	{/if}

	<Card.Root>
		<Card.Header>
			<Card.Title>Incident management</Card.Title>
			<Card.Action>
				{#if controller.incidentManagementEnabled}
					<Badge>On</Badge>
				{:else}
					<Badge variant="outline">Off</Badge>
				{/if}
			</Card.Action>
		</Card.Header>
		<Card.Content class="grid gap-3">
			{#if controller.toggleError}
				<InlineAlert bind:error={controller.toggleError} />
				<a class="text-sm underline" href={resolve("/incidents") + "?status=active"}>
					View unresolved incidents
				</a>
			{/if}
			<div class="flex items-start justify-between gap-4">
				<div class="grid gap-1">
					<Label for="incident-management-enabled">Run incidents in Rezible</Label>
					<p class="text-sm text-muted-foreground">
						Declare incidents, record milestones and assign roles, on the web and through incident
						integrations. Incidents and their reviews are always available.
					</p>
				</div>
				{#if controller.showEnableToggle}
					<Switch
						id="incident-management-enabled"
						bind:checked={
							() => controller.incidentManagementEnabled,
							(checked) => controller.setIncidentManagementEnabled(checked)
						}
						disabled={controller.saving}
					/>
				{/if}
			</div>
			{#if controller.showIntegrationsStayInstalled}
				<p class="text-sm text-muted-foreground">
					Incident integrations stay installed and reply that incident management is off.
				</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<LoadingQueryWrapper query={controller.metadataQuery}>
		{#snippet view()}
			<fieldset disabled={controller.saving} class="flex min-w-0 flex-col gap-4">
				<Card.Root>
					<Card.Header>
						<Card.Title>Severities</Card.Title>
						{#if controller.canEdit && !controller.addingSeverity}
							<Card.Action>
								<Button size="sm" variant="outline" onclick={controller.openNewSeverity}>
									<RiAddLine data-icon="inline-start" />
									Add severity
								</Button>
							</Card.Action>
						{/if}
					</Card.Header>
					<Card.Content class="grid">
						{#if controller.addingSeverity}
							<div class="grid gap-3 pb-3">
								{@render severityFields("new-severity", controller.newSeverity.value)}
								<SettingsFormActions
									onSave={() => controller.createSeverity()}
									onCancel={controller.closeNewSeverity}
									dirty={controller.newSeverity.dirty}
									pending={controller.saving}
									saveLabel="Add"
									cancelRequiresChanges={false}
								/>
							</div>
						{:else if controller.severityRows.length === 0}
							<p class="text-sm text-muted-foreground">No severities yet.</p>
						{/if}
						{#each controller.severityRows as severity (severity.id)}
							{#if controller.editingSeverityId === severity.id}
								{@const draft = controller.severities.get(severity.id)}
								<div class="grid gap-3 border-t py-3 first:border-t-0">
									{@render severityFields(
										`severity-${severity.id}`,
										draft.value.attributes
									)}
									{@render archiveField(`severity-${severity.id}`, draft.value.attributes)}
									<SettingsFormActions
										onSave={() => controller.updateSeverity(draft.value)}
										onCancel={controller.closeSeverityEditor}
										dirty={draft.dirty}
										pending={controller.saving}
										cancelRequiresChanges={false}
									/>
								</div>
							{:else}
								{@render severityRow(severity)}
							{/if}
						{/each}
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header>
						<Card.Title>Types</Card.Title>
						{#if controller.canEdit && !controller.addingType}
							<Card.Action>
								<Button size="sm" variant="outline" onclick={controller.openNewType}>
									<RiAddLine data-icon="inline-start" />
									Add type
								</Button>
							</Card.Action>
						{/if}
					</Card.Header>
					<Card.Content class="grid">
						{#if controller.addingType}
							<div class="grid gap-3 pb-3">
								<div class="grid max-w-md gap-1.5">
									<Label for="new-type-name">Name</Label>
									<Input id="new-type-name" bind:value={controller.newType.value} />
								</div>
								<SettingsFormActions
									onSave={() => controller.createType()}
									onCancel={controller.closeNewType}
									dirty={controller.newType.dirty}
									pending={controller.saving}
									saveLabel="Add"
									cancelRequiresChanges={false}
								/>
							</div>
						{:else if controller.types.savedItems.length === 0}
							<p class="text-sm text-muted-foreground">No types yet.</p>
						{/if}
						{#each controller.types.savedItems as type (type.id)}
							{#if controller.editingTypeId === type.id}
								{@const draft = controller.types.get(type.id)}
								<div class="grid gap-3 border-t py-3 first:border-t-0">
									<div class="grid max-w-md gap-1.5">
										<Label for={`type-${type.id}-name`}>Name</Label>
										<Input
											id={`type-${type.id}-name`}
											bind:value={draft.value.attributes.name}
										/>
									</div>
									{@render archiveField(`type-${type.id}`, draft.value.attributes)}
									<SettingsFormActions
										onSave={() => controller.updateType(draft.value)}
										onCancel={controller.closeTypeEditor}
										dirty={draft.dirty}
										pending={controller.saving}
										cancelRequiresChanges={false}
									/>
								</div>
							{:else}
								<div class="flex min-h-11 items-center gap-3 border-t py-2 first:border-t-0">
									<span class="min-w-0 flex-1 text-sm font-medium">
										{type.attributes.name}
									</span>
									{#if type.attributes.archived}
										<Badge variant="outline">Archived</Badge>
									{/if}
									{@render editButton(type.attributes.name, () =>
										controller.editType(type.id)
									)}
								</div>
							{/if}
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
								<div class="grid w-56 gap-1">
									<Input
										type="number"
										aria-label="Limit"
										placeholder="Any"
										bind:value={
											() => controller.newRole.value.maxAssignees,
											(value) =>
												(controller.newRole.value.maxAssignees = value ?? undefined)
										}
									/>
									<span class="text-xs text-muted-foreground">
										How many people can hold this role in one incident
									</span>
									{#if controller.newRoleLimitError}
										<span class="text-xs text-destructive">
											{controller.newRoleLimitError}
										</span>
									{/if}
								</div>
								<SettingsFormActions
									onSave={() => controller.createRole()}
									onCancel={controller.newRole.cancel}
									dirty={controller.newRole.dirty}
									pending={controller.saving}
									saveDisabled={!!controller.newRoleLimitError}
									saveLabel="Add"
								/>
							</div>
						{/if}
						{#each controller.roles.items as role (role.id)}
							{@const limitError = controller.roleLimitError(role.attributes.maxAssignees)}
							<div
								class="grid grid-cols-1 lg:grid-cols-[minmax(0,1fr)_auto_auto_auto] items-center gap-2 border-t pt-3"
							>
								<Input
									aria-label="Name"
									bind:value={role.attributes.name}
									disabled={!controller.canEdit}
								/>
								<div class="grid w-28 gap-1">
									<Input
										type="number"
										aria-label="Limit"
										placeholder="Any"
										bind:value={
											() => role.attributes.maxAssignees,
											(value) => (role.attributes.maxAssignees = value ?? undefined)
										}
										disabled={!controller.canEdit}
									/>
									{#if limitError}
										<span class="text-xs text-destructive">{limitError}</span>
									{/if}
								</div>
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
									saveDisabled={!!limitError}
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
									aria-label="Options, comma separated"
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
										controller.setFieldOptionsText(field.id, event.currentTarget.value)}
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
		{/snippet}
	</LoadingQueryWrapper>
</div>

{#snippet severityFields(idPrefix: string, fields: SeverityFields)}
	<div class="grid grid-cols-1 gap-3 sm:grid-cols-[minmax(0,1fr)_6rem_6rem]">
		<div class="grid gap-1.5">
			<Label for={`${idPrefix}-name`}>Name</Label>
			<Input id={`${idPrefix}-name`} bind:value={fields.name} />
		</div>
		<div class="grid gap-1.5">
			<Label for={`${idPrefix}-rank`}>Rank</Label>
			<Input id={`${idPrefix}-rank`} type="number" bind:value={fields.rank} />
		</div>
		<div class="grid gap-1.5">
			<Label for={`${idPrefix}-color`}>Color</Label>
			<Input id={`${idPrefix}-color`} type="color" bind:value={fields.color} />
		</div>
	</div>
	<div class="grid gap-1.5">
		<Label for={`${idPrefix}-description`}>Description</Label>
		<Textarea id={`${idPrefix}-description`} bind:value={fields.description} />
	</div>
{/snippet}

{#snippet archiveField(idPrefix: string, fields: { archived: boolean })}
	<label class="flex items-center gap-2 text-sm">
		<Checkbox id={`${idPrefix}-archived`} bind:checked={fields.archived} />
		Archive
	</label>
{/snippet}

{#snippet severityRow(severity: IncidentSeverity)}
	<div class="flex min-h-11 items-start gap-3 border-t py-2 first:border-t-0">
		<span
			class="mt-1.5 size-3 shrink-0 rounded-full border"
			style:background-color={severity.attributes.color}
			aria-hidden="true"
		></span>
		<div class="grid min-w-0 flex-1 gap-0.5">
			<div class="flex flex-wrap items-center gap-2">
				<span class="text-sm font-medium">{severity.attributes.name}</span>
				{#if severity.attributes.archived}
					<Badge variant="outline">Archived</Badge>
				{/if}
			</div>
			{#if severity.attributes.description}
				<p class="text-sm text-muted-foreground">{severity.attributes.description}</p>
			{/if}
		</div>
		<span class="mt-0.5 text-xs text-muted-foreground tabular-nums">Rank {severity.attributes.rank}</span>
		{@render editButton(severity.attributes.name, () => controller.editSeverity(severity.id))}
	</div>
{/snippet}

{#snippet editButton(name: string, onEdit: () => void)}
	{#if controller.canEdit}
		<Button size="sm" variant="ghost" aria-label={`Edit ${name}`} onclick={onEdit}>
			<RiPencilLine data-icon="inline-start" />
			Edit
		</Button>
	{/if}
{/snippet}
