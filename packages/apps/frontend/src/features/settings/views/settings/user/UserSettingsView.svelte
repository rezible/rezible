<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import LoadingIndicator from "$components/layout/loading-indicator/LoadingIndicator.svelte";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import * as NativeSelect from "$components/ui/native-select";
	import { Switch } from "$components/ui/switch";
	import { resetMode, setMode, userPrefersMode } from "mode-watcher";
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initUserSettingsController } from "./controller.svelte";

	const view = initUserSettingsController();

	const notificationRows = [
		{ key: "incidentUpdates", label: "Incident updates" },
		{ key: "incidentRoleAssignments", label: "Incident role assignments" },
		{ key: "agentRunResults", label: "Agent run results" },
		{ key: "integrationSyncFailures", label: "Integration sync failures" },
	] as const;

	registerPageDescriptor(() => ({
		title: "User",
		parents: [{ label: "Settings", path: resolve("/settings") }],
	}));
</script>

<div class="flex max-w-3xl flex-col gap-4">
	{#if view.loading}
		<div class="flex items-center gap-2 text-sm text-muted-foreground">
			<LoadingIndicator />
			<span>Loading settings...</span>
		</div>
	{:else if view.error}
		<InlineAlert error={view.error} />
	{:else}
		{#if view.saveError}
			<InlineAlert error={view.saveError} />
		{/if}

		<Card.Root>
			<Card.Header>
				<Card.Title>Profile</Card.Title>
			</Card.Header>
			<Card.Content class="grid gap-3">
				<div class="grid gap-1.5">
					<Label for="user-name">Name</Label>
					<Input id="user-name" bind:value={view.name} />
				</div>
				<div class="grid gap-1.5">
					<Label for="user-email">Email</Label>
					<Input id="user-email" value={view.email} disabled />
				</div>
				<div class="grid gap-1.5">
					<Label for="user-timezone">Timezone</Label>
					<Input id="user-timezone" bind:value={view.timezone} placeholder="Australia/Sydney" />
				</div>
			</Card.Content>
			<Card.Footer>
				<Button onclick={() => view.saveProfile()} disabled={view.saving}>Save profile</Button>
			</Card.Footer>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title>Appearance</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="grid max-w-xs gap-1.5">
					<Label for="user-theme">Theme</Label>
					<NativeSelect.Root
						id="user-theme"
						value={userPrefersMode.current}
						onchange={(event) => {
							const value = event.currentTarget.value;
							if (value === "system") resetMode();
							else if (value === "light" || value === "dark") setMode(value);
						}}
					>
						<NativeSelect.Option value="system">System</NativeSelect.Option>
						<NativeSelect.Option value="light">Light</NativeSelect.Option>
						<NativeSelect.Option value="dark">Dark</NativeSelect.Option>
					</NativeSelect.Root>
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title>Notifications</Card.Title>
			</Card.Header>
			<Card.Content class="divide-y">
				{#each notificationRows as row (row.key)}
					<div class="flex items-center justify-between gap-4 py-3 first:pt-0 last:pb-0">
						<Label for={`notification-${row.key}`}>{row.label}</Label>
						<Switch
							id={`notification-${row.key}`}
							checked={view.notifications[row.key]}
							onCheckedChange={(checked) => view.setNotification(row.key, checked)}
						/>
					</div>
				{/each}
			</Card.Content>
			<Card.Footer>
				<Button onclick={() => view.saveNotifications()} disabled={view.saving}
					>Save notifications</Button
				>
			</Card.Footer>
		</Card.Root>
	{/if}
</div>
