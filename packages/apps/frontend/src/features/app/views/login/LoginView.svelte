<script lang="ts">
	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";
    import { Input } from "$components/ui/input";
    import { Label } from "$components/ui/label";
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	
	import RiGoogleFill from "remixicon-svelte/icons/google-fill";
	import RiKeyLine from "remixicon-svelte/icons/key-2-line";
	import RiArrowLeft from "remixicon-svelte/icons/arrow-left-line";

	import { LoginViewController } from "./controller.svelte";
	import { cn } from "$src/lib/utils";

	const view = new LoginViewController();
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
				{#if view.showSSO}
					<div class="flex w-full max-w-sm flex-col gap-1.5">
						<Label for="sso-email">Email</Label>
						<Input type="email" id="sso-email" bind:value={view.ssoEmail} placeholder="Email" />
					</div>
				{:else}
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
					{:else}
						<Button color="primary" onclick={() => {view.startGoogleFlow()}} class="cursor-pointer w-full">
							<span class="flex items-center gap-2">
								<RiGoogleFill />
								Continue with Google
							</span>
						</Button>
					{/if}
				{/if}
			</Card.Content>

			<Card.Footer class={cn("p-2 flex", view.showSSO ? "justify-between" : "justify-end")}>
				<Button variant="ghost" onclick={() => {view.toggleSSO()}} class="">
					{#if view.showSSO}
						<RiArrowLeft /> 
						Go Back
					{:else}
						Continue with SSO
						<!-- <RiKeyLine /> -->
					{/if}
				</Button>
				{#if view.showSSO}
					<Button color="primary" class="cursor-pointer" 
						disabled={!view.ssoEmailValid}
						onclick={() => {view.startSSOFlow()}}
					>
						Sign In
					</Button>
				{/if}
			</Card.Footer>
		{/if}
	</Card.Root>
</div>