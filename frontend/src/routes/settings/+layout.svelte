<script lang="ts">
	import { page } from "$app/state";
	import { Icon } from "svelte-ux";
	import { cls } from '@layerstack/tailwind';
	import { mdiAccount, mdiAccountGroup, mdiCog, mdiFire, mdiLayers, mdiPuzzle } from "@mdi/js";
	import SplitPage from "$src/components/split-page/SplitPage.svelte";

	const { children } = $props();

	const pages = [
		{ label: "General", path: "", routeId: "/(general)", icon: mdiCog },
		{ label: "Oncall", path: "/oncall", icon: mdiFire },
		{ label: "Incidents", path: "/incidents", icon: mdiFire },
		{ label: "Integrations", path: "/integrations", icon: mdiPuzzle },
		{ label: "Users", path: "/users", icon: mdiAccount },
		{ label: "Teams", path: "/teams", icon: mdiAccountGroup },
	];
	const basePath = "/settings";

	const activePath = $derived(page.route.id?.replace("/settings", "") ?? "");
	const activePageIdx = $derived(pages.findIndex((p) => (p.routeId ?? p.path) === activePath));
</script>

<SplitPage>
	{#snippet nav()}
		<ul class="flex flex-col space-y-0 -mb-px w-full h-fit w-54 border rounded">
			{#each pages as p, i}
				{@const active = i === activePageIdx}
				<li class="group flex w-full" role="presentation">
					<a
						href="{basePath}{p.path}"
						class={cls(
							"w-full py-3 px-6 gap-2 text-lg text-center border-r-2",
							active
								? "text-primary-content border-primary bg-surface-100 active"
								: "text-surface-content border-transparent hover:text-primary-content hover:border-primary/50"
						)}
					>
						<span class="flex items-center gap-2">
							{#if p.icon}
								<!-- <Icon data={p.icon} size={24} /> -->
							{/if}
							{p.label}
						</span>
					</a>
				</li>
			{/each}
		</ul>
	{/snippet}

	{@render children()}
</SplitPage>

