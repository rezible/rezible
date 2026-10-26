import { join } from "node:path";
import type { JourneyPlan, Step } from "./types";

type WaitForState = "visible" | "absent";
const waitForStateEvals: Record<WaitForState, string> = {
	visible: `if (elements.length > 1) throw new Error("Selector matches multiple elements");
		return !!elements[0] && elements[0].checkVisibility({ checkOpacity: true, checkVisibilityCSS: true });`,
	absent: "return elements.length === 0;",
};
async function waitFor(view: Bun.WebView, selector: string, state: WaitForState) {
	const queryElements = `const elements = document.querySelectorAll(${JSON.stringify(selector)});`;
	const evalElementsState = waitForStateEvals[state];
	const evalScript = `(() => {${queryElements} ${evalElementsState}})()`;
	while (true) {
		if (await view.evaluate(evalScript)) return;
		await Bun.sleep(50);
	}
}

async function perform(view: Bun.WebView, step: Step, appUrl: URL, output: string, timeout: number) {
	if ("selector" in step && step.action !== "waitFor") {
		await waitFor(view, step.selector, "visible");
		await view.scrollTo(step.selector, { timeout });
	}
	switch (step.action) {
		case "navigate": {
			const url = new URL(step.path, appUrl);
			if (url.origin !== appUrl.origin) {
				throw new Error("Navigation must stay on APP_URL's origin");
			}
			for (const [key, value] of Object.entries(step.query ?? {})) {
				url.searchParams.set(key, value);
			}
			await view.navigate(url.href);
			break;
		}
		case "waitFor":
			await waitFor(view, step.selector, step.state);
			break;
		case "click":
			await view.click(step.selector, { timeout });
			break;
		case "fill":
			await view.click(step.selector, { timeout });
			await view.press("a", { modifiers: [process.platform === "darwin" ? "Meta" : "Control"] });
			await view.press("Backspace");
			await view.type(step.value);
			break;
		case "press":
			await view.press(step.key, { modifiers: step.modifiers });
			break;
		case "scrollTo":
			break;
		case "resize":
			await view.resize(step.width, step.height);
			break;
		case "screenshot":
			await view.evaluate(
				"document.fonts.ready.then(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))))"
			);
			await Bun.write(join(output, `${step.name}.png`), await view.screenshot());
			break;
		default:
			throw new Error(`Unknown action: ${JSON.stringify(step)}`);
	}
}

// CDP can capture the DOM even while a page-side evaluate() is still pending.
async function captureFailure(view: Bun.WebView, output: string) {
	const deadline = setTimeout(() => view.close(), 2000);
	try {
		await Promise.allSettled([
			(async () => {
				const document = (await view.cdp("DOM.getDocument")) as { root: { nodeId: number } };
				const { outerHTML } = (await view.cdp("DOM.getOuterHTML", {
					nodeId: document.root.nodeId,
				})) as { outerHTML: string };
				await Bun.write(join(output, "failure.html"), outerHTML);
			})(),
			(async () => {
				await Bun.write(join(output, "failure.png"), await view.screenshot());
			})(),
		]);
	} finally {
		clearTimeout(deadline);
	}
}

export async function execute(
	name: string,
	plan: JourneyPlan,
	appUrl: URL,
	output: string,
	profileDirectory: string
) {
	const report = {
		name,
		description: plan.description,
		prerequisites: plan.prerequisites,
		appUrl: appUrl.href,
		bunVersion: Bun.version,
		browser: null as unknown,
		startedAt: new Date().toISOString(),
		finishedAt: "",
		status: "running",
		error: "",
		viewport: plan.viewport,
		steps: [] as Array<Record<string, unknown>>,
		diagnostics: [] as Array<Record<string, unknown>>,
	};
	const save = () => Bun.write(join(output, "run.json"), JSON.stringify(report, null, 2));
	let view: Bun.WebView | undefined;
	let timer: ReturnType<typeof setTimeout> | undefined;
	let timedOut = false;
	let failureCapture: Promise<void> | undefined;
	const onTimeout = () => {
		timedOut = true;
		if (view) {
			const failedView = view;
			failureCapture = captureFailure(failedView, output).finally(() => failedView.close());
		}
	};
	try {
		await save();
		view = new Bun.WebView({
			// Fresh Chrome processes can refresh their verifier during startup and abort
			// HTTP/2 script requests with ERR_CERT_VERIFIER_CHANGED. Keep TLS validation
			// enabled and use HTTP/1.1 for these local development walkthroughs.
			backend: { type: "chrome", argv: ["--disable-http2"] },
			// Bun never removes Chrome's temporary profile; the parent process owns this one.
			dataStore: { directory: profileDirectory },
			...plan.viewport,
			console: (type, ...args) => {
				if (type === "error" || type === "warn") {
					report.diagnostics.push({ type, args });
				}
			},
		});
		// Establish the CDP session before navigating to the application.
		timer = setTimeout(onTimeout, plan.timeoutMs ?? 10000);
		await view.navigate("about:blank");
		// Chrome applies the constructor size to the window, not the viewport. Resize to correct it.
		await view.resize(plan.viewport.width, plan.viewport.height);
		report.browser = await view.cdp("Browser.getVersion");
		view.addEventListener("Network.responseReceived", (event) => {
			const { data } = event as unknown as MessageEvent<{ response: { status: number; url: string } }>;
			if (data.response.status >= 400) {
				report.diagnostics.push({ type: "http", ...data.response });
			}
		});
		view.addEventListener("Network.loadingFailed", (event) => {
			const { data } = event as unknown as MessageEvent<Record<string, unknown>>;
			report.diagnostics.push({ ...data, type: "network" });
		});
		view.addEventListener("Runtime.exceptionThrown", (event) => {
			const { data } = event as unknown as MessageEvent<Record<string, unknown>>;
			report.diagnostics.push({ ...data, type: "exception" });
		});
		await view.cdp("Network.enable");
		await view.cdp("Runtime.enable");
		clearTimeout(timer);
		for (const [index, step] of plan.steps.entries()) {
			console.log(`${name}: ${index + 1}/${plan.steps.length} ${step.action}`);
			const entry: Record<string, unknown> = { step: index + 1, ...step, status: "running" };
			report.steps.push(entry);
			await save();
			const timeout = step.timeoutMs ?? plan.timeoutMs ?? 10000;
			// Allow a bounded diagnostic capture before closing to cancel the timed-out operation.
			timer = setTimeout(onTimeout, timeout);
			try {
				await perform(view, step, appUrl, output, timeout);
				entry.url = await view.evaluate("location.href");
				entry.viewport = await view.evaluate("({width: innerWidth, height: innerHeight})");
				if (timedOut) {
					throw new Error("Browser operation timed out");
				}
				entry.status = "passed";
			} catch (error) {
				entry.status = "failed";
				throw error;
			} finally {
				clearTimeout(timer);
			}
			await save();
		}
		report.status = "passed";
	} catch (error) {
		report.status = "failed";
		report.error = timedOut ? "Browser operation timed out" : String(error);
		clearTimeout(timer);
		if (failureCapture) {
			await failureCapture;
		} else if (view) {
			await captureFailure(view, output);
		}
		throw new Error(`${name}, step ${report.steps.length}: ${report.error}`);
	} finally {
		clearTimeout(timer);
		view?.close();
		report.finishedAt = new Date().toISOString();
		await save();
	}
	console.log(`Saved ${output}`);
}
