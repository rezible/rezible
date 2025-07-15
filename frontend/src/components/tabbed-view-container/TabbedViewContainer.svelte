<script lang="ts" module>
	export type Tab = {
		label: string;
		path: string;
		component: Component;
	}

</script>

<script lang="ts">
	import { page } from "$app/state";

	import { cls } from "@layerstack/tailwind";
	import type { Component, Snippet } from "svelte";

	type Props = { 
		tabs: Tab[];
		pathBase: string;
		infoBar?: Component;
		tabSidebar?: Snippet;
	};
	const props: Props = $props();

	const activePath = $derived(page.url.pathname.replaceAll(props.pathBase, "").replace("/", ""));
	const activeIdx = $derived(Math.max(props.tabs.findIndex(t => activePath === t.path), 0));
	const activeTab = $derived(props.tabs[activeIdx]);
</script>

<div class="flex-1 flex flex-col h-full max-h-full min-h-0 overflow-hidden">
	<div class="w-full flex h-12 z-[1] justify-between">
		<div class="flex gap-1 self-end">
			{#each props.tabs as tab, i}
				{@const active = i === activeIdx}
				<a href="{props.pathBase}/{tab.path}" 
					class={cls(
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
			<props.infoBar></props.infoBar>
		{/if}
	</div>

	<div class="flex-1 flex border border-surface-100 rounded-b-lg rounded-tr-lg bg-surface-200">
		<div class="flex-1 min-h-0 max-h-full overflow-y-auto p-2">
			<activeTab.component></activeTab.component>
		</div>
		{@render props.tabSidebar?.()}
	</div>
</div>
