<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { differenceInHours } from "date-fns";
	import { mdiArrowRight } from "@mdi/js";
	import { getPreviousOncallShiftOptions, type OncallShift } from "$lib/api";
	import { Button, Card, Header, Icon, ProgressCircle } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";

	type Props = {
		shift: OncallShift;
	};
	const { shift }: Props = $props();

	const roster = $derived(shift.attributes.roster);

	const start = $derived(new Date(shift.attributes.startAt));
	const end = $derived(new Date(shift.attributes.endAt));
	const progress = $derived((100 * (Date.now() - start.valueOf())) / (end.valueOf() - start.valueOf()));
	const firstDay = $derived(differenceInHours(Date.now(), start) <= 24);

	// const previousShiftQuery = createQuery(() => getPreviousOncallShiftOptions({path: {id: shift.id}}))
	// const previousShift = $derived(previousShiftQuery.data?.data && false);
</script>

<div class="max-w-lg p-2 border border-success-900/50 bg-success-900/5 rounded-lg overflow-auto">
	<Header title="You are Currently Oncall" classes={{ title: "text-md" }}>
		<div slot="avatar">
			<ProgressCircle
				size={32}
				value={progress}
				track
				class="text-success [--track-color:theme(colors.success/10%)]"
			/>
		</div>
		<span slot="subheading" class="text-surface-content/70 inline-flex gap-1 items-center whitespace-pre">
			<span class="font-semibold text-lg">{shift.attributes.role}</span>
			<span class="ml-1 text-lg">for</span>
			<Button size="sm" href="/oncall/rosters/{roster.attributes.slug}" classes={{ root: "p-1 py-0" }}>
				<span class="font-bold text-base text-lg">{roster.attributes.name}</span>
				<div class="self-center ml-1">
					<Avatar id={"roster-id"} kind="roster" size={16} />
				</div>
			</Button>
		</span>

		<div slot="actions" class="flex flex-col gap-2 px-2 py-0 justify-end">
			<Button href="/oncall/shifts/{shift.id}" variant="fill" color="success">
				View
				<Icon data={mdiArrowRight} />
			</Button>
		</div>
	</Header>
</div>
