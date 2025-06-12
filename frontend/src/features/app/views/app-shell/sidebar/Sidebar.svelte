<script lang="ts">
	import { page } from "$app/state";
	import {
		mdiAccountGroup,
		mdiChartBox,
		mdiFire,
		mdiHome,
		mdiPhoneRing,
		mdiVideo,
	} from "@mdi/js";
	import { cls } from '@layerstack/tailwind';
	import Icon from "$components/icon/Icon.svelte";
	import { Collapse } from "svelte-ux";
	import { session } from "$lib/auth.svelte";
	import OmniSearch from "./omni-search/OmniSearch.svelte";
	import UserProfileMenu from "./UserProfileMenu.svelte";

	type SidebarItem = {
		path: string;
		route?: string;
		label: string;
		icon: string;
	};
	type SidebarNavItem = SidebarItem | SidebarItem & { children: SidebarItem[] };
	const routes: SidebarNavItem[] = [
		{ path: "/", route: "/(index)", label: "Home", icon: mdiHome },
		{ path: "/incidents", label: "Incidents", icon: mdiFire },
		{ path: "/oncall", label: "Oncall", icon: mdiPhoneRing },
		{ path: "/reports", label: "Reports", icon: mdiChartBox },
		{ path: "/meetings", label: "Meetings", icon: mdiVideo },
		// { path: "/services", label: "Services", icon: mdiVectorPolyline },
		// { path: "/wiki", label: "Wiki", icon: mdiBookshelf },
		{ path: "/teams", label: "Teams", icon: mdiAccountGroup },
	];

	const currentPath = $derived(page.route.id);
	const expandingHover = false;

    const preloadHome = $derived(session.error ? "tap" : "hover");
</script>

{#snippet navItem(r: SidebarItem)}
	{@const active = currentPath?.startsWith(r.route ?? r.path)}
	<a
		href={r.path}
		class={cls(
			"inline-block px-4 py-3 flex items-center gap-2 text-center border-none-2 rounded-lg",
			active
				? "text-neutral-content bg-primary-900"
				: "border-transparent hover:text-primary-content hover:border-primary/50 hover:bg-primary-900/50"
		)}
	>
		<Icon data={r.icon} classes={{ root: expandingHover ? "group-hover:mr-3" : "mr-3" }} />
		{r.label}
	</a>
{/snippet}

<aside
	class={cls(
		"h-full group flex flex-col overflow-hidden bg-surface-300 pb-2 pl-2",
		expandingHover ? "w-fit hover:w-60" : "w-60"
	)}
>
	<div class="h-16 flex items-center px-4">
		<a href="/" class="text-2xl flex items-center" 
			data-sveltekit-preload-data={preloadHome} 
			data-sveltekit-preload-code={preloadHome}
		>
			<img src="/images/logo.svg" alt="logo" class="h-10 w-10 fill-neutral" />
			<span class="pl-3 {expandingHover ? 'hidden group-hover:inline' : ''}">Rezible</span>
		</a>
	</div>

	<div class="overflow-y-auto flex flex-col flex-1 min-h-0 justify-between">
		<div class="flex flex-col gap-2 overflow-y-auto overflow-x-hidden">
			{#each routes as r (r.label)}
				{#if "children" in r}
					<Collapse>
						<div
							slot="trigger"
							class="inline-block p-4 flex flex-1 items-center gap-2 text-center"
						>
							<Icon
								data={r.icon}
								classes={{
									root: expandingHover ? "group-hover:mr-3" : "mr-3",
								}}
							/>
							{r.label}
						</div>
						<div class="ml-4">
							{#each r.children as rc (rc.label)}
								{@render navItem(rc)}
							{/each}
						</div>
					</Collapse>
				{:else}
					{@render navItem(r)}
				{/if}
			{/each}
		</div>
	</div>

	<div class="my-2">
		<OmniSearch />
	</div>

	<div class="">
		<UserProfileMenu />
	</div>
</aside>
