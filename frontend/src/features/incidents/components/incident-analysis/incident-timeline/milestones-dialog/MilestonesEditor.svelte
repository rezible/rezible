<script lang="ts">
	import { Button, Icon, ListItem } from "svelte-ux";
	import { mdiBabel, mdiFlagPlus, mdiPencil, mdiTrashCan } from "@mdi/js";
	import { deleteIncidentMilestoneMutation, listIncidentMilestonesOptions, type IncidentMilestone, type IncidentMilestoneAttributes } from "$lib/api";
	import MilestoneAttributesEditor from "./MilestoneAttributesEditor.svelte";
	import { milestonesDialog } from "./dialogState.svelte";
	import { incidentCtx } from "$src/features/incidents/lib/context";
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";

	const incident = incidentCtx.get();

	const queryClient = useQueryClient();
	const queryOpts = $derived(listIncidentMilestonesOptions({path: {id: incident.id}}));
	const listMilestonesQuery = createQuery(() => queryOpts);
	const invalidateQuery = () => queryClient.invalidateQueries(queryOpts);

	const onAddClick = () => {
		milestonesDialog.editingMilestone = undefined;
		milestonesDialog.editorOpen = true;
	}

	const onEditClick = (ms: IncidentMilestone) => {
		milestonesDialog.editingMilestone = $state.snapshot(ms);
		milestonesDialog.editorOpen = true;
	};

	const deleteMutation = createMutation(() => ({...deleteIncidentMilestoneMutation(), onSuccess: invalidateQuery}))

	const onDeleteClick = (ms: IncidentMilestone) => {
		if (!confirm("Are you sure you want to delete this milestone?")) return;
		deleteMutation.mutate({path: {id: ms.id}});
	};

	const onEditorClosed = () => {
		milestonesDialog.editingMilestone = undefined;
		milestonesDialog.editorOpen = false;
	};

	const onSaved = (ms: IncidentMilestone) => {
		onEditorClosed();
		invalidateQuery();
	}

	const getIconForKind = (kind: IncidentMilestoneAttributes["kind"]) => {

		return mdiBabel;
	}
</script>

{#snippet milestonesListView(milestones: IncidentMilestone[])}
	<div class="w-full h-full overflow-y-hidden flex flex-col gap-2 p-3">
		{#each milestones as ms (ms.id)}
			<ListItem
				title={ms.attributes.kind}
				subheading={ms.attributes.timestamp}
				icon={getIconForKind(ms.attributes.kind)}
				noShadow
				class="flex-1"
				classes={{ root: "border first:border-t rounded elevation-0" }}
			>
				<div slot="actions">
					<Button iconOnly icon={mdiPencil} on:click={() => {onEditClick(ms)}} />
					<Button iconOnly icon={mdiTrashCan} on:click={() => {onDeleteClick(ms)}} />
				</div>
			</ListItem>
		{/each}

		<Button variant="fill-light" on:click={onAddClick}>
			<span class="flex gap-2 items-center">
				Add Milestone
				<Icon data={mdiFlagPlus} />
			</span>
		</Button>
	</div>
{/snippet}

{#if milestonesDialog.editorOpen}
	<MilestoneAttributesEditor milestone={milestonesDialog.editingMilestone} onClose={onEditorClosed} {onSaved} />
{:else}
	<LoadingQueryWrapper query={listMilestonesQuery} view={milestonesListView} />
{/if}
