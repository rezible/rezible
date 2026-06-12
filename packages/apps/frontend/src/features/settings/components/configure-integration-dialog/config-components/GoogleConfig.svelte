<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";

	import { useConfigureIntegrationDialogController } from "../controller.svelte";
	import { watchOnce } from "runed";

	const ctrl = useConfigureIntegrationDialogController();

	let svcAccParseError = $state<string>();
	let svcAccFileName = $state<string>();

	const loadServiceAccountFile = async (file: File) => {
		svcAccParseError = undefined;
		svcAccFileName = undefined;
		try {
			const fileData = await file.text();
			const parsed = JSON.parse(fileData);
			if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
				svcAccParseError = "Service account file must be a valid JSON object.";
				return;
			}
			svcAccFileName = file.name;
			console.log("parsed", parsed);
		} catch {
			svcAccParseError = "Could not parse JSON file. Check that this is a valid service account credentials file.";
		}
	};

	watchOnce(() => ctrl.installation, inst => {
        const cfg = {
            displayName: "Google",
            config: {},
            preferences: {},
        };
        ctrl.setConfig(cfg, true);
		console.log("install", inst);
	});
</script>

<div class="flex flex-col gap-3">
	{#if !!ctrl.installation}
		<!--div class="flex items-center justify-between rounded-md border p-3">
			<div class="flex flex-col gap-1">
				<Label for="google-video-conference-toggle">Enable incident video conferences</Label>
				<p class="text-sm text-muted-foreground">Creates Google Meet links for incidents when supported.</p>
			</div>
			<Switch
				id="google-video-conference-toggle"
				checked={videoConferenceEnabled}
				onCheckedChange={onVideoConferenceChange}
			/>
		</div-->
	{:else}
		<div class="space-y-2">
			<Label for="google-service-account-file">Service account credentials</Label>
			<div
				role="region"
				class="rounded-md border border-dashed p-4 text-sm transition-colors [&.is-dragging]:border-primary [&.is-dragging]:bg-accent"
				ondragover={e => {
					e.preventDefault();
					e.currentTarget.classList.add("is-dragging");
				}}
				ondragleave={e => {
					e.preventDefault();
					e.currentTarget.classList.remove("is-dragging");
				}}
				ondrop={e => {
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
						onchange={e => {
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
	{/if}
</div>
