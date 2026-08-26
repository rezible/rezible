<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import * as NativeSelect from "$components/ui/native-select";
	import { Switch } from "$components/ui/switch";
	import { watch } from "runed";

	import { useIntegrationProviderConfigController } from "../controller.svelte";

	const ctrl = useIntegrationProviderConfigController();

	const incidentsInstall = $derived(ctrl.installationsFor("slack_incidents").at(0));
	let incidentSettings = $state({
		AnnouncementChannelID: "",
		ChannelNamePattern: "incident-{slug}",
		AutoCreateVideoConference: false,
		InviteMode: "assigned_users",
	});

	watch(
		() => incidentsInstall?.attributes.userSettings?.Incidents,
		(settings) => {
			if (settings && typeof settings === "object") {
				incidentSettings = { ...incidentSettings, ...(settings as typeof incidentSettings) };
			}
		}
	);

	const saveIncidentSettings = () => {
		if (!incidentsInstall) return;
		ctrl.setEditing("slack_incidents", incidentsInstall);
		ctrl.setUserSettings({ Incidents: incidentSettings });
		ctrl.saveInstall();
	};
</script>

<Card.Root>
	<Card.Header>
		<Card.Title>Slack Incident Management</Card.Title>
	</Card.Header>
	<Card.Content class="grid gap-3">
		{#if incidentsInstall}
			<div class="flex flex-wrap items-center justify-between gap-2">
				<Alert.Root class="flex-1">
					<Alert.Title>Installed</Alert.Title>
					<Alert.Description>{incidentsInstall.attributes.displayName}</Alert.Description>
				</Alert.Root>
				<Button variant="destructive" onclick={() => ctrl.disconnect(incidentsInstall.id)}
					>Disconnect</Button
				>
			</div>
			<div class="grid gap-3 md:grid-cols-2">
				<div class="grid gap-1.5">
					<Label for="slack-announcement-channel">Announcement channel</Label>
					<Input
						id="slack-announcement-channel"
						bind:value={incidentSettings.AnnouncementChannelID}
					/>
				</div>
				<div class="grid gap-1.5">
					<Label for="slack-channel-pattern">Channel pattern</Label>
					<Input id="slack-channel-pattern" bind:value={incidentSettings.ChannelNamePattern} />
				</div>
				<div class="grid gap-1.5">
					<Label for="slack-invite-mode">Invite mode</Label>
					<NativeSelect.Root
						id="slack-invite-mode"
						bind:value={incidentSettings.InviteMode}
						class="w-full"
					>
						<NativeSelect.Option value="assigned_users">Assigned users</NativeSelect.Option>
						<NativeSelect.Option value="all_users">All users</NativeSelect.Option>
					</NativeSelect.Root>
				</div>
				<div class="flex items-center justify-between gap-3">
					<Label for="slack-auto-video">Auto-create video conference</Label>
					<Switch id="slack-auto-video" bind:checked={incidentSettings.AutoCreateVideoConference} />
				</div>
			</div>
			<Button class="w-fit" onclick={saveIncidentSettings}>Save settings</Button>
		{:else}
			<Button onclick={() => ctrl.startOAuthFlow("slack_incidents")} variant="outline">
				Install Slack Incident Management
			</Button>
		{/if}
	</Card.Content>
</Card.Root>
