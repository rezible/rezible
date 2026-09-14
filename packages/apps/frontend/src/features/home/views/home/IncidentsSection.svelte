<script lang="ts">
	import { resolve } from "$app/paths";
	import { useHomeController } from "./controller.svelte";
	import DisplayTime from "./DisplayTime.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import IncidentSeverity from "$features/incidents/components/incident-severity/IncidentSeverity.svelte";
	import { Badge } from "$components/ui/badge";
	import {
		incidentPriorityClasses,
		incidentSeverityVariant,
	} from "$features/incidents/components/incident-severity/severity";
	import { Skeleton } from "$components/ui/skeleton";

	const controller = useHomeController();
	const query = $derived(controller.incidentsQuery);

	const numIncidents = $derived(query.data?.pagination.total ?? "-");
</script>

<section aria-labelledby="home-incidents" class="min-w-0">
	<header class="mb-3 flex items-center justify-between gap-3">
		<h2 id="home-incidents" class="text-lg font-semibold">
			Active incidents
			<span class="ml-2 text-sm font-normal text-muted-foreground">{numIncidents}</span>
		</h2>
		<a
			class="text-sm text-primary hover:underline focus-visible:outline-ring"
			href={"/incidents?status=active"}
		>
			View all
		</a>
	</header>
	<div class="overflow-hidden rounded-md border border-border bg-card">
		<LoadingQueryWrapper {query} feedbackOnly>
			{#snippet loading()}
				<div class="divide-y" aria-label="Loading incidents">
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
						<th scope="col">Incident</th>
						<th scope="col">Severity</th>
						<th scope="col">Services</th>
						<th scope="col" class="text-right">Updated</th>
					</tr>
				</thead>
				<tbody>
					{#each query.data.data as item, index (item.id)}
						{@const attrs = item.attributes}
						<tr
							class={[
								"border-t border-border",
								index === 0 &&
									incidentPriorityClasses[incidentSeverityVariant(attrs.severity)],
							]}
						>
							<td class="w-full">
								<a
									class="font-medium hover:underline focus-visible:outline-ring"
									href={resolve("/incidents/[slug]/[[view=incidentView]]", {
										slug: attrs.slug,
									})}
								>
									{attrs.title}
								</a>
								<p class="mt-1 line-clamp-2 text-muted-foreground">{attrs.summary}</p>
							</td>
							<td>
								<IncidentSeverity severity={attrs.severity} />
							</td>
							<td>
								<Badge variant="secondary" class="text-muted-foreground">TODO</Badge>
							</td>
							<td class="text-right text-xs text-muted-foreground">
								<DisplayTime value={attrs.updatedAt} />
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
			<ul class="divide-y divide-border @min-[760px]/main:hidden">
				{#each query.data.data as item, index (item.id)}
					{@const attrs = item.attributes}
					<li
						class={[
							"px-4 py-3",
							index === 0 && incidentPriorityClasses[incidentSeverityVariant(attrs.severity)],
						]}
					>
						<div class="flex items-start justify-between gap-3">
							<a
								class="min-w-0 break-words text-sm font-medium hover:underline focus-visible:outline-ring"
								href={resolve("/incidents/[slug]/[[view=incidentView]]", {
									slug: attrs.slug,
								})}
							>
								{attrs.title}
							</a>
							<IncidentSeverity severity={attrs.severity} />
						</div>
						<p class="mt-1 line-clamp-2 text-sm text-muted-foreground">{attrs.summary}</p>
						<div class="mt-3 flex items-center justify-between gap-3">
							<Badge variant="secondary" class="text-muted-foreground">TODO</Badge>
							<DisplayTime value={attrs.updatedAt} />
						</div>
					</li>
				{/each}
			</ul>
		{:else if query.data}
			<p class="p-6 text-center text-sm text-muted-foreground">No active incidents.</p>
		{/if}
	</div>
</section>
