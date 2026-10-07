<script lang="ts">
	import RiAlarmWarningLine from "remixicon-svelte/icons/alarm-warning-line";
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import RiFileCopyLine from "remixicon-svelte/icons/file-copy-line";

	import { resolve } from "$app/paths";
	import type { IntegrationInstallation } from "$lib/api";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import * as Empty from "$components/ui/empty";
	import { Spinner } from "$components/ui/spinner";

	import { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

	import IntegrationConnections from "../../integration-connections/IntegrationConnections.svelte";
	import { useIntegrationProviderController } from "../controller.svelte";
	import AlertmanagerGenerateDialog from "./AlertmanagerGenerateDialog.svelte";
	import AlertmanagerServiceLabelsForm from "./AlertmanagerServiceLabelsForm.svelte";
	import AlertmanagerWebhookUrl from "./AlertmanagerWebhookUrl.svelte";
	import { alertmanagerSettings, routingExample } from "./alertmanagerSettings";
	import {
		AlertmanagerProviderController,
		alertmanagerIntegrationName,
	} from "./alertmanagerProvider.svelte";

	const ctrl = useIntegrationProviderController();

	const installations = $derived(ctrl.integrations.installationsFor(alertmanagerIntegrationName));
	const hasInstallations = $derived(installations.length > 0);
	const canInstall = $derived(ctrl.integrations.canInstall(alertmanagerIntegrationName));
	const pending = $derived(ctrl.installPendingName === alertmanagerIntegrationName);
	const installError = $derived(ctrl.installErrorFor(alertmanagerIntegrationName));

	const forms = new IntegrationSettingsForms(alertmanagerSettings, () => installations);
	const alertmanager = new AlertmanagerProviderController(ctrl);
</script>

{#snippet installButton(label: string)}
	<Button
		class="w-fit"
		variant={hasInstallations ? "outline" : "default"}
		disabled={pending}
		onclick={alertmanager.install}
	>
		{#if pending}
			<Spinner />
		{:else}
			<RiAddLine />
		{/if}
		{label}
	</Button>
{/snippet}

{#snippet connectionBody(installation: IntegrationInstallation)}
	{@const example = routingExample(alertmanager.issuedUrl(installation.id))}
	<div class="flex flex-col gap-6 border-t pt-4">
		<div class="flex flex-col gap-3">
			<AlertmanagerWebhookUrl {installation} {alertmanager} />
		</div>

		<div class="flex flex-col gap-2">
			<div class="flex items-center justify-between gap-2">
				<h3 class="text-sm font-medium">Alertmanager configuration</h3>
				<Button variant="ghost" size="sm" onclick={() => alertmanager.copy(example, "Configuration")}>
					<RiFileCopyLine />
					Copy
				</Button>
			</div>
			<p class="text-sm text-muted-foreground">
				This wraps your existing routing tree; it is not a fragment to append to a route. Keep your
				receivers, copy the old root's grouping and timing settings to the new root, and move the
				entire old root, with its receiver and child routes, to the second child. Rezible receives
				every alert first, and the old tree still chooses the original receiver, including its
				fallback.
			</p>
			<pre class="overflow-x-auto border bg-muted p-3 font-mono text-xs">{example}</pre>
			<p class="text-sm text-muted-foreground">
				Keep the 4-hour repeat interval, and the inherited group interval, below Rezible's alert
				resolution timeout of 12 hours by default.
			</p>
		</div>

		<AlertmanagerServiceLabelsForm {installation} {forms} />
	</div>
{/snippet}

<Card.Root>
	<Card.Header>
		<Card.Title>Alertmanager receivers</Card.Title>
		<Card.Description>
			Each installation has its own webhook URL for one Alertmanager. There is no separate connection
			status: a firing alert that reaches Rezible appears under
			<a class="underline underline-offset-4" href={resolve("/signals")}>Signals</a>
		</Card.Description>
	</Card.Header>

	<Card.Content class="flex flex-col gap-4">
		<IntegrationConnections name={alertmanagerIntegrationName} settings={connectionBody}>
			{#snippet empty()}
				<Empty.Root class="border">
					<Empty.Header>
						<Empty.Media variant="icon">
							<RiAlarmWarningLine />
						</Empty.Media>
						<Empty.Title>Alertmanager is not installed</Empty.Title>
						{#if !canInstall}
							<Empty.Description>
								Alertmanager is not available in this deployment.
							</Empty.Description>
						{/if}
					</Empty.Header>
					{#if canInstall}
						<Empty.Content>
							{@render installButton("Install Alertmanager")}
						</Empty.Content>
					{/if}
				</Empty.Root>
			{/snippet}
		</IntegrationConnections>

		{#if hasInstallations && canInstall}
			{@render installButton("Add another Alertmanager")}
		{/if}

		{#if installError}
			<InlineAlert error={installError} dismissable={false} />
		{/if}
	</Card.Content>
</Card.Root>

<AlertmanagerGenerateDialog {alertmanager} />
