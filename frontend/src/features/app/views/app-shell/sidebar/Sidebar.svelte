<script lang="ts">
	import { page } from "$app/state";
	import {
	mdiAccountClock,
		mdiAccountGroup,
		mdiAlarm,
		mdiAlarmLight,
		mdiBookshelf,
		mdiChartBox,
		mdiDockLeft,
		mdiFire,
		mdiHome,
		mdiPhoneRing,
		mdiShape,
		mdiShield,
		mdiVectorPolyline,
		mdiVideo,
	} from "@mdi/js";
	import { cls } from '@layerstack/tailwind';
	import Icon from "$components/icon/Icon.svelte";
	import { Collapse } from "svelte-ux";
	import { session } from "$lib/auth.svelte";
	import OmniSearch from "./omni-search/OmniSearch.svelte";
	import UserProfileMenu from "./UserProfileMenu.svelte";

	type SidebarRoute = {
		path: string;
		route?: string;
		label: string;
		icon: string;
	};
	type SidebarGroup = {
		label: string;
	}
	type SidebarItem = SidebarRoute | SidebarGroup;
	const items: SidebarItem[] = [
		{ path: "/", route: "/(index)", label: "Home", icon: mdiHome },
		{ label: "Oncall" },
		{ path: "/incidents", label: "Incidents", icon: mdiFire },
		{ path: "/alerts", label: "Alerts", icon: mdiAlarmLight },
		{ path: "/rosters", label: "Rosters", icon: mdiShield },
		{ path: "/shifts", label: "Shifts", icon: mdiAccountClock },
		{ label: "System" },
		{ path: "/playbooks", label: "Playbooks", icon: mdiBookshelf },
		{ path: "/system", label: "Components", icon: mdiShape },
		{ label: "People" },
		{ path: "/teams", label: "Teams", icon: mdiAccountGroup },
		{ path: "/meetings", label: "Meetings", icon: mdiVideo },
	];

	const currentPath = $derived(page.route.id);
	let expanded = $state(true);

    const preloadHome = $derived(session.error ? "tap" : "hover");
</script>

{#snippet navRouteItem(r: SidebarRoute)}
	{@const active = currentPath?.startsWith(r.route ?? r.path)}
	<a
		href={r.path}
		class={cls(
			"inline-block px-4 py-2 flex items-center gap-2 text-center border-none-2 rounded-lg",
			active
				? "text-neutral-content bg-primary-900"
				: "border-transparent hover:text-primary-content hover:border-primary/50 hover:bg-primary-900/50"
		)}
	>
		<Icon data={r.icon} classes={{ root: "" }} />
		<span class={!expanded ? "hidden" : "pl-3"}>{r.label}</span>
	</a>
{/snippet}

<aside
	class={cls(
		"h-full group flex flex-col overflow-hidden bg-surface-300 pb-2 pl-2",
		expanded ? "w-60" : "w-fit"
	)}
>
	<div class="h-16 flex items-center justify-between px-4">
		<a href="/" class="text-2xl flex items-center" 
			data-sveltekit-preload-data={preloadHome} 
			data-sveltekit-preload-code={preloadHome}
		>
			<img src="/images/logo.svg" alt="logo" class="h-10 w-10 fill-neutral" />
			<span class={!expanded ? "hidden" : "pl-3"}>Rezible</span>
		</a>

		<!-- <Button icon={mdiDockLeft} iconOnly size="sm" classes={{root: "ml-2 text-surface-content/40"}} 
			on:click={() => {expanded = !expanded}} 
		/> -->
	</div>

	<div class="overflow-y-auto flex flex-col flex-1 min-h-0 justify-between">
		<div class="flex flex-col overflow-y-auto gap-1 overflow-x-hidden">
			{#each items as item}
				{#if "path" in item}
					{@render navRouteItem(item)}
				{:else}
					<span class="font-bold text-sm uppercase text-surface-content/40 pl-2 mt-3">{item.label}</span>
				{/if}
			{/each}
		</div>
	</div>

	{#if expanded}
		<div class="my-2">
			<OmniSearch />
		</div>

		<div class="">
			<UserProfileMenu />
		</div>
	{/if}
</aside>
