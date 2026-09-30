<script lang="ts">
	import RiCloseLine from "remixicon-svelte/icons/close-line";
	import RiFileUploadLine from "remixicon-svelte/icons/file-upload-line";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";
	import * as Field from "$components/ui/field";
	import { Input } from "$components/ui/input";
	import { Spinner } from "$components/ui/spinner";
	import { cn } from "$lib/utils";

	import { untrack } from "svelte";

	import type { IntegrationInstallation } from "$lib/api";

	import { useIntegrationProviderController } from "../controller.svelte";
	import { GoogleInstallForm } from "./googleInstallForm.svelte";

	type Props = {
		// Replace this connection's service account key instead of installing a new connection.
		replacing?: IntegrationInstallation;
		onDone?: () => void;
		onCancel?: () => void;
	};
	const { replacing, onDone, onCancel }: Props = $props();

	const integrationName = "google";

	const ctrl = useIntegrationProviderController();
	// Each form is created for one connection; the page mounts a new form to replace another connection's key.
	const form = new GoogleInstallForm(untrack(() => replacing?.attributes.providerInstallationRef));

	const submitLabel = $derived(replacing ? "Replace key" : "Install");

	const pending = $derived(ctrl.installPendingName === integrationName);
	const installError = $derived(ctrl.installErrorFor(integrationName));
	const showCustomerIdError = $derived(form.customerIdTouched && !!form.customerIdError);

	let dragging = $state(false);
	let fileInput = $state<HTMLInputElement | null>(null);

	const onDrop = (event: DragEvent) => {
		event.preventDefault();
		dragging = false;

		const files = event.dataTransfer?.files;
		const file = files?.item(0) ?? undefined;
		if (!files || !file) return;

		// Show the dropped file in the input, as if it had been chosen there.
		if (fileInput) {
			fileInput.files = files;
		}
		form.selectFile(file);
	};

	const install = async () => {
		form.customerIdTouched = true;
		const config = form.installConfig();
		if (!config) return;

		const successMessage = replacing ? "Replaced the service account key." : undefined;
		const installed = await ctrl.install(integrationName, config, successMessage);
		if (installed) {
			form.reset();
			onDone?.();
		}
	};
</script>

<div class="flex flex-col gap-4">
	<Field.Group>
		<Field.Field data-invalid={showCustomerIdError}>
			<Field.Label for="google-customer-id">Workspace customer ID</Field.Label>
			<Input
				id="google-customer-id"
				placeholder="C0123abcd"
				bind:value={form.customerId}
				readonly={form.replacing}
				class={cn(form.replacing && "bg-muted text-muted-foreground")}
				onblur={() => {
					form.customerIdTouched = true;
				}}
				aria-invalid={showCustomerIdError}
			/>
			{#if form.replacing}
				<Field.Description>
					The new key replaces the current one. Settings are kept.
				</Field.Description>
			{:else}
				<Field.Description>
					Find it in the Google Admin console under Account settings.
				</Field.Description>
			{/if}
			{#if showCustomerIdError}
				<Field.Error>{form.customerIdError}</Field.Error>
			{/if}
		</Field.Field>

		<Field.Field data-invalid={!!form.fileError}>
			<Field.Label for="google-service-account-file">Service account key</Field.Label>
			<div
				role="region"
				aria-label="Service account key file drop area"
				class={cn(
					"flex flex-col gap-3 border border-dashed p-4 text-sm transition-colors",
					dragging && "border-primary bg-muted"
				)}
				ondragover={(event) => {
					event.preventDefault();
					dragging = true;
				}}
				ondragleave={() => {
					dragging = false;
				}}
				ondrop={onDrop}
			>
				<div class="flex items-center gap-2 text-muted-foreground">
					<RiFileUploadLine class="size-4 shrink-0" />
					<span>Drop the JSON key file here, or choose it.</span>
				</div>
				<div class="flex items-center gap-2">
					{#key form.fileInputKey}
						<Input
							bind:ref={fileInput}
							id="google-service-account-file"
							type="file"
							accept=".json,application/json"
							aria-invalid={!!form.fileError}
							onchange={(event) =>
								form.selectFile(event.currentTarget.files?.item(0) ?? undefined)}
						/>
					{/key}
					{#if form.fileName}
						<Button
							size="icon"
							variant="ghost"
							aria-label="Remove key file"
							onclick={form.clearFile}
						>
							<RiCloseLine />
						</Button>
					{/if}
				</div>
			</div>
			<Field.Description>
				The service account needs domain-wide delegation to create Google Meet meetings.
			</Field.Description>
			{#if form.fileError}
				<Field.Error>{form.fileError}</Field.Error>
			{/if}
		</Field.Field>
	</Field.Group>

	{#if installError}
		<InlineAlert error={installError} dismissable={false} />
	{/if}

	<div class="flex flex-wrap gap-2">
		<Button disabled={!form.canSubmit || pending} onclick={install}>
			{#if pending}
				<Spinner />
			{/if}
			{submitLabel}
		</Button>
		{#if onCancel}
			<Button variant="outline" disabled={pending} onclick={onCancel}>Cancel</Button>
		{/if}
	</div>
</div>
