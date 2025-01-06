<script lang="ts">
	import { slide } from 'svelte/transition';
	import { mdiCalendar } from '@mdi/js';
	import { Field, Dialog, DateSelect, DateToken, getSettings, PeriodType, NumberStepper, Input } from 'svelte-ux';
	import type { DateTimeAnchor } from "$lib/api";
    import ConfirmChangeButtons from '$components/confirm-buttons/ConfirmButtons.svelte';

	type Props = {
		name?: string;
		label: string;
		current: DateTimeAnchor;
		exactTime?: boolean;
		onChange: (newValue: DateTimeAnchor) => void;
	}
	let {
		name = "",
		label,
		current,
		exactTime,
		onChange,
	}: Props = $props();

	const { format, localeSettings } = getSettings();
	// const dictionary = $derived($format.settings.dictionary);

	const getTimes = (a: DateTimeAnchor) => {
		const timeParts = a.time.split(":");
		const hour = Number.parseInt(timeParts[0]);
		const minute = Number.parseInt(timeParts[1]);
		const seconds = Number.parseInt(timeParts[2]);
		return {
			date: new Date(a.date),
			hour, 
			minute,
			seconds,
			amPm: hour >= 12 ? "pm" : "am",
			timezone: a.timezone,
		}
	}

	let open = $state(false);
	let value = $state(getTimes(current));

	const periodType = PeriodType.Day;
	const primaryFormat = [
		DateToken.Month_long,
		DateToken.DayOfMonth_withOrdinal,
		DateToken.Year_numeric,
	];
	let secondaryFormat = DateToken.DayOfWeek_long;

	const pad = (n: number) => String(n).padStart(2, "0");

	const formatHourMinute = $derived(`${value.hour}:${pad(value.minute)}`)
	const formatTime = $derived(formatHourMinute + value.amPm);
	const formatDayOfWeek = $derived($format(value.date, PeriodType.Day, { custom: secondaryFormat }));
	const formatDate = $derived($format(value.date, PeriodType.Day, { custom: primaryFormat }));

	const convertHour24 = () => {
		const hour = value.hour === 12 ? 0 : value.hour;
		return value.amPm === "am" ? hour : hour + 12;
	}
	const onConfirm = () => {
		const hour24 = convertHour24()
		const newValue = {
			date: value.date,// `${value.date.getFullYear()}-${pad(value.date.getMonth() + 1)}-${pad(value.date.getDate())}`,
			time: `${pad(hour24)}:${pad(value.minute)}:${pad(value.seconds)}`,
			timezone: value.timezone,
		};
		onChange(newValue);
		open = false;
	}

	const onClose = () => {
		value = getTimes(current);
		open = false;
	}

	const selectClasses = "py-2 px-3 block border-base-content rounded-lg text-md focus:border-accent focus:ring-accent-content dark:bg-neutral dark:border-base-100 dark:text-neutral-content dark:placeholder-base-content dark:focus:ring-neutral"
</script>

<div>
	<Field let:id icon={mdiCalendar} {label}>
		<span slot="prepend">
			<input type="hidden" {name} value={current} />
		</span>

		<button {id}
			type="button"
			class="text-sm min-h-[1.25rem] whitespace-nowrap w-full focus:outline-none"
			style="text-align: inherit"
			onclick={() => (open = true)}>
			{formatTime} {formatDayOfWeek} {formatDate}
		</button>
	</Field>
	
	<Dialog bind:open>
		{#if value.date}
			<div transition:slide
				class="flex flex-col justify-center bg-primary text-primary-content px-6 h-24">
				<div class="text-sm opacity-50">
					{$format(value.date, PeriodType.Day, { custom: secondaryFormat })}
				</div>
				<div class="text-3xl">
					{$format(value.date, PeriodType.Day, { custom: primaryFormat })}
				</div>
			</div>
		{/if}

		<div class="p-2 w-96">
			<DateSelect
				bind:selected={value.date}
				{periodType}
				on:dateChange={e => (value.date = e.detail)}
			/>

			<div class="flex items-center justify-center gap-2 border-t pt-2">
				{#if exactTime}
					<!-- https://time-picker.nouro.app/ -->
					<Input mask="hhmmss" replace="hms" on:change={e => {console.log(e.detail.value)}} class="text-md w-8 border text-center" />
				{:else}
					<select class={selectClasses} bind:value={value.hour}>
						{#each ["01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"] as h}
							{@const hourNum = Number.parseInt(h)}
							<option selected={value.hour === hourNum} value={hourNum}>{h}</option>
						{/each}
					</select>

					<span>:</span>

					<select class={selectClasses} bind:value={value.minute}>
						{#each ["00", "15", "30", "45"] as m}
							{@const minuteNum = Number.parseInt(m)}
							<option selected={value.minute === minuteNum} value={minuteNum}>{m}</option>
						{/each}
					</select>
				{/if}

				<select class={selectClasses} bind:value={value.amPm}>
					<option selected={value.amPm==="am"} value="am">AM</option>
					<option selected={value.amPm==="pm"} value="pm">PM</option>
				</select>
			</div>
		</div>

		<div slot="actions">
			<ConfirmChangeButtons {onConfirm} {onClose} />
		</div>
	</Dialog>
</div>