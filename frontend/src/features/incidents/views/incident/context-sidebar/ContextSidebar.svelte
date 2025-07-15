<script lang="ts">
	import { Button } from "svelte-ux";
	import { fly } from 'svelte/transition';
	import { mdiChevronLeft, mdiChevronRight, mdiCircleMedium } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";
	import { useIncidentCollaboration } from "../collaboration.svelte";
	import { cls } from "@layerstack/tailwind";
	import { WebSocketStatus } from "@hocuspocus/provider";
	import ComponentSelector from "./add-component-drawer/ComponentSelector.svelte";
	import Header from "$components/header/Header.svelte";

	const collab = useIncidentCollaboration();

	const ctxColor = $derived.by(() => {
		if (collab.error) return "fill-danger";
		switch (collab.connectionStatus) {
			case WebSocketStatus.Connecting: return "fill-default";
			case WebSocketStatus.Connected: return "fill-success";
			case WebSocketStatus.Disconnected: return "fill-warning";
		}
	})

	let drawer = $state<undefined | "add-component">();

	const drawerOpen = $derived(!!drawer);
</script>

<div class="w-64 border-l flex flex-col relative">
	<!-- <div class="absolute w-full h-full bg-surface-300/20 z-1" class:hidden={!drawerOpen}></div> -->

	<Header classes={{root: "p-2"}}>
		{#snippet title()}
			<span class="text-xl flex gap-1 items-center">
				Context
				<Icon data={mdiCircleMedium} classes={{root: "opacity-70", path: ctxColor}} />
			</span>
		{/snippet}
		{#snippet subheading()}
			{#if collab.error}
				<span class="text-danger-300">Connection Error: {collab.error.message}</span>
			{/if}
		{/snippet}
	</Header>

	<div class="flex-1 min-h-0 overflow-x-hidden relative">
		{#if drawerOpen}
			<div
				class="bg-surface-100 z-50 outline-none h-full w-full absolute transform top-0 left-2 border"
				in:fly|global={{x: "100%", y: 0}}
				out:fly={{x: "100%", y: 0}}
			>
				<ComponentSelector />
			</div>
		{/if}
	</div>
</div>
