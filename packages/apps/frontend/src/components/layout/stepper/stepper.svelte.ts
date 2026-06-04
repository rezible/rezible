import type { Component } from "svelte";

type StepperAction = () => void | Promise<void>;

export type StepperStep = {
	label: string;
	description?: string;
	component: Component;
	canContinue?: () => boolean;
	onNext?: StepperAction;
};

type StepperStepState = StepperStep & {
	key: string;
};

export type StepperConfig = {
	steps: StepperStep[];
	initialStepIndex?: number;
	onFinish?: StepperAction;
};

const messageFromError = (err: unknown) => {
	if (err instanceof Error && err.message) return err.message;
	if (typeof err === "string" && err) return err;
	return "Something went wrong. Try again.";
};

export class StepperController {
	steps = $state.raw<StepperStepState[]>([]);

	private currentIndexState = $state(0);
	private onFinish?: StepperAction;

	pending = $state(false);
	error = $state<string>();

	currentIndex = $derived(this.currentIndexState);
	currentStep = $derived(this.steps[this.currentIndexState]);
	isFirst = $derived(this.currentIndexState === 0);
	isLast = $derived(this.currentIndexState === this.steps.length - 1);
	canGoBack = $derived(!this.pending && !this.isFirst);
	progress = $derived(this.steps.length === 0 ? 0 : ((this.currentIndexState + 1) / this.steps.length) * 100);
	canContinue = $derived.by(() => {
		if (this.pending || !this.currentStep) return false;
		return this.currentStep.canContinue?.() ?? true;
	});

	constructor(cfg: StepperConfig) {
		if (cfg.steps.length === 0) {
			throw new Error("StepperController requires at least one step");
		}

		this.steps = cfg.steps.map((step, index) => ({
			...step,
			key: `${index}:${step.label}`,
		}));
		this.onFinish = cfg.onFinish;
		this.currentIndexState = Math.max(0, Math.min(cfg.initialStepIndex ?? 0, cfg.steps.length - 1));
	}

	next = async () => {
		if (!this.canContinue) return;

		if (this.isLast) {
			await this.finish();
			return;
		}

		const step = this.currentStep;
		await this.runTransition(async () => {
			await step.onNext?.();
			this.currentIndexState = Math.min(this.currentIndexState + 1, this.steps.length - 1);
		});
	};

	back = () => {
		if (!this.canGoBack) return;
		this.error = undefined;
		this.currentIndexState = Math.max(this.currentIndexState - 1, 0);
	};

	goTo = (index: number) => {
		if (this.pending) return;
		if (index < 0 || index > this.currentIndexState) return;

		this.error = undefined;
		this.currentIndexState = index;
	};

	reset = () => {
		if (this.pending) return;
		this.error = undefined;
		this.currentIndexState = 0;
	};

	private finish = async () => {
		if (!this.canContinue) return;

		const step = this.currentStep;
		await this.runTransition(async () => {
			await step.onNext?.();
			await this.onFinish?.();
		});
	};

	private runTransition = async (action: StepperAction) => {
		this.pending = true;
		this.error = undefined;

		try {
			await action();
		} catch (err) {
			this.error = messageFromError(err);
		} finally {
			this.pending = false;
		}
	};
}
