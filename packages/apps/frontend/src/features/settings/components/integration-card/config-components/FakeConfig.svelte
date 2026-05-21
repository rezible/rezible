<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import Button from "$components/ui/button/button.svelte";
	import { useIntegrationCardController } from "../controller.svelte";
	import { onMount } from "svelte";

	const ctrl = useIntegrationCardController();

    onMount(() => {
        ctrl.setConfig({
            displayName: "Fake Data Provider",
            externalRef: "fake",
            config: {},
        });
    })
</script>

<Alert.Root>
    <Alert.Title>Connect Fake</Alert.Title>
    <Alert.Description>Ready to connect</Alert.Description>
</Alert.Root>
