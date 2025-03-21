<script lang="ts">
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import { getUserOptions, listIncidentsOptions, listOncallShiftsOptions, listTeamsOptions } from "$lib/api";
	
	import UserProfile from "./UserProfile.svelte";
	import OncallStats from "./OncallStats.svelte";
	import IncidentParticipation from "./IncidentParticipation.svelte";
	import TeamMembership from "./TeamMembership.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import { page } from "$app/state";
	import LoadingIndicator from "$src/components/loader/LoadingIndicator.svelte";
	
	const userId = $derived(page.params.id);
	const userQuery = createQuery(() => getUserOptions({path: {id: userId}}));
	const user = $derived(userQuery.data?.data);
	
	const shiftsQuery = createQuery(() => listOncallShiftsOptions({query: {userId}}))
	const oncallShifts = $derived(shiftsQuery.data?.data);

	const incidentsQuery = createQuery(() => listIncidentsOptions({query: {}}));
	const incidents = $derived(incidentsQuery.data?.data);

	const teamsQuery = createQuery(() => listTeamsOptions({query: {}}));
	const teams = $derived(teamsQuery.data?.data);

	const loading = $derived(shiftsQuery.isLoading || incidentsQuery.isLoading || teamsQuery.isLoading);
	const error = $derived(shiftsQuery.error || incidentsQuery.error || teamsQuery.error);

	const userName = $derived(user?.attributes.name ?? "");
	setPageBreadcrumbs(() => [
		{ label: "Users", href: "/users" },
		{ label: userName, href: `/users/${userId}`},
	]);
</script>

<div class="p-2">
	{#if loading}
		<LoadingIndicator />
	{:else if error}
		<span>error</span>
	{:else if user && oncallShifts && incidents && teams}
		<UserProfile {user} />
		
		<OncallStats {user} shifts={oncallShifts} />
		<IncidentParticipation {user} {incidents} />
		
		<TeamMembership {user} {teams} />
	{/if}
</div>
