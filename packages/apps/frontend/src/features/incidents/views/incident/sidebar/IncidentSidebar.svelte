<script lang="ts">
	import { fly } from "svelte/transition";
	import RiCircleFill from "remixicon-svelte/icons/circle-fill";
	import Header from "$src/components/layout/header/Header.svelte";

	import { initIncidentSidebarController } from "./controller.svelte";

	const controller = initIncidentSidebarController();
</script>

<div class="w-64 flex flex-col relative border p-2 ml-2">
	<!-- <div class="absolute w-full h-full bg-background/20 z-1" class:hidden={!drawerOpen}></div> -->

	<Header>
		{#snippet title()}
			<span class="flex text-lg gap-1 items-center">
				Collaboration
				<RiCircleFill class="size-4" aria-hidden="true" />
			</span>
		{/snippet}
		{#snippet subheading()}
			{#if controller.connectionError}
				<span class="text-destructive/70">Connection Error: {controller.connectionError.message}</span
				>
			{/if}
		{/snippet}
	</Header>

	<div class="flex-1 min-h-0 overflow-x-hidden overflow-y-auto relative">
		{#if controller.drawerOpen}
			<div
				class="bg-card z-50 outline-none h-full w-full absolute transform top-0 left-2 border"
				in:fly|global={{ x: "100%", y: 0 }}
				out:fly={{ x: "100%", y: 0 }}
			>
				<span>component selector</span>
			</div>
		{/if}
	</div>
</div>
