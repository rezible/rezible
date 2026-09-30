<script lang="ts">
	import RiGoogleFill from "remixicon-svelte/icons/google-fill";
	import RiKey2Line from "remixicon-svelte/icons/key-2-line";

	import { resolve } from "$app/paths";
	import type { IntegrationInstallation } from "$lib/api";
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import * as Empty from "$components/ui/empty";
	import { useUserSessionState } from "$lib/user-session.svelte";

	import { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

	import IntegrationConnections from "../../integration-connections/IntegrationConnections.svelte";
	import { useIntegrationProviderController } from "../controller.svelte";
	import GoogleInstallForm from "./GoogleInstallForm.svelte";
	import GoogleSettingsForm from "./GoogleSettingsForm.svelte";
	import { googleSettings } from "./googleSettings";

	const integrationName = "google";

	const ctrl = useIntegrationProviderController();
	const session = useUserSessionState();

	const installations = $derived(ctrl.integrations.installationsFor(integrationName));
	const canInstall = $derived(ctrl.integrations.canInstall(integrationName));
	const incidentManagementEnabled = $derived(!!session.orgPreferences?.enableIncidentManagement);

	const forms = new IntegrationSettingsForms(googleSettings, () => installations);

	// The connection whose service account key is being replaced.
	let replacingId = $state<string>();

	const startReplacing = (installation: IntegrationInstallation) => {
		replacingId = installation.id;
	};

	const stopReplacing = () => {
		replacingId = undefined;
	};
</script>

{#snippet replaceKeyAction(installation: IntegrationInstallation)}
	{#if replacingId !== installation.id}
		<Button variant="outline" size="sm" onclick={() => startReplacing(installation)}>
			<RiKey2Line />
			Replace key
		</Button>
	{/if}
{/snippet}

{#snippet connectionBody(installation: IntegrationInstallation)}
	{#if replacingId === installation.id}
		<div class="border-t pt-4">
			<GoogleInstallForm replacing={installation} onDone={stopReplacing} onCancel={stopReplacing} />
		</div>
	{:else if incidentManagementEnabled}
		<GoogleSettingsForm {installation} {forms} />
	{:else}
		<Alert.Root>
			<Alert.Title>Incident management is turned off</Alert.Title>
			<Alert.Description>
				<span>
					Turn it on in
					<a class="underline underline-offset-4" href={resolve("/settings/incidents")}>
						incident management settings
					</a>
					to create Google Meet conferences for incidents.
				</span>
			</Alert.Description>
		</Alert.Root>
	{/if}
{/snippet}

<Card.Root>
	<Card.Header>
		<Card.Title>Workspace connection</Card.Title>
		<Card.Description>
			Rezible uses a service account with domain-wide delegation to create Google Meet meetings.
		</Card.Description>
	</Card.Header>

	<Card.Content class="flex flex-col gap-4">
		<IntegrationConnections
			name={integrationName}
			referenceLabel="Customer ID"
			connectionActions={replaceKeyAction}
			settings={connectionBody}
		>
			{#snippet empty()}
				{#if canInstall}
					<GoogleInstallForm />
				{:else}
					<Empty.Root class="border">
						<Empty.Header>
							<Empty.Media variant="icon">
								<RiGoogleFill />
							</Empty.Media>
							<Empty.Title>Google Workspace is not available</Empty.Title>
							<Empty.Description>It is not enabled in this deployment.</Empty.Description>
						</Empty.Header>
					</Empty.Root>
				{/if}
			{/snippet}
		</IntegrationConnections>
	</Card.Content>
</Card.Root>
