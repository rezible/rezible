<script lang="ts">
	import type { OncallShift } from "$lib/api";
	import { isFuture, formatDistanceToNowStrict, isPast } from "date-fns";

	import { Card, Header, Icon } from "svelte-ux";
	import { mdiChevronRight } from "@mdi/js";
	import Avatar from "$src/components/avatar/Avatar.svelte";
	import ShiftProgressCircle from "$features/oncall/components/shift-progress-circle/ShiftProgressCircle.svelte";

	type Props = {
		shift: OncallShift;
	}
	let { shift }: Props = $props();

	const attr = $derived(shift.attributes);

	const start = $derived(new Date(attr.startAt));
	const end = $derived(new Date(attr.endAt));

	const isUpcoming = $derived(isFuture(start));
	const isFinished = $derived(isPast(end));
	const isActive = $derived(!isUpcoming && !isFinished);
</script>

<a href="/oncall/shifts/{shift.id}" class="group w-96">
	<Card
		class="bg-success-900/20 border-success-100/10 group-hover:bg-success-900/50 group-hover:border-success-100/50"
		classes={{ headerContainer: "py-2" }}
	>
		<svelte:fragment slot="header">
			<Header>
				<div slot="title" class="flex gap-2 items-center">
					<Avatar kind="roster" size={20} id={attr.roster.id} />
					<span class="text-lg font-semibold">{attr.roster.attributes.name}</span>
				</div>
				<div slot="subheading" class="flex gap-2 items-center">
					<span class="text-sm">Ends in {formatDistanceToNowStrict(end)}</span>
				</div>
				<div slot="actions" class:hidden={!isActive}>
					<ShiftProgressCircle {shift} size={32} />
				</div>
			</Header>
		</svelte:fragment>

		<div slot="contents" class="flex gap-2 items-center border-t pt-2">
			<Avatar kind="user" size={32} id={attr.user.id} />
			<div class="flex flex-col">
				<span class="font-bold">{attr.user.attributes.name ?? "user"}</span>
				<span class="font-normal">{attr.role ?? "unknown role"}</span>
			</div>
		</div>

		<div slot="actions" class="flex justify-end items-center">
			<span class="flex items-center group-hover:text-success">
				View <Icon data={mdiChevronRight} />
			</span>
		</div>
	</Card>
</a>
