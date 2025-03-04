<script lang="ts" module>
	import type { Time as TimeType } from "@internationalized/date";
	import type { Period } from "./utils";
	import type { HTMLButtonAttributes } from "svelte/elements";

	export type PeriodSelectorProps = HTMLButtonAttributes & {
		period: Period;
		setPeriod?: (period: PeriodSelectorProps["period"]) => void;

		time: TimeType;
		setTime?: (time: TimeType) => void;

		onRightFocus?: () => void;
		onLeftFocus?: () => void;

		ref: HTMLElement | null;
	};
</script>

<script lang="ts">
	import { display12HourValue, setDateByType } from "./utils";
	import { onMount } from "svelte";

	let {
		period = $bindable(),
		time = $bindable(),
		ref,

		onLeftFocus,
		onRightFocus,
		setPeriod,
		setTime,
	}: PeriodSelectorProps = $props();

	const handlePeriod = () => {
		const tempTime = time.copy();
		const hours = display12HourValue(time.hour);
		const _time = setDateByType(tempTime, hours.toString(), "12hours", period);

		time = _time;
		setTime?.(_time);
	};

	const handleKeyDown = (e: KeyboardEvent) => {
		if (e.key === "ArrowRight") onRightFocus?.();
		if (e.key === "ArrowLeft") onLeftFocus?.();
	};

	function handleChange(value: Period) {
		period = value;
		setPeriod?.(value);
		if (time) handlePeriod();
	}
</script>

<select
	bind:value={period}
	onkeydown={handleKeyDown}
	onchange={(e) => {
		handleChange(e.currentTarget.value as Period);
	}}
	class="w-[48px] h-6 text-center font-mono text-base tabular-nums caret-transparent bg-surface-200 focus:bg-primary focus:text-primary-content [&::-webkit-inner-spin-button]:appearance-none"
>
	<option selected={period === "AM"} value="AM">AM</option>
	<option selected={period === "PM"} value="PM">PM</option>
</select>
