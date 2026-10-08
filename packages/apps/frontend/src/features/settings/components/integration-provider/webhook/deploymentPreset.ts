export const webhookIntegrationName = "webhook";

export type WebhookPreset = {
	value: string;
	label: string;
};

// Payload formats Rezible defines. An installation's preset is chosen when it is installed and never changes.
export const webhookPresets: WebhookPreset[] = [{ value: "deployment", label: "Deployments" }];

export const defaultWebhookPreset = "deployment";

// An installed webhook's preset comes from its metadata. An unknown preset is shown as sent.
export const installedPresetLabel = (metadata: Record<string, string> | undefined) => {
	const preset = metadata?.preset ?? "";
	return webhookPresets.find((candidate) => candidate.value === preset)?.label ?? preset;
};

export type DeploymentReportField = {
	name: string;
	rule: string;
};

export const deploymentReportGeneralRules = [
	"Each report is a JSON object posted to the webhook URL. Reports with the same id, service and environment describe one deployment as it moves from started to succeeded or failed.",
	"Send every field in each report. The latest report replaces earlier ones, so a field it leaves out is not kept from an earlier report.",
	"String values are trimmed. An empty string or null counts as missing: allowed for an optional field, rejected for a required one. Unknown fields are ignored.",
	"A rejected report gets a 400 response naming the field to fix, so it shows in the pipeline log.",
];

export const deploymentReportFields: DeploymentReportField[] = [
	{
		name: "id",
		rule: "Required. Your ID for this deployment, at most 200 characters, used exactly as sent.",
	},
	{
		name: "service",
		rule: "Required, at most 200 characters, and must contain a letter or digit. It is lower-cased and every run of other characters becomes a hyphen, so Checkout_API and checkout-api name one service.",
	},
	{
		name: "environment",
		rule: "Required, at most 200 characters, and must contain a letter or digit. Matched the same way as service.",
	},
	{ name: "status", rule: "Required. One of started, succeeded or failed." },
	{
		name: "repository",
		rule: "Optional. owner/name with exactly one slash, each part 1 to 100 letters, digits, '.', '_' or '-', such as acme/checkout. Case is ignored.",
	},
	{ name: "sha", rule: "Optional. The commit deployed, as 40 hexadecimal characters." },
	{ name: "version", rule: "Optional. The version or tag deployed, at most 200 characters." },
	{
		name: "url",
		rule: "Optional. An http or https link, such as the pipeline run, at most 2,000 characters.",
	},
	{
		name: "started_at",
		rule: "Optional RFC3339 time, no more than 5 minutes after Rezible receives the report. A started report without it starts when Rezible receives it.",
	},
	{
		name: "finished_at",
		rule: "Optional RFC3339 time, no more than 5 minutes after Rezible receives the report. Not accepted on a started report. A succeeded or failed report without it finishes when Rezible receives it. It must not be before started_at.",
	},
];

export const deploymentStepNotes = [
	"Store the webhook URL as the repository secret REZIBLE_DEPLOYMENTS_URL, then add the report step after your deploy step.",
	"The report runs only when the deploy step ran. A cancelled deploy reports failed.",
	"github.sha is the commit that triggered the workflow. If the build you deploy comes from another commit, send that commit instead.",
	"In a monorepo, send one report for each service deployed.",
	"Reported deployments will appear among a situation's changes once the changes view is available.",
];

export const githubActionsStep = `- name: Deploy
  id: deploy
  run: ./deploy.sh # your existing deploy step

- name: Report deployment to Rezible
  if: \${{ always() && steps.deploy.outcome != 'skipped' }}
  env:
    REZIBLE_DEPLOYMENTS_URL: \${{ secrets.REZIBLE_DEPLOYMENTS_URL }}
  run: |
    curl --fail-with-body -sS --retry 3 -X POST "$REZIBLE_DEPLOYMENTS_URL" \\
      -H 'Content-Type: application/json' \\
      -d '{
        "id": "\${{ github.run_id }}-\${{ github.run_attempt }}",
        "service": "checkout-api",
        "environment": "production",
        "status": "\${{ steps.deploy.outcome == 'success' && 'succeeded' || 'failed' }}",
        "repository": "\${{ github.repository }}",
        "sha": "\${{ github.sha }}",
        "url": "\${{ github.server_url }}/\${{ github.repository }}/actions/runs/\${{ github.run_id }}"
      }'
`;
