<script lang="ts">
  	import { toggleMode } from "mode-watcher";
	import { mdiCog, mdiUnfoldMoreHorizontal, mdiSunAngle, mdiMoonWaningCrescent, mdiDoorClosed, mdiDoor } from "@mdi/js";
	import { useAuthSessionState } from "$lib/auth.svelte";
	import Icon from "$components/icon/Icon.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { Button } from "$components/ui/button";
 	import * as Popover from "$components/ui/popover";
	import { Separator } from "$components/ui/separator";

	const accountLinks = [
		{ href: "/settings", title: "Settings", icon: mdiCog },
	];

	const session = useAuthSessionState();
	let open = $state(false);
</script>

<Popover.Root bind:open>
	<Popover.Trigger class="border-2 rounded-full border-accent p-0 cursor-pointer">
		<Avatar kind="user" id={session.user?.id || ""} size={32} />
	</Popover.Trigger>
	<Popover.Content class="mr-2 p-0">
		<div class="px-2 flex flex-col gap-2 pt-2">
			<span class="font-bold">{session.user?.attributes.name || "username"}</span>
			<span class="text-sm text-base-content/80">{session.user?.attributes.email || "email"}</span>
		</div>
		<Separator class="py-0"></Separator>
		<div class="flex flex-col gap-2 px-2 pt-0">
			<Button href="/settings" onclick={() => {open = false}} variant="outline">
				Settings
			</Button>
			<Button onclick={toggleMode} variant="outline">
				<span class="sr-only">Toggle theme</span>
				<Icon data={mdiSunAngle}
					classes={{root: "h-[1.2rem] w-[1.2rem] scale-100 rotate-0 !transition-all dark:scale-0 dark:-rotate-90"}}
				/>
				<Icon data={mdiMoonWaningCrescent}
					classes={{root: "absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 !transition-all dark:scale-100 dark:rotate-0"}}
				/>
			</Button>
		</div>
		<Separator class="py-0"></Separator>
		<div class="flex flex-col gap-2 p-2 pt-0">
			<Button onclick={() => {session.logout()}} variant="outline">
				Logout
				<Icon data={mdiDoor} />
			</Button>
		</div>
	</Popover.Content>
</Popover.Root>