<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import { initSystemMapViewController } from "./controller.svelte";
	import SystemMapControls from "./SystemMapControls.svelte";
	import SystemMapList from "./SystemMapList.svelte";
	import SystemMapDetails from "./SystemMapDetails.svelte";

	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { Panel } from "@xyflow/svelte";
	import { SystemDiagram } from "$components/system-diagram";
	import { Badge } from "$components/ui/badge";
	import { Spinner } from "$components/ui/spinner";
	import RiRestartLine from "remixicon-svelte/icons/restart-line";

	const view = initSystemMapViewController();
	registerPageDescriptor(() => ({ title: "System Map" }));
</script>

<section
	class="flex min-h-0 flex-1 flex-col overflow-hidden"
	{@attach view.setFocusFallback}
	tabindex="-1"
	aria-label="System map"
>
	<SystemMapControls />

	{#if view.error && view.hasGraph}
		<Alert.Root variant="destructive">
			<Alert.Title>Map refresh failed</Alert.Title>
			<Alert.Description>{view.error.detail}</Alert.Description>
			<Alert.Action>
				<Button variant="outline" size="sm" onclick={view.retry}>Retry</Button>
			</Alert.Action>
		</Alert.Root>
	{/if}

	{#if view.selectionLoading}
		<p role="status" class="px-4 py-2 text-sm text-muted-foreground">Loading selected subject…</p>
	{:else if view.selectionError}
		<Alert.Root variant="destructive">
			<Alert.Title>Selected subject unavailable</Alert.Title>
			<Alert.Description>{view.selectionError.detail || view.selectionError.title}</Alert.Description>
			<Alert.Action>
				<Button variant="outline" size="sm" onclick={view.retrySelection}>Retry</Button>
			</Alert.Action>
		</Alert.Root>
	{/if}

	{#if view.focusMissing && view.hasGraph}
		<Alert.Root>
			<Alert.Title>Focused subject unavailable</Alert.Title>
			<Alert.Description>
				The requested subject was not returned. Previously loaded relationships remain available.
			</Alert.Description>
			<Alert.Action>
				<Button variant="outline" size="sm" onclick={view.retry}>Retry</Button>
			</Alert.Action>
		</Alert.Root>
	{/if}

	<div class="bg-muted/20 relative min-h-64 flex-1">
		{#if view.error && !view.hasGraph}
			<div class="absolute inset-0 grid place-items-center">
				<Alert.Root variant="destructive" class="w-fit max-w-md">
					<Alert.Title>Could not load the system map</Alert.Title>
					<Alert.Description>{view.error.detail}</Alert.Description>
					<Alert.Action>
						<Button variant="outline" size="sm" onclick={view.retry}>Retry</Button>
					</Alert.Action>
				</Alert.Root>
			</div>
		{:else if view.focusMissing && !view.hasGraph}
			<div class="absolute inset-0 grid place-items-center">
				<Alert.Root class="w-fit max-w-md">
					<Alert.Title>That subject is not on the map</Alert.Title>
					<Alert.Description>
						The linked subject could not be found or you are not authorized to see it. Search for
						a subject to explore its neighborhood.
					</Alert.Description>
					<Alert.Action>
						<Button variant="outline" size="sm" onclick={() => view.reset()}>
							Start from the default view
						</Button>
					</Alert.Action>
				</Alert.Root>
			</div>
		{:else if !view.hasGraph && !view.loading}
			<div class="absolute inset-0 grid place-items-center">
				<Alert.Root class="w-fit max-w-md">
					<Alert.Title>No mapped systems yet</Alert.Title>
					<Alert.Description>
						Once providers and mappings are connected, systems and their relationships appear
						here. Search for a subject to focus the map on it.
					</Alert.Description>
					<Alert.Action>
						<Button variant="outline" size="sm" href="/settings/integrations">
							Review integrations
						</Button>
					</Alert.Action>
				</Alert.Root>
			</div>
		{:else if view.displayMode === "list" && view.hasGraph}
			<SystemMapList />
		{:else}
			<SystemDiagram controller={view.diagram}>
				<Panel position="bottom-left">
					<div class="flex flex-col gap-2">
						{#if view.loading}
							<div
								class="bg-background border-border flex items-center gap-2 border px-3 py-2 text-sm shadow-sm"
							>
								<Spinner />
								Loading neighborhood…
							</div>
						{/if}
						{#if view.truncated}
							<div
								class="bg-background border-border max-w-xs border px-3 py-2 text-sm shadow-sm"
							>
								<Badge variant="outline" class="mb-1">Partial coverage</Badge>
								<p class="text-muted-foreground text-xs">
									This neighborhood is larger than the map loads at once. Expand specific
									subjects to load more of it.
								</p>
							</div>
						{/if}
					</div>
				</Panel>
				<Panel position="bottom-right">
					<div class="flex gap-2">
						<Button variant="outline" size="sm" onclick={() => view.recenter()}>Recenter</Button>
						<Button variant="outline" size="sm" onclick={() => view.reset()}>
							<RiRestartLine data-icon="inline-start" />
							Reset
						</Button>
					</div>
				</Panel>
			</SystemDiagram>
		{/if}
		<SystemMapDetails />
	</div>
</section>
