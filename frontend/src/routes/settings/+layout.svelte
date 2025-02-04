<script lang="ts">
	import {
		mdiAccount,
		mdiAccountGroup,
		mdiCog,
		mdiFire,
		mdiLayers,
		mdiPuzzle,
	} from "@mdi/js";
	import PageContainer, {
		type PageTabsProps,
	} from "$components/page-container/PageContainer.svelte";

	const { children } = $props();

	const tabs = $derived<PageTabsProps>({
		pages: [
			{ label: "General", path: "", routeId: "/(general)", icon: mdiCog },
			{ label: "Incidents", path: "/incidents", icon: mdiFire },
			{ label: "Users", path: "/users", icon: mdiAccount },
			{ label: "Teams", path: "/teams", icon: mdiAccountGroup },
			{ label: "Integrations", path: "/integrations", icon: mdiPuzzle },
		],
		basePath: `/settings`,
		baseRouteId: "/settings",
	});
</script>

<PageContainer breadcrumbs={[{ label: "Settings" }]} {tabs}>
	{@render children()}
</PageContainer>

<!--div class="flex flex-row w-full max-h-full">
	<div class="w-fit h-full max-h-full min-h-0 rounded border">
		<ul class="flex flex-col space-y-0 -mb-px w-fit">
			{#each pages as p, i}
				<li class="group" role="presentation">
					<a
						href={p.href}
						class={cls(
							"inline-block p-4 flex items-center gap-2 text-sm font-medium text-center border-e-2", 
							i === activePageIdx 
								? "text-primary-content bg-primary-900 border-primary active" 
								: "text-neutral-content border-transparent hover:text-primary-content hover:border-primary/50"
						)}
					>
						<Icon data={p.icon} />
						{p.label}
					</a>
				</li>
			{/each}
		</ul>
	</div>

	<div class="h-px bg-gray-200 dark:bg-gray-700"></div>

	<div class="px-2 pb-2 grow max-h-full overflow-y-auto overflow-x-hidden">
		{@render children()}
	</div>
</div-->
