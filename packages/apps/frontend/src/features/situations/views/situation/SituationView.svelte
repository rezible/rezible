<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { initSituationController } from "./controller.svelte";

	type Props = {
		id: string;
	};

	let { id }: Props = $props();

	const controller = initSituationController(() => id);
	const attrs = $derived(controller.situation?.attributes);
	const report = $derived(controller.report);
	const reportSections = $derived.by(() => {
		if (!report) return [];

		return [
			{ title: "Recommended actions", items: report.recommendedActions },
			{ title: "Suggested checks", items: report.suggestedChecks },
			{ title: "Limitations", items: report.limitations },
		];
	});
	registerPageDescriptor(() => ({
		title: attrs?.title ?? "Situation",
		breadcrumbs: [{ title: "Situations", path: resolve("/situations") }],
	}));
</script>

<div class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
	<LoadingQueryWrapper query={controller.query} feedbackOnly />
	{#if controller.query.error?.status === 404}
		<p class="text-sm text-muted-foreground">This situation is unavailable.</p>
	{/if}
	{#if attrs}
		<header class="flex flex-wrap items-start justify-between gap-3">
			<div class="min-w-0 space-y-2">
				<h1 class="break-words text-xl font-semibold">{attrs.title}</h1>
				<div class="flex flex-wrap items-center gap-2">
					<Badge variant="outline" class="capitalize">{attrs.status}</Badge>
					{#if attrs.closeReason}
						<span class="text-sm capitalize text-muted-foreground">{attrs.closeReason}</span>
					{/if}
				</div>
			</div>
			<Button href={controller.mapHref} variant="outline" size="sm">Open System Map</Button>
		</header>
		<section class="space-y-3 rounded-lg border border-border bg-card p-4" aria-label="Situation summary">
			<p class="whitespace-pre-wrap text-sm">{attrs.summary || "No summary recorded."}</p>
			<dl class="grid gap-3 text-sm sm:grid-cols-3">
				<div>
					<dt class="text-muted-foreground">Opened</dt>
					<dd>
						<time datetime={attrs.openedAt}>{new Date(attrs.openedAt).toLocaleString()}</time>
					</dd>
				</div>
				<div>
					<dt class="text-muted-foreground">Updated</dt>
					<dd>
						<time datetime={attrs.updatedAt}>{new Date(attrs.updatedAt).toLocaleString()}</time>
					</dd>
				</div>
				{#if attrs.closedAt}
					<div>
						<dt class="text-muted-foreground">Closed</dt>
						<dd>
							<time datetime={attrs.closedAt}>{new Date(attrs.closedAt).toLocaleString()}</time>
						</dd>
					</div>
				{/if}
			</dl>
		</section>
		<section class="rounded-lg border border-border bg-card" aria-labelledby="signals-title">
			<h2 id="signals-title" class="p-4 font-semibold">
				Contributing signals
				<span class="font-normal text-muted-foreground">{attrs.alertEpisodes.length}</span>
			</h2>
			{#each attrs.alertEpisodes as episode (episode.id)}
				{@const epAttrs = episode.attributes}
				{@const alertDef = epAttrs.definition}
				<div class="space-y-2 border-t border-border p-4 text-sm">
					<div class="flex flex-wrap items-center gap-2">
						{#if alertDef}
							<a
								class="font-medium text-primary hover:underline"
								href={resolve("/signals/[id]/[[view=signalView]]", {id: alertDef.id})}
							>
								{alertDef.attributes.title || "Signal"}
							</a>
						{:else}
							<span class="break-all font-medium">Signal episode {episode.id}</span>
						{/if}
						<Badge variant="outline" class="capitalize">{epAttrs.status}</Badge>
					</div>
					<div class="flex flex-wrap gap-x-6 gap-y-1 text-xs text-muted-foreground">
						<span>Started {new Date(epAttrs.startedAt).toLocaleString()}</span>
						<span>
							Last observed {new Date(epAttrs.lastObservedAt).toLocaleString()}
						</span>
						{#if epAttrs.closedAt}
							<span>Closed {new Date(epAttrs.closedAt).toLocaleString()}</span>
						{/if}
					</div>
				</div>
			{:else}
				<p class="px-4 pb-4 text-sm text-muted-foreground">No contributing signals recorded.</p>
			{/each}
		</section>
		<section class="space-y-4 rounded-lg border border-border bg-card p-4" aria-labelledby="report-title">
			<div>
				<h2 id="report-title" class="font-semibold">Investigation report</h2>
				<p class="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
					<span>Current evidence revision: {attrs.evidenceRevision}</span>
					{#if report && attrs.investigation}
						<span>Report covers revision: {attrs.investigation.attributes.completedRevision}</span>
					{/if}
				</p>
			</div>
			{#if report}
				{#if attrs.investigation}
					<p class="text-xs text-muted-foreground">
						Updated {new Date(attrs.investigation.attributes.updatedAt).toLocaleString()}
					</p>
				{/if}
				{#if report.text}
					<p class="whitespace-pre-wrap text-sm">{report.text}</p>
				{/if}
				{#if report.likelyCause}
					<div>
						<h3 class="mb-1 text-sm font-medium">Likely cause</h3>
						<p class="whitespace-pre-wrap text-sm">{report.likelyCause}</p>
					</div>
				{/if}
				{#if report.bestNextStep}
					<div>
						<h3 class="mb-1 text-sm font-medium">Best next step</h3>
						<p class="whitespace-pre-wrap text-sm">{report.bestNextStep}</p>
					</div>
				{/if}
				{#each reportSections as section (section.title)}
					{#if section.items?.length}
						<div>
							<h3 class="mb-1 text-sm font-medium">{section.title}</h3>
							<ul class="list-disc space-y-1 pl-5 text-sm">
								{#each section.items as item, index (index)}
									<li class="whitespace-pre-wrap">
										{item}
									</li>
								{/each}
							</ul>
						</div>
					{/if}
				{/each}
			{:else}
				<p class="text-sm text-muted-foreground">No investigation report recorded.</p>
			{/if}
		</section>
	{/if}
</div>
