<script lang="ts">
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import RiFileCopyLine from "remixicon-svelte/icons/file-copy-line";
	import RiWebhookLine from "remixicon-svelte/icons/webhook-line";

	import type { IntegrationInstallation } from "$lib/api";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import * as Empty from "$components/ui/empty";
	import * as Field from "$components/ui/field";
	import * as Select from "$components/ui/select";
	import { Spinner } from "$components/ui/spinner";

	import IntegrationConnections from "../../integration-connections/IntegrationConnections.svelte";
	import WebhookUrl from "../../webhook-url/WebhookUrl.svelte";
	import { WebhookUrlController } from "../../webhook-url/webhookUrl.svelte";
	import WebhookUrlGenerateDialog from "../../webhook-url-generate-dialog/WebhookUrlGenerateDialog.svelte";
	import { useIntegrationProviderController } from "../controller.svelte";
	import {
		deploymentReportFields,
		deploymentReportGeneralRules,
		deploymentStepNotes,
		githubActionsStep,
		installedPresetLabel,
		webhookIntegrationName,
	} from "./deploymentPreset";
	import { WebhookProviderController } from "./webhookProvider.svelte";

	const ctrl = useIntegrationProviderController();

	const installations = $derived(ctrl.integrations.installationsFor(webhookIntegrationName));
	const hasInstallations = $derived(installations.length > 0);
	const canInstall = $derived(ctrl.integrations.canInstall(webhookIntegrationName));
	const pending = $derived(ctrl.installPendingName === webhookIntegrationName);
	const installError = $derived(ctrl.installErrorFor(webhookIntegrationName));

	const webhook = new WebhookProviderController(ctrl);
	const webhookUrls = new WebhookUrlController();
	const presetSelectId = $props.id();
</script>

{#snippet installForm(label: string)}
	<div class="flex flex-wrap items-end gap-2">
		<Field.Field class="w-56">
			<Field.Label for={presetSelectId}>Preset</Field.Label>
			<Select.Root type="single" bind:value={webhook.preset}>
				<Select.Trigger id={presetSelectId} class="w-full">
					{webhook.presetLabel}
				</Select.Trigger>
				<Select.Content>
					{#each webhook.presets as preset (preset.value)}
						<Select.Item value={preset.value}>{preset.label}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</Field.Field>
		<Button
			class="w-fit"
			variant={hasInstallations ? "outline" : "default"}
			disabled={pending}
			onclick={webhook.install}
		>
			{#if pending}
				<Spinner />
			{:else}
				<RiAddLine />
			{/if}
			{label}
		</Button>
	</div>
{/snippet}

{#snippet connectionBody(installation: IntegrationInstallation)}
	<div class="flex flex-col gap-4 border-t pt-4">
		<p class="text-sm">
			<span class="text-muted-foreground">Preset:</span>
			{installedPresetLabel(installation.attributes.metadata)}
		</p>
		<div class="flex flex-col gap-3">
			<WebhookUrl {installation} {webhookUrls} sender="Your pipelines report deployments" />
		</div>
	</div>
{/snippet}

{#snippet deploymentReference()}
	<div class="flex flex-col gap-2">
		<h3 class="text-sm font-medium">Report format</h3>
		{#each deploymentReportGeneralRules as rule (rule)}
			<p class="text-sm text-muted-foreground">{rule}</p>
		{/each}
		<dl class="grid grid-cols-[minmax(0,10rem)_1fr] gap-x-4 gap-y-2 border p-3 text-sm">
			{#each deploymentReportFields as field (field.name)}
				<dt class="font-mono text-xs">{field.name}</dt>
				<dd class="text-muted-foreground">{field.rule}</dd>
			{/each}
		</dl>
	</div>

	<div class="flex flex-col gap-2">
		<div class="flex items-center justify-between gap-2">
			<h3 class="text-sm font-medium">GitHub Actions step</h3>
			<Button variant="ghost" size="sm" onclick={() => webhookUrls.copy(githubActionsStep, "Step")}>
				<RiFileCopyLine />
				Copy
			</Button>
		</div>
		<pre class="overflow-x-auto border bg-muted p-3 font-mono text-xs">{githubActionsStep}</pre>
		<ul class="list-disc pl-5 text-sm text-muted-foreground">
			{#each deploymentStepNotes as note (note)}
				<li>{note}</li>
			{/each}
		</ul>
	</div>
{/snippet}

<Card.Root>
	<Card.Header>
		<Card.Title>Webhooks</Card.Title>
		<Card.Description>
			Each installation accepts one preset, a payload format Rezible defines, at its own webhook URL. A
			different preset is a new installation.
		</Card.Description>
	</Card.Header>

	<Card.Content class="flex flex-col gap-6">
		<div class="flex flex-col gap-4">
			<IntegrationConnections name={webhookIntegrationName} settings={connectionBody}>
				{#snippet empty()}
					<Empty.Root class="border">
						<Empty.Header>
							<Empty.Media variant="icon">
								<RiWebhookLine />
							</Empty.Media>
							<Empty.Title>No webhooks are installed</Empty.Title>
							{#if !canInstall}
								<Empty.Description>
									Webhooks are not available in this deployment.
								</Empty.Description>
							{/if}
						</Empty.Header>
						{#if canInstall}
							<Empty.Content>
								{@render installForm("Install webhook")}
							</Empty.Content>
						{/if}
					</Empty.Root>
				{/snippet}
			</IntegrationConnections>

			{#if hasInstallations && canInstall}
				{@render installForm("Add another webhook")}
			{/if}

			{#if installError}
				<InlineAlert error={installError} dismissable={false} />
			{/if}
		</div>

		{#if hasInstallations}
			{@render deploymentReference()}
		{/if}
	</Card.Content>
</Card.Root>

<WebhookUrlGenerateDialog {webhookUrls} users="Pipelines" />
