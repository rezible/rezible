<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import RiCloseLine from "remixicon-svelte/icons/close-line"
	import type { ErrorModel } from "$lib/api";

	type Props = {
		error: ErrorModel | undefined;
		onDismiss?: () => void;
		dismissable?: boolean;
	};
	let { error = $bindable(), onDismiss, dismissable = true }: Props = $props();
</script>

{#if !!error}
<Alert.Root variant="destructive">
	<Alert.Title class="font-semibold text-sm">{error?.title ?? "An Error Occurred"}</Alert.Title>
	<Alert.Description>{error?.detail ?? ""}</Alert.Description>

	{#if onDismiss}
		<Alert.Action>
			<Button size="icon-sm" variant="ghost" onclick={onDismiss}><RiCloseLine /></Button>
		</Alert.Action>
	{/if}

	{#if dismissable}
		<Alert.Action>
			<Button size="icon-sm" variant="ghost" onclick={() => {error = undefined}}><RiCloseLine /></Button>
		</Alert.Action>
	{/if}
</Alert.Root>
{/if}
