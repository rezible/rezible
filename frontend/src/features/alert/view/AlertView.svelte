<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import AlertOverview from "./overview/AlertOverview.svelte";
	import { AlertViewState, alertViewStateCtx } from "./viewState.svelte";

	type Props = {
		id: string;
	};
	const { id }: Props = $props();

	const viewState = new AlertViewState(() => id);
	alertViewStateCtx.set(viewState);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Alerts" },
		{ label: viewState.alertTitle, href: `/alerts/${id}` },
	]);
</script>

<div class="flex-1">
	<AlertOverview />
</div>