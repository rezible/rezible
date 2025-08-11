<script lang="ts">
	import Button from "$components/button/Button.svelte";
	import { mdiGhostOutline } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";
	import Header from "$components/header/Header.svelte";

	// let params = $state<ListUserAssignmentsData["query"]>({});
	// const query = createQuery(() => listUserAssignmentsOptions({ query: params }));

	const assignments: any[] = [];
	// const assignments = $derived(query.data?.data ?? []);
	// const incidents = $derived(assignments.filter((a) => a.itemType === "incident"));
	// const shifts = $derived(assignments.filter((a) => a.itemType === "oncall_shift"));
	// const tasks = $derived(assignments.filter((a) => a.itemType === "tasks"));
</script>

<div class="flex flex-col h-fit border rounded-lg overflow-y-auto">
	<Header title="Your Tasks" subheading="Backlog Items assigned to you" classes={{root: "p-2 w-full"}}>
		{#snippet actions()}
			<div class="justify-end flex gap-2 items-end">
				<Button>View All</Button>
			</div>
		{/snippet}
	</Header>

	<div class="flex flex-col gap-2 overflow-y-auto border-t p-2">
		{#each assignments as a, i (i)}
			{JSON.stringify(a)}
			<!--
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
			{/if}
			-->
		{:else}
			<div class="grid place-items-center">
				<div class="flex gap-2 items-center text-surface-content/80">
					<span class="leading-none">No Items Found</span>
					<Icon data={mdiGhostOutline} />
				</div>
			</div>
		{/each}
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
