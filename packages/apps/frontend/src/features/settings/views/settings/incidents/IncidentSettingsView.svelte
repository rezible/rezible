<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import { Checkbox } from "$components/ui/checkbox";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { Switch } from "$components/ui/switch";
	import { Textarea } from "$components/ui/textarea";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import { initIncidentSettingsController } from "./controller.svelte";

	const view = initIncidentSettingsController();

	setPageBreadcrumbs(() => [
		{ label: "Settings", path: "/settings" },
		{ label: "Incidents", path: "/settings/incidents" },
	]);
</script>

<div class="flex max-w-5xl flex-col gap-4">
	{#if view.saveError}
		<InlineAlert error={view.saveError} />
	{/if}

	<Card.Root>
		<Card.Header>
			<Card.Title>Incident Management</Card.Title>
			<Card.Action>
				{#if view.incidentIntegrationInstalled}
					<Badge>Managed by integration</Badge>
				{:else if view.incidentManagementEnabled}
					<Badge>Enabled</Badge>
				{:else}
					<Badge variant="outline">Disabled</Badge>
				{/if}
			</Card.Action>
		</Card.Header>
		<Card.Content>
			{#if view.showEnableToggle}
				<div class="flex items-center justify-between gap-4">
					<Label for="incident-management-enabled">Enable incidents</Label>
					<Switch
						id="incident-management-enabled"
						checked={view.incidentManagementEnabled}
						disabled={view.saving}
						onCheckedChange={(checked) => view.setIncidentManagementEnabled(checked)}
					/>
				</div>
			{:else}
				<div class="text-sm text-muted-foreground">
					{view.incidentIntegrationInstalled
						? "Incident management is controlled by the installed incident integration."
						: "Incident settings are read-only."}
				</div>
			{/if}
		</Card.Content>
	</Card.Root>

	{#if view.loading}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
			<span>Loading metadata...</span>
		</div>
	{:else if view.error}
		<InlineAlert error={view.error} />
	{:else if view.incidentIntegrationInstalled || view.incidentManagementEnabled}
		<Card.Root>
			<Card.Header><Card.Title>Severities</Card.Title></Card.Header>
			<Card.Content class="grid gap-3">
				{#if view.canEdit}
					<div class="grid grid-cols-[minmax(0,1fr)_5rem_6rem_auto] gap-2">
						<Input bind:value={view.newSeverity.name} placeholder="Name" />
						<Input type="number" bind:value={view.newSeverity.rank} />
						<Input type="color" bind:value={view.newSeverity.color} />
						<Button onclick={() => view.createSeverity()} disabled={view.saving}>Add</Button>
					</div>
					<Textarea bind:value={view.newSeverity.description} placeholder="Description" />
				{/if}
				{#each view.severities as severity (severity.id)}
					<div class="grid gap-2 border-t pt-3">
						<div class="grid grid-cols-[minmax(0,1fr)_5rem_6rem_auto_auto] items-center gap-2">
							<Input bind:value={severity.attributes.name} disabled={!view.canEdit} />
							<Input
								type="number"
								bind:value={severity.attributes.rank}
								disabled={!view.canEdit}
							/>
							<Input
								type="color"
								bind:value={severity.attributes.color}
								disabled={!view.canEdit}
							/>
							<label class="flex items-center gap-2 text-sm">
								<Checkbox
									bind:checked={severity.attributes.archived}
									disabled={!view.canEdit}
								/>
								Archive
							</label>
							<Button
								variant="outline"
								onclick={() => view.updateSeverity(severity)}
								disabled={!view.canEdit || view.saving}
							>
								Save
							</Button>
						</div>
						<Textarea bind:value={severity.attributes.description} disabled={!view.canEdit} />
					</div>
				{/each}
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header><Card.Title>Types</Card.Title></Card.Header>
			<Card.Content class="grid gap-3">
				{#if view.canEdit}
					<div class="grid grid-cols-[minmax(0,1fr)_auto] gap-2">
						<Input bind:value={view.newType} placeholder="Type name" />
						<Button onclick={() => view.createType()} disabled={view.saving}>Add</Button>
					</div>
				{/if}
				{#each view.types as type (type.id)}
					<div class="grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-2 border-t pt-3">
						<Input bind:value={type.attributes.name} disabled={!view.canEdit} />
						<label class="flex items-center gap-2 text-sm">
							<Checkbox bind:checked={type.attributes.archived} disabled={!view.canEdit} />
							Archive
						</label>
						<Button
							variant="outline"
							onclick={() => view.updateType(type)}
							disabled={!view.canEdit || view.saving}>Save</Button
						>
					</div>
				{/each}
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header><Card.Title>Roles</Card.Title></Card.Header>
			<Card.Content class="grid gap-3">
				{#if view.canEdit}
					<div class="grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-2">
						<Input bind:value={view.newRole.name} placeholder="Role name" />
						<label class="flex items-center gap-2 text-sm">
							<Checkbox bind:checked={view.newRole.required} />
							Required
						</label>
						<Button onclick={() => view.createRole()} disabled={view.saving}>Add</Button>
					</div>
				{/if}
				{#each view.roles as role (role.id)}
					<div
						class="grid grid-cols-[minmax(0,1fr)_auto_auto_auto] items-center gap-2 border-t pt-3"
					>
						<Input bind:value={role.attributes.name} disabled={!view.canEdit} />
						<label class="flex items-center gap-2 text-sm">
							<Checkbox bind:checked={role.attributes.required} disabled={!view.canEdit} />
							Required
						</label>
						<label class="flex items-center gap-2 text-sm">
							<Checkbox bind:checked={role.attributes.archived} disabled={!view.canEdit} />
							Archive
						</label>
						<Button
							variant="outline"
							onclick={() => view.updateRole(role)}
							disabled={!view.canEdit || view.saving}>Save</Button
						>
					</div>
				{/each}
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header><Card.Title>Tags</Card.Title></Card.Header>
			<Card.Content class="grid gap-3">
				{#if view.canEdit}
					<div class="grid grid-cols-[minmax(0,1fr)_auto] gap-2">
						<Input bind:value={view.newTag} placeholder="Tag value" />
						<Button onclick={() => view.createTag()} disabled={view.saving}>Add</Button>
					</div>
				{/if}
				{#each view.tags as tag (tag.id)}
					<div class="grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-2 border-t pt-3">
						<Input bind:value={tag.attributes.value} disabled={!view.canEdit} />
						<label class="flex items-center gap-2 text-sm">
							<Checkbox bind:checked={tag.attributes.archived} disabled={!view.canEdit} />
							Archive
						</label>
						<Button
							variant="outline"
							onclick={() => view.updateTag(tag)}
							disabled={!view.canEdit || view.saving}>Save</Button
						>
					</div>
				{/each}
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header><Card.Title>Fields</Card.Title></Card.Header>
			<Card.Content class="grid gap-3">
				{#if view.canEdit}
					<div class="grid grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)_auto] gap-2">
						<Input bind:value={view.newField.name} placeholder="Field name" />
						<Input bind:value={view.newField.options} placeholder="Options, comma separated" />
						<Button onclick={() => view.createField()} disabled={view.saving}>Add</Button>
					</div>
				{/if}
				{#each view.fields as field (field.id)}
					<div
						class="grid grid-cols-[minmax(0,1fr)_minmax(0,1.5fr)_auto_auto] items-center gap-2 border-t pt-3"
					>
						<Input bind:value={field.attributes.name} disabled={!view.canEdit} />
						<Input
							value={view.fieldOptionsText(field)}
							disabled={!view.canEdit}
							oninput={(event) => view.setFieldOptionsText(field.id, event.currentTarget.value)}
						/>
						<label class="flex items-center gap-2 text-sm">
							<Checkbox bind:checked={field.attributes.archived} disabled={!view.canEdit} />
							Archive
						</label>
						<Button
							variant="outline"
							onclick={() => view.updateField(field, view.fieldOptionsText(field))}
							disabled={!view.canEdit || view.saving}
						>
							Save
						</Button>
					</div>
				{/each}
			</Card.Content>
		</Card.Root>
	{/if}
</div>
