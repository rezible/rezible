import { tick } from "svelte";
import type { SourceTarget } from "./model";

export class SourceInspection {
	target = $state<SourceTarget>();
	private trigger: HTMLElement | undefined;

	open = (target: SourceTarget, trigger: HTMLElement) => {
		this.target = target;
		this.trigger = trigger;
	};

	openEvidence = (id: string, trigger: HTMLElement) => {
		this.open({ kind: "knowledgeEvidence", id }, trigger);
	};

	reset() {
		this.target = undefined;
		this.trigger = undefined;
	}

	close = async () => {
		const trigger = this.trigger;
		this.reset();
		await tick();
		if (trigger?.isConnected) {
			trigger.focus();
		}
	};
}
