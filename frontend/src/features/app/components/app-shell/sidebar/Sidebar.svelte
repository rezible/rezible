<script lang="ts">
	import { page } from "$app/state";
	import {
		mdiAccountClock,
		mdiAccountGroup,
		mdiAlarmLight,
		mdiBookshelf,
		mdiFire,
		mdiHome,
		mdiShape,
		mdiShield,
		mdiTimelineText,
		mdiVideo,
	} from "@mdi/js";
	import { cn } from '$lib/utils';
	import Icon from "$components/icon/Icon.svelte";
	import UserProfileMenu from "./UserProfileMenu.svelte";
	import LogoHeader from "./LogoHeader.svelte";

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
		{ path: "/events", label: "Events", icon: mdiTimelineText },
		{ path: "/incidents", label: "Incidents", icon: mdiFire },
		{ path: "/rosters", label: "Rosters", icon: mdiShield },
		{ path: "/shifts", label: "Shifts", icon: mdiAccountClock },
		{ label: "System" },
		{ path: "/alerts", label: "Alerts", icon: mdiAlarmLight },
		{ path: "/playbooks", label: "Playbooks", icon: mdiBookshelf },
		{ path: "/system", label: "Components", icon: mdiShape },
		{ label: "People" },
		{ path: "/teams", label: "Teams", icon: mdiAccountGroup },
		{ path: "/meetings", label: "Meetings", icon: mdiVideo },
	];

	const currentPath = $derived(page.route.id);
	let expanded = $state(true);
</script>

{#snippet navRouteItem(r: SidebarRoute)}
	{@const active = currentPath?.startsWith(r.route ?? r.path)}
	<a
		href={r.path}
		class={cn(
			"inline-block px-4 py-2 flex items-center gap-2 border-none-2 rounded-lg",
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
	class={cn(
		"h-full group flex flex-col overflow-hidden bg-surface-300",
		expanded ? "w-60" : "w-fit"
	)}
>
	<LogoHeader showText={expanded} />

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
			<span>search</span>
		</div>

		<div class="">
			<UserProfileMenu />
		</div>
	{/if}
</aside>
