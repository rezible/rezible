<script lang="ts">
	import * as AlertDialog from "$components/ui/alert-dialog";
	import * as Dialog from "$components/ui/dialog";
	import * as Field from "$components/ui/field";
	import * as RadioGroup from "$components/ui/radio-group";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import SituationPicker from "$features/situations/components/situation-picker/SituationPicker.svelte";
	import type { MuteReason, SituationController } from "./controller.svelte";

	type Props = { controller: SituationController };
	const { controller }: Props = $props();

	const muteReasons: { value: MuteReason; label: string; description: string }[] = [
		{
			value: "not_noteworthy",
			label: "Not noteworthy",
			description: "This activity does not need anyone's attention.",
		},
		{
			value: "expected",
			label: "Expected",
			description: "Planned or known activity, such as maintenance.",
		},
	];

	const mutePending = $derived(controller.muteMutation.isPending);
	const mergePending = $derived(controller.mergeMutation.isPending);
	const closePending = $derived(controller.closeMutation.isPending);
</script>

<Dialog.Root bind:open={controller.muteDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Mute situation</Dialog.Title>
			<Dialog.Description>
				A muted situation keeps its evidence and still gathers matching signals, but is not assessed,
				raised or investigated.
			</Dialog.Description>
		</Dialog.Header>

		<Field.Set>
			<Field.Legend variant="label">Reason</Field.Legend>
			<RadioGroup.Root value={controller.muteReason} onValueChange={controller.setMuteReason}>
				{#each muteReasons as reason (reason.value)}
					<Field.Field orientation="horizontal">
						<RadioGroup.Item value={reason.value} id={`mute-reason-${reason.value}`} />
						<Field.Content>
							<Field.Label for={`mute-reason-${reason.value}`}>{reason.label}</Field.Label>
							<Field.Description>{reason.description}</Field.Description>
						</Field.Content>
					</Field.Field>
				{/each}
			</RadioGroup.Root>
		</Field.Set>

		<Dialog.Footer>
			<Dialog.Close>
				{#snippet child({ props })}
					<Button {...props} variant="outline" disabled={mutePending}>Cancel</Button>
				{/snippet}
			</Dialog.Close>
			<Button disabled={mutePending} onclick={controller.submitMute}>
				{#if mutePending}
					<Spinner data-icon="inline-start" />
				{/if}
				Mute
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={controller.mergeDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Merge into another situation</Dialog.Title>
			<Dialog.Description>
				Use this for a duplicate: this situation's signals move to the one you choose, and this
				situation closes as merged.
			</Dialog.Description>
		</Dialog.Header>

		<SituationPicker
			label="Merge into"
			excludeId={controller.situationId}
			value={controller.mergeTargetId}
			onchange={controller.setMergeTarget}
		/>

		<Dialog.Footer>
			<Dialog.Close>
				{#snippet child({ props })}
					<Button {...props} variant="outline" disabled={mergePending}>Cancel</Button>
				{/snippet}
			</Dialog.Close>
			<Button disabled={!controller.mergeTargetId || mergePending} onclick={controller.submitMerge}>
				{#if mergePending}
					<Spinner data-icon="inline-start" />
				{/if}
				Merge
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<AlertDialog.Root bind:open={controller.closeDialogOpen}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Close situation?</AlertDialog.Title>
			<AlertDialog.Description>
				Closing means the noteworthy activity is over. A closed situation stays closed; later matching
				activity starts a new situation linked to this one.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={closePending}>Cancel</AlertDialog.Cancel>
			<Button disabled={closePending} onclick={controller.confirmClose}>
				{#if closePending}
					<Spinner data-icon="inline-start" />
				{/if}
				Close situation
			</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
