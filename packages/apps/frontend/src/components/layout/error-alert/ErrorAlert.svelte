<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import RiCloseLine from "remixicon-svelte/icons/close-line";
	import { errorDisplay, type ApiError } from "$lib/api";

	type Props = {
		error: ApiError | undefined;
		onDismiss?: () => void;
		dismissable?: boolean;
	};
	let { error = $bindable(), onDismiss, dismissable = true }: Props = $props();

	const display = $derived(error ? errorDisplay(error) : undefined);
</script>

{#if !!display}
	<Alert.Root variant="destructive">
		<Alert.Title class="font-semibold text-sm">{display.title}</Alert.Title>
		{#if display.detail}
			<Alert.Description>{display.detail}</Alert.Description>
		{/if}

		{#if onDismiss}
			<Alert.Action>
				<Button size="icon-sm" variant="ghost" aria-label="Dismiss error" onclick={onDismiss}>
					<RiCloseLine />
				</Button>
			</Alert.Action>
		{/if}

		{#if dismissable}
			<Alert.Action>
				<Button
					aria-label="Dismiss error"
					size="icon-sm"
					variant="ghost"
					onclick={() => {
						error = undefined;
					}}
				>
					<RiCloseLine />
				</Button>
			</Alert.Action>
		{/if}
	</Alert.Root>
{/if}
