<script lang="ts">
	import type { IntegrationInstallation } from "$lib/api";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as Field from "$components/ui/field";
	import { Input } from "$components/ui/input";

	import SettingsFormActions from "$features/settings/components/settings-form-actions/SettingsFormActions.svelte";
	import type { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

	import type { AlertmanagerSettings } from "./alertmanagerSettings";

	type Props = {
		installation: IntegrationInstallation;
		forms: IntegrationSettingsForms<AlertmanagerSettings>;
	};
	const { installation, forms }: Props = $props();

	const draft = $derived(forms.draftFor(installation.id));
	const saving = $derived(forms.savingId === installation.id);
	const saveError = $derived(forms.saveError(installation.id));
	const labelsError = $derived(forms.fieldErrors(installation.id).serviceLabels);

	const inputId = $derived(`alertmanager-service-labels-${installation.id}`);
</script>

{#if draft}
	<div class="flex flex-col gap-4">
		<Field.Field data-invalid={!!labelsError}>
			<Field.Label for={inputId}>Service labels</Field.Label>
			<Input
				id={inputId}
				class="font-mono text-xs"
				bind:value={draft.value.values.serviceLabels}
				aria-invalid={!!labelsError}
			/>
			<Field.Description>
				Labels that name an alert's service, tried in order. Leave empty to attach no services.
			</Field.Description>
			<Field.Description>
				Changing service labels affects subsequent deliveries. Some active alerts may remain under
				their previous service until they expire.
			</Field.Description>
			{#if labelsError}
				<Field.Error>{labelsError}</Field.Error>
			{/if}
		</Field.Field>

		{#if saveError}
			<InlineAlert error={saveError} dismissable={false} />
		{/if}

		<SettingsFormActions
			dirty={draft.dirty}
			pending={saving}
			saveDisabled={!!labelsError}
			onSave={() => forms.save(installation)}
			onCancel={() => forms.cancel(installation.id)}
		/>
	</div>
{/if}
