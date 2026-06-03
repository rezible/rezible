<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { useAvailableIntegrationCardController } from "../availableIntegrationController.svelte";

	const ctrl = useAvailableIntegrationCardController();
</script>

<Alert.Root>
	<Alert.Title>No custom settings UI yet</Alert.Title>
	<Alert.Description>
		<span>Configuration controls for <span class="font-medium capitalize">{ctrl.name}</span> will be added later.</span>
	</Alert.Description>
</Alert.Root>
