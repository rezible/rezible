<script lang="ts">
	import type { IntegrationInstallation } from "$lib/api";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as Field from "$components/ui/field";

	import SettingsFormActions from "$features/settings/components/settings-form-actions/SettingsFormActions.svelte";
	import type { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

	import SlackChannelPicker from "./SlackChannelPicker.svelte";
	import type { SlackAgentSettings } from "./slackAgentSettings";

	type Props = {
		installation: IntegrationInstallation;
		forms: IntegrationSettingsForms<SlackAgentSettings>;
	};
	const { installation, forms }: Props = $props();

	const draft = $derived(forms.draftFor(installation.id));
	const saving = $derived(forms.savingId === installation.id);
	const saveError = $derived(forms.saveError(installation.id));

	const idPrefix = $derived(`slack-agent-${installation.id}`);
</script>

{#if draft}
	{@const values = draft.value.values}
	<div class="flex flex-col gap-5 border-t pt-4">
		<Field.Set>
			<Field.Legend>Situation notices</Field.Legend>
			<Field.Group>
				<Field.Field>
					<Field.Label for="{idPrefix}-notices">Notice channel</Field.Label>
					<SlackChannelPicker
						id="{idPrefix}-notices"
						installationId={installation.id}
						bind:value={values.situationNoticeChannelId}
						noneLabel="No notice channel"
					/>
					<Field.Description>
						Optional. Each raised situation is posted here once, and the message is updated as the
						situation changes. Notices already posted stay where they are. Only public channels
						can be chosen.
					</Field.Description>
				</Field.Field>
			</Field.Group>
		</Field.Set>

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
