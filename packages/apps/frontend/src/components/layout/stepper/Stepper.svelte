<script lang="ts">
	import type { StepperController } from "./stepper.svelte";

    type Props = {
        controller: StepperController;
    }
    const { controller }: Props = $props();
</script>