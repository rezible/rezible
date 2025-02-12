<script lang="ts">
	import { slide } from "svelte/transition";
	import { mdiCalendar } from "@mdi/js";
	import {
		Field,
		Dialog,
		DateSelect,
		DateToken,
		getSettings,
		PeriodType,
		NumberStepper,
		Input,
	} from "svelte-ux";
	import type { DateTimeAnchor } from "$lib/api";
	import ConfirmChangeButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import TimePicker from "$components/time-picker/TimePicker.svelte";
	import { Time } from "@internationalized/date";

	type Props = {
		name?: string;
		label: string;
		current: DateTimeAnchor;
		exactTime?: boolean;
		onChange: (newValue: DateTimeAnchor) => void;
	};
	let { name = "", label, current, exactTime, onChange }: Props = $props();

	const { format, localeSettings } = getSettings();
	// const dictionary = $derived($format.settings.dictionary);

	type InternalValue = {
		date: Date;
		time: Time;
		period: "AM" | "PM";
		timezone: string;
	}
	const getTimes = (a: DateTimeAnchor): InternalValue => {
		const timeParts = a.time.split(":");
		const hour = Number.parseInt(timeParts[0]);
		const minute = Number.parseInt(timeParts[1]);
		const seconds = Number.parseInt(timeParts[2]);
		return {
			date: new Date(a.date),
			time: new Time(hour, minute, seconds),
			period: hour >= 12 ? "PM" : "AM",
			timezone: a.timezone,
		};
	};

	let open = $state(false);
	const currentValue = $derived(getTimes(current));
	let value = $state(getTimes($state.snapshot(current)));

	const periodType = PeriodType.Day;
	const primaryFormat = [DateToken.Month_long, DateToken.DayOfMonth_withOrdinal, DateToken.Year_numeric];
	let secondaryFormat = DateToken.DayOfWeek_long;

	const pad = (n: number) => String(n).padStart(2, "0");

	const formatHourMinute = $derived(`${currentValue.time.hour}:${pad(currentValue.time.minute)}`);
	const formatTime = $derived(formatHourMinute + currentValue.period.toLowerCase());
	const formatDayOfWeek = $derived($format(currentValue.date, PeriodType.Day, { custom: secondaryFormat }));
	const formatDate = $derived($format(currentValue.date, PeriodType.Day, { custom: primaryFormat }));

	const convertHour24 = () => {
		const hour = value.time.hour === 12 ? 0 : value.time.hour;
		return value.period === "AM" ? hour : hour + 12;
	};

	const onConfirm = () => {
		const hour24 = convertHour24();
		const newValue = {
			date: value.date,
			time: `${pad(hour24)}:${pad(value.time.minute)}:${pad(value.time.second)}`,
			timezone: value.timezone,
		};
		onChange(newValue);
		open = false;
	};

	const onClose = () => {
		value = getTimes(current);
		open = false;
	};

	const selectClasses =
		"py-2 px-3 block border-base-content rounded-lg text-md focus:border-accent focus:ring-accent-content dark:bg-neutral dark:border-base-100 dark:text-neutral-content dark:placeholder-base-content dark:focus:ring-neutral";
</script>

<div>
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
			{formatTime}
			{formatDayOfWeek}
			{formatDate}
		</button>
	</Field>

	<Dialog bind:open>
		{#if value.date}
			<div
				transition:slide
				class="flex flex-col justify-center bg-primary text-primary-content px-6 h-24"
			>
				<div class="text-sm opacity-50">
					{$format(value.date, PeriodType.Day, {
						custom: secondaryFormat,
					})}
				</div>
				<div class="text-3xl">
					{$format(value.date, PeriodType.Day, {
						custom: primaryFormat,
					})}
				</div>
			</div>
		{/if}

		<div class="p-2 w-96">
			<DateSelect
				selected={value.date}
				{periodType}
				on:dateChange={(e) => (value.date = e.detail)}
			/>

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
			<ConfirmChangeButtons {onConfirm} {onClose} />
		</div>
	</Dialog>
</div>
