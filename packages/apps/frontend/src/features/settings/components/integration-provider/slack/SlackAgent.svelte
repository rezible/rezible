<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";
	import { useUserSessionState } from "$lib/user-session.svelte";

	import { useIntegrationDataSyncController } from "../../integration-datasync-dialog/controller.svelte";
	import { useIntegrationProviderConfigController } from "../controller.svelte";
	import SlackIncidentManagement from "./SlackIncidentManagement.svelte";
	import type { IntegrationInstallation } from "@rezible/api-client-ts";

    type Props = {
        installation: IntegrationInstallation;
    };
    const { installation }: Props = $props();

	const ctrl = useIntegrationProviderConfigController();
	const sync = useIntegrationDataSyncController();
	
    const attrs = $derived(installation.attributes);
</script>

<Card.Root>
    <Card.Header>
        <Card.Title>Slack Agent</Card.Title>
    </Card.Header>
    <Card.Content>
        <div class="flex flex-wrap items-center justify-between gap-2">
            <Alert.Root class="flex-1">
                <Alert.Title>Installed</Alert.Title>
                <Alert.Description>{attrs.displayName}</Alert.Description>
            </Alert.Root>
            <div class="flex gap-2">
                {#if attrs.capabilities.includes("event_sync")}
                    <Button variant="outline" onclick={() => sync.openFor(installation)}>Sync</Button>
                {/if}
                <Button variant="destructive" onclick={() => ctrl.disconnect(installation.id)}
                    >Disconnect</Button
                >
            </div>
        </div>
    </Card.Content>
</Card.Root>
