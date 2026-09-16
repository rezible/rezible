<script lang="ts">
	import * as Card from "$components/ui/card";
	import * as Field from "$components/ui/field";
	import { Input } from "$components/ui/input";
	import { Button } from "$components/ui/button";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import TimezoneSelectField from "$components/forms/timezone-select-field/TimezoneSelectField.svelte";
	import { useInitialSetupController } from "./initialSetupController.svelte";

	const ctrl = useInitialSetupController();
</script>

<Card.Root class="w-full max-w-lg">
	<Card.Header>
		<h1 tabindex="-1" class="text-xl font-semibold outline-none">Set up your organization</h1>
		<Card.Description>Name your organization and choose its default timezone.</Card.Description>
	</Card.Header>

	<Card.Content>
		{#if ctrl.error}
			<div class="mb-4" role="alert">
				<InlineAlert error={ctrl.error} dismissable={false} />
			</div>
		{/if}

		{#if !ctrl.isAdmin}
			<p class="text-sm text-muted-foreground">An administrator must complete organization setup.</p>
		{:else}
			<form
				class="flex flex-col gap-5"
				onsubmit={(event) => {
					event.preventDefault();
					ctrl.finish();
				}}
			>
				<Field.FieldGroup>
					<Field.Field data-invalid={ctrl.name.trim().length === 0}>
						<Field.FieldLabel for="setup-org-name">Organization name</Field.FieldLabel>
						<Input
							id="setup-org-name"
							bind:value={ctrl.name}
							aria-invalid={ctrl.name.trim().length === 0}
							required
						/>
					</Field.Field>
					<Field.Field>
						<Field.FieldLabel for="setup-timezone">Default timezone</Field.FieldLabel>
						<TimezoneSelectField id="setup-timezone" bind:value={ctrl.timezone} />
					</Field.Field>
				</Field.FieldGroup>

				<Button type="submit" disabled={!ctrl.canSubmit}>
					{#if ctrl.saving}
						<Spinner data-icon="inline-start" />
					{:else}
						Save
					{/if}
				</Button>
			</form>
		{/if}
	</Card.Content>
</Card.Root>
