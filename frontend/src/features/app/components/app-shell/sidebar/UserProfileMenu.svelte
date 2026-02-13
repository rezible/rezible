<script lang="ts">
  	import { toggleMode } from "mode-watcher";
	import { mdiCog, mdiUnfoldMoreHorizontal, mdiSunAngle, mdiMoonWaningCrescent } from "@mdi/js";
	import { useAuthSessionState } from "$lib/auth.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { Button } from "$components/ui/button";

	const accountLinks = [
		{ href: "/settings", title: "Settings", icon: mdiCog },
	];
	let accountMenuOpen = $state(false);

	const session = useAuthSessionState();
</script>

{#if accountMenuOpen}
	<div class="border p-2">
		<Button onclick={toggleMode} variant="outline" size="icon">
			<Icon data={mdiSunAngle}
				classes={{root: "h-[1.2rem] w-[1.2rem] scale-100 rotate-0 !transition-all dark:scale-0 dark:-rotate-90"}}
			/>
			<Icon data={mdiMoonWaningCrescent}
				classes={{root: "absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 !transition-all dark:scale-100 dark:rotate-0"}}
			/>
			<span class="sr-only">Toggle theme</span>
		</Button>
	</div>
{/if}

<Button onclick={() => (accountMenuOpen = !accountMenuOpen)}> <!-- classes={{root: "w-full flex items-center gap-2 h-12 justify-start"}} -->
	<Avatar kind="user" id={session.user?.id || ""} size={24} />
	<div class="pl-3 flex flex-col items-start flex-1">
		<span class="text-surface-content">{session.user?.attributes.name}</span>
		<!-- <span class="text-surface-content/40">org</span> -->
	</div>
	<Icon data={mdiUnfoldMoreHorizontal} size={24} />
</Button>
