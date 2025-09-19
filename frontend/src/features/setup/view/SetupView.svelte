<script lang="ts">
	import { useAuthSessionState } from "$lib/auth.svelte";
	import { appShell } from "$features/app-shell";
	import Button from "$src/components/button/Button.svelte";
	import { createMutation } from "@tanstack/svelte-query";
	import { finishOrganizationSetupMutation } from "$src/lib/api";

	const session = useAuthSessionState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Initial Setup", href: "/setup" },
	]);

	const updateOrgMut = createMutation(() => finishOrganizationSetupMutation());
	const finishSetup = async () => {
		await updateOrgMut.mutateAsync({});
		session.refetch();
	}
</script>

<span>setup</span>

<Button onclick={finishSetup} loading={updateOrgMut.isPending}>Finish setup</Button>