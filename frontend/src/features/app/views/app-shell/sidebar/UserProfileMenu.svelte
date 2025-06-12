<script lang="ts">
	import { Popover, Button, ThemeSwitch } from "svelte-ux";
	import { mdiAccount, mdiCog, mdiUnfoldMoreHorizontal } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { session } from "$lib/auth.svelte";

	const accountLinks = [
		{ href: "/profile", title: "Profile", icon: mdiAccount },
		{ href: "/settings", title: "Settings", icon: mdiCog },
	];
	let accountMenuOpen = $state(false);
</script>

<Popover bind:open={accountMenuOpen}>
	<div class="bg-surface-100 border shadow flex flex-col gap-1 p-2 items-center">
		{#each accountLinks as l}
			<Button href={l.href} icon={l.icon} classes={{root: "w-full"}}>
				{l.title}
			</Button>
		{/each}
		<div class="w-full flex justify-center">
			<ThemeSwitch />
		</div>
	</div>
</Popover>

<Button onclick={() => (accountMenuOpen = !accountMenuOpen)} classes={{root: "w-full flex items-center gap-2 h-12 justify-start font-normal"}}>
	<Avatar kind="user" id={session.user?.id || ""} size={24} />
	<div class="pl-3 flex flex-col items-start flex-1">
		<span>{session.username}</span>
		<!-- <span class="text-surface-content/40">org</span> -->
	</div>
	<Icon data={mdiUnfoldMoreHorizontal} size={24} />
</Button>
