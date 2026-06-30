<script lang="ts">
	import { fly } from "svelte/transition";
	import { mdiCircleMedium } from "@mdi/js";
	import Icon from "$components/common/icon/Icon.svelte";
	import ComponentSelector from "./add-component-drawer/ComponentSelector.svelte";
	import Header from "$src/components/layout/header/Header.svelte";

	import { initIncidentSidebarController } from "./controller.svelte";

	const controller = initIncidentSidebarController();
</script>

<div class="w-64 flex flex-col relative border p-2 ml-2">
	<!-- <div class="absolute w-full h-full bg-surface-300/20 z-1" class:hidden={!drawerOpen}></div> -->

	<Header>
		{#snippet title()}
			<span class="flex text-lg gap-1 items-center">
				Collaboration
				<Icon data={mdiCircleMedium} classes={{ root: "opacity-70", path: controller.ctxColor }} />
			</span>
		{/snippet}
		{#snippet subheading()}
			{#if controller.connectionError}
				<span class="text-danger-300">Connection Error: {controller.connectionError.message}</span>
			{/if}
		{/snippet}
	</Header>

	<div class="flex-1 min-h-0 overflow-x-hidden overflow-y-auto relative">
		
		
		{#if controller.drawerOpen}
			<div
				class="bg-surface-100 z-50 outline-none h-full w-full absolute transform top-0 left-2 border"
				in:fly|global={{ x: "100%", y: 0 }}
				out:fly={{ x: "100%", y: 0 }}
			>
				<ComponentSelector />
			</div>
		{/if}
	</div>
</div>
