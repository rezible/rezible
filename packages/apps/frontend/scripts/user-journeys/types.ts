import type { Pathname } from "$app/types";

export type Step = (
	| { action: "navigate"; path: Pathname; query?: Record<string, string> }
	| { action: "waitFor"; selector: string; state: "visible" | "absent" }
	| { action: "click" | "scrollTo"; selector: string }
	| { action: "fill"; selector: string; value: string }
	| { action: "press"; key: Bun.WebView.VirtualKey; modifiers?: Bun.WebView.Modifier[] }
	| { action: "resize"; width: number; height: number }
	| { action: "screenshot"; name: string; caption: string }
) & { timeoutMs?: number };

export type JourneyPlan = {
	description: string;
	prerequisites: string[];
	viewport: { width: number; height: number };
	timeoutMs?: number;
	steps: Step[];
};
