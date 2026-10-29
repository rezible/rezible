<script lang="ts">
	import type { Snippet } from "svelte";
	import RiFilter3Line from "remixicon-svelte/icons/filter-3-line";
	import { Button } from "$components/ui/button";
	import * as Sheet from "$components/ui/sheet";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import { IsMobile } from "$lib/hooks/is-mobile.svelte";
	import { cn } from "$lib/utils";

	type Props = {
		filters?: Snippet;
		/** Results heading, such as "Incidents". */
		title?: string;
		/** Result total, shown after the title. Omit while unknown. */
		count?: number;
		/** Right-aligned actions beside the results heading. */
		actions?: Snippet;
		/** Shows a quiet refreshing note beside the results heading. */
		updating?: boolean;
		/** Number of filters that differ from their defaults. */
		activeFilters?: number;
		/** Resets every filter; shown when any filter is active. */
		onclear?: () => void;
		children: Snippet;
	};
	const {
		filters,
		title,
		count,
		actions,
		updating = false,
		activeFilters = 0,
		onclear,
		children,
	}: Props = $props();

	const id = $props.id();
	const isMobile = new IsMobile();
	let sheetOpen = $state(false);
</script>

{#snippet filterControls()}
	<div class="flex flex-col gap-5">
		{@render filters?.()}
	</div>
{/snippet}

{#snippet headingActions()}
	{#if updating}
		<span class="text-xs text-muted-foreground" aria-live="polite">Updating…</span>
	{/if}
	{@render actions?.()}
{/snippet}

{#snippet clearButton()}
	{#if onclear && activeFilters > 0}
		<Button variant="link" size="inline" onclick={onclear}>Clear filters</Button>
	{/if}
{/snippet}

<div class="mx-auto w-full max-w-[1280px] min-w-0 p-4 sm:p-6">
	<div
		class={cn(
			"flex flex-col gap-4",
			filters && "md:grid md:grid-cols-[240px_minmax(0,1fr)] md:items-start md:gap-6"
		)}
	>
		{#if filters}
			{#if isMobile.current}
				<Sheet.Root bind:open={sheetOpen}>
					<Sheet.Trigger>
						{#snippet child({ props })}
							<Button {...props} variant="outline" class="self-start">
								<RiFilter3Line aria-hidden="true" />
								Filters
								{#if activeFilters > 0}
									<span class="tabular-nums text-muted-foreground">({activeFilters})</span>
								{/if}
							</Button>
						{/snippet}
					</Sheet.Trigger>
					<Sheet.Content side="left" class="gap-5 p-4">
						<Sheet.Header class="flex-row items-center justify-between p-0 pr-8">
							<Sheet.Title>Filters</Sheet.Title>
							{@render clearButton()}
						</Sheet.Header>
						{@render filterControls()}
					</Sheet.Content>
				</Sheet.Root>
			{:else}
				<aside
					aria-labelledby={`${id}-filters`}
					class="sticky top-6 flex flex-col gap-4 rounded-lg border bg-card p-4"
				>
					<header class="flex min-h-6 items-center justify-between gap-2">
						<h2 id={`${id}-filters`} class="text-sm font-semibold">Filters</h2>
						{@render clearButton()}
					</header>
					{@render filterControls()}
				</aside>
			{/if}
		{/if}

		<section aria-labelledby={title ? `${id}-results` : undefined} class="flex min-w-0 flex-col gap-3">
			{#if title}
				<SectionHeading
					id={`${id}-results`}
					{title}
					{count}
					actions={updating || actions ? headingActions : undefined}
				/>
			{/if}
			{@render children()}
		</section>
	</div>
</div>
