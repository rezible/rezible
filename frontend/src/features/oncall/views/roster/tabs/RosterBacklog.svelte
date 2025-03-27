<script lang="ts">
	import { mdiChevronRight, mdiAccount } from "@mdi/js";
	import { Header, Icon } from "svelte-ux";
	import type { BacklogItem, OncallRoster } from "../types";
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { rosterIdCtx } from "$features/oncall/views/roster/context";
	
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

<div class="border rounded-lg flex gap-2 row-span-1 py-1">
	<div class="flex flex-col gap-2 w-full">
		<Header title="Backlog Items" classes={{root: "gap-2 text-lg font-medium px-2 pt-2"}}>
			
		</Header>

		<div class="flex-1 flex flex-col gap-1 overflow-y-auto">
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
</div>
