<script lang="ts">
	import { useAuthSessionState } from "$lib/auth.svelte";
	import * as Sidebar from "$components/ui/sidebar";
	import * as Avatar from "$components/ui/avatar";
	import UserAvatar from "$components/avatar/Avatar.svelte";
	import * as DropdownMenu from "$components/ui/dropdown-menu";

	import RiUserSettingsLine from "remixicon-svelte/icons/user-settings-line";
	import RiNotification2 from "remixicon-svelte/icons/notification-2-line";
	import RiLogoutBoxRLine from "remixicon-svelte/icons/logout-box-r-line";
	import RiExpandUpDownLine from "remixicon-svelte/icons/expand-up-down-line";

	const session = useAuthSessionState();
	const user = $derived(session.user);
</script>

{#snippet userMenuContent()}
	<DropdownMenu.Label class="p-0 font-normal">
		<div class="flex items-center gap-2 px-1 py-1.5 text-start text-sm">
			<Avatar.Root class="size-8 rounded-lg">
				<UserAvatar kind="user" id={session.user?.id || ""} size={32} />
			</Avatar.Root>
			<div class="grid flex-1 text-start text-sm leading-tight">
				<span class="truncate font-medium">{user?.attributes.name}</span>
				<span class="truncate text-xs">{user?.attributes.email}</span>
			</div>
		</div>
	</DropdownMenu.Label>
	
	<DropdownMenu.Separator />

	<DropdownMenu.Group>
		<DropdownMenu.Item>
			<RiUserSettingsLine />
			Preferences
		</DropdownMenu.Item>
		<DropdownMenu.Item>
			<RiNotification2 />
			Notifications
		</DropdownMenu.Item>
	</DropdownMenu.Group>

	<DropdownMenu.Separator />

	<a href="/api/auth/logout" class="cursor-pointer">
		<DropdownMenu.Item>
			<RiLogoutBoxRLine />
			Log out
		</DropdownMenu.Item>
	</a>
{/snippet}

<Sidebar.Menu>
	<Sidebar.MenuItem>
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Sidebar.MenuButton
						size="default"
						class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
						{...props}
					>
						<Avatar.Root class="size-5 [&>svg]:size-5 data-[state=open]:bg-white" loadingStatus="loaded">
							<UserAvatar kind="user" id={session.user?.id || ""} size={24} />
						</Avatar.Root>
						<span class="truncate font-medium">{user?.attributes.name}</span>
						<RiExpandUpDownLine class="ms-auto size-4" />
					</Sidebar.MenuButton>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content
				class="w-(--bits-dropdown-menu-anchor-width) min-w-48 rounded-lg"
				side="top"
				sideOffset={8}
				align="start"
			>
				{@render userMenuContent()}
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.MenuItem>
</Sidebar.Menu>

<!--Popover.Root bind:open>
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
			{#if session.isSetup}
			<Button href="/settings" onclick={() => {open = false}} variant="outline">
				Settings
			</Button>
			{/if}
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
</Popover.Root-->
