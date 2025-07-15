<script lang="ts">
	import type { OncallShift } from "$lib/api";
	import { isFuture, formatDistanceToNowStrict, isPast } from "date-fns";

	import Icon from "$components/icon/Icon.svelte";
	import { mdiChevronRight } from "@mdi/js";
	import Avatar from "$components/avatar/Avatar.svelte";
	import ShiftProgressCircle from "$features/oncall/components/shift-progress-circle/ShiftProgressCircle.svelte";
	import Header from "$components/header/Header.svelte";
	import Card from "$components/card/Card.svelte";

	type Props = {
		shift: OncallShift;
	};
	let { shift }: Props = $props();

	const attr = $derived(shift.attributes);

	const start = $derived(new Date(attr.startAt));
	const end = $derived(new Date(attr.endAt));

	const isUpcoming = $derived(isFuture(start));
	const isFinished = $derived(isPast(end));
	const isActive = $derived(!isUpcoming && !isFinished);
</script>

<a href="/shifts/{shift.id}" class="group w-96">
	<Card
		classes={{
			root: "bg-success-900/20 border-success-100/10 group-hover:bg-success-900/50 group-hover:border-success-100/50",
			headerContainer: "py-2",
		}}
	>
		{#snippet header()}
			<Header>
				{#snippet title()}
					<div class="flex gap-2 items-center">
						<Avatar kind="roster" size={20} id={attr.roster.id} />
						<span class="text-lg font-semibold">{attr.roster.attributes.name}</span>
					</div>
				{/snippet}
				{#snippet subheading()}
					<div class="flex gap-2 items-center">
						<span class="text-sm">Ends in {formatDistanceToNowStrict(end)}</span>
					</div>
				{/snippet}
				{#snippet actions()}
					<div class:hidden={!isActive}>
						<ShiftProgressCircle {shift} size={32} />
					</div>
				{/snippet}
			</Header>
		{/snippet}

		{#snippet contents()}
			<div class="flex gap-2 items-center border-t pt-2">
				<Avatar kind="user" size={32} id={attr.user.id} />
				<div class="flex flex-col">
					<span class="font-bold">{attr.user.attributes.name ?? "user"}</span>
					<span class="font-normal">{attr.role ?? "unknown role"}</span>
				</div>
			</div>
		{/snippet}

		{#snippet actions()}
			<div class="flex justify-end items-center">
				<span class="flex items-center group-hover:text-success">
					View <Icon data={mdiChevronRight} />
				</span>
			</div>
		{/snippet}
	</Card>
</a>
