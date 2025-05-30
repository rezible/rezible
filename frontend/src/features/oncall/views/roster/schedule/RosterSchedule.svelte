<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, Month } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiChevronRight } from "@mdi/js";
	import { getUserOncallInformationOptions, type OncallShift } from "$lib/api";
	import { formatDate, isFuture, isPast } from "date-fns";
	import { getLocalTimeZone, parseAbsoluteToLocal } from "@internationalized/date";
	import Header from "$src/components/header/Header.svelte";

	type Props = {};
	const {}: Props = $props();

	// TODO: use correct query
	const shiftsQuery = createQuery(() => getUserOncallInformationOptions());

	const currentShifts = $derived<OncallShift[]>(shiftsQuery.data?.data.activeShifts ?? []);
	const pastShifts = $derived<OncallShift[]>(shiftsQuery.data?.data.pastShifts ?? []);
	const upcomingShifts = $derived<OncallShift[]>(shiftsQuery.data?.data.upcomingShifts ?? []);

	const allShifts = $derived([...pastShifts, ...currentShifts, ...upcomingShifts]);

	const isActive = (s: OncallShift) => {
		return isPast(s.attributes.startAt) && isFuture(s.attributes.endAt);
	}

	const coverRequests: string[] = [];
</script>

<div class="grid grid-cols-4 gap-2 w-full h-full">
	<div class="col-span-2 border flex flex-col">
		<div class="flex flex-col p-2">
			<Header title="Schedule Details" classes={{ root: "text-lg font-medium" }} />

			<span>Every <span class="font-bold">Monday</span> at <span class="font-bold">9AM</span> in <span class="font-bold">{getLocalTimeZone()}</span></span>
		</div>

		<div class="py-2 border-y">
			<Month />
		</div>
	</div>

	<div class="flex flex-col border p-2">
		<Header title="Shifts" classes={{ root: "text-lg font-medium" }} />

		{#snippet shiftItem(shift: OncallShift)}
			{@const start = parseAbsoluteToLocal(shift.attributes.startAt).toDate()}
			<a href="/oncall/shifts/{shift.id}" class="block">
				<div
					class="flex items-center gap-4 bg-surface-100 hover:bg-surface-content/10 p-3 rounded-lg justify-between border"
					class:border-success-900={isActive(shift)}
				>
					<div class="flex flex-col flex-1">
						<span class="font-medium">{shift.attributes.user.attributes.name}</span>
						<div class="text-sm text-surface-600">{formatDate(start, "yyyy-LL-dd")}</div>
					</div>
					<div class="justify-items-end">
						<Icon data={mdiChevronRight} />
					</div>
				</div>
			</a>
		{/snippet}

		<div class="flex flex-col gap-2 flex-1 overflow-y-auto min-h-0">
			{#each allShifts as shift}
				{@render shiftItem(shift)}
			{/each}
		</div>
	</div>

	<div class="flex flex-col gap-2">
		<div class="flex flex-col border p-2">
			<Header title="Cover Requests" classes={{ root: "text-lg font-medium" }}>
				{#snippet actions()}
					<Button variant="fill-light">Request Cover</Button>
				{/snippet}
			</Header>

			{#each coverRequests as req}
				<span>{req}</span>
			{:else}
				<span>No Outstanding Requests</span>
			{/each}
		</div>
	</div>
</div>