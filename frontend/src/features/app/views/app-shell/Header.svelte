<script lang="ts">
	import { Popover, Button, Badge, Icon, ListItem, ThemeSelect } from 'svelte-ux';
	import { mdiAccount, mdiAbTesting, mdiChevronDown, mdiCog, mdiInbox } from '@mdi/js';
	import { session, notifications } from '$lib/auth.svelte';
	import Avatar from '$components/avatar/Avatar.svelte';
    import { getToastState } from "$components/toaster";

    import OmniSearch from '$features/app/components/omni-search/OmniSearch.svelte';

	const toastState = getToastState();

	const accountLinks = [
		{ href: '/profile', title: 'Profile', icon: mdiAccount },
		{ href: '/preferences', title: 'Preferences', icon: mdiCog }
	];
	let accountMenuOpen = $state(false);
	let inboxOpen = $state(false);
	
	const inbox = $derived(notifications.inbox);
</script>

<div class="flex flex-wrap justify-between items-center h-16 px-4 border-b">
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
				<Button iconOnly onclick={() => (inboxOpen = !inboxOpen)} classes={{ root: 'p-2' }}>
					<Icon size={28} data={mdiInbox} />
				</Button>
			</Badge>
		</div>

		<div class="inline-block">
			<Popover bind:open={accountMenuOpen}>
				<div class="bg-surface-100 border shadow flex flex-col gap-1 p-2">
					{#each accountLinks as l}
						<a href={l.href} class="">
							<ListItem title={l.title} icon={l.icon} classes={{ root: 'hover:bg-accent' }} />
						</a>
					{/each}
					<Button on:click={() => toastState.add("test", "an example toast", (Math.random() > .25 ? mdiAbTesting : undefined), 1000000)}>
						add toast
					</Button>
					<ThemeSelect />
				</div>
			</Popover>
			<Button iconOnly onclick={() => (accountMenuOpen = !accountMenuOpen)}>
				<Avatar kind="user" id={session.user?.id || ''} />
				<Icon data={mdiChevronDown} />
			</Button>
		</div>
	</div>
</div>