<script lang="ts">
	import { mdiGithub, mdiPlus, mdiSlack, mdiWeb } from "@mdi/js";
	import { Button, Icon, SelectField } from "svelte-ux";
	import Slack from "./data-sources/Slack.svelte";
	// import Github from "./data-sources/Github.svelte";
	import Url from "./data-sources/Url.svelte";

	type Props = {};
	const {}: Props = $props();

	const dataSources = [
		{ value: "slack", label: "Slack", icon: mdiSlack, component: Slack },
		// { value: "github", label: "Github", icon: mdiGithub, component: Github },
		{ value: "url", label: "Web URL", icon: mdiWeb, component: Url },
	];
	let eventData = $state<any[]>([]);

	type DataEvidence = {
		source: string;
		id: string;
	};
	let addingState = $state<DataEvidence>();
	const SourceComponent = $derived(dataSources.find((s) => s.value === addingState?.source)?.component);

	const onEvidenceLinked = (id: string) => {
		console.log(addingState?.source, id);
		addingState = undefined;
	};
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	{#if addingState}
		<SelectField bind:value={addingState.source} options={dataSources} label="Data Source" />

		{#if SourceComponent}
			<SourceComponent onLinked={onEvidenceLinked} />
		{/if}
	{:else}
		<Button
			class="text-surface-content/50 p-2"
			color="primary"
			variant="fill-light"
			on:click={() => {
				addingState = { source: "", id: "" };
			}}
		>
			<span class="flex items-center gap-2 text-primary-content">
				Add Evidence
				<Icon data={mdiPlus} />
			</span>
		</Button>
	{/if}
</div>
