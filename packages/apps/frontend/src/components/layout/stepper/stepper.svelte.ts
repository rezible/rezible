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

export class StepperController {
	steps = $state.raw<StepperStepState[]>([]);

	private onFinish?: StepperAction;
	private currentIndexState = $state(0);

    private setCurrentIndex(idx: number) {
		this.errorMessage = undefined;
        this.currentIndexState = Math.max(0, Math.min(idx, this.steps.length - 1))
    }

	pending = $state(false);
	errorMessage = $state<string>();

	currentIndex = $derived(this.currentIndexState);
	progress = $derived(this.steps.length === 0 ? 0 : ((this.currentIndex + 1) / this.steps.length) * 100);

	private isFirst = $derived(this.currentIndex === 0);
	private isLast = $derived(this.currentIndex === this.steps.length - 1);

    private currentStep = $derived(this.steps[this.currentIndex]);
    currentComponent = $derived(this.currentStep.component);

	canGoBack = $derived(!this.pending && !this.isFirst);
	canContinue = $derived.by(() => {
		if (this.pending || !this.currentStep) return false;
		return this.currentStep.canContinue?.() ?? true;
	});
    continueButtonText = $derived(this.pending ? "Saving" : (this.isLast ? "Finish" : "Next"));

	constructor(cfg: StepperConfig) {
		if (cfg.steps.length === 0) {
			throw new Error("StepperController requires at least one step");
		}

		this.steps = cfg.steps.map((step, index) => ({ ...step, key: `${index}:${step.label}` }));
		this.onFinish = cfg.onFinish;
        this.setCurrentIndex(cfg.initialStepIndex ?? 0);
	}

	async next() {
		if (!this.canContinue) return;

		if (this.isLast) {
			await this.finish();
			return;
		}

		const step = this.currentStep;
		await this.runTransition(async () => {
			await step.onNext?.();
            this.setCurrentIndex(this.currentIndex + 1);
		});
	};

	back() {
		if (!this.canGoBack) return;
        this.setCurrentIndex(this.currentIndex - 1);
	};

	goTo(index: number) {
		if (this.pending) return;
		if (index < 0 || index > this.currentIndexState) return;
        this.setCurrentIndex(index);
	};

	reset() {
		if (this.pending) return;
        this.setCurrentIndex(0);
	};

	private async finish() {
		if (!this.canContinue) return;

		const step = this.currentStep;
		await this.runTransition(async () => {
			await step.onNext?.();
			await this.onFinish?.();
		});
	};

	private async runTransition(action: StepperAction) {
		this.pending = true;
		this.errorMessage = undefined;

		try {
			await action();
		} catch (err) {
            let errMsg = "";
            if (err instanceof Error) errMsg = err.message;
            if (typeof err === "string") errMsg = err;
			this.errorMessage = !!errMsg ? errMsg : "Something went wrong. Try again.";
		} finally {
			this.pending = false;
		}
	};
}
