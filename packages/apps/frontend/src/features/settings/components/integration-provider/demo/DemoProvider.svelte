<script lang="ts">
	import RiFlaskLine from "remixicon-svelte/icons/flask-line";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import * as Empty from "$components/ui/empty";
	import { Spinner } from "$components/ui/spinner";

	import IntegrationConnections from "../../integration-connections/IntegrationConnections.svelte";
	import { useIntegrationProviderController } from "../controller.svelte";

	const integrationName = "demo";

	const ctrl = useIntegrationProviderController();

	const canInstall = $derived(ctrl.integrations.canInstall(integrationName));
	const pending = $derived(ctrl.installPendingName === integrationName);
	const installError = $derived(ctrl.installErrorFor(integrationName));

	const install = () => {
		ctrl.install(integrationName, {});
	};
</script>

<Card.Root>
	<Card.Header>
		<Card.Title>Sample data</Card.Title>
		<Card.Description>
			Demo supplies sample alerts, incidents, and code changes so you can explore Rezible without
			connecting real tools.
		</Card.Description>
	</Card.Header>

	<Card.Content class="flex flex-col gap-4">
		<IntegrationConnections name={integrationName}>
			{#snippet empty()}
				<Empty.Root class="border">
					<Empty.Header>
						<Empty.Media variant="icon">
							<RiFlaskLine />
						</Empty.Media>
						<Empty.Title>Demo data is not installed</Empty.Title>
						{#if !canInstall}
							<Empty.Description>
								Demo data is not available in this deployment.
							</Empty.Description>
						{/if}
					</Empty.Header>
					{#if canInstall}
						<Empty.Content>
							<Button disabled={pending} onclick={install}>
								{#if pending}
									<Spinner />
								{/if}
								Install demo data
							</Button>
						</Empty.Content>
					{/if}
				</Empty.Root>
			{/snippet}
		</IntegrationConnections>

		{#if installError}
			<InlineAlert error={installError} dismissable={false} />
		{/if}
	</Card.Content>
</Card.Root>
