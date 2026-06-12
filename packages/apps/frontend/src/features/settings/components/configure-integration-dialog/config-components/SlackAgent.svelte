<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import { watchOnce } from "runed";

	import { useConfigureIntegrationDialogController } from "../controller.svelte";

	const ctrl = useConfigureIntegrationDialogController();

	watchOnce(() => ctrl.installation, inst => {
        const cfg = {
            displayName: "Slack Agent",
            config: {},
            preferences: {},
        };
        ctrl.setConfig(cfg, true);
		console.log("install", inst);
	});
</script>

{#if ctrl.installation}
	<span>edit</span>
{:else}
	<div class="flex flex-row gap-6">
		<Alert.Root class="flex-1">
			<Alert.Description>
				Sign in with Slack to connect the workspace and grant chat/user data access.
			</Alert.Description>
		</Alert.Root>
		<div class="place-self-center">
			<Button
				onclick={() => ctrl.startOAuthFlow()}
				variant="ghost"
				class="w-fit h-fit cursor-pointer p-0"
			>
				<img
					alt="Add to Slack"
					width="139px"
					height="40px"
					src="https://platform.slack-edge.com/img/add_to_slack.png"
					srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x"
				/>
			</Button>
		</div>
	</div>
{/if}