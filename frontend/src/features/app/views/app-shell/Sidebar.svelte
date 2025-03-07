<script lang="ts">
	import { page } from "$app/state";
	import {
		mdiAccountGroup,
		mdiBookshelf,
		mdiChartBox,
		mdiCogBox,
		mdiFire,
		mdiHome,
		mdiPhoneRing,
		mdiVectorPolyline,
		mdiVideo,
	} from "@mdi/js";
	import { cls } from '@layerstack/tailwind';
	import { Icon, Collapse } from "svelte-ux";
	import HeaderLogo from "./HeaderLogo.svelte";

	type SidebarItem = {
		path: string;
		route?: string;
		label: string;
		icon: string;
	};
	type SidebarNavItem =
		| SidebarItem
		| {
				label: string;
				icon: string;
				children: SidebarItem[];
		  };
	const routes: SidebarNavItem[] = [
		{ path: "/", route: "/(index)", label: "Home", icon: mdiHome },
		{ path: "/incidents", label: "Incidents", icon: mdiFire },
		{ path: "/oncall", label: "Oncall", icon: mdiPhoneRing },
		{ path: "/reports", label: "Reports", icon: mdiChartBox },
		{ path: "/services", label: "Services", icon: mdiVectorPolyline },
		{ path: "/wiki", label: "Wiki", icon: mdiBookshelf },
		{ path: "/teams", label: "Teams", icon: mdiAccountGroup },
		{ path: "/meetings", label: "Meetings", icon: mdiVideo },
	];

	const currentPath = $derived(page.route.id);
	const expandingHover = false;
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
		"h-full group flex flex-col overflow-hidden bg-surface-200 pb-2",
		expandingHover ? "w-fit hover:w-60" : "w-60"
	)}
>
	<HeaderLogo {expandingHover} />

	<div class="overflow-y-auto flex flex-col flex-1 min-h-0 justify-between pl-2">
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

		<div class="">
			{#if currentPath !== "/oncall"}
				<!--a href="/oncall">
					<div
						class="border rounded-lg border-success-700 bg-success-600/20 {expandingHover
							? 'group-hover:p-2'
							: 'p-2'} m-2 flex justify-center"
					>
						<Header
							title="Currently Oncall"
							subheading="search-team"
							class={expandingHover ? "hidden group-hover:flex" : "flex"}
						>
							<svelte:fragment slot="actions">
								<Icon data={mdiCircleMedium} classes={{ root: "text-success" }} />
							</svelte:fragment>
						</Header>
						<Icon
							class={expandingHover ? "inline group-hover:hidden" : "hidden"}
							data={mdiCircleMedium}
							classes={{ root: "text-success" }}
						/>
					</div>
				</a-->
			{/if}

			{@render navItem({
				label: "Settings",
				path: "/settings",
				icon: mdiCogBox,
			})}
		</div>
	</div>
</aside>
