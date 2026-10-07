<script lang="ts">
	import type { IntegrationInstallation } from "$lib/api";

	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import * as Field from "$components/ui/field";
	import { Input } from "$components/ui/input";

	import SettingsFormActions from "$features/settings/components/settings-form-actions/SettingsFormActions.svelte";
	import type { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

	import { defaultServiceLabel, type GrafanaSettings } from "./grafanaSettings";

	type Props = {
		installation: IntegrationInstallation;
		forms: IntegrationSettingsForms<GrafanaSettings>;
	};
	const { installation, forms }: Props = $props();

	const draft = $derived(forms.draftFor(installation.id));
	const saving = $derived(forms.savingId === installation.id);
	const saveError = $derived(forms.saveError(installation.id));

	const fieldId = (field: string) => `grafana-${field}-${installation.id}`;
</script>

{#if draft}
	<div class="flex flex-col gap-4">
		<div class="flex flex-col gap-1">
			<h3 class="text-sm font-medium">Data sources</h3>
			<p class="text-sm text-muted-foreground">
				Each UID is the last part of the data source's address in Grafana's Connections page:
				<span class="font-mono text-xs">/connections/datasources/edit/&lt;uid&gt;</span>
			</p>
		</div>

		<Field.Group class="grid gap-4 sm:grid-cols-2">
			<Field.Field>
				<Field.Label for={fieldId("logs-uid")}>Loki data source UID</Field.Label>
				<Input
					id={fieldId("logs-uid")}
					class="font-mono text-xs"
					bind:value={draft.value.values.logsDataSourceUid}
				/>
			</Field.Field>

			<Field.Field>
				<Field.Label for={fieldId("metrics-uid")}>Prometheus data source UID</Field.Label>
				<Input
					id={fieldId("metrics-uid")}
					class="font-mono text-xs"
					bind:value={draft.value.values.metricsDataSourceUid}
				/>
			</Field.Field>

			<Field.Field>
				<Field.Label for={fieldId("log-label")}>Log service label</Field.Label>
				<Input
					id={fieldId("log-label")}
					class="font-mono text-xs"
					placeholder={defaultServiceLabel}
					bind:value={draft.value.values.logServiceLabel}
				/>
				<Field.Description>The Loki label that names a service.</Field.Description>
			</Field.Field>

			<Field.Field>
				<Field.Label for={fieldId("metric-label")}>Metric service label</Field.Label>
				<Input
					id={fieldId("metric-label")}
					class="font-mono text-xs"
					placeholder={defaultServiceLabel}
					bind:value={draft.value.values.metricServiceLabel}
				/>
				<Field.Description>The Prometheus label that names a service.</Field.Description>
			</Field.Field>
		</Field.Group>

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
