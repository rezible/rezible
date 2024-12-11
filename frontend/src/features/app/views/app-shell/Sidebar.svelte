<script lang="ts">
	import { page } from '$app/stores';
	import {
		mdiAccountGroup,
		mdiChartBox,
		mdiCircleMedium,
		mdiCogBox,
		mdiFire,
		mdiHome,
		mdiPhoneLog,
		mdiPhoneRing,
		mdiVideo
	} from '@mdi/js';
	import { cls, Header, Icon, Collapse } from 'svelte-ux';

	type SidebarItem = {
		path: string;
		route?: string;
		label: string;
		icon: string;
	}
	type SidebarNavItem = SidebarItem | {
		label: string;
		icon: string;
		children: SidebarItem[];
	}
	const routes: SidebarNavItem[] = [
		{ path: '/', route: "/(index)", label: 'Home', icon: mdiHome },
		{ path: '/incidents', label: 'Incidents', icon: mdiFire },
		{ path: '/oncall', label: 'Oncall', icon: mdiPhoneRing },
		{ path: '/meetings', label: 'Meetings', icon: mdiVideo },
		{ path: '/teams', label: 'Teams', icon: mdiAccountGroup },
		{ path: '/reports', label: 'Reports', icon: mdiChartBox }
	];

	const currentPath = $derived($page.route.id);
	const expandingHover = false;
</script>

{#snippet navItem(r: SidebarItem)}
	{@const active = currentPath?.startsWith(r.route ?? r.path)}
	<a href={r.path} 
		class={cls("inline-block px-4 py-3 flex items-center gap-2 text-center border-e-2", 
			active ? "text-primary-content bg-primary-900" 
				: "text-neutral-content border-transparent hover:text-primary-content hover:border-primary/50 hover:bg-primary-900/50"
		)}>
		<Icon data={r.icon} classes={{root: (expandingHover ? "group-hover:mr-3" : "mr-3")}} />
		{r.label}
	</a>
{/snippet}

<aside class="h-full {expandingHover ? "w-fit hover:w-60" : "w-60"} group border-e flex flex-col overflow-hidden bg-surface-200">
	<div class="h-16 flex items-center px-4">
		<a href="/" class="text-2xl flex items-center">
			<img src="/images/logo.svg" alt="logo" class="h-10 w-10 fill-neutral" />
			<span class="pl-3 {expandingHover ? "hidden group-hover:inline" : ""}">Rezible</span>
		</a>
	</div>
	<div class="overflow-y-auto flex flex-col flex-1 min-h-0 justify-between">
		<div class="flex flex-col gap-2 overflow-y-auto overflow-x-hidden">
			{#each routes as r (r.label)}
				{#if "children" in r}
					<Collapse>
						<div slot="trigger" class="inline-block p-4 flex flex-1 items-center gap-2 text-center">
							<Icon data={r.icon} classes={{root: (expandingHover ? "group-hover:mr-3" : "mr-3")}} />
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

			{@render navItem({label: "Settings", path: "/settings", icon: mdiCogBox})}
		</div>

		<div class="">
			{#if currentPath !== '/oncall'}
				<a href="/oncall">
					<div class="border rounded-lg border-success-700 bg-success-600/20 {expandingHover ? "group-hover:p-2" : "p-2"} m-2 flex justify-center">
						<Header title="Currently Oncall" subheading="search" class={expandingHover ? "hidden group-hover:flex" : "flex"}>
							<svelte:fragment slot="actions">
								<Icon data={mdiCircleMedium} classes={{ root: 'text-success' }} />
							</svelte:fragment>
						</Header>
						<Icon class={expandingHover ? "inline group-hover:hidden" : "hidden"} data={mdiCircleMedium} classes={{ root: 'text-success' }} />
					</div>
				</a>
			{/if}
		</div>
	</div>
</aside>