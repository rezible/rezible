<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import type { ErrorDisplay } from "$lib/api";

	import { LoginViewController } from "./controller.svelte";

	const view = new LoginViewController();
</script>

{#snippet notice(display: ErrorDisplay)}
	<Alert.Root variant="destructive">
		<Alert.Title class="font-semibold text-sm">{display.title}</Alert.Title>
		{#if display.detail}
			<Alert.Description>{display.detail}</Alert.Description>
		{/if}
	</Alert.Root>
{/snippet}

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
				{#if view.sessionNotice}
					{@render notice(view.sessionNotice)}
				{/if}

				{#if view.loginError}
					{@render notice(view.loginError)}
				{/if}

				<Button
					color="primary"
					onclick={() => {
						view.doLogin();
					}}
					class="cursor-pointer w-full"
				>
					<span class="flex items-center gap-2">Continue</span>
				</Button>
			</Card.Content>
		{/if}
	</Card.Root>
</div>
