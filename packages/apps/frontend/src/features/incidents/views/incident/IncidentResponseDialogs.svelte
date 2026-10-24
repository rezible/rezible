<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as AlertDialog from "$components/ui/alert-dialog";
	import { Button } from "$components/ui/button";
	import * as Dialog from "$components/ui/dialog";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { Spinner } from "$components/ui/spinner";
	import { Textarea } from "$components/ui/textarea";
	import * as ToggleGroup from "$components/ui/toggle-group";

	import IncidentDetailsFields from "$features/incidents/components/create-incident-dialog/IncidentDetailsFields.svelte";
	import SituationCombobox from "$features/situations/components/situation-combobox/SituationCombobox.svelte";

	import { recordableMilestoneKinds, useIncidentResponse } from "./response.svelte";

	const response = useIncidentResponse();

	const milestone = $derived(response.milestoneDialog);
	const resolving = $derived(milestone?.kind === "resolution" && !milestone.milestoneId);
</script>

{#snippet milestoneFields()}
	{#if milestone}
		{#if milestone.chooseKind}
			<div class="flex flex-col gap-2">
				<span class="text-sm font-medium">Kind</span>
				<ToggleGroup.Root
					type="single"
					variant="outline"
					bind:value={() => milestone.kind, response.setMilestoneKind}
					aria-label="Milestone kind"
				>
					{#each recordableMilestoneKinds as option (option.kind)}
						<ToggleGroup.Item value={option.kind}>{option.label}</ToggleGroup.Item>
					{/each}
				</ToggleGroup.Root>
			</div>
		{/if}
		<div class="flex flex-col gap-2">
			<Label for="incident-milestone-time">Time</Label>
			<Input id="incident-milestone-time" type="datetime-local" bind:value={milestone.time} />
		</div>
		<div class="flex flex-col gap-2">
			<Label for="incident-milestone-description">Description</Label>
			<Textarea
				id="incident-milestone-description"
				rows={3}
				placeholder={milestone.placeholder || "Optional"}
				bind:value={milestone.description}
			/>
		</div>
		<InlineAlert error={response.milestoneError} dismissable={false} />
	{/if}
{/snippet}

<Dialog.Root
	open={!!milestone && !resolving}
	onOpenChange={(open) => {
		if (!open) {
			response.closeMilestoneDialog();
		}
	}}
>
	<Dialog.Content class="sm:max-w-[480px]">
		<form
			class="flex flex-col gap-4"
			onsubmit={(event) => {
				event.preventDefault();
				response.saveMilestone();
			}}
		>
			<Dialog.Header>
				<Dialog.Title>{milestone?.title}</Dialog.Title>
			</Dialog.Header>
			{@render milestoneFields()}
			<Dialog.Footer>
				<Button type="button" variant="ghost" onclick={response.closeMilestoneDialog}>Cancel</Button>
				<Button type="submit" disabled={response.milestonePending}>
					{#if response.milestonePending}
						<Spinner data-icon="inline-start" />
					{/if}
					Save
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<AlertDialog.Root
	open={resolving}
	onOpenChange={(open) => {
		if (!open) {
			response.closeMilestoneDialog();
		}
	}}
>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Resolve incident?</AlertDialog.Title>
			<AlertDialog.Description>
				Resolving is final. A recurrence is a new incident. Linked situations can close once their
				signals have been quiet.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<div class="flex flex-col gap-4">
			{@render milestoneFields()}
		</div>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={response.milestonePending}>Cancel</AlertDialog.Cancel>
			<Button disabled={response.milestonePending} onclick={response.saveMilestone}>
				{#if response.milestonePending}
					<Spinner data-icon="inline-start" />
				{/if}
				Resolve
			</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>

<Dialog.Root bind:open={response.detailsOpen}>
	<Dialog.Content class="max-h-[calc(100dvh-2rem)] overflow-y-auto sm:max-w-[640px]">
		<form
			class="flex flex-col gap-4"
			onsubmit={(event) => {
				event.preventDefault();
				response.saveDetails();
			}}
		>
			<Dialog.Header>
				<Dialog.Title>Edit details</Dialog.Title>
			</Dialog.Header>
			<IncidentDetailsFields
				form={response.detailsForm}
				metadata={response.incidentMetadata}
				titleError={response.detailsTitleError}
				isAdmin={response.isAdmin}
				idPrefix="edit-incident"
			/>
			<InlineAlert error={response.detailsError} dismissable={false} />
			<Dialog.Footer>
				<Button type="button" variant="ghost" onclick={() => (response.detailsOpen = false)}>
					Cancel
				</Button>
				<Button type="submit" disabled={response.detailsPending || !!response.detailsTitleError}>
					{#if response.detailsPending}
						<Spinner data-icon="inline-start" />
					{/if}
					Save
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<AlertDialog.Root bind:open={response.archiveOpen}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Archive incident?</AlertDialog.Title>
			<AlertDialog.Description>
				It will be hidden from incident lists and its situations will be unlinked. Its retrospective
				and milestones are kept.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<InlineAlert error={response.archiveError} dismissable={false} />
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={response.archivePending}>Cancel</AlertDialog.Cancel>
			<Button variant="destructive" disabled={response.archivePending} onclick={response.archive}>
				{#if response.archivePending}
					<Spinner data-icon="inline-start" />
				{/if}
				Archive
			</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>

<Dialog.Root bind:open={response.linkSituationOpen}>
	<Dialog.Content class="sm:max-w-[480px]">
		<Dialog.Header>
			<Dialog.Title>Link situation</Dialog.Title>
			<Dialog.Description>Open and closed situations can be linked.</Dialog.Description>
		</Dialog.Header>
		<SituationCombobox linkedIds={response.linkedSituationIds} onSelect={response.linkSituation} />
	</Dialog.Content>
</Dialog.Root>
