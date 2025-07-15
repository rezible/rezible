<script lang="ts">
	import { Button } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { formatDistanceToNow } from "date-fns";
	import { rosterViewCtx } from "../viewState.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import {
		getUserOncallInformationOptions,
		type OncallShift,
	} from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { cls } from "@layerstack/tailwind";
	import { parseAbsoluteToLocal } from "@internationalized/date";
	import Header from "$src/components/header/Header.svelte";
	import { mdiArrowRight } from "@mdi/js";

	const viewCtx = rosterViewCtx.get();
	const rosterId = $derived(viewCtx.rosterId);

	// TODO: use correct query
	const shiftsQuery = createQuery(() => getUserOncallInformationOptions({ query: {} }));
	const shifts = $derived(shiftsQuery.data?.data);
	const prevShift = $derived(shifts?.pastShifts.at(0));
	const activeShift = $derived(shifts?.activeShifts.at(0));
	const nextShift = $derived(shifts?.upcomingShifts.at(0));
</script>

<div class="flex flex-col h-full border border-surface-content/10 rounded">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Shifts" classes={{root: "", title: "text-xl"}}>
			{#snippet actions()}
				<Button variant="fill-light" href={`/rosters/${rosterId}/shifts`}>
					View All
					<Icon data={mdiArrowRight} classes={{root: "ml-1 h-4 w-4"}} />
				</Button>
			{/snippet}
		</Header>
	</div>

	<div class="flex-1 flex flex-col px-0 overflow-y-auto">
		{#snippet shiftCard(shift: OncallShift, status: "previous" | "active" | "next")}
			{@const user = shift.attributes.user}
			{@const isActive = status === "active"}
			<div
				class={cls(
					"border-y p-4 flex items-center justify-between gap-2 min-w-72",
					isActive
						? "border-success-900 bg-success-900/10"
						: "border-neutral-content/10 bg-neutral-900/30"
				)}
			>
				<div class="flex flex-col gap-2">
					<div class="flex items-center gap-2">
						<Avatar id={user.id} kind="user" size={30} />
						<span class="text-lg">{user.attributes.name}</span>
					</div>

					{#if status === "previous"}
						<span
							>Ended {formatDistanceToNow(
								parseAbsoluteToLocal(shift.attributes.endAt).toDate()
							)} ago</span
						>
					{:else if status === "active"}
						<span>Active Now</span>
					{:else}
						<span
							>Starts in {formatDistanceToNow(
								parseAbsoluteToLocal(shift.attributes.startAt).toDate()
							)}</span
						>
					{/if}
				</div>

				<div class="flex self-center">
					<Button
						variant="fill"
						color={isActive ? "success" : "neutral"}
						href={`/shifts/${shift.id}`}
					>
						View
					</Button>
				</div>
			</div>
		{/snippet}
		
		{#if nextShift}{@render shiftCard(nextShift, "next")}{/if}
		{#if activeShift}{@render shiftCard(activeShift, "active")}{/if}
		{#if prevShift}{@render shiftCard(prevShift, "previous")}{/if}
	</div>
</div>