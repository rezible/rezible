<script lang="ts">
	import { type User, type OncallRoster, getUserOncallDetailsOptions } from "$lib/api";
	import { mdiChevronRight, mdiCalendarClock, mdiAlertCircle, mdiCheckCircle, mdiClockOutline, mdiGraph, mdiAccount } from "@mdi/js";
	import { Header, Icon, Button } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { createQuery } from "@tanstack/svelte-query";

	type Props = {};
	const {}: Props = $props();

	const users = $state<User[]>([
		{ id: "u1", attributes: { name: "Jane Doe", email: "jane@example.com" } },
		{ id: "u2", attributes: { name: "John Smith", email: "john@example.com" } },
		{ id: "u3", attributes: { name: "Alex Johnson", email: "alex@example.com" } },
		{ id: "u4", attributes: { name: "Sarah Williams", email: "sarah@example.com" } }
	]);

	// TODO: use correct query
	const shiftsQuery = createQuery(() => getUserOncallDetailsOptions());
	const currentShifts = $derived(shiftsQuery.data?.data.activeShifts ?? []);
	const pastShifts = $derived(shiftsQuery.data?.data.pastShifts ?? []);
	const upcomingShifts = $derived(shiftsQuery.data?.data.upcomingShifts ?? []);
	
	const services = $state([
		{ id: "s1", name: "API Gateway", description: "", status: "healthy" },
		{ id: "s2", name: "Database Cluster", description: "", status: "healthy" },
		{ id: "s3", name: "Authentication Service", description: "", status: "warning" }
	]);
	
	function getStatusColor(status: string): string {
		switch (status) {
			case "healthy": return "text-green-500";
			case "warning": return "text-yellow-500";
			case "critical": return "text-red-500";
			default: return "text-gray-500";
		}
	}
</script>

<div class="border rounded-lg flex gap-2 row-span-1 py-1">
	<div class="flex flex-col gap-2 w-full">
		<Header title="Members" classes={{root: "gap-2 text-lg font-medium px-2 pt-2"}}>
			<div slot="avatar">
				<Icon data={mdiAccount} class="" />
			</div>
		</Header>

		<div class="flex-1 flex flex-col gap-1 overflow-y-auto">
			{#each users as usr}
				<a href="/users/{usr.id}" class="block w-full">
					<div class="flex items-center bg-surface-100 hover:bg-accent-800/40 p-2 w-full">
						<div class="flex items-center gap-2">
							<Avatar kind="user" size={20} id={usr.id} />
							<span class="font-medium">{usr.attributes.name}</span>
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
</div>

<div class="border rounded-lg flex gap-2 row-span-1 py-1">
	<div class="flex flex-col gap-2 w-full">
		<Header title="Services" classes={{root: "gap-2 text-lg font-medium px-2 pt-2"}}>
			<div slot="avatar">
				<Icon data={mdiGraph} class="" />
			</div>
		</Header>

		<div class="flex-1 flex flex-col gap-1 overflow-y-auto">
			{#each services as service}
				<a href="/services/{service.id}" class="block w-full">
					<div class="flex items-center bg-surface-100 hover:bg-accent-800/40 p-2 w-full">
						<div class="flex items-center gap-2">
							<span class="font-medium">{service.name}</span>
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
</div>

<div class="col-span-2 row-span-2 flex flex-col p-2 border rounded-lg overflow-y-auto">
	<Header title="Oncall Shifts" classes={{root: "gap-2 text-lg font-medium"}} />

	<div class="flex flex-col gap-2 border rounded p-2 border-success-900/40">
		<Header title="Active" classes={{root: "gap-2 text-lg font-medium"}}>
			<div slot="avatar">
				<Icon data={mdiClockOutline} class="text-success-800" />
			</div>
		</Header>
			
		<div class="flex flex-col gap-2">
			{#each currentShifts as shift}
				<a href="/oncall/shifts/{shift.id}" class="block">
					<div class="flex items-center gap-4 bg-success-900/40 hover:bg-success-800/50 p-3 rounded-lg">
						<Avatar kind="user" size={40} id={shift.attributes.user.id} />
						<div class="flex flex-col">
							<span class="text-lg font-medium">{shift.attributes.user.attributes.name}</span>
							<span class="text-sm text-surface-600"></span>
						</div>
						<div class="flex-1 grid justify-items-end">
							<Icon data={mdiChevronRight} />
						</div>
					</div>
				</a>
			{/each}
		</div>
	</div>
	
	<div class="flex-1 grid grid-cols-2 gap-4">
		<div class="flex flex-col">
			<Header title="Recent" classes={{root: "gap-2 text-lg font-medium mt-2"}}>
				<svelte:fragment slot="actions">
					<Button variant="text" href="/oncall/shifts">View All</Button>
				</svelte:fragment>
				<div slot="avatar">
					<Icon data={mdiCalendarClock} class="text-accent-500" />
				</div>
			</Header>

			<div class="flex flex-col gap-2 flex-1 overflow-y-auto min-h-0">
				{#each [...pastShifts, ...pastShifts, ...pastShifts, ...pastShifts] as shift}
					<a href="/oncall/shifts/{shift.id}" class="block">
						<div class="flex items-center gap-4 bg-surface-100 hover:bg-accent-800/50 p-3 rounded-lg justify-between">
							<div class="flex flex-col flex-1">
								<span class="font-medium">{shift.attributes.user.attributes.name}</span>
								<div class="text-sm text-surface-600">
									time
								</div>
							</div>
							<div class="justify-items-end">
								<Icon data={mdiChevronRight} />
							</div>
						</div>
					</a>
				{/each}
			</div>
		</div>

		<div class="">
			<Header title="Upcoming Shifts" classes={{root: "gap-2 text-lg font-medium mt-2"}}>
				<svelte:fragment slot="actions">
					<Button variant="text" href="/oncall/shifts">View All</Button>
				</svelte:fragment>
				<div slot="avatar">
					<Icon data={mdiCalendarClock} class="text-accent-500" />
				</div>
			</Header>

			<div class="flex flex-col gap-2 min-h-32 overflow-y-auto">
				{#each upcomingShifts as shift}
					<div class="flex items-center gap-4 bg-surface-100 p-3 rounded-lg">
						<div class="flex flex-col flex-1">
							<span class="font-medium">{shift.attributes.user.attributes.name}</span>
							<div class="text-sm text-surface-600">
								time
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>