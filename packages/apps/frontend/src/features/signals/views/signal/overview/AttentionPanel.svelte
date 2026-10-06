<script lang="ts">
	import * as Field from "$components/ui/field";
	import * as RadioGroup from "$components/ui/radio-group";
	import { Spinner } from "$components/ui/spinner";
	import Timestamp from "$components/common/timestamp/Timestamp.svelte";
	import { attentionLevelOptions } from "$features/signals/lib/attention";
	import { useAlertViewController } from "../controller.svelte";

	const view = useAlertViewController();
	const pending = $derived(view.attentionMutation.isPending);
	const setAt = $derived(view.attention?.setAt);
</script>

<section aria-labelledby="attention-title" class="flex flex-col gap-3 rounded-lg border bg-card p-4">
	<div class="flex items-center gap-2">
		<h2 id="attention-title" class="region-label">Signal attention</h2>
		{#if pending}
			<Spinner aria-label="Saving" />
		{/if}
	</div>

	<Field.Set>
		<Field.Legend class="sr-only">How this definition's alerts may raise situations</Field.Legend>
		<RadioGroup.Root
			value={view.attentionLevel}
			onValueChange={view.setAttentionLevel}
			disabled={pending}
		>
			{#each attentionLevelOptions as option (option.value)}
				<Field.Field orientation="horizontal">
					<RadioGroup.Item value={option.value} id={`attention-${option.value}`} />
					<Field.Content>
						<Field.Label for={`attention-${option.value}`}>{option.label}</Field.Label>
						<Field.Description>{option.help}</Field.Description>
					</Field.Content>
				</Field.Field>
			{/each}
		</RadioGroup.Root>
	</Field.Set>

	<p class="text-xs text-muted-foreground">
		{#if setAt}
			Changed <Timestamp value={setAt} />.
		{:else}
			Never changed.
		{/if}
		Rezible reassesses affected situations after a change.
	</p>
</section>
