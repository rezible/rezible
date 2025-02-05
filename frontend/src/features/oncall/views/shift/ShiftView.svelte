<script lang="ts">
	import { Button, Header, Icon } from "svelte-ux";
	import { mdiArrowRight, mdiChevronRight } from "@mdi/js";
	import type { OncallShift } from "$lib/api";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { session } from "$lib/auth.svelte";
	import { buildShiftTimeDetails } from "$features/oncall/lib/shift";
	import ShiftStats from "./ShiftStats.svelte";
	import ShiftAnnotationsList from "$features/oncall/components/shift-annotations/ShiftAnnotationsList.svelte";

	type Props = {
		shift: OncallShift;
	};
	const { shift }: Props = $props();

	const attr = $derived(shift.attributes);
	const roster = $derived(attr.roster);
	const user = $derived(attr.user);

	const timeDetails = $derived(buildShiftTimeDetails(shift));
	const isSessionUser = $derived(session.userId == user.id);
</script>

{#if timeDetails.status === "active" && isSessionUser}
	<div class="mb-2 w-full">
		<Button
			href="/oncall/shifts/{shift.id}/handover"
			variant="fill"
			color="success"
			classes={{ root: "w-full" }}
		>
			<span>Edit Handover</span>
			<Icon data={mdiArrowRight} />
		</Button>
	</div>
{/if}

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

		<ShiftStats {shift} />
	</div>

	<div class="flex flex-col gap-1 h-full min-h-0 border rounded-lg p-2">
		<ShiftAnnotationsList editable={false} shiftId={shift.id} />
	</div>
</div>
