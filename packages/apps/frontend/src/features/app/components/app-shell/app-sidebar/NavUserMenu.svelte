<script lang="ts">
	import { useUserSessionState } from "$src/lib/user-session.svelte";
	import * as Sidebar from "$components/ui/sidebar";
	import * as Avatar from "$components/ui/avatar";
	import UserAvatar from "$components/common/entity-avatar/EntityAvatar.svelte";
	import * as DropdownMenu from "$components/ui/dropdown-menu";
	import { useAppSidebarController } from "./controller.svelte";

	import RiUserSettingsLine from "remixicon-svelte/icons/user-settings-line";
	import RiLogoutBoxRLine from "remixicon-svelte/icons/logout-box-r-line";
	import RiArrowUpSLine from "remixicon-svelte/icons/arrow-drop-up-fill";
	import RiSettings3Line from "remixicon-svelte/icons/settings-3-line";

	const auth = useUserSessionState();
	const controller = useAppSidebarController();

	const user = $derived(auth.user);
	const org = $derived(auth.org);
	const userName = $derived(user?.attributes.name ?? "");
	const menuLabel = $derived(userName ? `Account menu for ${userName}` : "Account menu");
</script>

{#snippet userMenuContent()}
	<DropdownMenu.Label class="p-0 font-normal">
		<div class="flex items-center gap-2 px-1 py-1.5 text-start text-sm">
			<Avatar.Root class="size-7 rounded-lg">
				<UserAvatar kind="user" id={auth.user?.id || ""} size={28} />
			</Avatar.Root>
			<div class="grid flex-1 text-start text-sm leading-tight">
				<span class="text-muted-foreground text-xs">{org?.attributes.name}</span>
				<span class="truncate font-medium">{userName}</span>
				<span class="truncate text-xs">{user?.attributes.email}</span>
			</div>
		</div>
	</DropdownMenu.Label>

	<DropdownMenu.Separator />

	<DropdownMenu.Group>
		<DropdownMenu.Item>
			{#snippet child({ props })}
				<a {...props} href="/settings">
					<RiSettings3Line />
					Settings
				</a>
			{/snippet}
		</DropdownMenu.Item>
		<DropdownMenu.Item>
			{#snippet child({ props })}
				<a {...props} href="/settings/user">
					<RiUserSettingsLine />
					Preferences
				</a>
			{/snippet}
		</DropdownMenu.Item>
	</DropdownMenu.Group>

	<DropdownMenu.Separator />

	<DropdownMenu.Item onSelect={() => auth.logout()}>
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
						aria-label={menuLabel}
						class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
						{...props}
					>
						<Avatar.Root class="size-7 shrink-0 [&_svg]:size-7!" loadingStatus="loaded">
							<UserAvatar kind="user" id={auth.user?.id || ""} size={28} />
						</Avatar.Root>
						{#if controller.expanded}
							<span class="truncate font-medium">{userName}</span>
							<RiArrowUpSLine class="ms-auto size-2" />
						{/if}
					</Sidebar.MenuButton>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content class="min-w-56" side="right" align="end" sideOffset={8}>
				{@render userMenuContent()}
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.MenuItem>
</Sidebar.Menu>
