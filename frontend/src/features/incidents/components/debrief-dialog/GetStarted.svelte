<script lang="ts">
	import { Button } from "svelte-ux";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";
	import {
		getIncidentDebriefOptions,
		getIncidentUserDebriefOptions,
		updateIncidentDebriefMutation,
		type Incident,
		type IncidentDebrief,
	} from "$lib/api";
	import { useQueryClient, createMutation } from "@tanstack/svelte-query";

	interface Props {
		debrief: IncidentDebrief;
	}

	let { debrief }: Props = $props();

	const queryClient = useQueryClient();
	const start = createMutation(() => ({
		...updateIncidentDebriefMutation(),
		onSuccess: () => {
			// this is the query used to fetch the debrief in +page.svelte
			const debriefQueryOpts = getIncidentUserDebriefOptions({
				path: { id: debrief.attributes.incidentId },
			});
			return queryClient.invalidateQueries(debriefQueryOpts);
		},
	}));

	const startDebrief = () =>
		start.mutate({
			path: { id: debrief.id },
			body: { attributes: { status: "started" } },
		});
</script>

<div class="p-2 text-surface-content overflow-y-auto shrink" class:hidden={start.isSuccess}>
	<p class="">
		A post-incident debrief brings teams together to learn from service disruptions and make our systems
		more resilient.
	</p>

	<p>We'll work through this together in a blameless environment focused on improvement.</p>

	<h2 class="text-xl font-semibold mt-6 mb-3">What We'll Cover</h2>
	<ul class="list-disc pl-6 mb-4">
		<li>What happened and when</li>
		<li>Understand contributing factors</li>
		<li>Identify what worked well and what didn't</li>
	</ul>

	<h2 class="text-xl font-semibold mt-6 mb-3">Why This Matters</h2>
	<ul class="list-disc pl-6 mb-4">
		<li>Strengthen our systems through honest analysis</li>
		<li>Share knowledge across teams</li>
		<li>Improve our response processes</li>
		<li>Prevent similar incidents</li>
	</ul>
</div>

<div class="bg-success-900/50 textcontent p-4 rounded-lg">
	<p class="text-sm">Best Practice: Complete the debrief within 72 hours while details are fresh.</p>
</div>

<div class="border-t h-0 my-2"></div>

<div class="w-fit mx-auto">
	<Button
		size="lg"
		variant="fill"
		color="primary"
		disabled={start.isPending}
		on:click={startDebrief}
		loading={start.isPending}
	>
		Start Debrief
	</Button>
</div>

{#if start.isPending}
	<div class="flex items-center gap-2 w-fit shrink overflow-hidden">
		<span class="text-accent-100">Thinking</span>
		<LoadingIndicator />
	</div>
{:else if start.isError}
	<p class="text-error">Failed: {start.error.detail}</p>
{/if}
