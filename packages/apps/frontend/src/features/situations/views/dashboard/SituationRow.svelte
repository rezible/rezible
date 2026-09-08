<script lang="ts">
	import { formatDistanceToNow, parseISO } from "date-fns";
	import { Badge } from "$components/ui/badge";
	import * as Button from "$components/ui/button";
	import RiCrosshair2Line from "remixicon-svelte/icons/crosshair-2-line";
	import RiSparkling2Line from "remixicon-svelte/icons/sparkling-2-line";
	import type { Situation } from "$lib/api";
	import {
		attentionReasonLabels,
		deriveSituationPresentation,
	} from "$features/situations/lib/presentation";

	type Props = {
		situation: Situation;
		showCloseReason?: boolean;
	};
	const { situation, showCloseReason = false }: Props = $props();

	const attrs = $derived(situation.attributes);
	const presentation = $derived(deriveSituationPresentation(situation));

	const mapHref = $derived(`/map?focus=${attrs.knowledgeEntityId}`);

	const timeAgo = (iso: string | undefined) => {
		if (!iso) return undefined;
		return formatDistanceToNow(parseISO(iso), { addSuffix: true });
	};

	const signalLabel = $derived.by(() => {
		if (presentation.signalState === "active") {
			return `${presentation.openSignalCount} active signal${presentation.openSignalCount === 1 ? "" : "s"}`;
		}
		if (presentation.signalState === "quiet") return "Signals quiet";
		return "No contributing signals recorded";
	});
</script>

<div class="border-border bg-card hover:border-primary/40 flex flex-col gap-2 border p-3 transition-colors sm:flex-row sm:items-start sm:gap-4">
	<div class="min-w-0 flex-1 space-y-1.5">
		<div class="flex flex-wrap items-center gap-2">
			<span class="truncate text-sm font-medium">{attrs.title}</span>
			<Badge
				variant="outline"
				class={attrs.status === "open" ? "border-primary/40 text-primary" : "text-muted-foreground"}
			>
				{attrs.status === "open" ? "Open" : `Closed · ${attrs.closeReason ?? "unknown reason"}`}
			</Badge>
			{#if showCloseReason && attrs.status === "closed"}
				<span class="text-muted-foreground text-xs">
					Closed {timeAgo(attrs.closedAt)}
				</span>
			{/if}
		</div>

		{#if attrs.summary}
			<p class="text-muted-foreground line-clamp-2 text-xs">{attrs.summary}</p>
		{/if}

		<div class="flex flex-wrap items-center gap-1.5">
			<Badge
				variant="outline"
				class={presentation.signalState === "active" ? "border-destructive/50 text-destructive" : "text-muted-foreground"}
			>
				{signalLabel}
			</Badge>
			{#each presentation.attentionReasons as reason (reason)}
				<Badge variant="secondary" class="text-xs">{attentionReasonLabels[reason]}</Badge>
			{/each}
			{#if presentation.reportLimitations.length > 0}
				<Badge variant="outline" class="text-muted-foreground text-xs">
					Investigation coverage limited ({presentation.reportLimitations.length})
				</Badge>
			{/if}
		</div>

		<div class="text-muted-foreground flex flex-wrap gap-x-4 gap-y-0.5 text-xs">
			<span>Opened {timeAgo(attrs.openedAt)}</span>
			{#if presentation.lastObservedAt}
				<span>Last observed {timeAgo(presentation.lastObservedAt)}</span>
			{/if}
			<span>Updated {timeAgo(attrs.updatedAt)}</span>
			<span>
				{#if presentation.investigationRunning}
					Investigation in progress
				{:else if presentation.reportStale}
					Report out of date with current evidence
				{:else if presentation.hasReport}
					Report current (evidence revision {attrs.evidenceRevision})
				{:else if presentation.hasInvestigation}
					Investigation started; no report yet
				{:else}
					Not investigated
				{/if}
			</span>
		</div>
	</div>

	<div class="flex shrink-0 flex-row items-center gap-2 sm:flex-col sm:items-end">
		<Button.Root variant="outline" size="sm" href={mapHref}>
			<RiCrosshair2Line />
			Map neighborhood
		</Button.Root>
		<span class="text-muted-foreground flex items-center gap-1 text-xs sm:justify-end">
			<RiSparkling2Line class="size-3.5" />
			{presentation.totalSignalCount} signal{presentation.totalSignalCount === 1 ? "" : "s"} linked
		</span>
	</div>
</div>
