<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import { Switch } from "$components/ui/switch";
	import { useUserSessionState } from "$lib/user-session.svelte";
	import { watch } from "runed";

	import { useIntegrationProviderConfigController } from "../controller.svelte";

	const ctrl = useIntegrationProviderConfigController();
	const session = useUserSessionState();

	const installation = $derived(ctrl.installations.length > 0 ? ctrl.installations[0] : undefined);
	const incidentManagementEnabled = $derived(Boolean(session.orgPreferences?.enableIncidentManagement));

	let svcAccParseError = $state<string>();
	let svcAccFileName = $state<string>();
	let videoConferenceEnabled = $state(Boolean(ctrl.userSettings.EnableVideoConference));

	const updateVideoConferenceSetting = (enabled: boolean) => {
		videoConferenceEnabled = enabled;
		ctrl.setUserSettings({ EnableVideoConference: enabled });
	};

	watch(
		() => installation?.attributes.userSettings?.EnableVideoConference,
		(enabled) => {
			videoConferenceEnabled = Boolean(enabled);
		}
	);

	const loadServiceAccountFile = async (file: File) => {
		svcAccParseError = undefined;
		svcAccFileName = undefined;
		ctrl.setInstallConfig({}, false);
		try {
			const fileData = await file.text();
			const parsed = JSON.parse(fileData);
			if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
				svcAccParseError = "Service account file must be a valid JSON object.";
				return;
			}
			svcAccFileName = file.name;
			ctrl.setInstallConfig({ ServiceAccountCredentials: parsed }, true);
		} catch {
			svcAccParseError =
				"Could not parse JSON file. Check that this is a valid service account credentials file.";
		}
	};

	const install = async () => {
		ctrl.setEditing("google");
		await ctrl.saveInstall();
	};

	const saveSettings = async () => {
		if (!installation) return;
		ctrl.setEditing("google", installation);
		ctrl.setUserSettings({ EnableVideoConference: videoConferenceEnabled });
		await ctrl.saveInstall();
	};
</script>

<div class="flex flex-col gap-3">
	<Card.Root>
		<Card.Header>
			<Card.Title>Google Workspace</Card.Title>
		</Card.Header>
		<Card.Content class="grid gap-3">
			{#if !!installation}
				<Alert.Root>
					<Alert.Title>Connected</Alert.Title>
					<Alert.Description>{installation.attributes.displayName}</Alert.Description>
				</Alert.Root>
				<div>
					<Button variant="destructive" onclick={() => ctrl.disconnect(installation.id)}
						>Disconnect</Button
					>
				</div>
			{:else}
				<div class="space-y-2">
					<Label for="google-service-account-file">Service account credentials</Label>
					<div
						role="region"
						class="rounded-md border border-dashed p-4 text-sm transition-colors [&.is-dragging]:border-primary [&.is-dragging]:bg-accent"
						ondragover={(e) => {
							e.preventDefault();
							e.currentTarget.classList.add("is-dragging");
						}}
						ondragleave={(e) => {
							e.preventDefault();
							e.currentTarget.classList.remove("is-dragging");
						}}
						ondrop={(e) => {
							e.preventDefault();
							e.currentTarget.classList.remove("is-dragging");
							const file = e.dataTransfer?.files.item(0);
							if (!!file) loadServiceAccountFile(file);
						}}
					>
						<p>Drag and drop JSON credentials here, or choose a file.</p>
						<div class="mt-3">
							<Input
								id="google-service-account-file"
								type="file"
								accept=".json,application/json"
								onchange={(e) => {
									e.preventDefault();
									const file = e.currentTarget.files?.item(0);
									if (!!file) loadServiceAccountFile(file);
								}}
							/>
						</div>
						{#if svcAccFileName}
							<p class="mt-2 text-sm text-muted-foreground">Selected: {svcAccFileName}</p>
						{/if}
					</div>
				</div>

				{#if svcAccParseError}
					<Alert.Root variant="destructive">
						<Alert.Title>Invalid credentials file</Alert.Title>
						<Alert.Description>{svcAccParseError}</Alert.Description>
					</Alert.Root>
				{/if}

				<Button class="w-fit" onclick={install} disabled={!ctrl.installConfigValid}>Install</Button>
			{/if}
		</Card.Content>
	</Card.Root>

	{#if installation && incidentManagementEnabled}
		<Card.Root>
			<Card.Header>
				<Card.Title>Video Conferencing</Card.Title>
			</Card.Header>
			<Card.Content class="grid gap-3">
				<div class="flex items-center justify-between gap-3">
					<Label for="google-video-conference-toggle">Enable Google Meet for incidents</Label>
					<Switch
						id="google-video-conference-toggle"
						checked={videoConferenceEnabled}
						onCheckedChange={(checked) => updateVideoConferenceSetting(checked)}
					/>
				</div>
				<Button class="w-fit" onclick={saveSettings}>Save settings</Button>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
