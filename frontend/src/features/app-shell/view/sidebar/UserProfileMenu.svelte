<script lang="ts">
	import { Popover, ThemeSwitch } from "svelte-ux";
	import { mdiCog, mdiUnfoldMoreHorizontal } from "@mdi/js";
	import { session } from "$lib/auth.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import Button from "$components/button/Button.svelte";

	const accountLinks = [
		{ href: "/settings", title: "Settings", icon: mdiCog },
	];
	let accountMenuOpen = $state(false);
</script>

<Popover bind:open={accountMenuOpen}>
	<div class="bg-surface-100 border shadow flex flex-col gap-2 p-2 items-center min-w-32">
		{#each accountLinks as l}
			<Button href={l.href} classes={{root: "w-full justify-between px-1"}}>
				{l.title}
				<Icon data={l.icon} size={24} />
			</Button>
		{/each}
		<div class="w-full flex justify-center">
			<ThemeSwitch />
		</div>
	</div>
</Popover>

<Button onclick={() => (accountMenuOpen = !accountMenuOpen)} classes={{root: "w-full flex items-center gap-2 h-12 justify-start"}}>
	<Avatar kind="user" id={session.user?.id || ""} size={24} />
	<div class="pl-3 flex flex-col items-start flex-1">
		<span class="text-surface-content">{session.username}</span>
		<!-- <span class="text-surface-content/40">org</span> -->
	</div>
	<Icon data={mdiUnfoldMoreHorizontal} size={24} />
</Button>
