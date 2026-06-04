<!-- adapted from https://time-picker.nouro.app/ -->

<script lang="ts">
	import { Time } from '@internationalized/date';
	import TimePickerInput from './TimePickerInput.svelte';
	import TimePeriodSelect from "./TimePeriodSelect.svelte";
	import { cn } from '$lib/utils';
	import type { Period } from './utils';

	type Props = {
		time?: Time;
		period?: Period;
		view?: 'labels' | 'dotted';
		rangeMin?: Time;
		rangeMax?: Time;
		setTime?: (time: Time) => void;
		setPeriod?: (period: Period) => void;
	};

	let {
		time = $bindable(new Time(0, 0, 0)), 
		period = $bindable("AM"),
		view = "labels",
		setTime,
		setPeriod,
	}: Props = $props();

	let minuteRef = $state<HTMLInputElement | null>(null);
	let hourRef = $state<HTMLInputElement | null>(null);
	let secondRef = $state<HTMLInputElement | null>(null);
	let periodRef = $state<HTMLInputElement | null>(null);
</script>

<div class={cn('flex items-center gap-2', view === 'dotted' && 'gap-1')}>
	<div class="grid gap-1 text-center">
		{#if view === 'labels'}
			<label for="hours" class="text-xs">Hours</label>
		{/if}

		<TimePickerInput
			picker="12hours"
			bind:time
			bind:ref={hourRef}
			{period}
			{setTime}
			onRightFocus={() => minuteRef?.focus()}
		/>
	</div>

	{#if view === 'dotted'}
		<span class="-translate-y-[2px]">:</span>
	{/if}

	<div class="grid gap-1 text-center">
		{#if view === 'labels'}
			<label for="minutes" class="text-xs">Minutes</label>
		{/if}

		<TimePickerInput
			picker="minutes"
			bind:time
			bind:ref={minuteRef}
			{setTime}
			onLeftFocus={() => hourRef?.focus()}
			onRightFocus={() => secondRef?.focus()}
		/>
	</div>

	{#if view === 'dotted'}
		<span class="-translate-y-[2px]">:</span>
	{/if}

	<div class="grid gap-1 text-center">
		{#if view === 'labels'}
			<label for="seconds" class="text-xs">Seconds</label>
		{/if}

		<TimePickerInput
			picker="seconds"
			bind:time
			bind:ref={secondRef}
			{setTime}
			onLeftFocus={() => minuteRef?.focus()}
		/>
	</div>

	<div class="grid gap-1 text-center">
		{#if view === 'labels'}
			<label for="seconds" class="text-xs">Period</label>
		{/if}

		<TimePeriodSelect
			bind:period
			bind:time
			{setPeriod}
			{setTime}
			ref={periodRef}
			onLeftFocus={() => secondRef?.focus()}
		/>
	</div>
</div>
