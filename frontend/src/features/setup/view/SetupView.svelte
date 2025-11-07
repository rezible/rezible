<script lang="ts">
	import { useAuthSessionState } from "$lib/auth.svelte";
	import { appShell } from "$features/app-shell";
	import Button from "$src/components/button/Button.svelte";
	import { createMutation } from "@tanstack/svelte-query";
	import { finishOrganizationSetupMutation } from "$src/lib/api";
	import Header from "$src/components/header/Header.svelte";

	const session = useAuthSessionState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Setup", href: "/setup" },
	]);

	const updateOrgMut = createMutation(() => finishOrganizationSetupMutation());
	const finishSetup = async () => {
		await updateOrgMut.mutateAsync({});
		session.refetch();
	}

	let setupComplete = $state(false);
</script>

<div class="grid h-full w-full place-items-center">
	<div class="flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Organization Setup" classes={{ root: "gap-2", title: "text-2xl" }}>
			{#snippet avatar()}
				<img src="/images/logo.svg" alt="logo" class="size-12 fill-neutral" />
			{/snippet}
		</Header>

		<a href="https://slack.com/oauth/v2/authorize?client_id=7924786649253.7927692114210&scope=app_mentions:read,assistant:write,channels:history,channels:join,channels:read,chat:write,chat:write.customize,chat:write.public,commands,groups:history,groups:read,im:history,im:read,im:write,im:write.topic,incoming-webhook,metadata.message:read,mpim:history,pins:read,reactions:read,usergroups:read,users.profile:read,users:read,users:read.email,channels:write.topic,channels:manage,channels:write.invites&user_scope=">
			<img alt="Add to Slack" height="40" width="139" 
				src="https://platform.slack-edge.com/img/add_to_slack.png" 
				srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" />
		</a>
		
		<Button 
			color="primary" 
			variant="fill"
			disabled={!setupComplete}
			onclick={finishSetup} 
			loading={updateOrgMut.isPending}
		>Finish setup</Button>
	</div>
</div>