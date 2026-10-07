<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as AlertDialog from "$components/ui/alert-dialog";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";

	import type { AlertmanagerProviderController } from "./alertmanagerProvider.svelte";

	type Props = {
		alertmanager: AlertmanagerProviderController;
	};
	const { alertmanager }: Props = $props();

	const issuing = $derived(!!alertmanager.issuingId);
	const issueError = $derived(alertmanager.confirmingError);
</script>

<AlertDialog.Root
	bind:open={
		() => !!alertmanager.confirmingId,
		(open) => {
			if (!open) {
				alertmanager.cancelGenerate();
			}
		}
	}
>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Generate a webhook URL?</AlertDialog.Title>
			<AlertDialog.Description>
				Any URL generated earlier for this installation stops working at once. Alertmanager receivers
				using it must be updated with the new URL.
			</AlertDialog.Description>
		</AlertDialog.Header>

		{#if issueError}
			<InlineAlert error={issueError} dismissable={false} />
		{/if}

		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={issuing}>Cancel</AlertDialog.Cancel>
			<Button disabled={issuing} onclick={alertmanager.confirmGenerate}>
				{#if issuing}
					<Spinner />
				{/if}
				Generate
			</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
