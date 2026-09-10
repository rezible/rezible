<script lang="ts">
	import RiFlagLine from "remixicon-svelte/icons/flag-line";
	import RiPencilLine from "remixicon-svelte/icons/pencil-line";
	import RiDeleteBinLine from "remixicon-svelte/icons/delete-bin-line";
	import {
		deleteIncidentMilestoneMutation,
		listIncidentMilestonesOptions,
		type IncidentMilestone,
	} from "$lib/api";

	import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";

	import { useIncidentView } from "$features/incidents/views/incident";
	import { useMilestonesDialog } from "./controller.svelte";

	import MilestoneAttributesEditor from "./MilestoneAttributesEditor.svelte";
	import { orderedMilestones } from "./milestones";

	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { Button } from "$components/ui/button";

	const milestonesDialog = useMilestonesDialog();

	const incidentView = useIncidentView();
	const incidentId = $derived(incidentView.incident?.id ?? "");

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
		otherMilestones={milestones.filter((m) => m.id !== milestonesDialog.editingMilestone?.id)}
		onClose={onEditorClosed}
		{onSaved}
	/>
{:else}
	<LoadingQueryWrapper query={listMilestonesQuery}>
		{#snippet view(milestones: IncidentMilestone[])}
			<div class="w-full h-full overflow-y-hidden flex flex-col gap-2 p-3">
				{#each orderedMilestones(milestones) as ms (ms.id)}
					<span>milestone: {ms.attributes.kind}</span>
					<!-- <ListItem
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
								icon={RiPencilLine}
								onclick={() => {
									onEditClick(ms);
								}}
							/>
							<Button
								iconOnly
								icon={RiDeleteBinLine}
								onclick={() => {
									onDeleteClick(ms);
								}}
							/>
						</div>
					</ListItem> -->
				{/each}

				<Button onclick={onAddClick}>
					<span class="flex gap-2 items-center">
						Add Milestone
						<RiFlagLine class="" aria-hidden="true" />
					</span>
				</Button>
			</div>
		{/snippet}
	</LoadingQueryWrapper>
{/if}
