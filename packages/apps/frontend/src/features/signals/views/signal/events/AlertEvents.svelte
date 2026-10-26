<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import PageCanvas from "$components/layout/page-canvas/PageCanvas.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { AlertEventsViewController } from "./alertEventsViewController.svelte";

	const events = new AlertEventsViewController();
</script>

<PageCanvas width="reading">
	<LoadingQueryWrapper query={events.query} feedbackOnly />
	<PaginatedQueryListBox {...events.paginatedEventsQuery}>
		{#if events.rows}
			<ul class="divide-y rounded-lg border bg-card">
				{#each events.rows as row (row.id)}
					<li class="flex min-h-[60px] flex-col gap-1 px-4 py-3">
						<div class="flex flex-wrap items-center justify-between gap-x-3 gap-y-1">
							<div class="flex min-w-0 flex-wrap items-center gap-2">
								<Badge variant="neutral">{row.kindLabel}</Badge>
								<span class="min-w-0 text-sm font-medium wrap-anywhere">{row.title}</span>
							</div>
							<Timestamp
								value={row.occurredAt}
								format="absolute"
								class="text-xs text-muted-foreground"
							/>
						</div>
						{#if row.identifier}
							<span class="font-mono text-xs text-muted-foreground wrap-anywhere">
								{row.identifier}
							</span>
						{/if}
					</li>
				{:else}
					<li class="p-6 text-sm text-muted-foreground">No recent activity.</li>
				{/each}
			</ul>
		{/if}
	</PaginatedQueryListBox>
</PageCanvas>
