<script lang="ts">
	import type { ConfigComponentProps } from '../types';
	import * as Alert from "$components/ui/alert";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { Switch } from "$components/ui/switch";

	const { configured, onConfigChange, onPreferencesChange }: ConfigComponentProps = $props();
	const videoConferencingPrefKey = "video_conferencing";

	let videoConferenceEnabled = $state(true);

	const onVideoConferenceChange = (checked: boolean) => {
		videoConferenceEnabled = checked;
		onPreferencesChange({[videoConferencingPrefKey]: checked});
	};

	let serviceAccountInput = $state<HTMLInputElement>(null!);
	let svcAccDragActive = $state(false);
	let svcAccParseError = $state("");
	let svcAccFileName = $state("");

	const onFileInputChange = async () => {
		if (!serviceAccountInput.files || serviceAccountInput.files.length === 0) return;
		await loadServiceAccountFile(serviceAccountInput.files[0]);
	};

	const onFileDrop = async (evt: DragEvent) => {
		evt.preventDefault();
		svcAccDragActive = false;

		const file = evt.dataTransfer?.files.item(0);
		if (file) await loadServiceAccountFile(file);
	};

	const loadServiceAccountFile = async (file: File) => {
		svcAccParseError = "";
		svcAccFileName = "";
		try {
			const fileData = await file.text();
			const parsed = JSON.parse(fileData);
			if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
				svcAccParseError = "Service account file must be a valid JSON object.";
				return;
			}
			svcAccFileName = file.name;
			onConfigChange({"ServiceAccountCredentials": parsed});
		} catch {
			svcAccParseError = "Could not parse JSON file. Check that this is a valid service account credentials file.";
		}
	};
</script>

<div class="flex flex-col gap-3">
	{#if configured}
		<div class="flex items-center justify-between rounded-md border p-3">
			<div class="flex flex-col gap-1">
				<Label for="google-video-conference-toggle">Enable incident video conferences</Label>
				<p class="text-sm text-muted-foreground">Creates Google Meet links for incidents when supported.</p>
			</div>
			<Switch
				id="google-video-conference-toggle"
				checked={videoConferenceEnabled}
				onCheckedChange={onVideoConferenceChange}
			/>
		</div>
	{:else}
		<div class="space-y-2">
			<Label for="google-service-account-file">Service account credentials</Label>
			<div
				role="region"
				class="rounded-md border border-dashed p-4 text-sm transition-colors"
				class:border-primary={svcAccDragActive}
				class:bg-accent={svcAccDragActive}
				ondragover={(evt) => {
					evt.preventDefault();
					svcAccDragActive = true;
				}}
				ondragleave={() => {svcAccDragActive = false}}
				ondrop={onFileDrop}
			>
				<p>Drag and drop JSON credentials here, or choose a file.</p>
				<div class="mt-3">
					<Input
						id="google-service-account-file"
						bind:ref={serviceAccountInput}
						type="file"
						accept=".json,application/json"
						onchange={onFileInputChange}
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
