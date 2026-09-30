<script lang="ts">
	import type { IntegrationInstallation } from "$lib/api";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as Field from "$components/ui/field";
	import { Input } from "$components/ui/input";
	import * as NativeSelect from "$components/ui/native-select";
	import { Switch } from "$components/ui/switch";

	import SettingsFormActions from "$features/settings/components/settings-form-actions/SettingsFormActions.svelte";
	import type { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

	import { inviteModeOptions, type SlackIncidentSettings } from "./slackIncidentSettings";

	type Props = {
		installation: IntegrationInstallation;
		forms: IntegrationSettingsForms<SlackIncidentSettings>;
	};
	const { installation, forms }: Props = $props();

	const draft = $derived(forms.draftFor(installation.id));
	const errors = $derived(forms.fieldErrors(installation.id));
	const hasErrors = $derived(Object.keys(errors).length > 0);
	const saving = $derived(forms.savingId === installation.id);
	const saveError = $derived(forms.saveError(installation.id));

	const idPrefix = $derived(`slack-incidents-${installation.id}`);
</script>

{#if draft}
	{@const values = draft.value.values}
	<div class="flex flex-col gap-5 border-t pt-4">
		<Field.Set>
			<Field.Legend>Incident channels</Field.Legend>
			<Field.Group>
				<Field.Field data-invalid={!!errors.ChannelNamePattern}>
					<Field.Label for="{idPrefix}-pattern">Channel name pattern</Field.Label>
					<Input
						id="{idPrefix}-pattern"
						bind:value={values.ChannelNamePattern}
						aria-invalid={!!errors.ChannelNamePattern}
					/>
					<Field.Description>
						Use <code>{"{slug}"}</code>
						,
						<code>{"{id}"}</code>
						, or
						<code>{"{title}"}</code>
						to include incident details.
					</Field.Description>
					<Field.Error>{errors.ChannelNamePattern}</Field.Error>
				</Field.Field>

				<Field.Field data-invalid={!!errors.AnnouncementChannelID}>
					<Field.Label for="{idPrefix}-announcement">Announcement channel ID</Field.Label>
					<Input
						id="{idPrefix}-announcement"
						placeholder="C0123456789"
						bind:value={values.AnnouncementChannelID}
						aria-invalid={!!errors.AnnouncementChannelID}
					/>
					<Field.Description>
						Optional. New incidents are announced in this channel.
					</Field.Description>
					<Field.Error>{errors.AnnouncementChannelID}</Field.Error>
				</Field.Field>

				<Field.Field data-invalid={!!errors.InviteMode}>
					<Field.Label for="{idPrefix}-invite">Invite to incident channels</Field.Label>
					<NativeSelect.Root id="{idPrefix}-invite" class="w-full" bind:value={values.InviteMode}>
						{#each inviteModeOptions as option (option.value)}
							<NativeSelect.Option value={option.value}>{option.label}</NativeSelect.Option>
						{/each}
					</NativeSelect.Root>
					<Field.Error>{errors.InviteMode}</Field.Error>
				</Field.Field>

				<Field.Field orientation="horizontal">
					<Field.Content>
						<Field.Label for="{idPrefix}-video">Create a video conference</Field.Label>
						<Field.Description>
							Requires an installed video conferencing integration, such as Google Workspace.
						</Field.Description>
					</Field.Content>
					<Switch id="{idPrefix}-video" bind:checked={values.AutoCreateVideoConference} />
				</Field.Field>
			</Field.Group>
		</Field.Set>

		{#if saveError}
			<InlineAlert error={saveError} dismissable={false} />
		{/if}

		<SettingsFormActions
			dirty={draft.dirty}
			pending={saving}
			saveDisabled={hasErrors}
			onSave={() => forms.save(installation)}
			onCancel={() => forms.cancel(installation.id)}
		/>
	</div>
{/if}
