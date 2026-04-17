<script lang="ts">
	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";

	import { LoginViewController } from "./controller.svelte";

	const view = new LoginViewController();

	const isError = $derived(!!view.authSessionError || !!view.loginError);
</script>

<div class="grid h-full w-full place-items-center">
	<Card.Root class="min-w-84">
		<Card.Header class="gap-0">
			<Card.Title class="text-lg">{view.titleText}</Card.Title>
			<Card.Description>{view.descriptionText}</Card.Description>
			<Card.Action>
				<img src="/images/logo.svg" alt="logo" class="size-10 fill-neutral" />
			</Card.Action>
		</Card.Header>

		{#if view.inFlow || !view.loaded}
			<Card.Content>
				<Spinner />
			</Card.Content>
		{:else}
			<Card.Content class="flex flex-col gap-2">
			
				{#if isError}
					{#if !!view.authSessionError}
						<InlineAlert bind:error={view.authSessionError} />
					{/if}

					{#if !!view.loginError}
						<InlineAlert bind:error={view.loginError} />
					{/if}
					
					{#if view.showLogout}
						<Button onclick={() => {view.doLogout()}} color="primary">
							Logout
						</Button>
					{/if}
				{:else}
					<Button color="primary" onclick={() => {view.doLogin()}} class="cursor-pointer w-full">
						<span class="flex items-center gap-2">
							Continue
						</span>
					</Button>
				{/if}
			</Card.Content>
		{/if}
	</Card.Root>
</div>