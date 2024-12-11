<script lang="ts">
	import type { OncallShift } from '$lib/api';
    import { session } from '$lib/auth.svelte';
    import { addMinutes, isFuture, differenceInMinutes, formatDistanceToNowStrict, isPast } from 'date-fns';

	import { cls, Card, Button, Header, ProgressCircle, Tooltip, Icon } from 'svelte-ux';
    import { mdiCircleMedium, mdiChevronRight } from '@mdi/js';

	interface Props {
		shift: OncallShift;
	};
	let { shift }: Props = $props();

	const attr = $derived(shift.attributes);

	const start = $derived(new Date(attr.start_at));
	const end = $derived(new Date(attr.end_at));
	const progress = $derived((Date.now() - start.valueOf()) / (end.valueOf() - start.valueOf()));
	const minutesLeft = $derived(differenceInMinutes(end, Date.now()));

	const isUpcoming = $derived(isFuture(start));
	const isFinished = $derived(isPast(end));
	const isActive = $derived(!isUpcoming && !isFinished);
</script>

<a href="/oncall/shifts/{shift.id}" class="group max-w-lg w-full">
	<Card
		class="w-full bg-success-900/20 border-success-100/10 group-hover:bg-success-900/50 group-hover:border-success-100/50"
		classes={{ headerContainer: 'py-2' }}>
		<svelte:fragment slot="header">
			{@render cardHeader()}
		</svelte:fragment>
		<div class="flex flex-col" slot="contents">
			<span class="">Ends in {formatDistanceToNowStrict(end)}</span>
		</div>
		<div slot="actions" class="flex justify-end items-center">
			<span class="flex items-center group-hover:text-success">View <Icon data={mdiChevronRight} /></span>
		</div>
	</Card>
</a>

{#snippet cardHeader()}
	<Header>
		<div slot="title">
			<span class="text-lg font-semibold">{attr.roster.attributes.name}</span>
		</div>
		<div slot="subheading" class="flex flex-col">
			<span class="">{attr.role ?? "unknown role"}</span>
		</div>
		<div slot="actions" class:hidden={!isActive}>
			<Tooltip>
				<ProgressCircle
					size={32}
					value={100 * progress}
					track
					class="text-success [--track-color:theme(colors.success/10%)]"
				>
					<Icon data={mdiCircleMedium} class="animate-pulse" />
				</ProgressCircle>
				<div slot="title" class="bg-neutral border text-sm text-surface-content p-2 rounded-lg">
					<div class="flex flex-col gap-2">
						<span>{minutesLeft} minutes left</span>
						<div>
							<span>Start: </span>
							<span>{start.toLocaleString()}</span>
						</div>
						<div class="">
							<span>End: </span>
							<span>{end.toLocaleString()}</span>
						</div>
					</div>
				</div>
			</Tooltip>
		</div>
	</Header>
{/snippet}
