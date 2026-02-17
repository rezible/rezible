<script lang="ts">
	import type { ConfigComponentProps } from '../types';
	import * as Alert from "$components/ui/alert";

	const { integration }: ConfigComponentProps = $props();
</script>

<Alert.Root>
	<Alert.Title>No custom settings UI yet</Alert.Title>
	<Alert.Description>
		Configuration controls for <span class="font-medium capitalize">{integration.name}</span> will be added later.
	</Alert.Description>
</Alert.Root>
