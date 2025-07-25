<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getLocalTimeZone } from "@internationalized/date";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import { getUserOptions, listIncidentsOptions, listOncallRostersOptions, listOncallShiftsOptions, listTeamsOptions } from "$lib/api";

	import Avatar from "$components/avatar/Avatar.svelte";
	import OncallStats from "./OncallStats.svelte";
	import Header from "$components/header/Header.svelte";

	type Props = {
		userId: string;
	};
	const { userId }: Props = $props();
	
	const userQuery = createQuery(() => getUserOptions({path: {id: userId}}));
	const user = $derived(userQuery.data?.data);
	const userName = $derived(user?.attributes.name ?? "");

	appShell.setPageBreadcrumbs(() => [
		{ label: "Users", href: "/users" },
		{ label: userName, href: `/users/${userId}`, avatar: {kind: "user", id: userId}},
	]);

	const timeZone = $derived(getLocalTimeZone());
	const userLocalTime = $derived(new Date().toLocaleTimeString([], {timeZone, hour: "2-digit", minute: "2-digit"}));
	
	const shiftsQuery = createQuery(() => listOncallShiftsOptions({query: {userId}}))
	const oncallShifts = $derived(shiftsQuery.data?.data);

	const incidentsQuery = createQuery(() => listIncidentsOptions({query: {}}));
	const incidents = $derived(incidentsQuery.data?.data);

	const teamsQuery = createQuery(() => listTeamsOptions({query: {}}));
	const teams = $derived(teamsQuery.data?.data);

	const rostersQuery = createQuery(() => listOncallRostersOptions({query: {}}));
	const rosters = $derived(rostersQuery.data?.data);
</script>

<div class="grid grid-cols-3 gap-2 h-full">
	<div class="flex flex-col gap-2">
		<div class="border p-2">
			<Header title="Information" classes={{title: "text-xl"}} />
		
			<div class="">
				<div class="">
					<span class="">🌐</span>
					<span>{timeZone}</span>
					<span class="text-surface-content">({userLocalTime})</span>
				</div>
				<div class="">
					<span class="">✉️</span>
					<a href="mailto:{user?.attributes.email}">{user?.attributes.email}</a>
				</div>
			</div>
		</div>

		<div class="flex-1 grid grid-cols-2 gap-2">
			<div class="flex flex-col p-2 border">
				<Header title="Teams" classes={{title: "text-xl"}} />
			
				<div class="flex flex-col gap-2">
					{#if !teams}
						<span>loading</span>
					{:else}
						{#each teams as team}
							{@const attr = team.attributes}
							<a href="/teams/{team.id}" class="rounded-lg border bg-neutral p-2">
								<div class="flex items-center gap-2">
									<Avatar kind="team" id={team.id} />
									<span class="text-lg">{attr.name}</span>
								</div>
							</a>
						{:else}
							<p>Not a member of any teams</p>
						{/each}
					{/if}
				</div>
			</div>

			<div class="flex flex-col p-2 border">
				<Header title="Rosters" classes={{title: "text-xl"}} />
			
				<div class="flex flex-col gap-2">
					{#if !rosters}
						<span>loading</span>
					{:else}
						{#each rosters as roster}
							{@const attr = roster.attributes}
							<a href="/rosters/{roster.id}" class="rounded-lg border bg-neutral p-2">
								<div class="flex items-center gap-2">
									<Avatar kind="roster" id={roster.id} />
									<span class="text-lg">{attr.name}</span>
								</div>
							</a>
						{:else}
							<p>Not a member of any oncall roster</p>
						{/each}
					{/if}
				</div>
			</div>
		</div>
	</div>

	<div class="col-span-2 border">
		{#if oncallShifts}
			<OncallStats shifts={oncallShifts} />
		{/if}
	</div>
</div>
