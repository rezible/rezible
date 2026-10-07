<script lang="ts">
	import RiCheckboxCircleLine from "remixicon-svelte/icons/checkbox-circle-line";
	import RiErrorWarningLine from "remixicon-svelte/icons/error-warning-line";
	import RiPulseLine from "remixicon-svelte/icons/pulse-line";

	import type { IntegrationInstallation } from "$lib/api";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";

	import type { GrafanaHealthCheck } from "./grafanaHealthCheck.svelte";

	type Props = {
		installation: IntegrationInstallation;
		health: GrafanaHealthCheck;
	};
	const { installation, health }: Props = $props();

	const result = $derived(health.resultFor(installation));
</script>

<div class="flex flex-col gap-3">
	<div class="flex flex-col gap-1">
		<h3 class="text-sm font-medium">Connection</h3>
		<p class="text-sm text-muted-foreground">
			Checks with the saved settings that Rezible can read each data source, and that its service label
			has values in the last 15 minutes.
		</p>
	</div>

	{#if result?.kind === "ok"}
		<Alert.Root>
			<RiCheckboxCircleLine />
			<Alert.Title>Rezible read both data sources</Alert.Title>
		</Alert.Root>
	{:else if result?.kind === "failed"}
		<Alert.Root variant="destructive">
			<RiErrorWarningLine />
			<Alert.Title>The check failed</Alert.Title>
			<Alert.Description class="whitespace-pre-line">{result.message}</Alert.Description>
		</Alert.Root>
	{:else if result?.kind === "error"}
		<InlineAlert error={result.error} dismissable={false} />
	{/if}

	<Button
		class="w-fit"
		variant="outline"
		disabled={health.checking}
		onclick={() => health.check(installation)}
	>
		{#if health.checking}
			<Spinner />
		{:else}
			<RiPulseLine />
		{/if}
		Check connection
	</Button>
</div>
