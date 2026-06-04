<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { useAvailableIntegrationCardController } from "../availableIntegrationController.svelte";
	import { onMount } from "svelte";

	const ctrl = useAvailableIntegrationCardController();

    onMount(() => {
        ctrl.setConfig({
            displayName: "Fake Data Provider",
            config: {},
            preferences: {},
        });
    })
</script>

<Alert.Root>
    <Alert.Title>Connect Fake</Alert.Title>
    <Alert.Description>Ready to connect</Alert.Description>
</Alert.Root>
