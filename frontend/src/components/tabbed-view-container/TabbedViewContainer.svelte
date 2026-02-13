<script lang="ts" module>
	export type Tab<View> = {
		label: string;
		view: View;
		component: Component;
	}

</script>

<script lang="ts" generics="TabView">
	import { page } from "$app/state";
	import { cn } from "$lib/utils";
	import type { Component, Snippet } from "svelte";

	type Props = { 
		tabs: Tab<TabView>[];
		path: string;
		infoBar?: Component;
		tabSidebar?: Snippet;
	};
	const props: Props = $props();

	const activeViewPath = $derived(page.url.pathname.replaceAll(props.path, "").replace("/", ""));
	const activeView = $derived(activeViewPath === "" ? undefined : activeViewPath);
	const activeIdx = $derived(Math.max(props.tabs.findIndex(t => (activeView === t.view)), 0));
	const activeTab = $derived(props.tabs[activeIdx]);
</script>

<div class="flex-1 flex flex-col h-full max-h-full min-h-0 overflow-auto">
	<div class="w-full flex h-12 z-[1] justify-between">
		<div class="flex gap-1 self-end">
			{#each props.tabs as tab, i}
				{@const active = i === activeIdx}
				<a href="{props.path}/{tab.view}" 
					class={cn(
						"inline-flex self-end h-12 p-4 py-3 text-lg border border-surface-100 border-b-0 rounded-t-lg relative", 
						active && "bg-surface-200 text-secondary",
					)}>
					<span class="leading-none self-center">
						{tab.label}
					</span>
					<div class="absolute bottom-0 left-0 -mb-px w-full border-b border-surface-200" class:hidden={!active}></div>
				</a>
			{/each}
		</div>

		{#if props.infoBar}
			<div class="flex gap-4 h-12 max-h-14 overflow-y-hidden justify-between pb-1">
				<props.infoBar></props.infoBar>
			</div>
		{/if}
	</div>

	<div class="flex-1 flex border border-surface-100 rounded-b-lg rounded-tr-lg bg-surface-200 overflow-y-auto">
		<div class="flex-1 min-h-0 max-h-full overflow-y-auto p-2">
			<activeTab.component></activeTab.component>
		</div>
		{@render props.tabSidebar?.()}
	</div>
</div>
