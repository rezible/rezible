<script lang="ts">
	import { Button } from "$components/ui/button";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { getLocalTimeZone } from "@internationalized/date";
	import { createQuery } from "@tanstack/svelte-query";
	import { listUsersOptions } from "$lib/api";
	import Header from "$components/header/Header.svelte";
	import Card from "$components/card/Card.svelte";

	const usersQuery = createQuery(() => listUsersOptions());
	const users = $derived(usersQuery.data?.data ?? []);

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
				{#snippet header()}
					<Header title={usr.attributes.name} classes={{ root: "w-full", title: "font-medium" }}>
						{#snippet avatar()}
							<Avatar kind="user" size={32} id={usr.id} />
						{/snippet}
						{#snippet actions()}
							<div class="flex flex-col text-surface-content">
								<div class="">{userTz}</div>
							</div>
						{/snippet}
					</Header>
				{/snippet}

				{#snippet contents()}
					<div class="w-full p-2 border">
						info
					</div>
				{/snippet}

				{#snippet actions()}
					<div class="flex-1 grid justify-items-end">
						<Button href="/users/{usr.id}">
							View
						</Button>
					</div>
				{/snippet}
			</Card>
		{:else}
			<div class="text-surface-600 italic p-2">No users assigned to this roster</div>
		{/each}
	</div>

	<div class="col-span-2">
		<div class="h-[420px] w-[862px] m-2">
			timezone map
		</div>
	</div>
</div>