<script lang="ts" module>
	export type TabPage = {
		label: string;
		path: string;
		routeId?: string;
		icon?: string;
	};

	export type PageTabsProps = {
		pages: TabPage[];
		baseRouteId: string;
		basePath: string;
		end?: boolean;
	};
</script>

<script lang="ts">
	import { page } from "$app/state";
	import { cls } from '@layerstack/tailwind';
	import { Icon } from "svelte-ux";

	const { pages, baseRouteId, basePath, end = false }: PageTabsProps = $props();

	const activePath = $derived(page.route.id?.replace(baseRouteId, "") ?? "");
	const activePageIdx = $derived(pages.findIndex((p) => (p.routeId ?? p.path) === activePath));
</script>

<ul class="flex space-y-0 -mb-px w-fit" class:self-end={end}>
	{#each pages as p, i}
		{@const active = i === activePageIdx}
		<li class="group flex" role="presentation">
			<a
				href="{basePath}{p.path}"
				class={cls(
					"inline-block pt-2 pb-1 px-6 gap-2 text-xl text-center border-b-2",
					active
						? "text-primary-content border-primary bg-surface-100 active"
						: "text-surface-content border-transparent hover:text-primary-content hover:border-primary/50"
				)}
			>
				<span class="flex items-stretch gap-2">
					{#if p.icon}
						<Icon data={p.icon} size={24} />
					{/if}
					{p.label}
				</span>
			</a>
		</li>
	{/each}
</ul>
