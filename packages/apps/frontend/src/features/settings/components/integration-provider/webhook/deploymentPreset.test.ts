import { describe, expect, test } from "bun:test";

import {
	deploymentReportFields,
	deploymentReportGeneralRules,
	deploymentStepNotes,
	githubActionsStep,
	installedPresetLabel,
	webhookPresets,
} from "./deploymentPreset";

// Components load as inert stubs in unit tests and there is no rendering harness. These tests check the
// content WebhookProvider.svelte renders, and source-wiring checks confirm the component's markup uses it;
// they read the source text and do not render the page.
const component = await Bun.file(new URL("./WebhookProvider.svelte", import.meta.url)).text();

const ruleFor = (name: string) => {
	const field = deploymentReportFields.find((candidate) => candidate.name === name);
	expect(field).toBeDefined();
	return field!.rule;
};

describe("webhook presets", () => {
	test("the only preset is Deployments", () => {
		expect(webhookPresets).toEqual([{ value: "deployment", label: "Deployments" }]);
	});

	test("an installation is labelled with the preset in its metadata", () => {
		expect(installedPresetLabel({ preset: "deployment" })).toBe("Deployments");
		expect(installedPresetLabel({ preset: "incident" })).toBe("incident");
		expect(installedPresetLabel(undefined)).toBe("");
	});
});

describe("deployment report reference", () => {
	test("names every field of the deployment preset", () => {
		const names = deploymentReportFields.map((field) => field.name);
		expect(names).toEqual([
			"id",
			"service",
			"environment",
			"status",
			"repository",
			"sha",
			"version",
			"url",
			"started_at",
			"finished_at",
		]);
	});

	test("states each field's rules", () => {
		expect(ruleFor("id")).toMatch(/Required.*200 characters/);
		expect(ruleFor("service")).toMatch(
			/Required, at most 200 characters, and must contain a letter or digit/
		);
		expect(ruleFor("environment")).toMatch(
			/Required, at most 200 characters, and must contain a letter or digit/
		);
		expect(ruleFor("status")).toMatch(/started, succeeded or failed/);
		expect(ruleFor("repository")).toMatch(
			/exactly one slash, each part 1 to 100 letters, digits, '\.', '_' or '-'/
		);
		expect(ruleFor("sha")).toMatch(/40 hexadecimal characters/);
		expect(ruleFor("version")).toMatch(/at most 200 characters/);
		expect(ruleFor("url")).toMatch(/http or https.*2,000 characters/);
	});

	test("says which time defaults for each status", () => {
		expect(ruleFor("started_at")).toMatch(/A started report without it starts when Rezible receives it/);
		expect(ruleFor("finished_at")).toMatch(/Not accepted on a started report/);
		expect(ruleFor("finished_at")).toMatch(
			/A succeeded or failed report without it finishes when Rezible receives it/
		);
		expect(ruleFor("finished_at")).toMatch(/must not be before started_at/);
		for (const name of ["started_at", "finished_at"]) {
			expect(ruleFor(name)).toMatch(
				/RFC3339 time, no more than 5 minutes after Rezible receives the report/
			);
		}
	});

	test("explains trimming, empty and null values, and unknown fields", () => {
		const general = deploymentReportGeneralRules.join(" ");
		expect(general).toMatch(/trimmed/);
		expect(general).toMatch(/empty string or null counts as missing/);
		expect(general).toMatch(/Unknown fields are ignored/);
		expect(general).toMatch(/Send every field in each report\. The latest report replaces earlier ones/);
	});
});

describe("GitHub Actions step", () => {
	test("reports only when the deploy step ran, mapping its outcome to a status", () => {
		expect(githubActionsStep).toContain("  id: deploy\n");
		expect(githubActionsStep).toContain("if: ${{ always() && steps.deploy.outcome != 'skipped' }}");
		expect(githubActionsStep).toContain(
			`"status": "\${{ steps.deploy.outcome == 'success' && 'succeeded' || 'failed' }}"`
		);
	});

	test("posts to the URL stored as a repository secret", () => {
		expect(githubActionsStep).toContain(
			"REZIBLE_DEPLOYMENTS_URL: ${{ secrets.REZIBLE_DEPLOYMENTS_URL }}"
		);
		expect(githubActionsStep).toContain(
			`curl --fail-with-body -sS --retry 3 -X POST "$REZIBLE_DEPLOYMENTS_URL"`
		);
	});

	test("sends the run, repository and commit", () => {
		expect(githubActionsStep).toContain(`"id": "\${{ github.run_id }}-\${{ github.run_attempt }}"`);
		expect(githubActionsStep).toContain(`"repository": "\${{ github.repository }}"`);
		expect(githubActionsStep).toContain(`"sha": "\${{ github.sha }}"`);
		expect(githubActionsStep).toContain(
			`"url": "\${{ github.server_url }}/\${{ github.repository }}/actions/runs/\${{ github.run_id }}"`
		);
	});

	test("comes with the required notes", () => {
		const notes = deploymentStepNotes.join(" ");
		expect(notes).toMatch(/repository secret REZIBLE_DEPLOYMENTS_URL/);
		expect(notes).toMatch(/runs only when the deploy step ran/);
		expect(notes).toMatch(/cancelled deploy reports failed/);
		expect(notes).toMatch(/github\.sha is the commit that triggered the workflow/);
		expect(notes).toMatch(/monorepo, send one report for each service/);
		expect(notes).toMatch(/situation's changes/);
	});
});

describe("WebhookProvider.svelte source wiring", () => {
	test("the install form asks for the preset with a select, and is rendered", () => {
		expect(component).toContain('<Select.Root type="single" bind:value={webhook.preset}>');
		expect(component).toContain("{#each webhook.presets as preset (preset.value)}");
		expect(component).toContain('{@render installForm("Install webhook")}');
	});

	test("each installation shows the preset from its metadata", () => {
		expect(component).toContain("settings={connectionBody}");
		expect(component).toContain("{installedPresetLabel(installation.attributes.metadata)}");
	});

	test("the reference, step and notes are rendered", () => {
		expect(component).toContain("{#each deploymentReportGeneralRules as rule (rule)}");
		expect(component).toContain("{#each deploymentReportFields as field (field.name)}");
		expect(component).toContain("{githubActionsStep}");
		expect(component).toContain("{#each deploymentStepNotes as note (note)}");
		expect(component).toContain("{@render deploymentReference()}");
	});
});
