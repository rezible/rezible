<script lang="ts">
	import { Popover, Button, Badge, Icon } from "svelte-ux";
	import { mdiInbox } from "@mdi/js";
	import { notifications } from "$lib/auth.svelte";

	import OmniSearch from "$features/app/components/omni-search/OmniSearch.svelte";
	import UserProfileMenu from "$features/app/components/user-profile-menu/UserProfileMenu.svelte";

	let inboxOpen = $state(false);

	const inbox = $derived(notifications.inbox);
</script>

<div class="flex flex-wrap justify-between items-center h-16 px-4">
	<div class="flex-1 flex justify-around">
		<OmniSearch />
	</div>

	<div class="flex-0 flex items-center gap-3 h-16">
		<div class="inline-block">
			<Popover bind:open={inboxOpen}>
				<div class="bg-surface-100 border shadow flex flex-col gap-1 p-2">
					{#each inbox as item (item.id)}
						<span>{item.attributes.text}</span>
					{/each}
					{#if inbox.length === 0}
						<span>empty</span>
					{/if}
				</div>
			</Popover>
			<Badge value={inbox.length}>
				<Button iconOnly onclick={() => (inboxOpen = !inboxOpen)} classes={{ root: "p-2" }}>
					<Icon size={28} data={mdiInbox} />
				</Button>
			</Badge>
		</div>

		<div class="inline-block">
			<UserProfileMenu />
		</div>
	</div>
</div>
