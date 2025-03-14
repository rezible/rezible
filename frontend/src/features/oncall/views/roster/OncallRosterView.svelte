<script lang="ts">
	import { type User, type OncallRoster } from "$lib/api";
	import { mdiChevronRight, mdiCalendarClock, mdiAlertCircle, mdiCheckCircle, mdiClockOutline } from "@mdi/js";
	import { Header, Icon, Button, Badge } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import PageActions from "./PageActions.svelte";
	import RosterStats from "./RosterStats.svelte";

	type Props = { roster: OncallRoster };
	const { roster }: Props = $props();

	appShell.setPageActions(PageActions, false);

	let users = $state<User[]>([
		{ id: "u1", attributes: { name: "Jane Doe", email: "jane@example.com" } },
		{ id: "u2", attributes: { name: "John Smith", email: "john@example.com" } },
		{ id: "u3", attributes: { name: "Alex Johnson", email: "alex@example.com" } },
		{ id: "u4", attributes: { name: "Sarah Williams", email: "sarah@example.com" } }
	]);
	let currentShifts = $state([
		{ id: "1", user: "Jane Doe", startTime: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000), endTime: new Date(Date.now() - 6 * 24 * 60 * 60 * 1000) },
	]);
	
	// Mock data for demonstration - replace with actual API calls
	let recentShifts = $state([
		{ id: "1", user: "Jane Doe", startTime: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000), endTime: new Date(Date.now() - 6 * 24 * 60 * 60 * 1000) },
		{ id: "2", user: "John Smith", startTime: new Date(Date.now() - 6 * 24 * 60 * 60 * 1000), endTime: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000) },
		{ id: "3", user: "Alex Johnson", startTime: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000), endTime: new Date() }
	]);
	
	let upcomingShifts = $state([
		{ id: "4", user: "Sarah Williams", startTime: new Date(), endTime: new Date(Date.now() + 1 * 24 * 60 * 60 * 1000) },
		{ id: "5", user: "Mike Brown", startTime: new Date(Date.now() + 1 * 24 * 60 * 60 * 1000), endTime: new Date(Date.now() + 2 * 24 * 60 * 60 * 1000) }
	]);
	
	let services = $state([
		{ id: "s1", name: "API Gateway", description: "", status: "healthy" },
		{ id: "s2", name: "Database Cluster", description: "", status: "healthy" },
		{ id: "s3", name: "Authentication Service", description: "", status: "warning" }
	]);
	
	function formatDate(date: Date): string {
		return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}
	
	function getStatusColor(status: string): string {
		switch (status) {
			case "healthy": return "text-green-500";
			case "warning": return "text-yellow-500";
			case "critical": return "text-red-500";
			default: return "text-gray-500";
		}
	}
</script>

<div class="grid grid-cols-4 gap-4 h-full max-h-full min-h-0 overflow-hidden">
	<div class="grid grid-rows-2 gap-4 h-full min-h-0">
		{@render rosterDetails()}
	</div>

	<div class="flex flex-col gap-4 h-full min-h-0">
		{@render rosterShifts()}
	</div>

	<div class="col-span-2 flex flex-col gap-4 h-full overflow-y-auto min-h-0 border rounded-lg p-4">
		<RosterStats {roster} />
	</div>
</div>

{#snippet rosterDetails()}
	<div class="border rounded-lg p-4 flex flex-col gap-2">
		<div class="flex items-center justify-between">
			<Header title="Members" />
		</div>

		<div class="flex flex-col gap-2 overflow-y-auto flex-1">
			{#each users as usr}
				<a href="/users/{usr.id}" class="block">
					<div class="flex items-center gap-4 bg-surface-100 hover:bg-accent-800/40 p-3 rounded-lg">
						<Avatar kind="user" size={32} id={usr.id} />
						<div class="flex flex-col">
							<span class="font-medium">{usr.attributes.name}</span>
							<span class="text-sm text-surface-600">{usr.attributes.email}</span>
						</div>
						<div class="flex-1 grid justify-items-end">
							<Icon data={mdiChevronRight} />
						</div>
					</div>
				</a>
			{:else}
				<div class="text-surface-600 italic p-2">No users assigned to this roster</div>
			{/each}
		</div>
	</div>

	<div class="border rounded-lg p-4 flex flex-col gap-2">
		<Header title="Services" />

		<div class="flex flex-col gap-2">
			{#each services as service}
				<a href="/services/{service.id}" class="block">
					<div class="flex items-center gap-4 bg-surface-100 hover:bg-accent-800/40 p-3 rounded-lg">
						<div class="flex flex-col flex-1">
							<span class="font-medium">{service.name}</span>
							<span class="text-sm text-surface-600">{service.description}</span>
						</div>
						<div class="flex-1 flex gap-2 justify-end">
							<Icon data={service.status === 'healthy' ? mdiCheckCircle : mdiAlertCircle} classes={{root: getStatusColor(service.status)}} />
							<Icon data={mdiChevronRight} />
						</div>
					</div>
				</a>
			{:else}
				<div class="text-surface-600 italic p-2">No services assigned to this roster</div>
			{/each}
		</div>
	</div>
{/snippet}

{#snippet rosterShifts()}
	<div class="border rounded-lg p-4 flex flex-col gap-4 flex-1 overflow-y-auto">
		{#if currentShifts.length > 0}
			<Header title="Active Shifts" classes={{root: "gap-2 text-lg font-medium"}}>
				<div slot="avatar">
					<Icon data={mdiClockOutline} class="text-success-800" />
				</div>
			</Header>
				
			<div class="flex flex-col gap-2">
				{#each currentShifts as shift}
					<a href="/oncall/shifts/{shift.id}" class="block">
						<div class="flex items-center gap-4 bg-success-900/40 hover:bg-success-800/50 p-3 rounded-lg">
							<Avatar kind="user" size={40} id={shift.user} />
							<div class="flex flex-col">
								<span class="text-lg font-medium">{shift.user}</span>
								<span class="text-sm text-surface-600"></span>
							</div>
							<div class="flex-1 grid justify-items-end">
								<Icon data={mdiChevronRight} />
							</div>
						</div>
					</a>
				{/each}
			</div>
		{:else}
			<Header title="No Active Shifts" classes={{root: "gap-2 text-lg font-medium text-yellow-500"}}>
				<div slot="avatar">
					<Icon data={mdiClockOutline} class="text-yellow-500" />
				</div>
			</Header>
		{/if}
		
		<Header title="Recent Shifts" classes={{root: "gap-2 text-lg font-medium"}}>
			<div slot="avatar">
				<Icon data={mdiCalendarClock} class="text-accent-500" />
			</div>
		</Header>

		<div class="flex flex-col gap-2 min-h-32 overflow-y-auto">
			{#each recentShifts as shift}
				<a href="/oncall/shifts/{shift.id}" class="block">
					<div class="flex items-center gap-4 bg-surface-100 hover:bg-accent-800/50 p-3 rounded-lg justify-between">
						<div class="flex flex-col flex-1">
							<span class="font-medium">{shift.user}</span>
							<div class="text-sm text-surface-600">
								{formatDate(shift.startTime)} - {formatDate(shift.endTime)}
							</div>
						</div>
						<div class="justify-items-end">
							<Icon data={mdiChevronRight} />
						</div>
					</div>
				</a>
			{/each}
		</div>

		<Header title="Upcoming Shifts" classes={{root: "gap-2 text-lg font-medium"}}>
			<div slot="avatar">
				<Icon data={mdiCalendarClock} class="text-accent-500" />
			</div>
		</Header>

		<div class="flex flex-col gap-2 min-h-32 overflow-y-auto">
			{#each upcomingShifts as shift}
				<div class="flex items-center gap-4 bg-surface-100 p-3 rounded-lg">
					<div class="flex flex-col flex-1">
						<span class="font-medium">{shift.user}</span>
						<div class="text-sm text-surface-600">
							{formatDate(shift.startTime)} - {formatDate(shift.endTime)}
						</div>
					</div>
				</div>
			{/each}
		</div>
	</div>
{/snippet}
