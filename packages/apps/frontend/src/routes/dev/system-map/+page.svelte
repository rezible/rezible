<script lang="ts">
	import { dev } from "$app/environment";
	import type { Component } from "svelte";

	let FixtureView = $state<Component>();

	if (dev) {
		void import("$features/dev/views/system-map-fixture").then(({ default: component }) => {
			FixtureView = component;
		});
	}
</script>

{#if FixtureView}
	<FixtureView />
{:else}
	<div class="grid min-h-0 flex-1 place-items-center">
		<p class="text-muted-foreground text-sm" role="status">Loading system map fixture…</p>
	</div>
{/if}
