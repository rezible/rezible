<script lang="ts">
	import { page } from "$app/state";
	import { cls, Icon } from "svelte-ux";
	import { mdiAccount, mdiAccountGroup, mdiCog, mdiFire, mdiLayers, mdiPuzzle } from "@mdi/js";
	import PageContainer, { type PageTabsProps } from "$components/page-container/PageContainer.svelte";
	import SplitPage from "$src/components/split-page/SplitPage.svelte";

	const { children } = $props();

	const pages = [
		{ label: "General", path: "", routeId: "/(general)", icon: mdiCog },
		{ label: "Incidents", path: "/incidents", icon: mdiFire },
		{ label: "Users", path: "/users", icon: mdiAccount },
		{ label: "Teams", path: "/teams", icon: mdiAccountGroup },
		{ label: "Integrations", path: "/integrations", icon: mdiPuzzle },
	];
	const basePath = "/settings";

	const activePath = $derived(page.route.id?.replace("/settings", "") ?? "");
	const activePageIdx = $derived(pages.findIndex((p) => (p.routeId ?? p.path) === activePath));
</script>

{#snippet pagesNav()}
	<ul class="flex flex-col gap-2 space-y-0 -mb-px w-full h-fit">
		{#each pages as p, i}
			{@const active = i === activePageIdx}
			<li class="group flex w-full" role="presentation">
				<a
					href="{basePath}{p.path}"
					class={cls(
						"w-full p-1 px-6 gap-2 text-lg text-center border-r-2",
						active
							? "text-primary-content border-primary bg-surface-100 active"
							: "text-surface-content border-transparent hover:text-primary-content hover:border-primary/50"
					)}
				>
					<span class="flex items-center gap-2">
						{#if p.icon}
							<Icon data={p.icon} size={24} />
						{/if}
						{p.label}
					</span>
				</a>
			</li>
		{/each}
	</ul>
{/snippet}

<PageContainer breadcrumbs={[{ label: "Settings" }]}>
	<div class="flex w-full h-full gap-2">
		<div class="w-64 h-fit border rounded-lg py-2">
			{@render pagesNav()}
		</div>
		<div class="flex-1 block">
			{@render children()}
		</div>
	</div>
</PageContainer>
