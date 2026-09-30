<script lang="ts">
	import type { IntegrationInstallation } from "$lib/api";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as Field from "$components/ui/field";
	import { Switch } from "$components/ui/switch";

	import SettingsFormActions from "$features/settings/components/settings-form-actions/SettingsFormActions.svelte";
	import type { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

	import type { GoogleSettings } from "./googleSettings";

	type Props = {
		installation: IntegrationInstallation;
		forms: IntegrationSettingsForms<GoogleSettings>;
	};
	const { installation, forms }: Props = $props();

	const draft = $derived(forms.draftFor(installation.id));
	const saving = $derived(forms.savingId === installation.id);
	const saveError = $derived(forms.saveError(installation.id));

	const switchId = $derived(`google-video-conference-${installation.id}`);
</script>

{#if draft}
	<div class="flex flex-col gap-4 border-t pt-4">
		<Field.Field orientation="horizontal">
			<Field.Content>
				<Field.Label for={switchId}>Create Google Meet conferences for incidents</Field.Label>
				<Field.Description>New incidents get a Google Meet link.</Field.Description>
			</Field.Content>
			<Switch id={switchId} bind:checked={draft.value.values.EnableVideoConference} />
		</Field.Field>

		{#if saveError}
			<InlineAlert error={saveError} dismissable={false} />
		{/if}

		<SettingsFormActions
			dirty={draft.dirty}
			pending={saving}
			onSave={() => forms.save(installation)}
			onCancel={() => forms.cancel(installation.id)}
		/>
	</div>
{/if}
