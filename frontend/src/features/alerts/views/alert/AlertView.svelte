<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
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

<div class="flex flex-1 items-center justify-around">
	<div class="w-42 h-full border">
		hello
	</div>
</div>