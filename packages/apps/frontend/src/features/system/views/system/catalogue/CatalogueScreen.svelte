<script lang="ts">
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { Switch } from "$components/ui/switch";
	import * as Tabs from "$components/ui/tabs";
	import ConfirmProposalDialog from "./confirm-proposal-dialog/ConfirmProposalDialog.svelte";
	import { initCatalogueController } from "./controller.svelte";
	import EntityRow from "./entity-row/EntityRow.svelte";
	import ProposalCard from "./proposal-card/ProposalCard.svelte";

	const controller = initCatalogueController();

	// The catalogue scrolls inside its own panel, so bring a newly linked proposal into view once it loads.
	let linkedSection = $state<HTMLElement>();
	let scrolledToProposalId: string | undefined;
	$effect(() => {
		const proposalId = controller.focusedProposalId;
		if (!proposalId || !controller.focusedProposal || !linkedSection) {
			return;
		}
		if (scrolledToProposalId === proposalId) {
			return;
		}
		scrolledToProposalId = proposalId;
		linkedSection.scrollIntoView({ block: "start" });
	});
</script>

{#snippet moreProposals(shown: number, total: number)}
	{#if total > shown}
		<p class="text-xs text-muted-foreground">Showing {shown} of {total}, newest first.</p>
	{/if}
{/snippet}

<section class="flex min-h-0 min-w-0 flex-1 flex-col gap-6 overflow-y-auto p-4" aria-label="System catalogue">
	{#if controller.focusedProposalId}
		<section bind:this={linkedSection} class="space-y-2" aria-labelledby="focused-proposal-heading">
			<h2 id="focused-proposal-heading" class="text-sm font-semibold">Linked match</h2>
			<LoadingQueryWrapper query={controller.focusedProposalQuery}>
				{#snippet view()}
					{#if controller.focusedProposal}
						<ProposalCard proposal={controller.focusedProposal} />
					{/if}
				{/snippet}
			</LoadingQueryWrapper>
		</section>
	{/if}

	<section class="space-y-2" aria-labelledby="proposed-matches-heading">
		<div class="flex flex-wrap items-center justify-between gap-2">
			<h2 id="proposed-matches-heading" class="text-sm font-semibold">Proposed matches</h2>
			<div class="flex items-center gap-2">
				<Switch id="show-rejected" bind:checked={controller.showRejected} />
				<Label for="show-rejected" class="text-xs">Show rejected</Label>
			</div>
		</div>

		<LoadingQueryWrapper query={controller.pendingQuery} isEmpty={(proposals) => proposals.length === 0}>
			{#snippet empty()}
				<p class="text-sm text-muted-foreground">No proposed matches.</p>
			{/snippet}
			{#snippet view()}
				<div class="space-y-3">
					{#each controller.pendingViews as proposal (proposal.id)}
						<ProposalCard {proposal} />
					{/each}
					{@render moreProposals(controller.pendingViews.length, controller.pendingTotal)}
				</div>
			{/snippet}
		</LoadingQueryWrapper>

		{#if controller.showRejected}
			<h3 class="pt-2 text-xs font-semibold text-muted-foreground">Rejected</h3>
			<LoadingQueryWrapper
				query={controller.rejectedQuery}
				isEmpty={(proposals) => proposals.length === 0}
			>
				{#snippet empty()}
					<p class="text-sm text-muted-foreground">No rejected matches.</p>
				{/snippet}
				{#snippet view()}
					<div class="space-y-3">
						{#each controller.rejectedViews as proposal (proposal.id)}
							<ProposalCard {proposal} />
						{/each}
						{@render moreProposals(controller.rejectedViews.length, controller.rejectedTotal)}
					</div>
				{/snippet}
			</LoadingQueryWrapper>
		{/if}
	</section>

	<Tabs.Root value={controller.tab} onValueChange={controller.setTab} class="flex min-h-96 flex-col gap-3">
		<div class="flex flex-wrap items-end justify-between gap-2">
			<Tabs.List aria-label="Catalogue">
				<Tabs.Trigger value="services">Services</Tabs.Trigger>
				<Tabs.Trigger value="repositories">Repositories</Tabs.Trigger>
			</Tabs.List>
			<div class="grid w-full max-w-xs gap-1">
				<Label for="catalogue-search" class="sr-only">Filter by name</Label>
				<Input id="catalogue-search" bind:value={controller.search} placeholder="Filter by name" />
			</div>
		</div>

		<Tabs.Content value={controller.tab} class="flex min-h-0 flex-1 flex-col">
			<PaginatedQueryListBox {...controller.entities}>
				<LoadingQueryWrapper
					query={controller.entitiesQuery}
					isEmpty={(entities) => entities.length === 0}
				>
					{#snippet loading()}
						<ListSkeleton label="Loading the catalogue" />
					{/snippet}
					{#snippet empty()}
						<p class="p-6 text-sm text-muted-foreground">{controller.emptyEntitiesMessage}</p>
					{/snippet}
					{#snippet view()}
						<ul class="divide-y rounded-lg border bg-card">
							{#each controller.entityRows as entity (entity.id)}
								<EntityRow {entity} />
							{/each}
						</ul>
					{/snippet}
				</LoadingQueryWrapper>
			</PaginatedQueryListBox>
		</Tabs.Content>
	</Tabs.Root>
</section>

<ConfirmProposalDialog />
