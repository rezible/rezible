<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as Empty from "$components/ui/empty";
	import ErrorAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { useIncidentView } from "../controller.svelte";

	const incident = useIncidentView();
	const mutation = $derived(incident.startRetrospectiveMutation);
</script>

<Empty.Root>
	<Empty.Header>
		<Empty.Title>No retrospective yet</Empty.Title>
		<Empty.Description>
			{#if incident.canStartRetrospective}
				Start a retrospective to analyse this incident and write its report.
			{:else}
				A retrospective will be available once this incident is resolved.
			{/if}
		</Empty.Description>
	</Empty.Header>
	<Empty.Content>
		{#if mutation.error}
			<ErrorAlert error={mutation.error} dismissable={false} />
		{/if}
		{#if incident.canStartRetrospective}
			<Button onclick={incident.startRetrospective} disabled={mutation.isPending}>
				{#if mutation.isPending}
					Starting retrospective…
				{:else}
					Start retrospective
				{/if}
			</Button>
		{/if}
	</Empty.Content>
</Empty.Root>
