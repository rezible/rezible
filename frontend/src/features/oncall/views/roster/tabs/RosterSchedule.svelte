<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Header, Icon, Button, Month } from "svelte-ux";
	import { mdiChevronRight, mdiCalendarClock, mdiClockOutline } from "@mdi/js";
	import { getUserOncallDetailsOptions } from "$lib/api";
	import Avatar from "$components/avatar/Avatar.svelte";
	import type { OncallShift } from "../types";
	import { addDays, subDays } from "date-fns";

	type Props = {};
	const {}: Props = $props();

	// TODO: use correct query
	const shiftsQuery = createQuery(() => getUserOncallDetailsOptions());

	const currentShifts = $derived<OncallShift[]>(shiftsQuery.data?.data.activeShifts ?? []);
	const pastShifts = $derived<OncallShift[]>(shiftsQuery.data?.data.pastShifts ?? []);
	const upcomingShifts = $derived<OncallShift[]>(shiftsQuery.data?.data.upcomingShifts ?? []);
</script>

<div class="grid grid-cols-2 w-full h-full">
	<div class="flex flex-col gap-2 border rounded p-2 border-success-900/40">
		<Header title="Active" classes={{ root: "gap-2 text-lg font-medium" }}>
			<div slot="avatar">
				<Icon data={mdiClockOutline} class="text-success-800" />
			</div>
		</Header>

		<div class="flex flex-col gap-2">
			{#each currentShifts as shift}
				<a href="/oncall/shifts/{shift.id}" class="block">
					<div
						class="flex items-center gap-4 bg-success-900/40 hover:bg-success-800/50 p-3 rounded-lg"
					>
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

		<div class="flex-1 grid grid-cols-2 gap-4">
			<div class="flex flex-col">
				<Header title="Recent" classes={{ root: "gap-2 text-lg font-medium mt-2" }}>
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
							<div
								class="flex items-center gap-4 bg-surface-100 hover:bg-accent-800/50 p-3 rounded-lg justify-between"
							>
								<div class="flex flex-col flex-1">
									<span class="font-medium">{shift.attributes.user.attributes.name}</span>
									<div class="text-sm text-surface-600">time</div>
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
				<Header title="Upcoming Shifts" classes={{ root: "gap-2 text-lg font-medium mt-2" }}>
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
								<div class="text-sm text-surface-600">time</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>
	<div class="border">
		<Month />
	</div>
</div>
