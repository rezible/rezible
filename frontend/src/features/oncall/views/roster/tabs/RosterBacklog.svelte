<script lang="ts">
	import type { BacklogItem } from "../types";
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { rosterIdCtx } from "$features/oncall/views/roster/context";
	import { Header } from "svelte-ux";
	
	const rosterId = rosterIdCtx.get();

	const mockBacklogItems: BacklogItem[] = [
		// TODO: fill this with some items
	];

	const getRosterBacklogOptions = (options: {path: {id: string}}) => {
		return queryOptions({
			queryFn: async ({ queryKey, signal }) => {
				return {data: mockBacklogItems};
			},
			queryKey: ["getUserOncallDetailsOptions", options.path.id]
		});
	};

	const backlogQuery = createQuery(() => getRosterBacklogOptions({path: {id: rosterId}}));
</script>

<div class="grid grid-cols-2 gap-2 h-full">
	<div class="flex flex-col p-2">
		<Header title="Tickets" class="text-lg font-medium" />

		<div class="flex-1 flex flex-col gap-1 overflow-y-auto border">
			<LoadingQueryWrapper query={backlogQuery}>
				{#snippet view(items: BacklogItem[])}
					{#each items as item, i}
						<!-- TODO: create this list -->
						<span>{JSON.stringify(item)}</span>
					{:else}
						<div class="text-surface-600 italic p-2">Backlog empty!</div>
					{/each}
				{/snippet}
			</LoadingQueryWrapper>
		</div>
	</div>

	<div class="">
<pre>Burndown chart (candlesticks?)
Time spent on toil work (by category/label/service?)
Trends</pre>
	</div>
</div>