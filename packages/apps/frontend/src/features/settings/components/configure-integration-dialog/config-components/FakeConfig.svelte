<script lang="ts">
	import { onMount } from "svelte";
	import { useConfigureIntegrationDialogController } from "../controller.svelte";

	const ctrl = useConfigureIntegrationDialogController();

    onMount(() => {
        const cfg = {
            displayName: "Fake Data Provider",
            config: {},
            preferences: {},
        };
        ctrl.setConfig(cfg, true);
    })
</script>

<p>No config needed</p>
