<script lang="ts">
	import { resolve } from "$app/paths";
	import { useHomeController } from "./controller.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import SituationStatus from "$features/situations/components/situation-status/SituationStatus.svelte";
	import { Badge } from "$components/ui/badge";
	import { Skeleton } from "$components/ui/skeleton";

	const controller = useHomeController();
	const query = $derived(controller.situationsQuery);
</script>

<section aria-labelledby="home-situations" class="min-w-0">
	<header class="mb-3 flex items-center justify-between gap-3">
		<h2 id="home-situations" class="text-lg font-semibold">
			Situations <span class="ml-2 text-sm font-normal text-muted-foreground">
				{query.data?.pagination.total ?? "—"}
			</span>
		</h2>
		<a
			class="text-sm text-primary hover:underline focus-visible:outline-ring"
			href={"/situations?status=active"}
		>
			View all
		</a>
	</header>
	<div class="overflow-hidden rounded-md border border-border bg-card">
		<LoadingQueryWrapper {query} feedbackOnly>
			{#snippet loading()}
				<div class="divide-y" aria-label="Loading situations">
					{#each [1, 2, 3] as row (row)}
						<div class="space-y-2 px-4 py-3">
							<Skeleton class="h-4 w-3/5" />
							<Skeleton class="h-3 w-4/5" />
						</div>
					{/each}
				</div>
			{/snippet}
		</LoadingQueryWrapper>

		{#if query.data?.data.length}
			<table
				class="hidden w-full text-left text-sm @min-[760px]/main:table [&_th]:px-4 [&_th]:py-3 [&_th]:align-top [&_th]:font-medium [&_td]:px-4 [&_td]:py-3 [&_td]:align-top [&_td:first-child]:wrap-anywhere"
			>
				<thead class="bg-muted/40 text-xs text-muted-foreground">
					<tr>
						<th scope="col">Situation</th>
						<th scope="col">Service</th>
						<th scope="col">Signals</th>
						<th scope="col">Status</th>
					</tr>
				</thead>
				<tbody>
					{#each query.data.data as item (item.id)}
						{@const attrs = item.attributes}
						<tr class="border-t border-border">
							<td class="w-full">
								<a
									class="font-medium hover:underline focus-visible:outline-ring"
									href={resolve("/situations/[id]/[[view=situationView]]", { id: item.id })}
								>
									{attrs.title}
								</a>
								<p class="mt-1 line-clamp-2 text-muted-foreground">{attrs.summary}</p>
							</td>
							<td>
								<Badge variant="secondary" class="text-muted-foreground">TODO</Badge>
							</td>
							<td class="text-xs text-muted-foreground">
								<span class="whitespace-nowrap">
									{attrs.signalCount} contributing {attrs.signalCount === 1
										? "signal"
										: "signals"}
								</span>
							</td>
							<td>
								<SituationStatus attributes={attrs} />
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
			<ul class="divide-y divide-border @min-[760px]/main:hidden">
				{#each query.data.data as item (item.id)}
					{@const attrs = item.attributes}
					<li class="px-4 py-3">
						<div class="flex items-start justify-between gap-3">
							<a
								class="min-w-0 break-words text-sm font-medium hover:underline focus-visible:outline-ring"
								href={resolve("/situations/[id]/[[view=situationView]]", { id: item.id })}
							>
								{attrs.title}
							</a>
							<SituationStatus attributes={attrs} />
						</div>
						<p class="mt-1 line-clamp-2 text-sm text-muted-foreground">{attrs.summary}</p>
						<div class="mt-3 flex items-center gap-3">
							<Badge variant="secondary" class="text-muted-foreground">TODO</Badge>
							<span class="text-xs text-muted-foreground">
								{attrs.signalCount} contributing {attrs.signalCount === 1
									? "signal"
									: "signals"}
							</span>
						</div>
					</li>
				{/each}
			</ul>
		{:else if query.data}
			<p class="p-6 text-center text-sm text-muted-foreground">No active situations.</p>
		{/if}
	</div>
</section>
