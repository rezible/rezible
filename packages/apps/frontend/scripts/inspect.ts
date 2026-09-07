import { mkdir } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { pathToFileURL } from "node:url";

type Interaction = (view: Bun.WebView) => Promise<void>;

type Options = {
	route: string;
	script?: string;
	waitFor: string;
	width: number;
	height: number;
	output: string;
};

function parseOptions(args: string[]): Options {
	const route = args.shift();
	if (!route)
		throw new Error(
			"usage: inspect ROUTE [--script FILE] [--wait-for SELECTOR] [--width N] [--height N] [--output FILE]"
		);
	const options: Options = {
		route,
		waitFor: "main",
		width: 1440,
		height: 900,
		output: resolve(import.meta.dir, "../.artifacts/ui/latest.png"),
	};
	while (args.length > 0) {
		const flag = args.shift();
		const value = args.shift();
		if (!value) throw new Error(`missing value for ${flag}`);
		switch (flag) {
			case "--script":
				options.script = resolve(value);
				break;
			case "--wait-for":
				options.waitFor = value;
				break;
			case "--width":
				options.width = Number(value);
				break;
			case "--height":
				options.height = Number(value);
				break;
			case "--output":
				options.output = resolve(value);
				break;
			default:
				throw new Error(`unknown option: ${flag}`);
		}
	}
	if (!Number.isInteger(options.width) || !Number.isInteger(options.height))
		throw new Error("viewport dimensions must be integers");
	return options;
}

function environmentArguments(): string[] {
	const route = process.env.INSPECT_ROUTE;
	if (!route) return Bun.argv.slice(2);
	const args = [
		route,
		"--wait-for",
		process.env.INSPECT_WAIT_FOR ?? "main",
		"--width",
		process.env.INSPECT_WIDTH ?? "1440",
		"--height",
		process.env.INSPECT_HEIGHT ?? "900",
	];
	if (process.env.INSPECT_SCRIPT) args.push("--script", process.env.INSPECT_SCRIPT);
	if (process.env.INSPECT_OUTPUT) args.push("--output", process.env.INSPECT_OUTPUT);
	return args;
}

async function inspect(options: Options): Promise<void> {
	const requestedUrl = new URL(options.route, process.env.APP_URL);
	const view = new Bun.WebView({
		width: options.width,
		height: options.height,
		backend: { type: "chrome", url: false },
		dataStore: "ephemeral",
		console: globalThis.console,
	});
	view.onNavigationFailed = (error) => console.error("navigation failed:", error.message);
	try {
		console.log(`navigating: ${requestedUrl.href}`);
		await view.navigate(requestedUrl.href);
		console.log(`navigated: ${view.url}`);
		if (options.script) {
			console.log(`interaction: ${options.script}`);
			const module = (await import(pathToFileURL(options.script).href)) as { default?: Interaction };
			if (typeof module.default !== "function")
				throw new Error(`${options.script} must default-export an async interaction function`);
			await module.default(view);
		}
		const selector = JSON.stringify(options.waitFor);
		console.log(`waiting for: ${options.waitFor}`);
		await view.evaluate(`new Promise((resolve, reject) => {
			const deadline = Date.now() + 30000;
			const check = () => {
				const element = document.querySelector(${selector});
				if (element) {
					const style = getComputedStyle(element);
					const rect = element.getBoundingClientRect();
					if (style.visibility !== "hidden" && style.display !== "none" && rect.width > 0 && rect.height > 0) return resolve(true);
				}
				if (Date.now() >= deadline) return reject(new Error("timed out waiting for visible selector ${options.waitFor.replaceAll('"', '\\"')}"));
				requestAnimationFrame(check);
			};
			check();
		})`);
		console.log(`visible: ${options.waitFor}`);
		await view.evaluate(
			"document.fonts.ready.then(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))))"
		);
		const finalUrl = new URL(view.url);
		if (finalUrl.pathname !== requestedUrl.pathname)
			throw new Error(
				`requested ${requestedUrl.pathname}, but browser finished at ${finalUrl.pathname}`
			);
		await mkdir(dirname(options.output), { recursive: true });
		await Bun.write(options.output, await view.screenshot());
		console.log(`screenshot: ${options.output}`);
		console.log(`page: ${view.url}`);
	} finally {
		view.close();
	}
}

const options = parseOptions(environmentArguments());
let timeout: ReturnType<typeof setTimeout>;
try {
	await Promise.race([
		inspect(options),
		new Promise<never>((_, reject) => {
			timeout = setTimeout(() => reject(new Error("browser work timed out after 30 seconds")), 30_000);
		}),
	]);
} finally {
	clearTimeout(timeout!);
}
