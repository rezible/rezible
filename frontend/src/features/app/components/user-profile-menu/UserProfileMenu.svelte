<script lang="ts">
	import { Popover, Button, Icon, ListItem, ThemeSelect } from 'svelte-ux';
	import { mdiAccount, mdiAbTesting, mdiChevronDown, mdiCog } from '@mdi/js';
	import { session } from '$lib/auth.svelte';
	import Avatar from '$components/avatar/Avatar.svelte';
    import { getToastState } from "$components/toaster";

	const toastState = getToastState();

	const accountLinks = [
		{ href: '/profile', title: 'Profile', icon: mdiAccount },
		{ href: '/preferences', title: 'Preferences', icon: mdiCog }
	];
	let accountMenuOpen = $state(false);
</script>

<Popover bind:open={accountMenuOpen}>
	<div class="bg-surface-100 border shadow flex flex-col gap-1 p-2">
		<span>{session.username}</span>
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