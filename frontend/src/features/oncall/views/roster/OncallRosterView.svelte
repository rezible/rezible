<script lang="ts">
	import { type User, type OncallRoster } from "$lib/api";
	import Avatar from "$src/components/avatar/Avatar.svelte";
	import { mdiChevronRight, mdiCalendarClock, mdiAccountGroup, mdiAlertCircle, mdiCheckCircle, mdiClockOutline } from "@mdi/js";
	import { Header, Icon, Button, Badge } from "svelte-ux";
	import { onMount } from "svelte";
	import { api } from "$lib/api";

	type Props = { roster: OncallRoster };
	const { roster }: Props = $props();

	const attr = $derived(roster.attributes);

	let users = $state<User[]>([]);
	let currentOncallers = $state<User[]>([]);
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	
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
		{ id: "s1", name: "API Gateway", status: "healthy" },
		{ id: "s2", name: "Database Cluster", status: "healthy" },
		{ id: "s3", name: "Authentication Service", status: "warning" }
	]);
	
	let metrics = $state({
		alertsLast24h: 3,
		averageResponseTime: "5m 12s",
		escalationsLast30d: 2,
		handoversCompleted: 12
	});

	onMount(async () => {
		try {
			// In a real implementation, fetch users assigned to this roster
			// const response = await api.listUsers({ rosterId: roster.id });
			// users = response.data;
			
			// For now, using mock data
			users = [
				{ id: "u1", attributes: { name: "Jane Doe", email: "jane@example.com" } },
				{ id: "u2", attributes: { name: "John Smith", email: "john@example.com" } },
				{ id: "u3", attributes: { name: "Alex Johnson", email: "alex@example.com" } },
				{ id: "u4", attributes: { name: "Sarah Williams", email: "sarah@example.com" } }
			] as User[];
			
			// Set current oncallers (would come from API)
			currentOncallers = [users[2]] as User[];
			
			isLoading = false;
		} catch (e) {
			error = "Failed to load roster data";
			isLoading = false;
		}
	});
	
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

{#if isLoading}
	<div class="flex justify-center items-center h-full">
		<span class="text-lg">Loading roster details...</span>
	</div>
{:else if error}
	<div class="flex justify-center items-center h-full">
		<span class="text-lg text-red-500">{error}</span>
	</div>
{:else}
	<div class="grid grid-cols-3 gap-4 h-full max-h-full min-h-0 overflow-hidden">
		<div class="flex flex-col gap-4 h-full min-h-0">
			{@render rosterDetails()}
		</div>

		<div class="col-span-2 flex flex-col gap-4 h-full min-h-0">
			<div class="border rounded-lg p-4 bg-surface-50">
				<Header title="Current Oncall Status" />
				
				<div class="mt-2">
					{#if currentOncallers.length > 0}
						<div class="flex flex-col gap-2">
							<div class="flex items-center gap-2">
								<Icon data={mdiClockOutline} class="text-accent-500" />
								<span class="font-semibold">Currently on call:</span>
							</div>
							
							{#each currentOncallers as user}
								<div class="flex items-center gap-4 bg-surface-100 p-3 rounded-lg">
									<Avatar kind="user" size={40} id={user.id} />
									<div class="flex flex-col">
										<span class="text-lg font-medium">{user.attributes.name}</span>
										<span class="text-sm text-surface-600">{user.attributes.email}</span>
									</div>
									<div class="ml-auto">
										<Badge color="accent">Active</Badge>
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<div class="text-yellow-500 flex items-center gap-2">
							<Icon data={mdiAlertCircle} />
							<span>No active oncall users assigned!</span>
						</div>
					{/if}
				</div>
			</div>
			
			<div class="border rounded-lg p-4 flex flex-col gap-2 flex-1">
				<Header title="Oncall Schedule" />
				
				<div class="flex gap-2 mb-2">
					<Button variant="outline" size="sm">Today</Button>
					<Button variant="outline" size="sm">This Week</Button>
					<Button variant="outline" size="sm">This Month</Button>
				</div>
				
				<div class="flex flex-col gap-4">
					<div>
						<h3 class="text-lg font-medium mb-2 flex items-center gap-2">
							<Icon data={mdiCalendarClock} class="text-accent-500" />
							<span>Current & Upcoming Shifts</span>
						</h3>
						<div class="flex flex-col gap-2">
							{#each upcomingShifts as shift}
								<div class="flex items-center gap-4 bg-surface-100 p-3 rounded-lg">
									<div class="flex flex-col flex-1">
										<span class="font-medium">{shift.user}</span>
										<div class="text-sm text-surface-600">
											{formatDate(shift.startTime)} - {formatDate(shift.endTime)}
										</div>
									</div>
									<Badge color="green">Active</Badge>
								</div>
							{/each}
						</div>
					</div>
					
					<div>
						<h3 class="text-lg font-medium mb-2 flex items-center gap-2">
							<Icon data={mdiCalendarClock} class="text-accent-500" />
							<span>Recent Shifts</span>
						</h3>
						<div class="flex flex-col gap-2">
							{#each recentShifts as shift}
								<div class="flex items-center gap-4 bg-surface-100 p-3 rounded-lg">
									<div class="flex flex-col flex-1">
										<span class="font-medium">{shift.user}</span>
										<div class="text-sm text-surface-600">
											{formatDate(shift.startTime)} - {formatDate(shift.endTime)}
										</div>
									</div>
									<Badge color="surface">Completed</Badge>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>
			
			<div class="border rounded-lg p-4">
				<Header title="Roster Stats" />
				
				<div class="grid grid-cols-2 gap-4 mt-2">
					<div class="bg-surface-100 p-3 rounded-lg">
						<div class="text-sm text-surface-600">Alerts (Last 24h)</div>
						<div class="text-2xl font-semibold">{metrics.alertsLast24h}</div>
					</div>
					<div class="bg-surface-100 p-3 rounded-lg">
						<div class="text-sm text-surface-600">Avg. Response Time</div>
						<div class="text-2xl font-semibold">{metrics.averageResponseTime}</div>
					</div>
					<div class="bg-surface-100 p-3 rounded-lg">
						<div class="text-sm text-surface-600">Escalations (30d)</div>
						<div class="text-2xl font-semibold">{metrics.escalationsLast30d}</div>
					</div>
					<div class="bg-surface-100 p-3 rounded-lg">
						<div class="text-sm text-surface-600">Handovers Completed</div>
						<div class="text-2xl font-semibold">{metrics.handoversCompleted}</div>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}

{#snippet rosterDetails()}
	<div class="border rounded-lg p-4 bg-surface-50">
		<Header title="Roster Details" />
		
		<div class="mt-2 flex flex-col gap-2">
			<div class="flex items-center gap-2">
				<span class="font-semibold">Name:</span>
				<span>{attr.name}</span>
			</div>
			<div class="flex items-center gap-2">
				<span class="font-semibold">Slug:</span>
				<span>{attr.slug || 'N/A'}</span>
			</div>
		</div>
	</div>

	<div class="border rounded-lg p-4 flex flex-col gap-2">
		<div class="flex items-center justify-between">
			<Header title="Team Members" />
			<Badge color="accent">{users.length}</Badge>
		</div>

		<div class="flex flex-col gap-2 overflow-y-auto max-h-60">
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
				<div class="flex items-center gap-4 bg-surface-100 p-3 rounded-lg">
					<div class="flex flex-col flex-1">
						<span class="font-medium">{service.name}</span>
					</div>
					<div class={getStatusColor(service.status)}>
						<Icon data={service.status === 'healthy' ? mdiCheckCircle : mdiAlertCircle} />
					</div>
				</div>
			{:else}
				<div class="text-surface-600 italic p-2">No services assigned to this roster</div>
			{/each}
		</div>
	</div>
	
	<div class="border rounded-lg p-4 flex flex-col gap-2">
		<Header title="Actions" />
		
		<div class="flex flex-col gap-2">
			<Button variant="primary">Edit Roster</Button>
			<Button variant="outline">Manage Shifts</Button>
			<Button variant="outline">Add User</Button>
		</div>
	</div>
{/snippet}
