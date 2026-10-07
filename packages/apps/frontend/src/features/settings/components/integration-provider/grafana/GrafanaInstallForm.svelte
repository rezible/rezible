<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";
	import * as Field from "$components/ui/field";
	import { Input } from "$components/ui/input";
	import { Spinner } from "$components/ui/spinner";
	import { cn } from "$lib/utils";

	import { untrack } from "svelte";

	import type { IntegrationInstallation } from "$lib/api";

	import { useIntegrationProviderController } from "../controller.svelte";
	import { GrafanaInstallForm, grafanaIntegrationName } from "./grafanaInstallForm.svelte";

	type Props = {
		// Replace this installation's token instead of installing a new one.
		replacing?: IntegrationInstallation;
		onDone?: () => void;
		onCancel?: () => void;
	};
	const { replacing, onDone, onCancel }: Props = $props();

	const ctrl = useIntegrationProviderController();
	// Each form is created for one installation; the page mounts a new form to replace a token.
	const form = new GrafanaInstallForm(
		ctrl,
		untrack(() => replacing?.attributes.providerInstallationRef)
	);

	const submitLabel = $derived(replacing ? "Replace token" : "Install");

	const pending = $derived(ctrl.installPendingName === grafanaIntegrationName);
	const installError = $derived(ctrl.installErrorFor(grafanaIntegrationName));
	const showUrlError = $derived(form.urlTouched && !!form.urlError);
	const showTokenError = $derived(form.tokenTouched && !!form.tokenError);

	const install = async () => {
		const installed = await form.install();
		if (installed) {
			onDone?.();
		}
	};
</script>

<div class="flex flex-col gap-4">
	<Field.Group>
		<Field.Field data-invalid={showUrlError}>
			<Field.Label for="grafana-url">Grafana URL</Field.Label>
			<Input
				id="grafana-url"
				placeholder="https://grafana.example.com"
				bind:value={form.url}
				readonly={form.replacing}
				class={cn(form.replacing && "bg-muted text-muted-foreground")}
				onblur={() => {
					form.urlTouched = true;
				}}
				aria-invalid={showUrlError}
			/>
			{#if form.replacing}
				<Field.Description>
					The new token replaces the current one. Settings are kept.
				</Field.Description>
			{:else}
				<Field.Description>The address Rezible uses to reach Grafana.</Field.Description>
			{/if}
			{#if showUrlError}
				<Field.Error>{form.urlError}</Field.Error>
			{/if}
		</Field.Field>

		<Field.Field data-invalid={showTokenError}>
			<Field.Label for="grafana-token">Service account token</Field.Label>
			<Input
				id="grafana-token"
				type="password"
				autocomplete="off"
				bind:value={form.token}
				onblur={() => {
					form.tokenTouched = true;
				}}
				aria-invalid={showTokenError}
			/>
			<Field.Description>
				Create the token for a service account with the Viewer role. Rezible only reads from Grafana.
			</Field.Description>
			{#if showTokenError}
				<Field.Error>{form.tokenError}</Field.Error>
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
