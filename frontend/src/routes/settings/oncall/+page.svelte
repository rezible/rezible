<script lang="ts">
	import { Weekdays, type Weekday } from "$lib/scheduling";
	import { isEqualDay } from "@internationalized/date";
	import { addDays, isBefore, isEqual, isSameDay, subDays } from "date-fns";
	import { Button, Card, Field, Header, Month, NumberStepper, ToggleGroup, ToggleOption } from "svelte-ux";
	import { SvelteSet } from "svelte/reactivity";

	let repeats = $state<"weekly" | "monthly">("monthly");
	let repetitionStep = $state(1);
	const pluralSuffix = $derived(repetitionStep > 1 ? "s" : "");

	let monthlyOn = $state<"same_day" | "same_weekday">("same_day");

	const today = new Date(Date.now());
	const getWeekDay = (d: Date) => Weekdays[d.getDay()].value;

	let starting = $state<Date>(today);
	let selectedWeekday = $state<Weekday>(getWeekDay(today));

	const setStarting = (date: Date) => {
		starting = date;
		selectedWeekday = getWeekDay(date);
	};
</script>

<Card
	title="Oncall Time Report"
	subheading="Set up scheduled exporting of an oncall hours report"
	class="p-4 w-full"
	classes={{ headerContainer: "px-0 pt-0", content: "bg-surface-200" }}
>
	<div class="flex flex-col border p-2 rounded-lg">
		<Header title="Schedule" />

		<div class="flex flex-row gap-2">
			<div class="flex flex-col gap-2 w-72">
				<Field label="Starting">
					<div class="block w-full">
						<Month
							selected={starting} 
							on:dateChange={e => (setStarting(e.detail))} 
							disabledDates={(date) => (!isSameDay(date, today) && isBefore(date, today))} />
					</div>
				</Field>
			</div>

			<div class="flex flex-col gap-2 w-1/3">
				<Field label="Repeats">
					<ToggleGroup variant="fill" inset class="w-fit" bind:value={repeats}>
						<ToggleOption value="weekly">Weeky</ToggleOption>
						<ToggleOption value="monthly">Monthly</ToggleOption>
					</ToggleGroup>
				</Field>

				{#if repeats === "weekly" || repeats === "monthly"}
					<Field label="Every" classes={{ input: "gap-2 mb-2" }} dense>
						<NumberStepper
							min={1}
							max={repeats === "weekly" ? 4 : 6}
							bind:value={repetitionStep}
						/>
						{repeats === "weekly" ? "Week" : "Month"}{pluralSuffix}
					</Field>
				{/if}

				{#if repeats === "monthly"}
					<Field label="On the Same">
						<ToggleGroup variant="fill" inset class="w-fit" bind:value={monthlyOn}>
							<ToggleOption value="same_day">Day of the Month ({starting.getDate()})</ToggleOption>
							<ToggleOption value="same_weekday">Weekday ({selectedWeekday})</ToggleOption>
						</ToggleGroup>
					</Field>
				{/if}
			</div>
		</div>
	</div>

	<div slot="actions" class="pt-2 flex flex-row-reverse gap-2">
		<Button variant="fill" color="primary" disabled>Save Schedule</Button>

		<Button variant="fill" color="secondary">Export Now</Button>
	</div>
</Card>