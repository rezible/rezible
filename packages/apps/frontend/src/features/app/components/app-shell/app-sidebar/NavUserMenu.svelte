<script lang="ts">
	import { useAuthSessionState } from "$src/lib/auth-session.svelte";
	import * as Sidebar from "$components/ui/sidebar";
	import * as Avatar from "$components/ui/avatar";
	import UserAvatar from "$components/avatar/Avatar.svelte";
	import * as DropdownMenu from "$components/ui/dropdown-menu";

	import RiUserSettingsLine from "remixicon-svelte/icons/user-settings-line";
	import RiNotification2 from "remixicon-svelte/icons/notification-2-line";
	import RiLogoutBoxRLine from "remixicon-svelte/icons/logout-box-r-line";
	import RiExpandUpDownLine from "remixicon-svelte/icons/expand-up-down-line";

	const auth = useAuthSessionState();
	const user = $derived(auth.user);
</script>

{#snippet userMenuContent()}
	<DropdownMenu.Label class="p-0 font-normal">
		<div class="flex items-center gap-2 px-1 py-1.5 text-start text-sm">
			<Avatar.Root class="size-7 rounded-lg">
				<UserAvatar kind="user" id={auth.user?.id || ""} size={28} />
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

	<DropdownMenu.Item onSelect={() => {auth.logout()}}>
		<RiLogoutBoxRLine />
		Log out
	</DropdownMenu.Item>
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
							<UserAvatar kind="user" id={auth.user?.id || ""} size={24} />
						</Avatar.Root>
						<span class="truncate font-medium">{user?.attributes.name}</span>
						<RiExpandUpDownLine class="ms-auto size-4" />
					</Sidebar.MenuButton>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content
				class="w-(--bits-dropdown-menu-anchor-width) min-w-48"
				side="top"
				sideOffset={8}
				align="start"
			>
				{@render userMenuContent()}
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.MenuItem>
</Sidebar.Menu>
