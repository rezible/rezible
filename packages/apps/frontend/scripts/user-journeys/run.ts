import { exists, mkdir, mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import type { JourneyPlan } from "./types";
import { execute } from "./execute";

const directory = import.meta.dir;
const safeName = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

const validateViewportDimensions = (width: number, height: number) => {
	if (![width, height].every((value) => Number.isInteger(value) && value >= 1 && value <= 16384)) {
		throw new Error("Viewport dimensions must be integers between 1 and 16384");
	}
}

const main = async () => {
	const args = Bun.argv.slice(2);
	const child = args[0] === "--execute";
	if (child) {
		args.shift();
	}
	const available = Array.from(new Bun.Glob("*.ts").scanSync(join(directory, "plans")))
		.map((file) => file.slice(0, -3))
		.sort();
	const names = args.length ? args : available;
	if (!names.length) {
		throw new Error("No journey plans found");
	}
	if (new Set(names).size !== names.length) {
		throw new Error("Duplicate journey names");
	}
	if (!process.env.APP_URL) {
		throw new Error("APP_URL is required");
	}
	const appUrl = new URL(process.env.APP_URL);
	if (!["http:", "https:"].includes(appUrl.protocol)) {
		throw new Error("APP_URL must use HTTP or HTTPS");
	}

	const plans = new Map<string, JourneyPlan>();
	for (const name of names) {
		if (!safeName.test(name) || !available.includes(name)) {
			throw new Error(`Unknown journey: ${name}`);
		}
		const plan: JourneyPlan = (await import(join(directory, "plans", `${name}.ts`))).default;
		validateViewportDimensions(plan.viewport.width, plan.viewport.height);
		if (!plan.steps.length) {
			throw new Error(`${name}: plan has no steps`);
		}
		const screenshots = new Set<string>();
		for (const step of plan.steps) {
			if (
				![
					"navigate",
					"waitFor",
					"click",
					"fill",
					"press",
					"scrollTo",
					"resize",
					"screenshot",
				].includes(step.action)
			) {
				throw new Error(`${name}: unknown action`);
			}
			const timeout = step.timeoutMs ?? plan.timeoutMs ?? 10000;
			if (!Number.isFinite(timeout) || timeout <= 0) {
				throw new Error(`${name}: invalid timeout`);
			}
			if (step.action === "resize") {
				validateViewportDimensions(step.width, step.height);
			}
			if (step.action === "screenshot") {
				if (!safeName.test(step.name) || step.name === "failure" || screenshots.has(step.name)) {
					throw new Error(`${name}: invalid or duplicate screenshot name: ${step.name}`);
				}
				screenshots.add(step.name);
			}
		}
		plans.set(name, plan);
	}
	if (child) {
		if (names.length !== 1) {
			throw new Error("--execute requires exactly one plan");
		}
		const name = names[0];
		const profileDirectory = process.env.JOURNEY_PROFILE_DIR;
		if (!profileDirectory) {
			throw new Error("--execute requires JOURNEY_PROFILE_DIR");
		}
		const output = join(directory, "output", name);
		await rm(output, { recursive: true, force: true });
		await mkdir(output, { recursive: true });
		try {
			await execute(name, plans.get(name)!, appUrl, output, profileDirectory);
		} finally {
			Bun.WebView.closeAll();
		}
		return;
	}
	for (const name of names) {
		// Chrome holds its profile directory until the child exits, so the parent removes it.
		const profileDirectory = await mkdtemp(join(tmpdir(), "rezible-journey-"));
		try {
			const childProcess = Bun.spawn([process.execPath, import.meta.path, "--execute", name], {
				stdin: "ignore",
				stdout: "inherit",
				stderr: "inherit",
				env: { ...process.env, JOURNEY_PROFILE_DIR: profileDirectory },
			});
			if ((await childProcess.exited) !== 0) {
				throw new Error(`Journey ${name} failed; remaining plans were not run`);
			}
		} finally {
			// Chrome's helper processes can still write to the profile briefly after the child exits.
			do {
				await rm(profileDirectory, { recursive: true, force: true });
				await Bun.sleep(250);
			} while (await exists(profileDirectory));
		}
	}
}

try {
	await main();
} catch (error) {
	console.error(error instanceof Error ? error.message : String(error));
	process.exitCode = 1;
}
