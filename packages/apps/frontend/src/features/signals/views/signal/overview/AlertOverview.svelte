<script lang="ts">
	import PageCanvas from "$components/layout/page-canvas/PageCanvas.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import Avatar from "$components/common/entity-avatar/EntityAvatar.svelte";
	import { initAlertOverviewController } from "./controller.svelte";
	import AttentionPanel from "./AttentionPanel.svelte";
	import IdentityLabelsPanel from "./IdentityLabelsPanel.svelte";

	const overview = initAlertOverviewController();
</script>

{#snippet settings()}
	<AttentionPanel />
	<IdentityLabelsPanel />
{/snippet}

{#snippet metricsList()}
	<dl class="divide-y rounded-lg border bg-card">
		{#each overview.metricRows as metric (metric.key)}
			<div class="flex items-baseline justify-between gap-4 px-3 py-2">
				<dt class="text-sm text-muted-foreground">{metric.label}</dt>
				<dd class="text-sm font-medium tabular-nums">{metric.value}</dd>
			</div>
		{/each}
	</dl>
{/snippet}

{#if overview.definition}
	<PageCanvas context={settings} contextLabel="Signal settings">
		<header class="flex flex-col gap-3">
			<h1 class="text-[28px] leading-9 font-semibold tracking-tight wrap-anywhere">
				{overview.title}
			</h1>
			{#if overview.description}
				<p
					class="max-w-[72ch] text-[15px] leading-6 whitespace-pre-wrap text-muted-foreground wrap-anywhere"
				>
					{overview.description}
				</p>
			{:else}
				<p class="text-sm text-muted-foreground">No description provided.</p>
			{/if}
		</header>

		<section aria-labelledby="signal-activity-title" class="flex max-w-xl flex-col gap-2">
			<div class="flex items-baseline justify-between gap-2">
				<h2 id="signal-activity-title" class="region-label">Activity</h2>
				<span class="text-xs text-muted-foreground">{overview.metricsPeriodLabel}</span>
			</div>
			<LoadingQueryWrapper query={overview.metricsQuery} view={metricsList} />
		</section>

		{#if overview.definitionText}
			<section aria-labelledby="signal-definition-title" class="flex flex-col gap-2">
				<h2 id="signal-definition-title" class="region-label">Definition</h2>
				<p
					class="rounded-lg border bg-card px-3 py-2 font-mono text-sm whitespace-pre-wrap wrap-anywhere"
				>
					{overview.definitionText}
				</p>
			</section>
		{/if}

		{#if overview.rosterId}
			<section aria-labelledby="signal-roster-title" class="flex flex-col gap-2">
				<h2 id="signal-roster-title" class="region-label">Roster</h2>
				<div class="flex items-center gap-2">
					<Avatar kind="roster" size={24} id={overview.rosterId} />
					<span class="text-sm font-medium">{overview.rosterName}</span>
				</div>
			</section>
		{/if}
	</PageCanvas>
{/if}
