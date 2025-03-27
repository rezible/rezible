<script lang="ts">
	import { mdiChevronRight, mdiAccount } from "@mdi/js";
	import { Button, Card, Header, Icon } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import type { User } from "../types";
	import { session } from "$lib/auth.svelte";
	import TimezoneMap from "$components/viz/timezone-map/TimezoneMap.svelte";
	import { getLocalTimeZone } from "@internationalized/date";
	import { onMount, tick } from "svelte";

	const makeFakeUsers = (): User[] => {
		// todo
		return session.user ? [session.user] : [];
	}
	const users: User[] = makeFakeUsers();

	let showTimezone = $state(false);
	onMount(() => {
		setTimeout(() => (showTimezone = true), 10);
	});

	let hoveringTimezone = $state<string>();
</script>

<div class="flex gap-2 w-full h-full">
	<div class="flex flex-col gap-2 w-96">
		{#each users as usr}
		{@const userTz = getLocalTimeZone()}
		<Card classes={{root: "border-surface-content/20 bg-neutral/30"}} 
			onmouseenter={() => (hoveringTimezone = userTz)}
			onmouseleave={() => (hoveringTimezone = undefined)}
		>
			<div slot="header" class="flex items-center gap-2 w-full">
				<Header title={usr.attributes.name} classes={{ root: "w-full", title: "font-medium" }}>
					<div slot="avatar">
						<Avatar kind="user" size={32} id={usr.id} />
					</div>
					<div slot="actions" class="flex flex-col text-surface-content">
						<div class="">{userTz}</div>
					</div>
				</Header>
			</div>

			<div slot="contents" class="w-full p-2 border">
				info
			</div>

			<div slot="actions" class="flex-1 grid justify-items-end">
				<Button href="/users/{usr.id}">
					View
				</Button>
			</div>
		</Card>
		{:else}
			<div class="text-surface-600 italic p-2">No users assigned to this roster</div>
		{/each}
	</div>

	<div class="col-span-2">
		<div class="h-[420px] w-[862px] m-2">
			{#if showTimezone}
				<TimezoneMap highlightTz={hoveringTimezone} />
			{/if}
		</div>
	</div>
</div>