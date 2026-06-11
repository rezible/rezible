<script lang="ts">
	import { onMount } from "svelte";
	import * as Alert from "$components/ui/alert";
	import { useConfigureIntegrationDialogController } from "../controller.svelte";

	const ctrl = useConfigureIntegrationDialogController();

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
