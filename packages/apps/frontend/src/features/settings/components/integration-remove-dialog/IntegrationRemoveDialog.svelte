<script lang="ts">
	import { watch } from "runed";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as AlertDialog from "$components/ui/alert-dialog";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";

	import { useIntegrationProviderController } from "../integration-provider/controller.svelte";

	const ctrl = useIntegrationProviderController();

	// Keep the last name while the dialog animates closed.
	let connectionName = $state("");
	watch(
		() => ctrl.removalCandidate,
		(candidate) => {
			if (candidate) {
				connectionName = candidate.attributes.displayName;
			}
		}
	);
</script>

<AlertDialog.Root
	bind:open={
		() => !!ctrl.removalCandidate,
		(open) => {
			if (!open) {
				ctrl.cancelRemoval();
			}
		}
	}
>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Disconnect {connectionName}?</AlertDialog.Title>
			<AlertDialog.Description>
				Rezible will stop using this connection and delete its saved settings. Apps installed on the
				provider's side, such as in Slack or GitHub, stay installed until you uninstall them there.
			</AlertDialog.Description>
		</AlertDialog.Header>

		{#if ctrl.removeError}
			<InlineAlert error={ctrl.removeError} dismissable={false} />
		{/if}

		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={ctrl.removing}>Cancel</AlertDialog.Cancel>
			<Button variant="destructive" disabled={ctrl.removing} onclick={ctrl.confirmRemoval}>
				{#if ctrl.removing}
					<Spinner />
				{/if}
				Disconnect
			</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
