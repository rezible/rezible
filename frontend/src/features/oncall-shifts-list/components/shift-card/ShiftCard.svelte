<script lang="ts">
	import Avatar from "$src/components/avatar/Avatar.svelte";
	import Card from "$src/components/card/Card.svelte";
	import Header from "$src/components/header/Header.svelte";
	import type { OncallShift } from "$src/lib/api";
	import { formatDistanceToNowStrict, isFuture, isPast, formatDuration, minutesToHours, differenceInMinutes } from "date-fns";
	import ShiftProgressCircle from "./ShiftProgressCircle.svelte";
	import Icon from "$src/components/icon/Icon.svelte";
	import { mdiChevronRight } from "@mdi/js";
	import { cls } from "@layerstack/tailwind";
	import { settings } from "$src/lib/settings.svelte";
	import { PeriodType } from "@layerstack/utils";
	import { Button } from "svelte-ux";

	type Props = {
		shift: OncallShift;
		hideRoster?: boolean;
	};
	const {
		shift,
		hideRoster = false,
	}: Props = $props();

	const attr = $derived(shift.attributes);
	const roster = $derived(attr.roster);
	const user = $derived(attr.user);

	const start = $derived(new Date(attr.startAt));
	const end = $derived(new Date(attr.endAt));

	const isUpcoming = $derived(isFuture(start));
	const isFinished = $derived(isPast(end));
	const isActive = $derived(!isUpcoming && !isFinished);

	const formatShiftDuration = () => {
		const minutes = differenceInMinutes(end, start);
		if (minutes < 60) return `${minutes} minutes`;
		const hours = minutesToHours(minutes);
		const remainingMinutes = minutes - hours * 60;
		if (hours < 24)
			return formatDuration({ hours, minutes: remainingMinutes }, { format: ["hours", "minutes"] });
		const days = Math.floor(hours / 24);
		const remainingHours = hours - days * 24;
		return formatDuration(
			{ days, hours: remainingHours, minutes: remainingMinutes },
			{ format: ["days", "hours", "minutes"] }
		);
	};

	const defaultClasses = "";
	const activeClasses = "bg-success-900/20 border-success-100/10 hover:bg-success-900/30 hover:border-success-100/20";
</script>

<div class={cls("p-2 flex flex-col gap-2 rounded border", isActive ? activeClasses : defaultClasses)}>
	<div class="flex items-center justify-between border-b pb-1">
		<div class="flex flex-col gap-1">
			{#if !hideRoster}
				<a href="/rosters/{roster.id}" class="flex gap-2 items-center">
					<Avatar kind="roster" size={24} id={roster.id} />
					<span class="text-xl font-semibold">{roster.attributes.name}</span>
				</a>
			{/if}

			<div class="text-sm uppercase font-semibold text-surface-content/80">
				<span>{settings.format(start, PeriodType.Day)}</span>
				-
				<span>{settings.format(end, PeriodType.Day)}</span>
			</div>
		</div>

		<div class="flex flex-col items-end">
			<div class="flex gap-2 items-center">
				<span class={cls("text-sm uppercase font-bold text-surface-content/60", isActive && "text-success-600")}>
					{#if isActive}
						Active
					{:else if isUpcoming}
						Starts in {formatDistanceToNowStrict(end)}
					{:else}
						Ended {formatDistanceToNowStrict(end)} ago
					{/if}
				</span>

				{#if isActive}
					<div>
						<ShiftProgressCircle {shift} size={20} />
					</div>
				{/if}
			</div>
		</div>
	</div>

	<div class="flex flex-col">
		<a href="/users/{user.id}" class="flex gap-2 items-center w-fit">
			<Avatar kind="user" size={18} id={user.id} />
			<span class="font-semibold">{user.attributes.name ?? "user"}</span>
		</a>
		<span class="font-normal text-sm uppercase text-surface-content/60">{attr.role ?? "role"}</span>
	</div>

	<div class="flex justify-end">
		<Button href="/shifts/{shift.id}" classes={{root: "max-w-lg"}}>
			<span class="flex items-center group-hover:text-success">
				View
				<Icon data={mdiChevronRight} />
			</span>
		</Button>
	</div>
</div>