<script lang="ts">
	import RiFileCopyLine from "remixicon-svelte/icons/file-copy-line";
	import RiLinkM from "remixicon-svelte/icons/link-m";

	import type { IntegrationInstallation } from "$lib/api";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { Button } from "$components/ui/button";
	import * as Field from "$components/ui/field";
	import { Input } from "$components/ui/input";

	import type { AlertmanagerProviderController } from "./alertmanagerProvider.svelte";

	type Props = {
		installation: IntegrationInstallation;
		alertmanager: AlertmanagerProviderController;
	};
	const { installation, alertmanager }: Props = $props();

	const issuedUrl = $derived(alertmanager.issuedUrl(installation.id));
	const issueError = $derived(alertmanager.issueError(installation.id));
	const inputId = $derived(`alertmanager-webhook-url-${installation.id}`);
</script>

<Field.Field>
	<Field.Label for={inputId}>Webhook URL</Field.Label>
	{#if issuedUrl}
		<div class="flex items-center gap-2">
			<Input id={inputId} value={issuedUrl} readonly class="font-mono text-xs" />
			<Button
				variant="outline"
				size="icon"
				aria-label="Copy webhook URL"
				onclick={() => alertmanager.copy(issuedUrl, "Webhook URL")}
			>
				<RiFileCopyLine />
			</Button>
		</div>
		<Field.Description>
			Copy it now. It contains a secret and is not shown again after you leave this page.
		</Field.Description>
	{:else}
		<Field.Description>
			Alertmanager sends alerts to this URL. Rezible shows it once, when it is generated.
		</Field.Description>
	{/if}
</Field.Field>

{#if issueError}
	<InlineAlert error={issueError} dismissable={false} />
{/if}

<Button class="w-fit" variant="outline" onclick={() => alertmanager.requestGenerate(installation.id)}>
	<RiLinkM />
	Generate webhook URL
</Button>
