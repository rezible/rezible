<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Progress } from "$components/ui/progress";
	import { cn } from "$lib/utils";
	import type { StepperController } from "./stepper.svelte";

	type Props = {
		controller: StepperController;
	};

	const { controller }: Props = $props();

	const ActiveComponent = $derived(controller.currentStep?.component);
</script>

<div class="flex w-full max-w-3xl flex-col gap-4">
	<div class="flex flex-col gap-3">
		<ol class="flex flex-col gap-2 sm:flex-row" aria-label="Setup progress">
			{#each controller.steps as step, i (step.key)}
				{@const isActive = i === controller.currentIndex}
				{@const isCompleted = i < controller.currentIndex}
				<li class="min-w-0 flex-1" aria-current={isActive ? "step" : undefined}>
					<button
						type="button"
						class="flex w-full min-w-0 items-start gap-2 text-left disabled:pointer-events-none"
						disabled={!isCompleted || controller.pending}
						onclick={() => controller.goTo(i)}
					>
						<span
							class={cn(
								"grid size-7 shrink-0 place-items-center border text-sm font-medium",
								isActive && "border-primary bg-primary text-primary-foreground",
								isCompleted && "border-primary text-primary",
								!isActive && !isCompleted && "border-border bg-muted text-muted-foreground",
							)}
						>
							{i + 1}
						</span>
						<span class="min-w-0 space-y-0.5">
							<span
								class={cn(
									"block truncate text-sm font-medium",
									isActive ? "text-foreground" : "text-muted-foreground",
								)}
							>
								{step.label}
							</span>
							{#if step.description}
								<span class="block text-sm leading-snug text-muted-foreground">{step.description}</span>
							{/if}
						</span>
					</button>
				</li>
			{/each}
		</ol>

		<Progress value={controller.progress} />
	</div>

	<section class="min-h-0 border border-border bg-card p-4 text-card-foreground">
		{#if ActiveComponent}
			<ActiveComponent />
		{/if}
	</section>

	{#if controller.error}
		<p class="text-sm text-destructive" role="alert">{controller.error}</p>
	{/if}

	<div class="flex items-center justify-between gap-3">
		<span class="text-sm text-muted-foreground">
			Step {controller.currentIndex + 1} of {controller.steps.length}
		</span>

		<div class="flex items-center gap-2">
			<Button variant="outline" onclick={controller.back} disabled={!controller.canGoBack}>Back</Button>
			<Button onclick={controller.next} disabled={!controller.canContinue}>
				{#if controller.pending}
					Saving
				{:else if controller.isLast}
					Finish
				{:else}
					Next
				{/if}
			</Button>
		</div>
	</div>
</div>
