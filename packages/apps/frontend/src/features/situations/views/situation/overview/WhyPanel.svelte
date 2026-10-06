<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import { Skeleton } from "$components/ui/skeleton";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { situationHref } from "$features/situations/lib/routes";
	import type { WhyAssessment, WhyClosure, WhyOrigin, WhyRaise } from "$features/situations/lib/why";
	import { cn } from "$lib/utils";
	import RiCheckboxBlankCircleLine from "remixicon-svelte/icons/checkbox-blank-circle-line";
	import RiCheckboxCircleFill from "remixicon-svelte/icons/checkbox-circle-fill";
	import { useSituationOverviewController } from "./controller.svelte";

	const controller = useSituationOverviewController();
	const why = $derived(controller.why);
	const mergedInto = $derived(why && "closure" in why ? why.closure?.mergedInto : undefined);

	const raiseActor: Record<WhyRaise["by"], string> = {
		person: "By a person",
		rezible: "Automatically, by this assessment",
		system: "By a system action",
	};
</script>

{#snippet originFact(origin: WhyOrigin | undefined)}
	<div class="flex flex-col gap-1 sm:contents">
		<dt class="text-muted-foreground">What started it</dt>
		<dd class="flex min-w-0 flex-col gap-0.5">
			{#if origin}
				<span class="font-medium wrap-anywhere">{origin.title}</span>
				<span class="text-xs text-muted-foreground">
					Started
					<Timestamp value={origin.startedAt} />
				</span>
			{:else if controller.originLoading}
				<Skeleton class="h-4 w-1/2" aria-label="Loading the original signal" />
			{:else if mergedInto}
				<span class="text-muted-foreground">
					Its signals moved to
					<a
						class="font-medium text-foreground hover:underline"
						href={situationHref(mergedInto.id)}
					>
						{mergedInto.title}
					</a>
					when it was merged.
				</span>
			{:else}
				<span class="text-muted-foreground">The original signal's details are unavailable.</span>
			{/if}
		</dd>
	</div>
{/snippet}

{#snippet raiseFact(raise: WhyRaise)}
	<div class="flex flex-col gap-1 sm:contents">
		<dt class="text-muted-foreground">Raised</dt>
		<dd class="flex min-w-0 flex-col gap-0.5">
			<span>
				{raiseActor[raise.by]} ·
				<Timestamp value={raise.at} />
			</span>
			{#if raise.reason}
				<span class="text-xs text-muted-foreground wrap-anywhere">Reason: {raise.reason}</span>
			{:else if raise.by !== "rezible"}
				<span class="text-xs text-muted-foreground">No reason was recorded.</span>
			{/if}
		</dd>
	</div>
{/snippet}

{#snippet closureFact(closure: WhyClosure)}
	<div class="flex flex-col gap-1 sm:contents">
		<dt class="text-muted-foreground">Closed</dt>
		<dd>
			<Timestamp value={closure.at} />
			{#if closure.reasonLabel}
				· {closure.reasonLabel}
			{/if}
		</dd>
	</div>
{/snippet}

{#snippet assessment(assessment: WhyAssessment, note?: string)}
	<div class="flex flex-col gap-3 border-t pt-4">
		<div class="flex flex-wrap items-center justify-between gap-2">
			<h3 class="text-sm font-medium">
				{assessment.title} recorded
				<Timestamp value={assessment.judgedAt} format="absolute" />
			</h3>
			<StatusBadge status={assessment.verdict} />
		</div>
		{#if note}
			<p class="text-xs text-muted-foreground">{note}</p>
		{/if}
		{#if assessment.notice}
			<p class="text-sm">{assessment.notice}</p>
		{/if}

		<ul aria-label="Reasons as recorded" class="grid gap-x-6 gap-y-3 sm:grid-cols-2">
			{#each assessment.reasons as reason (reason.key)}
				<li class="flex min-w-0 items-start gap-2">
					{#if reason.met}
						<RiCheckboxCircleFill
							class="mt-0.5 size-4 shrink-0 text-foreground"
							aria-hidden="true"
						/>
					{:else}
						<RiCheckboxBlankCircleLine
							class="mt-0.5 size-4 shrink-0 text-muted-foreground"
							aria-hidden="true"
						/>
					{/if}
					<div class="flex min-w-0 flex-col gap-0.5">
						<p class="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
							<span
								class={cn(
									reason.cited && "font-semibold",
									!reason.met && "text-muted-foreground"
								)}
							>
								{reason.label}
							</span>
							<span class="sr-only">{reason.met ? "met" : "not met"}</span>
							{#if reason.cited}
								<Badge variant="secondary">Cited</Badge>
							{/if}
						</p>
						{#if reason.detail}
							<p class="text-xs text-muted-foreground wrap-anywhere">{reason.detail}</p>
						{/if}
					</div>
				</li>
			{/each}
		</ul>

		{#if assessment.explanation}
			<div class="flex flex-col gap-1">
				<h4 class="region-label">Explanation</h4>
				{#if assessment.explanationContext}
					<p class="text-xs text-muted-foreground">{assessment.explanationContext}</p>
				{/if}
				<p class="max-w-[72ch] text-sm whitespace-pre-wrap wrap-anywhere">{assessment.explanation}</p>
			</div>
		{/if}

		<p class="flex flex-wrap items-center gap-x-1.5 text-xs text-muted-foreground">
			<span>{assessment.assessedBy}</span>
			{#if assessment.checksAgain}
				<span aria-hidden="true">·</span>
				<span>Rezible checks again as evidence, context and time change.</span>
			{/if}
		</p>
	</div>
{/snippet}

{#if why}
	<section aria-labelledby="why-title" class="flex flex-col gap-4 rounded-lg border bg-card p-5">
		<h2 id="why-title" class="region-label">{why.heading}</h2>

		<dl class="grid grid-cols-1 gap-x-4 gap-y-3 text-sm sm:grid-cols-[9rem_minmax(0,1fr)]">
			{@render originFact(why.origin)}
			{#if why.kind === "muted"}
				<div class="flex flex-col gap-1 sm:contents">
					<dt class="text-muted-foreground">Muted</dt>
					<dd>
						<Timestamp value={why.mutedAt} />
						{#if why.reasonLabel}
							· {why.reasonLabel}
						{/if}
					</dd>
				</div>
				{#if why.lastAssessedAt}
					<div class="flex flex-col gap-1 sm:contents">
						<dt class="text-muted-foreground">Last assessed</dt>
						<dd><Timestamp value={why.lastAssessedAt} format="absolute" /></dd>
					</div>
				{/if}
			{:else if why.kind === "raised"}
				{@render raiseFact(why.raise)}
			{:else if why.kind === "assessed" || why.kind === "closed-unassessed"}
				{#if why.raise}
					{@render raiseFact(why.raise)}
				{/if}
				{#if why.closure}
					{@render closureFact(why.closure)}
				{/if}
			{/if}
		</dl>

		{#if why.kind === "pending"}
			<p class="text-sm text-muted-foreground">
				Rezible has not assessed this yet. It checks as evidence, context and time change.
			</p>
		{:else if why.kind === "muted"}
			<p class="text-sm text-muted-foreground">Muted; not being assessed.</p>
		{:else if why.kind === "closed-unassessed"}
			<p class="text-sm text-muted-foreground">No automatic assessment was recorded.</p>
		{:else if why.kind === "raised"}
			{#if why.earlierAssessment}
				{@render assessment(
					why.earlierAssessment,
					"Recorded before the raise; it did not raise this situation."
				)}
			{:else}
				<p class="text-sm text-muted-foreground">No automatic assessment was recorded.</p>
			{/if}
		{:else}
			{@render assessment(why.assessment)}
		{/if}
	</section>
{/if}
