<script lang="ts">
	import { mdiCheck, mdiChevronLeft, mdiChevronRight, mdiClose } from "@mdi/js";
	import { PeriodType, getDateFuncsByPeriodType } from "@layerstack/utils";
	import { getDateRangePresets, type DateRange as DateRangeType } from "@layerstack/utils/dateRange";
	import { cls } from "@layerstack/tailwind";
	import { settings } from "$lib/settings.svelte";
	import { DateRange, DateRangeDisplay, DateRangeField, Dialog, Field, getComponentSettings } from "svelte-ux";
	import type { ComponentProps } from "svelte";
	import Button from "$components/button/Button.svelte";

	const _defaultValue: DateRangeType = {
		from: null,
		to: null,
		periodType: null,
	};

	const _defaultPeriodTypes: PeriodType[] = [
		PeriodType.Day,
		PeriodType.Week,
		PeriodType.BiWeek1,
		// PeriodType.BiWeek2,
		PeriodType.Month,
		PeriodType.Quarter,
		PeriodType.CalendarYear,
		PeriodType.FiscalYearOctober,
	];

	type Props = {
		value?: DateRangeType;
		onChange?: (v: DateRangeType) => void;
		onClear?: () => void;
		labelPlacement?: ComponentProps<Field>["labelPlacement"];
	} & ComponentProps<DateRangeField>;
	let {
		value = _defaultValue,
		onChange,
		onClear,
		stepper = false,
		center = false,
		periodTypes = _defaultPeriodTypes,
		getPeriodTypePresets = getDateRangePresets,
		disabledDates,
		classes = {},
		label = null,
		error = "",
		hint = "",
		disabled = false,
		clearable = false,
		base = false,
		rounded = false,
		dense = false,
		icon = null,
	}: Props = $props();
	
	const { format, localeSettings } = $derived(settings.uxSettings);
	const { classes: settingsClasses, defaults } = getComponentSettings("DateRangeField");

	let open = $state(false);

	let currentValue = $state(value);

	const onStepLeftClicked = () => {
		if (value && value.from && value.to && value.periodType) {
			const { difference, start, end, add } = getDateFuncsByPeriodType(
				$localeSettings,
				value.periodType
			);
			const offset = difference(value.from, value.to) - 1;
			value = {
				from: start(add(value.from, offset)),
				to: end(add(value.to, offset)),
				periodType: value.periodType,
			};
			onChange?.(value);
		}
	};

	const onStepRightClicked = () => {
		if (value && value.from && value.to && value.periodType) {
			const { difference, start, end, add } = getDateFuncsByPeriodType(
				$localeSettings,
				value.periodType
			);
			const offset = difference(value.to, value.from) + 1;
			value = {
				from: start(add(value.from, offset)),
				to: end(add(value.to, offset)),
				periodType: value.periodType,
			};
			onChange?.(value);
		}
	};

	const onClearClicked = () => {
		value = _defaultValue;
		onClear?.();
		onChange?.(value);
	}

	const onConfirmClicked = () => {
		open = false;
		value = currentValue;
		onChange?.(value);
	};

	const onCancelClicked = () => {
		open = false;
		currentValue = value;
	};
</script>

<Field
	label={label ?? (value.periodType ? $format.getPeriodTypeName(value.periodType) : "")}
	{icon}
	{error}
	{hint}
	{disabled}
	{base}
	{rounded}
	{dense}
	{center}
	classes={classes.field}
	let:id
>
	<span slot="prepend" class="flex items-center">
		<!-- <slot name="prepend" /> -->

		{#if stepper}
			<Button
				icon={mdiChevronLeft}
				class="p-2"
				on:click={onStepLeftClicked}
			/>
		{/if}
	</span>

	<button
		type="button"
		class={cls(
			"text-sm whitespace-nowrap w-full focus:outline-none",
			center ? "text-center" : "text-left"
		)}
		onclick={() => (open = true)}
		{id}
	>
		<DateRangeDisplay {value} />
	</button>

	<div slot="append" class="flex items-center">
		{#if clearable && (value?.periodType || value?.from || value?.to)}
			<Button
				icon={mdiClose}
				class="text-surface-content/50 p-1"
				on:click={onClearClicked}
			/>
		{/if}

		<!-- <slot name="append" /> -->

		{#if stepper}
			<Button
				icon={mdiChevronRight}
				class="p-2"
				on:click={onStepRightClicked}
			/>
		{/if}
	</div>
</Field>

<Dialog
	classes={{
		...classes.dialog,
		dialog: cls("max-h-[90vh] grid grid-rows-[auto,1fr,auto]", classes.dialog?.dialog),
	}}
	bind:open
>
	<div class="flex flex-col justify-center bg-primary text-primary-content px-6 h-24">
		<div class="text-sm opacity-50">
			{currentValue.periodType ? $format.getPeriodTypeName(currentValue.periodType) : ""}&nbsp;
		</div>
		<div class="text-xl sm:text-2xl">
			<DateRangeDisplay value={currentValue} />
		</div>
	</div>

	<div class="p-2 border-b overflow-auto">
		<DateRange
			bind:selected={currentValue}
			{periodTypes}
			{getPeriodTypePresets}
			{disabledDates}
			class="h-full"
		/>
	</div>

	<div slot="actions" class="flex items-center gap-2">
		<Button
			icon={mdiCheck}
			on:click={onConfirmClicked}
			color="primary"
			variant="fill"
		>
			{$localeSettings.dictionary.Ok}
		</Button>

		<Button on:click={onCancelClicked}>
			{$localeSettings.dictionary.Cancel}
		</Button>
	</div>
</Dialog>
