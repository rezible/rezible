<script lang="ts">
	import { Input } from "$components/ui/input";
	import { Textarea } from "$components/ui/textarea";
	import * as Field from "$components/ui/field";
	import * as Select from "$components/ui/select";
	import { getLocalTimeZone, parseDateTime, toCalendarDateTime, toZoned } from "@internationalized/date";
	import { useEventDialog } from "../controller.svelte";

	const editor = useEventDialog();
	const eventKinds = ["observation", "action", "decision", "context", "finding", "recommendation"] as const;
</script>

<Field.Group>
	<Field.Field>
		<Field.Label for="analysis-entry-title">Title</Field.Label>
		<Input id="analysis-entry-title" bind:value={editor.title} required disabled={editor.loading} />
	</Field.Field>
	<Field.Field>
		<Field.Label for="analysis-entry-kind">Kind</Field.Label>
		<Select.Root type="single" bind:value={editor.kind} disabled={editor.loading}>
			<Select.Trigger id="analysis-entry-kind">{editor.kind}</Select.Trigger>
			<Select.Content>
				<Select.Group>
					{#each eventKinds as kind (kind)}<Select.Item value={kind}>{kind}</Select.Item>{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	</Field.Field>
	{#if editor.hasTimestamp}
		<Field.Field>
			<Field.Label for="analysis-entry-time">Time ({getLocalTimeZone()})</Field.Label>
			<Input
				id="analysis-entry-time"
				type="datetime-local"
				step="1"
				value={toCalendarDateTime(editor.timestamp).toString()}
				oninput={(event) => {
					if (event.currentTarget.value)
						editor.timestamp = toZoned(
							parseDateTime(event.currentTarget.value),
							getLocalTimeZone()
						);
				}}
				required
				disabled={editor.loading}
			/>
		</Field.Field>
	{/if}
	<Field.Field>
		<Field.Label for="analysis-entry-body">Description</Field.Label>
		<Textarea id="analysis-entry-body" bind:value={editor.body} rows={6} disabled={editor.loading} />
	</Field.Field>
</Field.Group>
