<script lang="ts" module>
	import type { Time as TimeType } from '@internationalized/date';
	import type { Period } from './utils';
	import type { HTMLButtonAttributes } from 'svelte/elements';

	export type PeriodSelectorProps = HTMLButtonAttributes & {
		period: Period;
		setPeriod?: (period: PeriodSelectorProps['period']) => void;

		time: TimeType | undefined;
		setTime?: (time: TimeType) => void;

		onRightFocus?: () => void;
		onLeftFocus?: () => void;

		ref: HTMLElement | null;
	};
</script>

<script lang="ts">
	import { display12HourValue, setDateByType } from './utils';
	import { Time } from '@internationalized/date';
	import { onMount } from 'svelte';

	let {
		period = $bindable('PM'),
		time = $bindable(new Time(0, 0)),
		ref,

		onLeftFocus,
		onRightFocus,
		setPeriod,
		setTime
	}: PeriodSelectorProps = $props();

	function handlePeriod() {
		const tempTime = time.copy();
		const hours = display12HourValue(time.hour);
		const _time = setDateByType(
			tempTime,
			hours.toString(),
			'12hours',
			period === 'AM' ? 'PM' : 'AM'
		);

		time = _time;
		setTime?.(_time);
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key === 'ArrowRight') onRightFocus?.();
		if (e.key === 'ArrowLeft') onLeftFocus?.();
	}

	function handleValueChange(value: Period) {
		period = value;
		setPeriod?.(value);

		/**
		 * trigger an update whenever the user switches between AM and PM;
		 * otherwise user must manually change the hour each time
		 */
		if (time) {
			handlePeriod();
		}
	}

	onMount(() => {
		handlePeriod();
	});
</script>

<select bind:value={period} onchange={e => {handleValueChange(e.currentTarget.value as Period)}} class="w-[48px] h-6 text-center font-mono text-base tabular-nums caret-transparent bg-surface-200 focus:bg-primary focus:text-primary-content [&::-webkit-inner-spin-button]:appearance-none">
	<option selected={period === "AM"} value="AM">AM</option>
	<option selected={period === "PM"} value="PM">PM</option>
</select>