<script lang="ts">
	import { Button, ListItem } from "svelte-ux";
	import { mdiFlagPlus, mdiPencil, mdiTrashCan } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";
	import {
		deleteIncidentMilestoneMutation,
		listIncidentMilestonesOptions,
		type IncidentMilestone,
	} from "$lib/api";
	
	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";

	import { useIncidentViewState } from "../../../viewState.svelte";
	import { useMilestonesDialog } from "./dialogState.svelte";

	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import MilestoneAttributesEditor from "./MilestoneAttributesEditor.svelte";
	import { getIconForIncidentMilestoneKind, orderedMilestones } from "./milestones";

	const milestonesDialog = useMilestonesDialog();

	const incidentViewState = useIncidentViewState();
	const incidentId = $derived(incidentViewState.incident?.id ?? "");

	const queryClient = useQueryClient();
	const listMilestonesQueryOpts = $derived(listIncidentMilestonesOptions({ path: { id: incidentId } }));
	const listMilestonesQuery = createQuery(() => ({ ...listMilestonesQueryOpts, enabled: !!incidentId }));
	const milestones = $derived(listMilestonesQuery.data?.data || []);
	const invalidateQuery = () => queryClient.invalidateQueries(listMilestonesQueryOpts);

	const onAddClick = () => {
		milestonesDialog.editingMilestone = undefined;
		milestonesDialog.editorOpen = true;
	};

	const onEditClick = (ms: IncidentMilestone) => {
		milestonesDialog.editingMilestone = $state.snapshot(ms);
		milestonesDialog.editorOpen = true;
	};

	const deleteMutation = createMutation(() => ({
		...deleteIncidentMilestoneMutation(),
		onSuccess: invalidateQuery,
	}));

	const onDeleteClick = (ms: IncidentMilestone) => {
		if (!confirm("Are you sure you want to delete this milestone?")) return;
		deleteMutation.mutate({ path: { id: ms.id } });
	};

	const onEditorClosed = () => {
		milestonesDialog.editingMilestone = undefined;
		milestonesDialog.editorOpen = false;
	};

	const onSaved = (ms: IncidentMilestone) => {
		onEditorClosed();
		invalidateQuery();
	};
</script>

{#if milestonesDialog.editorOpen}
	<MilestoneAttributesEditor
		milestone={milestonesDialog.editingMilestone}
		otherMilestones={milestones.filter(m => m.id !== milestonesDialog.editingMilestone?.id)}
		onClose={onEditorClosed}
		{onSaved}
	/>
{:else}
	<LoadingQueryWrapper query={listMilestonesQuery}>
		{#snippet view(milestones: IncidentMilestone[])}
			<div class="w-full h-full overflow-y-hidden flex flex-col gap-2 p-3">
				{#each orderedMilestones(milestones) as ms (ms.id)}
					<ListItem
						title={ms.attributes.kind}
						subheading={ms.attributes.timestamp}
						icon={getIconForIncidentMilestoneKind(ms.attributes.kind)}
						noShadow
						class="flex-1"
						classes={{ root: "border first:border-t rounded elevation-0" }}
					>
						<div slot="actions">
							<Button
								iconOnly
								icon={mdiPencil}
								on:click={() => {
									onEditClick(ms);
								}}
							/>
							<Button
								iconOnly
								icon={mdiTrashCan}
								on:click={() => {
									onDeleteClick(ms);
								}}
							/>
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
	</LoadingQueryWrapper>
{/if}
