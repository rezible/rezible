import { describe, expect, test } from "bun:test";

import { alertmanagerSettings, defaultServiceLabels, routingExample } from "./alertmanagerSettings";

const fieldError = (serviceLabels: string) => {
	const result = alertmanagerSettings.schema.safeParse({ serviceLabels });
	if (result.success) {
		return undefined;
	}
	return result.error.issues.at(0)?.message;
};

describe("alertmanager service labels", () => {
	test("missing settings show the backend defaults", () => {
		expect(alertmanagerSettings.read({}).serviceLabels).toBe(defaultServiceLabels.join(", "));
	});

	test("a saved empty list stays empty", () => {
		expect(alertmanagerSettings.read({ service_labels: [] }).serviceLabels).toBe("");
	});

	test("saving writes an ordered list and keeps other settings", () => {
		const saved = alertmanagerSettings.write({ serviceLabels: "job,\n service  app" }, { other: true });
		expect(saved).toEqual({ other: true, service_labels: ["job", "service", "app"] });
	});

	test("saving an empty field turns service attachment off", () => {
		expect(alertmanagerSettings.write({ serviceLabels: " " }, {})).toEqual({ service_labels: [] });
	});

	test("duplicate and invalid label names are rejected before saving", () => {
		expect(fieldError("service, app")).toBeUndefined();
		expect(fieldError("")).toBeUndefined();
		expect(fieldError("service, service")).toBeDefined();
		expect(fieldError("service, app.name")).toBeDefined();
		expect(fieldError("1service")).toBeDefined();
	});
});

describe("alertmanager routing example", () => {
	test("fills in the webhook URL once generated", () => {
		const url = "https://api.test/webhooks/alertmanager/token";
		expect(routingExample(url)).toContain(`- url: ${url}`);
		expect(routingExample()).toContain("- url: <webhook url>");
	});

	test("routes every alert to Rezible first and keeps the old tree's receiver", () => {
		const example = routingExample();
		expect(example).toContain("send_resolved: true");
		expect(example).toContain("max_alerts: 0");
		const rezibleRoute = example.indexOf("- receiver: rezible");
		const previousRoot = example.lastIndexOf("- receiver: existing-default");
		expect(rezibleRoute).toBeGreaterThan(-1);
		expect(previousRoot).toBeGreaterThan(rezibleRoute);
		expect(example.slice(rezibleRoute, previousRoot)).toContain("continue: true");
	});
});
