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
	import type { Component } from "svelte";

	type Props = { 
		tabs: Tab[];
		pathBase: string;
		infoBar?: Component;
	};
	const props: Props = $props();

	const activePath = $derived(page.url.pathname.replaceAll(props.pathBase, "").replace("/", ""));
	const activeIdx = $derived(Math.max(props.tabs.findIndex(t => activePath === t.path), 0));
	const activeTab = $derived(props.tabs[activeIdx]);
</script>

<div class="flex flex-col h-full max-h-full min-h-0 overflow-hidden">
	<div class="w-full flex h-14 z-[1] justify-between">
		<div class="flex gap-2 self-end">
			{#each props.tabs as tab, i}
				{@const active = i === activeIdx}
				<a href="{props.pathBase}/{tab.path}" 
					class={cls(
						"inline-flex self-end h-14 p-4 py-3 text-lg border border-b-0 rounded-t-lg relative", 
						active && "bg-surface-100 text-secondary",
					)}>
					{tab.label}
					<div class="absolute bottom-0 -mb-px left-0 w-full border-b border-surface-100" class:hidden={!active}></div>
				</a>
			{/each}
		</div>

		{#if props.infoBar}
			<props.infoBar></props.infoBar>
		{/if}
	</div>

	<div class="flex-1 min-h-0 max-h-full overflow-y-auto border rounded-b-lg rounded-tr-lg p-2 bg-surface-100">
		<activeTab.component></activeTab.component>
	</div>
</div>
