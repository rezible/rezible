<script lang="ts">
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import { getLocalTimeZone, parseAbsolute, ZonedDateTime } from "@internationalized/date";
	import { cls } from "@layerstack/tailwind";
	import { differenceInCalendarDays } from "date-fns";
	import { v4 as uuidv4 } from "uuid";
	import type { OncallShift } from "$lib/api";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import { type ShiftEvent } from "$src/features/oncall/lib/utils";
	import { shiftCtx } from "$features/oncall/lib/context.svelte";
	import PageActions from "./PageActions.svelte";
	import ShiftDetailsBar from "./ShiftDetailsBar.svelte";
	import ShiftDetails from "./shift-details/ShiftDetails.svelte";
	import ShiftEvents from "./shift-events/ShiftEvents.svelte";

	type Props = { shift: OncallShift };
	const { shift }: Props = $props();

	appShell.setPageActions(PageActions, false);

	shiftCtx.set(shift);

	// TODO: default to shift timezone & allow choosing timezone
	let eventTimezone = $state(getLocalTimeZone());
	const shiftStart = $derived(parseAbsolute(shift.attributes.startAt, eventTimezone));
	const shiftEnd = $derived(parseAbsolute(shift.attributes.endAt, eventTimezone));

	// TODO: Implement this properly
	const makeFakeShiftEvent = (date: ZonedDateTime): ShiftEvent => {
		const isAlert = Math.random() > 0.25;
		const eventType = isAlert ? "alert" : "incident";
		const hour = Math.floor(Math.random() * 24);
		const minute = Math.floor(Math.random() * 60);
		const timestamp = date.copy().set({ hour, minute });
		return { id: uuidv4(), timestamp, eventType, description: "description", annotation: "annotation" };
	};
	
	const shiftEventsQuery = createQuery(() => queryOptions({
		queryKey: ["shiftEvents", shift.id],
		queryFn: async () => {
			const shiftDays = differenceInCalendarDays(shiftEnd.toDate(), shiftStart.toDate());
			let events: ShiftEvent[] = [];
			for (let day = 0; day < shiftDays; day++) {
				const dayDate = shiftStart.add({ days: day });
				const numDayEvents = Math.floor(Math.random() * 10);
				const dayEvents = Array.from({ length: numDayEvents }, () => makeFakeShiftEvent(dayDate));
				events = events.concat(dayEvents);
			}
			return { data: events };
		}
	}));

	const shiftEvents = $derived(shiftEventsQuery.data?.data || []);

	type ShiftViewTab = "details" | "events" | "handover";
	const tabs: {value: ShiftViewTab, label: string}[] = [
		{label: "Overview", value: "details"},
		{label: "Events", value: "events"},
		{label: "Handover", value: "handover"},
	];

	let currentTab = $state<ShiftViewTab>("details");
</script>

<div class="flex flex-col h-full max-h-full min-h-0 overflow-hidden">
	<div class="w-full flex justify-between h-16 z-[1]">
		<div class="flex gap-2 self-end">
			{#each tabs as tab}
				{@const active = tab.value === currentTab}
				<button 
					class={cls(
						"inline-flex self-end h-14 p-4 py-3 text-lg border border-b-0 rounded-t-lg relative", 
						active && "bg-surface-100 text-secondary",
					)}
					onclick={() => (currentTab = tab.value)}>
					{tab.label}
					<div class="absolute bottom-0 -mb-px left-0 w-full border-b border-surface-100" class:hidden={!active}></div>
				</button>
			{/each}
		</div>

		<ShiftDetailsBar {shift} {shiftStart} {shiftEnd} />
	</div>

	<div class="flex-1 min-h-0 max-h-full overflow-y-auto border rounded-b-lg rounded-tr-lg p-2 bg-surface-100">
		{#if currentTab === "details"}
			<ShiftDetails {shift} {shiftEvents} />
		{:else if currentTab === "events"}
			<ShiftEvents events={shiftEvents} {shiftStart} {shiftEnd} />
		{/if}
	</div>
</div>
