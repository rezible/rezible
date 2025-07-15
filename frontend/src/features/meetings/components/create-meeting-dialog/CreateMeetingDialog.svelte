<script lang="ts">
	import { createMutation } from "@tanstack/svelte-query";
	import { mdiArrowRight } from "@mdi/js";
	import {
		Button,
		Dialog,
		Field,
		NumberStepper,
		TextField,
		ToggleGroup,
		ToggleOption,
		DatePickerField,
	} from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { createMeetingScheduleMutation, createMeetingSessionMutation, type ErrorModel } from "$lib/api";
	import Header from "$components/header/Header.svelte";
	import ConfirmChangeButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import DateTimePickerField from "$components/date-time-field/DateTimePickerField.svelte";
	import {
		CreateMeetingFormSchema,
		getEmptyForm,
		type CreateMeetingFormData,
	} from "$features/meetings/lib/meetings";
	import UserPickerField from "./UserPickerField.svelte";
	import { Weekdays, type Weekday } from "$lib/scheduling";

	type Props = {
		open: boolean;
		onCreated: () => void;
	};
	let { open = $bindable(false), onCreated }: Props = $props();

	let formData = $state<CreateMeetingFormData>(getEmptyForm());
	const resetMeetingState = () => (formData = getEmptyForm());

	const toggleWeekday = (d: Weekday) => {
		if (!formData.weekDays.delete(d)) formData.weekDays.add(d);
		formData.weekDays = structuredClone(formData.weekDays);
	};

	const dayOfWeek = $derived(Weekdays[formData.start.toDate().getDay()].label);
	const daySelected = $derived<boolean[]>(Weekdays.map((v) => formData.weekDays.has(v.value)));
	const pluralSuffix = $derived(formData.repeats !== "daily" && formData.repetitionStep > 1 ? "s" : "");

	const parsedForm = $derived(open ? CreateMeetingFormSchema.safeParse(formData) : null);

	const onSuccess = () => {
		onCreated();
		resetMeetingState();
	};
	const onError = (resp: ErrorModel) => {
		const err = resp as Error;
		const model = JSON.parse(err.message) as ErrorModel;
		// TODO: handle this
		console.log(model);
	};
	const createScheduleMutation = createMutation(() => ({
		...createMeetingScheduleMutation(),
		onSuccess,
		onError,
	}));
	const createSessionMutation = createMutation(() => ({
		...createMeetingSessionMutation(),
		onSuccess,
		onError,
	}));

	const isPending = $derived(createScheduleMutation.isPending || createSessionMutation.isPending);
	// const error = $derived(createScheduleMutation.error || createSessionMutation.error);

	const tryCreateMeeting = () => {
		if (!parsedForm?.data) return;
		const { requestType, body } = parsedForm.data;
		if (requestType === "schedule") createScheduleMutation.mutate({ body });
		if (requestType === "session") createSessionMutation.mutate({ body });
	};

	const onUntilDateChange = (d: Date) => {
		console.log("onUntilDateChange", d);
		const newUntilDate = formData.untilDate.copy().set({
			day: d.getDate(), 
			month: d.getMonth(), 
			year: d.getFullYear(),
		});
		formData.untilDate = newUntilDate;
	}
</script>

<Dialog
	bind:open
	loading={isPending}
	persistent
	on:close={resetMeetingState}
	classes={{ root: "py-2", dialog: "max-h-full min-h-0 flex flex-col" }}
>
	<div slot="header" class="border-b p-2 flex justify-between items-center">
		<span class="text-xl flex-1">Create Meeting</span>
		<!--div class="">
			<Button variant="fill-light" color="secondary">
				<span class="flex gap-2 items-center">
					AI Draft
				</span>
			</Button>
		</div-->
	</div>

	<div class="flex flex-row gap-2 overflow-y-auto p-2 w-fit">
		<div class="w-fit flex flex-col gap-1 max-h-full">
			<!-- <Header title="Details" /> -->

			<TextField label="Title" bind:value={formData.name} />

			<TextField label="Description" bind:value={formData.description} />

			<DateTimePickerField
				label="Starting at"
				current={formData.start}
				onChange={(newStart) => {
					formData.start = newStart;
				}}
			/>

			<div class="flex flex-row gap-2 items-center">
				<Field label="Repeating">
					<ToggleGroup variant="outline" inset class="w-full" bind:value={formData.repeats}>
						<ToggleOption value="once">Once</ToggleOption>
						<ToggleOption value="daily">Daily</ToggleOption>
						<ToggleOption value="weekly">Weeky</ToggleOption>
						<ToggleOption value="monthly">Monthly</ToggleOption>
					</ToggleGroup>
				</Field>
				{#if formData.repeats !== "once"}
					<Icon data={mdiArrowRight} classes={{ root: "text-secondary" }} />
				{/if}
			</div>

			<UserPickerField value={[]} />
		</div>

		<div class="flex flex-col gap-1">
			{#if formData.repeats !== "once"}
				{@render repeatingFields()}
			{/if}
		</div>
	</div>

	<div slot="actions">
		<ConfirmChangeButtons
			loading={isPending}
			saveEnabled={parsedForm?.success}
			onConfirm={() => tryCreateMeeting()}
			onClose={() => {
				open = false;
			}}
		/>
	</div>
</Dialog>

{#snippet repeatingFields()}
	<Header title="Repeating" />

	{#if formData.repeats === "weekly" || formData.repeats === "monthly"}
		<Field label="Every" classes={{ input: "gap-2" }} dense>
			<NumberStepper
				min={1}
				max={formData.repeats === "weekly" ? 4 : 6}
				bind:value={formData.repetitionStep}
			/>
			{formData.repeats === "weekly" ? "Week" : "Month"}{pluralSuffix}
		</Field>
	{/if}

	{#if formData.repeats === "weekly"}
		<Field label="On Day(s)">
			<div class="flex gap-1">
				{#each Weekdays as day, i}
					<Button
						color={daySelected[i] ? "primary" : "default"}
						variant={daySelected[i] ? "fill" : "fill-light"}
						on:click={() => toggleWeekday(day.value)}
					>
						{day.value}
					</Button>
				{/each}
			</div>
		</Field>
	{:else if formData.repeats === "monthly"}
		<Field label="On the Same">
			<ToggleGroup variant="outline" inset class="w-full" bind:value={formData.monthlyOn}>
				<ToggleOption value="same_day">Day of the Month</ToggleOption>
				<ToggleOption value="same_weekday">Weekday ({dayOfWeek})</ToggleOption>
			</ToggleGroup>
		</Field>
	{/if}

	<Field label="Until">
		<ToggleGroup variant="outline" inset class="w-full" bind:value={formData.untilType}>
			<ToggleOption value="indefinite">Indefinitely</ToggleOption>
			<ToggleOption value="num_repetitions">Number of Repetitions</ToggleOption>
			<ToggleOption value="date">Date Reached</ToggleOption>
		</ToggleGroup>
	</Field>

	{#if formData.untilType === "num_repetitions"}
		<Field label="Number of Repetitions">
			<NumberStepper min={2} bind:value={formData.numRepetitions} />
		</Field>
	{:else if formData.untilType === "date"}
		<div>
			<DatePickerField 
				label="End after"
				value={formData.untilDate.toDate()}
				on:change={e => (onUntilDateChange(e.detail))}
			/>
		</div>
	{/if}
{/snippet}
