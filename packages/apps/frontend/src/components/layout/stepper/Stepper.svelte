<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Progress } from "$components/ui/progress";
	import { cn } from "$lib/utils";
	import { Spinner } from "$src/components/ui/spinner";
	import type { StepperController } from "./stepper.svelte";

    type Props = {
        controller: StepperController;
    }
	const { controller }: Props = $props();

</script>

<div class={cn("flex w-full gap-4 flex-col")}>
	<div class={cn("flex flex-col gap-3 w-full")}>
		<Progress value={controller.progress} />
        
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
								"grid size-8 shrink-0 place-items-center border text-sm font-medium",
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
	</div>

    <div class="flex flex-col gap-4">
        <section class="min-h-0 border border-border bg-card p-4 text-card-foreground">
            {#if !!controller.currentComponent}
                <controller.currentComponent />
            {/if}
        </section>

        {#if controller.errorMessage}
            <p class="text-sm text-destructive" role="alert">{controller.errorMessage}</p>
        {/if}

        <div class="flex self-end items-center gap-2">
            <Button variant="outline" onclick={() => {controller.back()}} disabled={!controller.canGoBack}>Back</Button>
            <Button onclick={() => {controller.next()}} disabled={!controller.canContinue}>
                {#if controller.pending}
                    <Spinner />
                {/if}
                {controller.continueButtonText}
            </Button>
        </div>
    </div>
</div>
