<script lang="ts" module>
	import type { HTMLInputAttributes } from 'svelte/elements';
	import type { Time as TimeType } from '@internationalized/date';

	export type TimePickerInputProps = HTMLInputAttributes & {
		type?: string;
		value?: string;
		name?: string;
		picker: TimePickerType;
		time: TimeType;
		setTime?: (time: TimeType) => void;
		period?: Period;
		onRightFocus?: () => void;
		onLeftFocus?: () => void;
		ref: HTMLElement | null;
	};
</script>

<script lang="ts">
	import {
		type Period,
		type TimePickerType,
		getArrowByType,
		getDateByType,
		setDateByType
	} from './utils';
	import { cls } from 'svelte-ux';

	let {
		class: className,
		type = 'tel',
		value,
		id,
		name,
		time = $bindable(),
		setTime,
		picker,
		period,
		onLeftFocus,
		onRightFocus,

		onkeydown,
		onchange,

		ref = $bindable<HTMLElement | null>(null),

		...restProps
	}: TimePickerInputProps = $props();

	let flag = $state<boolean>(false);
	let intKey = $state<string>('0');

	let calculatedValue = $derived(getDateByType(time, picker));

	$effect(() => {
		if (flag) {
			const timer = setTimeout(() => {
				flag = false;
			}, 2000);

			return () => clearTimeout(timer);
		}
	});

	function calculateNewValue(key: string) {
		/*
		 * If picker is '12hours' and the first digit is 0, then the second digit is automatically set to 1.
		 * The second entered digit will break the condition and the value will be set to 10-12.
		 */
		if (picker === '12hours') {
			if (flag && calculatedValue.slice(1, 2) === '1' && intKey === '0') return '0' + key;
		}

		return !flag ? '0' + key : calculatedValue.slice(1, 2) + key;
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key === 'Tab') return;

		e.preventDefault();

		if (e.key === 'ArrowRight') onRightFocus?.();
		if (e.key === 'ArrowLeft') onLeftFocus?.();

		if (['ArrowUp', 'ArrowDown'].includes(e.key)) {
			const step = e.key === 'ArrowUp' ? 1 : -1;
			const newValue = getArrowByType(calculatedValue, step, picker);

			if (flag) flag = false;

			const tempTime = time.copy();

			time = setDateByType(tempTime, newValue, picker, period);
			setTime?.(time);
		}
		if (e.key >= '0' && e.key <= '9') {
			if (picker === '12hours') intKey = e.key;

			const newValue = calculateNewValue(e.key);
			if (flag) onRightFocus?.();
			flag = !flag;

			const tempTime = time.copy();
			time = setDateByType(tempTime, newValue, picker, period);
			setTime?.(time);
		}
	}
</script>

<input
	bind:this={ref}
	id={id || picker}
	name={name || picker}
	class={cls(
		'w-[48px] text-center font-mono text-base tabular-nums caret-transparent focus:bg-primary focus:text-primary-foreground [&::-webkit-inner-spin-button]:appearance-none',
		className
	)}
	value={value || calculatedValue}
	onchange={(e) => {
		e.preventDefault();
		onchange?.(e);
	}}
	{type}
	inputmode="decimal"
	onkeydown={(e) => {
		onkeydown?.(e);
		handleKeyDown(e);
	}}
	{...restProps}
/>
