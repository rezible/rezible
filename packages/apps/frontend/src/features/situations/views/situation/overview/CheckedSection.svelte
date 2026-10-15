<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import type { CheckedState } from "$features/situations/lib/brief";
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";
	import RiArrowRightSLine from "remixicon-svelte/icons/arrow-right-s-line";
	import ProviderReadCitation from "../investigation/ProviderReadCitation.svelte";
	import { useSituationOverviewController, type ReadExpansion } from "./controller.svelte";

	type Props = { state: CheckedState };
	const { state }: Props = $props();

	const controller = useSituationOverviewController();
</script>

{#snippet readResult(expansion: ReadExpansion)}
	{#if expansion.kind === "loading"}
		<div class="flex items-center gap-2 text-xs text-muted-foreground">
			<Spinner />
			Loading what the read returned
		</div>
	{:else if expansion.kind === "error"}
		<div role="alert" class="flex flex-wrap items-center gap-3 text-xs">
			<span>What the read returned could not be loaded.</span>
			<Button variant="outline" size="sm" onclick={expansion.retry}>Retry</Button>
		</div>
	{:else if expansion.kind === "failed"}
		<p class="text-xs text-muted-foreground">Failed</p>
	{:else if expansion.kind === "unavailable"}
		<p class="text-xs text-muted-foreground">What this read returned is unavailable.</p>
	{:else if expansion.kind === "log_search"}
		{#if expansion.labels.length > 0}
			<p class="text-xs text-muted-foreground wrap-anywhere">
				Every line: {expansion.labels.join(" ")}
			</p>
		{/if}
		{#if expansion.lines.length === 0}
			<p class="text-xs text-muted-foreground">The search returned no lines.</p>
		{:else}
			<ol
				class="flex max-h-80 flex-col gap-1 overflow-auto rounded-md bg-muted/50 p-2 font-mono text-xs"
			>
				{#each expansion.lines as line (line.key)}
					<li class="flex flex-col">
						<div class="flex flex-wrap items-baseline gap-x-2">
							<Timestamp value={line.time} format="clock" class="text-muted-foreground" />
							{#if line.repeats}
								<span class="text-muted-foreground">{line.repeats}</span>
							{/if}
							<span class="whitespace-pre-wrap wrap-anywhere">{line.line}</span>
						</div>
						{#if line.labels.length > 0}
							<span class="text-muted-foreground wrap-anywhere">{line.labels.join(" ")}</span>
						{/if}
					</li>
				{/each}
			</ol>
		{/if}
	{:else}
		<div class="flex flex-col gap-1">
			<div
				class="flex h-16 items-end gap-px rounded-md bg-muted/50 p-1"
				role="img"
				aria-label="Log lines counted over time"
			>
				{#each expansion.points as point (point.key)}
					<div
						class="min-w-px flex-1 rounded-t-sm bg-primary/70"
						style:height={`${point.heightPercent}%`}
						title={`${point.value} at ${point.time}`}
					></div>
				{/each}
			</div>
			<p class="text-xs text-muted-foreground">
				{expansion.range}
				{#if expansion.step}
					· one bar per {expansion.step}
				{/if}
			</p>
		</div>
	{/if}
{/snippet}

<section aria-labelledby="checked-title" class="flex flex-col gap-3">
	<SectionHeading id="checked-title" title="What Rezible checked" />

	{#if state.kind === "no_report"}
		<p class="text-sm text-muted-foreground">No report has been published, so no reads are cited yet.</p>
	{:else if state.kind === "none_cited"}
		<p class="text-sm text-muted-foreground">The latest report cites no provider reads.</p>
	{:else}
		<ul class="flex flex-col gap-3">
			{#each state.reads as checked (checked.id)}
				{@const expanded = controller.expandedReadIds.has(checked.id)}
				{@const expansion = controller.readExpansions.get(checked.id)}
				<li class="flex flex-col gap-2">
					<div class="flex items-start gap-1">
						<Button
							variant="ghost"
							size="icon-sm"
							aria-expanded={expanded}
							aria-label={expanded
								? "Hide what the read returned"
								: "Show what the read returned"}
							onclick={() => controller.toggleRead(checked.id)}
						>
							{#if expanded}
								<RiArrowDownSLine />
							{:else}
								<RiArrowRightSLine />
							{/if}
						</Button>
						<ProviderReadCitation number={checked.number} read={checked.read} />
					</div>
					{#if expanded && expansion}
						<div class="pl-9">
							{@render readResult(expansion)}
						</div>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</section>
