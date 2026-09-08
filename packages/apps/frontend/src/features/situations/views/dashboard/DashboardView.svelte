<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import type { ErrorModel } from "$lib/api";
	import * as Button from "$components/ui/button";
	import * as Input from "$components/ui/input";
	import * as Select from "$components/ui/select";
	import * as Tabs from "$components/ui/tabs";
	import * as Card from "$components/ui/card";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import RiRefreshLine from "remixicon-svelte/icons/refresh-line";
	import RiRadarLine from "remixicon-svelte/icons/radar-line";
	import {
		initSituationsDashboardController,
		openedRanges,
		situationViews,
		type SituationView,
	} from "./controller.svelte";
	import SituationRow from "./SituationRow.svelte";

	registerPageDescriptor(() => ({ title: "Situations" }));

	const controller = initSituationsDashboardController();

	const viewLabels: Record<SituationView, string> = {
		"needs-attention": "Needs attention",
		active: "Active",
		history: "History",
	};
</script>

<section class="flex min-h-0 flex-1 flex-col gap-3 p-4">
	<header class="flex flex-wrap items-center justify-between gap-3">
		<Tabs.Root
			value={controller.view}
			onValueChange={(value) => controller.setView(value as SituationView)}
		>
			<Tabs.List>
				{#each situationViews as view (view)}
					<Tabs.Trigger value={view}>{viewLabels[view]}</Tabs.Trigger>
				{/each}
			</Tabs.List>
		</Tabs.Root>

		<div class="flex items-center gap-2">
			<Input.Root
				type="search"
				class="w-64"
				placeholder="Search situations…"
				value={controller.search}
				oninput={(event) => controller.setSearch(event.currentTarget.value)}
			/>
			<Select.Root
				type="single"
				value={controller.openedRange}
				onValueChange={(value) => value && controller.setOpenedRange(value as keyof typeof openedRanges)}
			>
				<Select.Trigger class="w-44">
					{openedRanges[controller.openedRange].label}
				</Select.Trigger>
				<Select.Content>
					{#each Object.entries(openedRanges) as [value, range] (value)}
						<Select.Item value={value}>{range.label}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
			<Button.Root variant="outline" size="icon-sm" aria-label="Refresh situations" onclick={controller.refresh}>
				<RiRefreshLine />
			</Button.Root>
		</div>
	</header>

	<LoadingQueryWrapper query={controller.query}>
		{#snippet error(error: ErrorModel)}
			<div class="flex flex-col items-start gap-2">
				<InlineAlert {error} />
				<Button.Root variant="outline" size="sm" onclick={controller.refresh}>Try again</Button.Root>
			</div>
		{/snippet}
		{#snippet view(situations)}
			<div class="flex min-h-0 flex-1 flex-col gap-2 overflow-y-auto">
				{#if controller.hasFilters && situations.length === 0}
					<Card.Root>
						<Card.Content class="flex flex-col items-center gap-2 py-10 text-center">
							<Card.Title class="text-base">No matching situations</Card.Title>
							<Card.Description>
								No situations match the current search and time range.
							</Card.Description>
							<Button.Root variant="outline" size="sm" onclick={() => controller.clearFilters()}>
								Clear filters
							</Button.Root>
						</Card.Content>
					</Card.Root>
				{:else if situations.length === 0 && controller.view === "history"}
					<Card.Root>
						<Card.Content class="flex flex-col items-center gap-2 py-10 text-center">
							<Card.Title class="text-base">No closed situations yet</Card.Title>
							<Card.Description>
								Situations appear here once they have been stabilized or dismissed.
							</Card.Description>
						</Card.Content>
					</Card.Root>
				{:else if situations.length === 0}
					<Card.Root class="mx-auto w-full max-w-md">
						<Card.Content class="flex flex-col items-center gap-3 py-10 text-center">
							<RiRadarLine class="text-muted-foreground size-10" />
							<Card.Title class="text-lg">
								{controller.view === "needs-attention"
									? "Nothing needs your attention"
									: "No active situations"}
							</Card.Title>
							<Card.Description>
								Situations are detected automatically from your alert signals. Connect a
								provider or explore how your systems relate to get started.
							</Card.Description>
							<div class="flex gap-2">
								<Button.Root variant="outline" href="/settings/integrations">
									Connect a provider
								</Button.Root>
								<Button.Root variant="outline" href="/map">Open the System Map</Button.Root>
							</div>
						</Card.Content>
					</Card.Root>
				{:else}
					{#each controller.situations as situation (situation.id)}
						<SituationRow {situation} showCloseReason={controller.view === "history"} />
					{/each}
				{/if}
			</div>
		{/snippet}
	</LoadingQueryWrapper>
</section>
