<script lang="ts" module>
	export type Tab = {
		key: string;
		label: string;
		href: string;
	}

</script>

<script lang="ts">
	import { cls } from "@layerstack/tailwind";
	import type { Snippet, Component } from "svelte";

	type Props = { 
		tabs: Tab[];
		activeKey: string | undefined;
		content: Snippet;
		actionsBar?: Snippet;
	};
	const { tabs, activeKey, content, actionsBar }: Props = $props();

	const activeIdx = $derived(activeKey ? tabs.findIndex(t => t.key === activeKey) : 0);
</script>

<div class="flex flex-col h-full max-h-full min-h-0 overflow-hidden">
	<div class="w-full flex h-16 z-[1] justify-between">
		<div class="flex gap-2 self-end">
			{#each tabs as tab, i}
				{@const active = i === activeIdx}
				<a href={tab.href} 
					class={cls(
						"inline-flex self-end h-14 p-4 py-3 text-lg border border-b-0 rounded-t-lg relative", 
						active && "bg-surface-100 text-secondary",
					)}>
					{tab.label}
					<div class="absolute bottom-0 -mb-px left-0 w-full border-b border-surface-100" class:hidden={!active}></div>
				</a>
			{/each}
		</div>

		{#if actionsBar}
			{@render actionsBar()}
		{/if}
	</div>

	<div class="flex-1 min-h-0 max-h-full overflow-y-auto border rounded-b-lg rounded-tr-lg p-2 bg-surface-100">
		{@render content()}
	</div>
</div>
