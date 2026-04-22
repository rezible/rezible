<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { useIntegrationConfigController } from "../controller.svelte";

	const ctrl = useIntegrationConfigController();
</script>

<Alert.Root>
	<Alert.Title>No custom settings UI yet</Alert.Title>
	<Alert.Description>
		<span>Configuration controls for <span class="font-medium capitalize">{ctrl.name}</span> will be added later.</span>
	</Alert.Description>
</Alert.Root>
