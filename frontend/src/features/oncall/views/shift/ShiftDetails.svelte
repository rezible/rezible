<script lang="ts">
	import type { OncallShift } from "$lib/api";
	import { Header, Icon } from "svelte-ux";
	import { mdiArrowUp, mdiChevronRight } from "@mdi/js";
	import Avatar from "$components/avatar/Avatar.svelte";
	import ShiftEventsStats from "./shift-events-stats/ShiftEventsStats.svelte";
	import ShiftAnnotationsList from "$features/oncall/components/shift-annotations/ShiftAnnotationsList.svelte";
	import { parseAbsoluteToLocal } from "@internationalized/date";
	import { makeFakeShiftEvents } from "$features/oncall/lib/shift";


	type Props = {
		shift: OncallShift;
	};
	const { shift }: Props = $props();

	const attr = $derived(shift.attributes);
	const roster = $derived(attr.roster);
	const user = $derived(attr.user);

	const shiftStart = $derived(parseAbsoluteToLocal(shift.attributes.startAt));
	const shiftEnd = $derived(parseAbsoluteToLocal(shift.attributes.endAt));
	const shiftEvents = $derived(makeFakeShiftEvents(shiftStart, shiftEnd));
</script>

<div class="grid grid-cols-2 gap-2 h-full max-h-full min-h-0 overflow-hidden">
	<div class="flex flex-col gap-1 h-full min-h-0">
		<div class="grid grid-cols-2 gap-2 auto-rows-min">
			<div class="flex flex-col gap-1 border rounded-lg p-2">
				<Header title="User" />

				<a href="/users/{user.id}" class="flex-1">
					<div class="flex items-center gap-4 bg-surface-100 hover:bg-accent-800/40 h-full p-2">
						<Avatar kind="user" size={32} id={user.id} />
						<div class="flex flex-col">
							<span class="text-lg">{user.attributes.name}</span>
							<span>{attr.role}</span>
						</div>
						<div class="flex-1 grid justify-items-end">
							<Icon data={mdiChevronRight} />
						</div>
					</div>
				</a>
			</div>

			<div class="flex flex-col gap-1 border rounded-lg p-2">
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
		</div>

		<div class="flex flex-col gap-1 border rounded-lg p-2">
			<Header title="Shift Burden" />

			<div class="flex items-center gap-4 bg-surface-100 h-full p-2">
				<Icon data={mdiArrowUp} />
				<div class="flex flex-col">
					<span class="text-lg">High</span>
					<span>23% more than usual for roster</span>
				</div>
			</div>
		</div>

		<ShiftEventsStats {shiftStart} {shiftEnd} {shiftEvents} />
	</div>

	<div class="flex flex-col gap-1 h-full min-h-0 border rounded-lg p-2">
		<ShiftAnnotationsList editable={false} shiftId={shift.id} />
	</div>
</div>