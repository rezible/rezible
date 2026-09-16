<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Input } from "$components/ui/input";
	import { Spinner } from "$components/ui/spinner";
	import * as Field from "$components/ui/field";
	import { useSituationInvestigationController } from "./controller.svelte";

	const controller = useSituationInvestigationController();
</script>

<form
	class="max-w-2xl"
	onsubmit={(e) => {
		e.preventDefault();
		void controller.runInvestigation();
	}}
>
	<Field.FieldGroup>
		<Field.Field>
			<Field.FieldLabel for="investigation-query">Question (optional)</Field.FieldLabel>
			<div class="flex flex-wrap gap-2">
				<Input
					id="investigation-query"
					class="min-w-40 flex-1"
					bind:value={controller.form.query}
					disabled={controller.form.pending}
					placeholder="What would you like to understand?"
				/>
				<Button type="submit" disabled={controller.form.pending}>
					{#if controller.form.pending}
						<Spinner data-icon="inline-start" />
					{/if}
					{controller.form.pending ? "Requesting investigation…" : "Run investigation"}
				</Button>
			</div>
		</Field.Field>
	</Field.FieldGroup>
</form>
