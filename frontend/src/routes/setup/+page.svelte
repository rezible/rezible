<script lang="ts">
	import { setInitialSetupViewState } from "$features/settings/lib/initialSetupViewState.svelte";
	import InitialSetupView from "$features/settings/view/initial-setup/InitialSetupView.svelte";

    setInitialSetupViewState();
</script>

<InitialSetupView />