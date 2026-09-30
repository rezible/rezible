<script lang="ts">
	import { resolve } from "$app/paths";
	import { Badge } from "$components/ui/badge";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { initIncidentOverviewController } from "./controller.svelte";

	const overview = initIncidentOverviewController();
	const attrs = $derived(overview.attributes);
</script>

<div class="min-h-0 flex-1 overflow-y-auto">
	<div class="mx-auto grid w-full max-w-[1240px] gap-8 p-6 lg:grid-cols-[minmax(0,1fr)_320px]">
		<main class="flex min-w-0 flex-col gap-6">
			<h1 class="text-2xl font-semibold">{attrs?.title ?? "Incident"}</h1>
			{#if attrs}
				<p class="whitespace-pre-wrap text-muted-foreground">{attrs.summary}</p>
				<dl class="grid gap-4 text-sm sm:grid-cols-3">
					<div>
						<dt class="text-muted-foreground">Response state</dt>
						<dd class="capitalize">{attrs.responseState}</dd>
					</div>
					<div>
						<dt class="text-muted-foreground">Severity</dt>
						<dd>{attrs.severity?.attributes?.name ?? "Unspecified"}</dd>
					</div>
					<div>
						<dt class="text-muted-foreground">Type</dt>
						<dd>{attrs.type?.attributes?.name ?? "Unspecified"}</dd>
					</div>
					<div>
						<dt class="text-muted-foreground">Opened</dt>
						<dd>{overview.dateLabel(attrs.openedAt)}</dd>
					</div>
					<div>
						<dt class="text-muted-foreground">Resolved</dt>
						<dd>{overview.dateLabel(attrs.resolvedAt)}</dd>
					</div>
					<div>
						<dt class="text-muted-foreground">Updated</dt>
						<dd>{overview.dateLabel(attrs.updatedAt)}</dd>
					</div>
					{#each attrs.fieldSelections as field (field.fieldId)}
						<div>
							<dt class="text-muted-foreground">{field.fieldName}</dt>
							<dd>{field.option.attributes.value}</dd>
						</div>
					{/each}
				</dl>
				{#if attrs.tags.length}
					<div class="flex flex-wrap gap-2" aria-label="Tags">
						{#each attrs.tags as tag (tag.id)}
							<Badge variant="secondary">{tag.attributes.key}: {tag.attributes.value}</Badge>
						{/each}
					</div>
				{/if}
				{#if overview.milestones.length}
					<section class="flex flex-col gap-3 border-t pt-4" aria-labelledby="incident-milestones">
						<h2 id="incident-milestones" class="font-semibold">Milestones</h2>
						<ol class="flex flex-col gap-3 text-sm">
							{#each overview.milestones as milestone (milestone.id)}
								<li>
									<div class="flex flex-wrap justify-between gap-2">
										<span class="font-medium capitalize">
											{milestone.attributes.kind}
										</span>
										<time datetime={milestone.attributes.timestamp}>
											{overview.dateLabel(milestone.attributes.timestamp)}
										</time>
									</div>
									<p class="whitespace-pre-wrap text-muted-foreground">
										{milestone.attributes.description}
									</p>
								</li>
							{/each}
						</ol>
					</section>
				{/if}
				{#if overview.impacts.length}
					<section class="flex flex-col gap-3 border-t pt-4" aria-labelledby="incident-impacts">
						<h2 id="incident-impacts" class="font-semibold">Impacts</h2>
						{#each overview.impacts as impact (impact.id)}
							<div class="text-sm">
								<p class="font-medium">{impact.name}</p>
								<p class="whitespace-pre-wrap">{impact.note}</p>
								{#if impact.source}
									<p class="text-muted-foreground">Source: {impact.source}</p>
								{/if}
							</div>
						{/each}
					</section>
				{/if}
			{/if}
		</main>
		<aside class="flex min-w-0 flex-col gap-6 text-sm">
			<section class="flex flex-col gap-3" aria-labelledby="incident-retrospective">
				<h2 id="incident-retrospective" class="font-semibold">Retrospective</h2>
				{#if overview.view.incidentRetrospectiveId}
					<LoadingQueryWrapper query={overview.view.retrospectiveQuery} feedbackOnly />
					{#if overview.retrospective && attrs}
						<p class="capitalize">{overview.retrospectiveState}</p>
						<a
							class="underline"
							href={resolve("/incidents/[slug]/[[view=incidentView]]", {
								slug: attrs.slug,
								view: "report",
							})}
						>
							Open report
						</a>
					{/if}
				{:else}
					<p class="text-muted-foreground">No retrospective is associated with this incident.</p>
				{/if}
			</section>
			{#if overview.roles.length}
				<section class="flex flex-col gap-3 border-t pt-4" aria-labelledby="incident-roles">
					<h2 id="incident-roles" class="font-semibold">Roles</h2>
					{#each overview.roles as assignment (assignment.id)}
						<div>
							<p>{assignment.user}</p>
							<p class="text-muted-foreground">{assignment.role}</p>
						</div>
					{/each}
				</section>
			{/if}
			{#if overview.view.situations.length}
				<section class="flex flex-col gap-3 border-t pt-4" aria-labelledby="incident-situations">
					<h2 id="incident-situations" class="font-semibold">Linked situations</h2>
					{#each overview.view.situations as situation (situation.id)}
						<div>
							<a
								class="underline"
								href={resolve("/situations/[id]/[[view=situationView]]", {
									id: situation.id,
								})}
							>
								{situation.title}
							</a>
							<p class="text-muted-foreground">{situation.summary}</p>
						</div>
					{/each}
				</section>
			{/if}
			{#if overview.ticketUrl}
				<a class="underline" href={overview.ticketUrl} target="_blank" rel="noreferrer">
					{attrs?.externalTicket?.title || "Open external ticket"}
				</a>
			{/if}
		</aside>
	</div>
</div>
