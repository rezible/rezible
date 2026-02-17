<script lang="ts">
	import type { ConfigComponentProps } from '../types';
	import * as Alert from "$components/ui/alert";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { Switch } from "$components/ui/switch";

	const { configured, onChange }: ConfigComponentProps = $props();
	const videoConferencePrefKey = "incident_video_conferences";

	const initialVideoConferenceEnabled = $derived.by(() => {
		const prefVal = configured?.attributes.preferences?.[videoConferencePrefKey];
		if (typeof prefVal === "boolean") return prefVal;
		if (typeof prefVal === "string") return prefVal !== "false";
		return true;
	});

	let serviceAccountInput = $state<HTMLInputElement>(null!);
	let dragActive = $state(false);
	let parseError = $state("");
	let selectedFileName = $state("");
	let videoConferenceEnabled = $state(true);
	let initializedPreference = $state(false);

	$effect(() => {
		if (initializedPreference) return;
		videoConferenceEnabled = initialVideoConferenceEnabled;
		initializedPreference = true;
	});

	const onVideoConferenceChange = (checked: boolean) => {
		videoConferenceEnabled = checked;
		onChange({ preferences: { [videoConferencePrefKey]: checked } });
	};

	const onFileInputChange = async () => {
		if (!serviceAccountInput.files || serviceAccountInput.files.length === 0) return;
		await loadServiceAccountFile(serviceAccountInput.files[0]);
	};

	const loadServiceAccountFile = async (file: File) => {
		parseError = "";
		selectedFileName = "";
		try {
			const fileData = await file.text();
			const parsed = JSON.parse(fileData);
			if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
				parseError = "Service account file must be a valid JSON object.";
				return;
			}

			selectedFileName = file.name;
			onChange({
				config: {
					UserConfig: {
						ServiceAccountCredentials: parsed,
					},
				},
			});
		} catch {
			parseError = "Could not parse JSON file. Check that this is a valid service account credentials file.";
		}
	};

	const onFileDrop = async (evt: DragEvent) => {
		evt.preventDefault();
		dragActive = false;

		const file = evt.dataTransfer?.files.item(0);
		if (!file) return;
		await loadServiceAccountFile(file);
	};
</script>

<div class="flex flex-col gap-3">
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

	<div class="space-y-2">
		<Label for="google-service-account-file">Service account credentials</Label>
		<div
			role="region"
			class="rounded-md border border-dashed p-4 text-sm transition-colors"
			class:border-primary={dragActive}
			class:bg-accent={dragActive}
			ondragover={(evt) => {
				evt.preventDefault();
				dragActive = true;
			}}
			ondragleave={() => {
				dragActive = false;
			}}
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
			{#if selectedFileName}
				<p class="mt-2 text-sm text-muted-foreground">Selected: {selectedFileName}</p>
			{/if}
		</div>
	</div>

	{#if parseError}
		<Alert.Root variant="destructive">
			<Alert.Title>Invalid credentials file</Alert.Title>
			<Alert.Description>{parseError}</Alert.Description>
		</Alert.Root>
	{/if}
</div>
