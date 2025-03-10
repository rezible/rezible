<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallShiftOptions, type OncallShift } from "$lib/api";
	import { appShell, setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import { formatShiftDates } from "$features/oncall/lib/shift";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import PageActionButtons from "./PageActionButtons.svelte";
	import { Button, Header, Icon } from "svelte-ux";
	import { mdiArrowRight, mdiArrowUp, mdiChevronRight } from "@mdi/js";
	import Avatar from "$components/avatar/Avatar.svelte";
	import ShiftStats from "./shift-stats/ShiftStats.svelte";
	import ShiftAnnotationsList from "$features/oncall/components/shift-annotations/ShiftAnnotationsList.svelte";

	type Props = {
		shiftId: string;
	};
	const { shiftId }: Props = $props();

	const query = createQuery(() => getOncallShiftOptions({ path: { id: shiftId } }));
	const shift = $derived(query.data?.data);
	const shiftDates = $derived(shift ? formatShiftDates(shift) : "");

	setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Shifts", href: "/oncall/shifts" },
		{ label: shiftDates, href: `/oncall/shifts/${shiftId}` },
	]);

	appShell.setPageActions(PageActionButtons, false);
</script>

<LoadingQueryWrapper {query} view={shiftView} />

{#snippet shiftView(shift: OncallShift)}
	{@const attr = shift.attributes}
	{@const roster = attr.roster}
	{@const user = attr.user}
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

			<ShiftStats {shift} />
		</div>

		<div class="flex flex-col gap-1 h-full min-h-0 border rounded-lg p-2">
			<ShiftAnnotationsList editable={false} shiftId={shift.id} />
		</div>
	</div>
{/snippet}