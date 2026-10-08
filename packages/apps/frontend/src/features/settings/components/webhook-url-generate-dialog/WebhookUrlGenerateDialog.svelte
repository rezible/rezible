<script lang="ts">
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as AlertDialog from "$components/ui/alert-dialog";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";

	import type { WebhookUrlController } from "../webhook-url/webhookUrl.svelte";

	type Props = {
		webhookUrls: WebhookUrlController;
		// What uses the URL and must be updated, such as "Alertmanager receivers".
		users: string;
	};
	const { webhookUrls, users }: Props = $props();

	const issuing = $derived(!!webhookUrls.issuingId);
	const issueError = $derived(webhookUrls.confirmingError);
</script>

<AlertDialog.Root
	bind:open={
		() => !!webhookUrls.confirmingId,
		(open) => {
			if (!open) {
				webhookUrls.cancelGenerate();
			}
		}
	}
>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Replace the webhook URL?</AlertDialog.Title>
			<AlertDialog.Description>
				The current URL for this installation stops working at once. {users} using it must be updated with
				the new URL.
			</AlertDialog.Description>
		</AlertDialog.Header>

		{#if issueError}
			<InlineAlert error={issueError} dismissable={false} />
		{/if}

		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={issuing}>Cancel</AlertDialog.Cancel>
			<Button disabled={issuing} onclick={webhookUrls.confirmGenerate}>
				{#if issuing}
					<Spinner />
				{/if}
				Replace
			</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
