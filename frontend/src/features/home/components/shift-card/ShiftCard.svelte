<script lang="ts">
	import { Button, Card, cls, Header, Icon, ProgressCircle } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import PreviousShiftOverview from "./PreviousShiftOverview.svelte";
	import { mdiArrowRight, mdiPhone } from "@mdi/js";
	import type { OncallShift } from "$src/lib/api";
	import { differenceInHours } from "date-fns";

	type Props = {
		shifts: OncallShift[];
	};
	const { shifts }: Props = $props();

	const shift = $derived(shifts[0]);
	const roster = $derived(shift.attributes.roster);

	const start = $derived(new Date(shift.attributes.startAt));
	const end = $derived(new Date(shift.attributes.endAt));
	const progress = $derived((100 * (Date.now() - start.valueOf())) / (end.valueOf() - start.valueOf()));
	const firstDay = $derived(differenceInHours(Date.now(), start) <= 24);

	let expanded = $state(false);
</script>

<Card
	class="w-full max-h-full border-success-900/50 bg-success-900/5 rounded-lg overflow-auto"
	classes={{
		content: "min-h-0 flex",
		headerContainer: "pb-2",
	}}
>
	<Header title="You are Currently Oncall" slot="header" classes={{ title: "text-xl" }}>
		<div slot="avatar">
			<ProgressCircle
				size={32}
				value={progress}
				track
				class="text-success [--track-color:theme(colors.success/10%)]"
			/>
		</div>
		<span slot="subheading" class="text-surface-content/70 inline-flex gap-1 items-center whitespace-pre">
			<span class="font-semibold">{shift.attributes.role}</span>
			<span class="ml-1">for</span>
			<Button size="sm" href="/oncall/rosters/{roster.attributes.slug}" classes={{ root: "p-1 py-0" }}>
				<span class="font-bold text-base">{roster.attributes.name}</span>
				<div class="self-center ml-1">
					<Avatar id={"roster-id"} kind="roster" size={16} />
				</div>
			</Button>
		</span>
	</Header>

	<svelte:fragment slot="contents">
		<PreviousShiftOverview bind:expanded />
	</svelte:fragment>

	<div slot="actions" class="flex gap-2 px-2 py-0 justify-end">
		<Button href="/oncall/shifts/{shift.id}" variant="fill-light" color="secondary">View Handover</Button>

		<Button href="/oncall/shifts/{shift.id}" variant="fill" color="primary">
			View Shift
			<Icon data={mdiArrowRight} />
		</Button>
	</div>
</Card>
