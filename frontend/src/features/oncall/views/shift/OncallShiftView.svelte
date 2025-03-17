<script lang="ts">
	import { Header, Icon } from "svelte-ux";
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import { mdiArrowUp, mdiChevronRight } from "@mdi/js";
	import { getLocalTimeZone, parseAbsolute } from "@internationalized/date";
	import { differenceInCalendarDays } from "date-fns";
	import { v4 as uuidv4 } from "uuid";
	import type { OncallShift } from "$lib/api";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import { shiftEventMatchesFilter, type ShiftEvent, type ShiftEventFilterKind } from "$features/oncall/lib/utils";
	import { shiftCtx } from "$features/oncall/lib/context.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import ShiftEvents from "./shift-events/ShiftEvents.svelte";
	import PageActions from "./PageActions.svelte";

	type Props = { shift: OncallShift };
	const { shift }: Props = $props();

	appShell.setPageActions(PageActions, false);

	shiftCtx.set(shift);

	const role = $derived(shift.attributes.role);
	const roster = $derived(shift.attributes.roster);
	const user = $derived(shift.attributes.user);

	// TODO: default to shift timezone & allow choosing timezone
	let eventTimezone = $state(getLocalTimeZone());
	const shiftStart = $derived(parseAbsolute(shift.attributes.startAt, eventTimezone));
	const shiftEnd = $derived(parseAbsolute(shift.attributes.endAt, eventTimezone));

	// TODO: Implement this properly
	const shiftEventsQuery = createQuery(() => queryOptions({
		queryKey: ["shiftEvents", shift.id],
		queryFn: async () => {
			const shiftDays = differenceInCalendarDays(shiftEnd.toDate(), shiftStart.toDate());
			let events: ShiftEvent[] = [];
			for (let day = 0; day < shiftDays; day++) {
				const dayDate = shiftStart.add({ days: day });
				const makeFakeShiftEvent = (): ShiftEvent => {
					const isAlert = Math.random() > 0.25;
					const eventType = isAlert ? "alert" : "incident";
					const hour = Math.floor(Math.random() * 24);
					const minute = Math.floor(Math.random() * 60);
					const timestamp = dayDate.copy().set({ hour, minute });
					return { id: uuidv4(), timestamp, eventType, description: "description", notes: "some notes" };
				};
				const numDayEvents = Math.floor(Math.random() * 10);
				const dayEvents = Array.from({ length: numDayEvents }, makeFakeShiftEvent);
				events = events.concat(dayEvents);
			}
			return { data: events };
		}
	}));

	const shiftEventsData = $derived(shiftEventsQuery.data?.data || []);

	const burdenScore = $derived(0.23);
	const burdenRating = $derived("High");
</script>

<div class="flex gap-2 h-full max-h-full min-h-0 overflow-hidden">
	<div class="flex flex-col gap-2">
		{@render shiftDetails()}
	</div>

	<div class="flex-1 min-h-0 max-h-full overflow-y-auto border rounded-lg p-2">
		<ShiftEvents events={shiftEventsData} {shiftStart} {shiftEnd} />
	</div>
</div>

{#snippet shiftDetails()}
	<div class="flex flex-col gap-1 border rounded-lg p-2 min-w-72">
		<Header title="Users" />

		<a href="/users/{user.id}" class="flex-1">
			<div class="flex items-center gap-4 bg-surface-100 hover:bg-accent-800/40 h-full p-2">
				<Avatar kind="user" size={32} id={user.id} />
				<div class="flex flex-col">
					<span class="text-lg">{user.attributes.name}</span>
					<span>{role}</span>
				</div>
				<div class="flex-1 grid justify-items-end">
					<Icon data={mdiChevronRight} />
				</div>
			</div>
		</a>
	</div>

	<div class="flex flex-col gap-1 border rounded-lg p-2 min-w-72">
		<Header title="Roster" />

		<a href="/oncall/rosters/{roster.id}" class="flex-1">
			<div class="flex items-center gap-4 bg-surface-100 hover:bg-accent-800/40 h-full p-2">
				<Avatar kind="roster" size={32} id={roster.id} />
				<span class="text-lg">{roster.attributes.name}</span>
				<div class="flex-1 grid justify-items-end">
					<Icon data={mdiChevronRight} />
				</div>
			</div>
		</a>
	</div>

	<div class="flex flex-col gap-1 border rounded-lg p-2 min-w-72">
		<Header title="Shift Burden" />

		<div class="flex items-center gap-4 bg-danger-400/20 h-full p-2">
			<Icon data={mdiArrowUp} />
			<div class="flex flex-col">
				<span class="text-lg">{burdenRating}</span>
				<span>{burdenScore * 100}% more than usual for roster</span>
			</div>
		</div>
	</div>
{/snippet}