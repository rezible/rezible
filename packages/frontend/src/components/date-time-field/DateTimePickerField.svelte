<script lang="ts">
	import { CalendarDateTime, parseAbsolute, parseZonedDateTime, toZoned, ZonedDateTime } from "@internationalized/date";
	import { convertTime } from "./format.svelte";
	import { differenceInCalendarDays } from "date-fns/differenceInCalendarDays";
	import { isSameDay } from "date-fns";

	type Props = {
		name?: string;
		label: string;
		current: ZonedDateTime;
		exactTime?: boolean;
		rangeMin?: ZonedDateTime;
		rangeMax?: ZonedDateTime;
		onChange: (newValue: ZonedDateTime) => void;
	};
	let { name = "", label, current, exactTime, rangeMin, rangeMax, onChange }: Props = $props();

	let open = $state(false);

	let value = $state(convertTime(current));
	const currentVal = $derived(convertTime(current));

	const onConfirm = () => {
		const d = value.date;
		const t = value.time;
		const val = new CalendarDateTime(d.getFullYear(), d.getMonth() + 1, d.getDate(), t.hour, t.minute, t.second);
		const newValue = toZoned(val, value.timezone);
		// const valStr = `${value.date.getFullYear()}T${value.time}[${value.timezone}]`;
		// const newValue = parseAbsolute(valStr, value.timezone);
		// console.log(valStr, newValue.toDate());
		onChange(newValue);
		open = false;
	};

	const onClose = () => {
		value = convertTime(current);
		open = false;
	};

	const selectClasses =
		"py-2 px-3 block border-base-content rounded-lg text-md focus:border-accent focus:ring-accent-content dark:bg-neutral dark:border-base-100 dark:text-neutral-content dark:placeholder-base-content dark:focus:ring-neutral";

	const rangeMinDate = $derived(rangeMin?.toDate());
	const isMinDate = $derived(rangeMinDate && isSameDay(value.date, rangeMinDate));
	const rangeMaxDate = $derived(rangeMax?.toDate());
	const isMaxDate = $derived(rangeMaxDate && isSameDay(value.date, rangeMaxDate));

	const timeTooEarly = $derived(rangeMin && isMinDate && value.time.compare(rangeMin) < 0);
	const timeTooLate = $derived(rangeMax && isMaxDate && value.time.compare(rangeMax) > 0);
	const timeValid = $derived(!timeTooEarly && !timeTooLate);

	const disabledDates = (date: Date) => {
		if (rangeMinDate && differenceInCalendarDays(date, rangeMinDate) < 0) return true;
		if (rangeMaxDate && differenceInCalendarDays(date, rangeMaxDate) > 0) return true;
		return false;
	}
</script>

<!--div>
	<Field let:id icon={mdiCalendar} {label}>
		<span slot="prepend">
			<input type="hidden" {name} value={current} />
		</span>

		<button
			{id}
			type="button"
			class="text-sm min-h-[1.25rem] whitespace-nowrap w-full focus:outline-none"
			style="text-align: inherit"
			onclick={() => (open = true)}
		>
			{format.asTime(currentVal.time)}
			{format.asWeekday(currentVal.date)}
			{format.asCalendarDate(currentVal.date)}
		</button>
	</Field>

	<Dialog bind:open>
		{#if value.date}
			<div
				transition:slide
				class="flex flex-col justify-center bg-primary text-primary-content px-6 h-24"
			>
				<div class="text-sm opacity-50">{format.asWeekday(value.date)}</div>
				<div class="text-3xl">{format.asCalendarDate(value.date)}</div>
			</div>
		{/if}

		<div class="p-2 w-96">
			<DateSelect selected={value.date} periodType={PeriodType.Day} on:dateChange={(e) => {value.date = e.detail; console.log(e)}} {disabledDates} />

			<div class="flex items-center justify-center gap-2 border-t pt-2">
				{#if exactTime}
					<TimePicker bind:time={value.time} bind:period={value.period} />
				{:else}
					<select class={selectClasses} bind:value={value.time.hour}>
						{#each ["01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"] as h}
							{@const hourNum = Number.parseInt(h)}
							<option selected={value.time.hour === hourNum} value={hourNum}>{h}</option>
						{/each}
					</select>

					<span>:</span>

					<select class={selectClasses} bind:value={value.time.minute}>
						{#each ["00", "15", "30", "45"] as m}
							{@const minuteNum = Number.parseInt(m)}
							<option selected={value.time.minute === minuteNum} value={minuteNum}>{m}</option>
						{/each}
					</select>

					<select class={selectClasses} bind:value={value.period}>
						<option selected={value.period === "AM"} value="AM">AM</option>
						<option selected={value.period === "PM"} value="PM">PM</option>
					</select>
				{/if}
			</div>
		</div>

		<div slot="actions">
			<ConfirmChangeButtons {onConfirm} {onClose} saveEnabled={timeValid} />
		</div>
	</Dialog>
</div-->
