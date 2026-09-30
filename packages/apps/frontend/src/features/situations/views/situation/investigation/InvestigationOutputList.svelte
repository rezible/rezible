<script lang="ts">
	import { Button } from "$components/ui/button";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import RichTextView from "$components/rich-text-view/RichTextView.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import { cn } from "$lib/utils";
	import CitationChips from "./CitationChips.svelte";
	import { OUTPUT_PAGE_SIZE, useSituationInvestigationController } from "./controller.svelte";

	type Props = { kind: "findings" | "hypotheses" };
	const { kind }: Props = $props();

	const controller = useSituationInvestigationController();
	const list = $derived(controller.outputList(kind));
</script>

<LoadingQueryWrapper query={list.query} feedback="quiet" isEmpty={() => list.rows.length === 0}>
	{#snippet loading()}
		<ListSkeleton label={`Loading ${list.name}`} />
	{/snippet}
	{#snippet error()}
		<span>Could not load {list.name}.</span>
	{/snippet}
	{#snippet empty()}
		<p class="text-sm text-muted-foreground">No {list.name} have been published yet.</p>
	{/snippet}
	{#snippet view()}
		<ol aria-label={list.label} class="divide-y rounded-lg border bg-card">
			{#each list.rows as row (row.id)}
				{@const expanded = controller.expandedOutputIds.has(row.id)}
				{@const hasDetails = row.markdown.trim().length > 0}
				<li
					class="grid min-h-[60px] grid-cols-[28px_minmax(0,1fr)] gap-x-3 gap-y-1 px-4 py-3 sm:grid-cols-[28px_minmax(0,1fr)_auto]"
				>
					<span
						aria-hidden="true"
						class="grid size-7 place-items-center rounded-full bg-muted text-xs font-semibold tabular-nums"
					>
						{row.number}
					</span>
					<div class="flex min-w-0 flex-col gap-1">
						<div class="flex flex-wrap items-center gap-x-2 gap-y-1">
							<p
								class={cn(
									"text-[15px] leading-[22px] font-medium wrap-anywhere",
									row.muted &&
										"text-muted-foreground line-through decoration-muted-foreground/60"
								)}
							>
								{row.title}
							</p>
							{#each row.badges as badge (badge.label)}
								<StatusBadge status={badge} />
							{/each}
						</div>
						{#if expanded}
							<RichTextView
								markdown={row.markdown}
								size="compact"
								headingOffset={2}
								label={`${row.title} details`}
							/>
						{:else if row.summary}
							<p class="line-clamp-2 text-sm text-muted-foreground">{row.summary}</p>
						{:else if hasDetails}
							<p class="text-sm text-muted-foreground">Details available.</p>
						{/if}
						{#if hasDetails}
							<Button
								variant="link"
								size="inline"
								class="self-start"
								aria-expanded={expanded}
								onclick={() => controller.toggleOutput(row.id)}
							>
								{expanded ? "Show less" : "Show more"}
							</Button>
						{/if}
						{#if row.relations.length}
							<p class="text-xs text-muted-foreground">{row.relations.join(" · ")}</p>
						{/if}
						<div class="sm:hidden">
							<CitationChips citations={row.citations} />
						</div>
					</div>
					<div class="max-sm:hidden">
						<CitationChips citations={row.citations} />
					</div>
				</li>
			{/each}
		</ol>
		{#if list.total > OUTPUT_PAGE_SIZE}
			<p class="text-xs text-muted-foreground">Showing the first {OUTPUT_PAGE_SIZE} of {list.total}.</p>
		{/if}
	{/snippet}
</LoadingQueryWrapper>
