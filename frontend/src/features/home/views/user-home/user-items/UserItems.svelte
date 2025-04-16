<script lang="ts">
	import { Icon, Header, ListItem, Button } from "svelte-ux";
	import { mdiGhostOutline, mdiFileDocument, mdiChevronRight, mdiVideo } from "@mdi/js";
	import { formatDistanceToNow } from "date-fns";

	// let params = $state<ListUserAssignmentsData["query"]>({});
	// const query = createQuery(() => listUserAssignmentsOptions({ query: params }));

	// const assignments = $derived(query.data?.data ?? []);
	// const incidents = $derived(assignments.filter((a) => a.itemType === "incident"));
	// const shifts = $derived(assignments.filter((a) => a.itemType === "oncall_shift"));
	// const tasks = $derived(assignments.filter((a) => a.itemType === "tasks"));
</script>

<div class="flex flex-col h-fit gap-2 border p-2 rounded-lg overflow-y-auto">
	<Header title="Assigned to You" class="border-b" />

	<div class="flex flex-col gap-2 overflow-y-auto">
		<span class="flex items-center gap-2">Nothing! <Icon data={mdiGhostOutline} /></span>
		<!--
		{#each assignments as a, i (i)}
			{#if a.itemType == "retrospective"}
				<a href="/incidents/{a.itemId}/retrospective">
					<ListItem
						title="{a.role} the Retrospective for {a.title}"
						subheading="Due in {formatDistanceToNow(a.deadline)}"
						classes={{ root: "hover:bg-primary-900/80" }}
						icon={mdiFileDocument}
					>
						<div slot="actions">
							<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
						</div>
					</ListItem>
				</a>
			{:else if a.itemType === "incident_review"}
				<a href="/meetings/sessions/{a.itemId}">
					<ListItem
						title="{a.role} {a.title} at Incident Review"
						subheading="In {formatDistanceToNow(a.deadline)}"
						classes={{ root: "hover:bg-primary-900/80" }}
						icon={mdiVideo}
					>
						<div slot="actions">
							<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
						</div>
					</ListItem>
				</a>
			{:else}
				{JSON.stringify(a)}
			{/if}
		{:else}
			<div class="flex gap-2 items-center">
				<span class="leading-none">No Assigned Items</span>
				<Icon data={mdiGhostOutline} />
			</div>
		{/each}
		-->
	</div>
</div>

<!--div class="flex flex-col h-fit gap-2 border p-2 rounded-lg overflow-y-auto">
	<Header class="border-b">
		<div class="flex items-center gap-2" slot="title">
			<span class="text-lg">Oncall Backlog for Search</span>
			<Avatar kind="roster" id="roster-id" size={22} />
		</div>
	</Header>

	<div class="flex flex-col gap-2 overflow-y-auto">
		<ListItem 
			title="Investigate Elevated Memory Usage in AP-South-1 Cluster" 
			subheading="Medium Priority"
			classes={{ root: 'hover:bg-primary-900/80' }}
		>
		<div slot="avatar" class="text-surface-content/80">
			SRCH-2245
		</div>
			<div slot="actions">
				<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
			</div>
		</ListItem>

		{#if assignments.length === 0}
			<div class="flex gap-2 items-center">
				<span class="leading-none">No Assigned Items</span>
				<Icon data={mdiGhostOutline} />
			</div>
		{/if}
	</div>
</div-->
