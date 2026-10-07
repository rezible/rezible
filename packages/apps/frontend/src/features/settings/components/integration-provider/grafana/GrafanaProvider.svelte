<script lang="ts">
	import RiKey2Line from "remixicon-svelte/icons/key-2-line";
	import RiLineChartLine from "remixicon-svelte/icons/line-chart-line";

	import type { IntegrationInstallation } from "$lib/api";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import * as Empty from "$components/ui/empty";
	import * as Field from "$components/ui/field";
	import { Input } from "$components/ui/input";

	import { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

	import IntegrationConnections from "../../integration-connections/IntegrationConnections.svelte";
	import { useIntegrationProviderController } from "../controller.svelte";
	import GrafanaConnectionCheck from "./GrafanaConnectionCheck.svelte";
	import GrafanaInstallForm from "./GrafanaInstallForm.svelte";
	import GrafanaSettingsForm from "./GrafanaSettingsForm.svelte";
	import { GrafanaHealthCheck } from "./grafanaHealthCheck.svelte";
	import { grafanaIntegrationName } from "./grafanaInstallForm.svelte";
	import { grafanaSettings } from "./grafanaSettings";

	const ctrl = useIntegrationProviderController();

	const installations = $derived(ctrl.integrations.installationsFor(grafanaIntegrationName));
	const canInstall = $derived(ctrl.integrations.canInstall(grafanaIntegrationName));

	const forms = new IntegrationSettingsForms(grafanaSettings, () => installations);
	const health = new GrafanaHealthCheck();

	// The installation whose token is being replaced.
	let replacingId = $state<string>();

	const startReplacing = (installation: IntegrationInstallation) => {
		replacingId = installation.id;
	};

	const stopReplacing = () => {
		replacingId = undefined;
	};

	// A check made before installing, with another token or another Grafana, no longer describes the connection.
	const tokenReplaced = () => {
		health.clear();
		stopReplacing();
	};
</script>

{#snippet replaceTokenAction(installation: IntegrationInstallation)}
	{#if replacingId !== installation.id}
		<Button variant="outline" size="sm" onclick={() => startReplacing(installation)}>
			<RiKey2Line />
			Replace token
		</Button>
	{/if}
{/snippet}

{#snippet connectionUrl(installation: IntegrationInstallation)}
	{@const urlId = `grafana-installed-url-${installation.id}`}
	<Field.Field>
		<Field.Label for={urlId}>Grafana URL</Field.Label>
		<Input
			id={urlId}
			value={installation.attributes.providerInstallationRef}
			readonly
			class="bg-muted text-muted-foreground"
		/>
		<Field.Description>
			To connect a different Grafana, disconnect this one and install again.
		</Field.Description>
	</Field.Field>
{/snippet}

{#snippet connectionBody(installation: IntegrationInstallation)}
	<div class="flex flex-col gap-6 border-t pt-4">
		{#if replacingId === installation.id}
			<GrafanaInstallForm replacing={installation} onDone={tokenReplaced} onCancel={stopReplacing} />
		{:else}
			{@render connectionUrl(installation)}
		{/if}

		<GrafanaSettingsForm {installation} {forms} />

		<GrafanaConnectionCheck {installation} {health} />
	</div>
{/snippet}

<Card.Root>
	<Card.Header>
		<Card.Title>Grafana connection</Card.Title>
		<Card.Description>
			Rezible reads logs from Loki and metrics from Prometheus through Grafana's data source proxy, with
			a service account token.
		</Card.Description>
	</Card.Header>

	<Card.Content class="flex flex-col gap-4">
		<IntegrationConnections
			name={grafanaIntegrationName}
			connectionActions={replaceTokenAction}
			settings={connectionBody}
		>
			{#snippet empty()}
				{#if canInstall}
					<GrafanaInstallForm onDone={health.clear} />
				{:else}
					<Empty.Root class="border">
						<Empty.Header>
							<Empty.Media variant="icon">
								<RiLineChartLine />
							</Empty.Media>
							<Empty.Title>Grafana is not available</Empty.Title>
							<Empty.Description>It is not enabled in this deployment.</Empty.Description>
						</Empty.Header>
					</Empty.Root>
				{/if}
			{/snippet}
		</IntegrationConnections>
	</Card.Content>
</Card.Root>
