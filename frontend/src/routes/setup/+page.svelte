<script lang="ts">
	import { setSetupViewState } from "$features/setup/lib/viewState.svelte";
	import { SetupView } from "$features/setup";

    setSetupViewState();
</script>

<SetupView />