<script lang="ts">
	import { differenceInHours } from "date-fns";
	import { mdiArrowRight } from "@mdi/js";
	import { type OncallShift } from "$lib/api";
	import { Button, Icon, ProgressCircle } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { cls } from "@layerstack/tailwind";
	import Header from "$components/header/Header.svelte";

	type Props = {
		shift: OncallShift;
		isUser?: boolean;
	};
	const { shift, isUser }: Props = $props();

	const user = $derived(shift.attributes.user);
	const roster = $derived(shift.attributes.roster);

	const start = $derived(new Date(shift.attributes.startAt));
	const end = $derived(new Date(shift.attributes.endAt));
	const progress = $derived((100 * (Date.now() - start.valueOf())) / (end.valueOf() - start.valueOf()));
	const firstDay = $derived(differenceInHours(Date.now(), start) <= 24);
</script>

<div class={cls("max-w-lg p-2 border rounded-lg overflow-auto flex items-center", isUser ? "border-success-900/50 bg-success-900/5" : "")}>
	<Header>
		{#snippet avatar()}
			<ProgressCircle
				size={32}
				value={progress}
				track
				class={isUser ? "text-success [--track-color:theme(colors.success/10%)]" : "text-neutral [--track-color:theme(colors.neutral/50%)]"}
			/>
		{/snippet}
		{#snippet title()}
			<div>
				{#if isUser}
					<span class="text-md">You are Currently Oncall</span>
				{:else}
					<Button size="sm" href="/users/{user.id}" classes={{ root: "p-1 py-0 flex w-fit items-center" }}>
						<div class="self-center mr-1">
							<Avatar id={user.id} kind="user" size={16} />
						</div>
						<span class="font-bold text-base text-lg">{user.attributes.name}</span>
					</Button>
				{/if}
			</div>
		{/snippet}
		{#snippet subheading()}
			<span class="text-surface-content/70 inline-flex gap-1 items-center whitespace-pre">
				<span class="text-lg">{shift.attributes.role}</span>
				<span class="ml-1 text-lg">for</span>
				<Button size="sm" href="/oncall/rosters/{roster.attributes.slug}" classes={{ root: "p-1 py-0" }}>
					<span class="font-bold text-base text-lg">{roster.attributes.name}</span>
					<div class="self-center ml-1">
						<Avatar id={roster.id} kind="roster" size={16} />
					</div>
				</Button>
			</span>
		{/snippet}
		{#snippet actions()}
			<div class="flex flex-col gap-2 px-2 py-0 justify-end">
				<Button href="/oncall/shifts/{shift.id}" variant="fill" color={isUser ? "success" : "default"}>
					View
					<Icon data={mdiArrowRight} />
				</Button>
			</div>
		{/snippet}
	</Header>
</div>
