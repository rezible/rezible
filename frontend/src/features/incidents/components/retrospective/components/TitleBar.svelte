<script lang="ts">
	import Avatar from '$components/avatar/Avatar.svelte';
	import { Header } from 'svelte-ux';
	import { collaborationState } from '../../../lib/collaboration.svelte';
</script>

<Header title="Incident Title" classes={{ root: 'h-10', title: 'text-2xl' }}>
	<div slot="actions">
		<div class="flex items-center gap-2">
			<div class="flex items-center -space-x-2">
				{#each collaborationState.awareness as usr, i}
					{#if i < 3}
						<Avatar kind="user" id={usr.name} badge={usr.active} />
					{/if}
				{/each}
			</div>
			<!--
            {#if $collaboration.connectionStatus !== WebSocketStatus.Disconnected}
                <div class="w-4 h-4 rounded-full {connecting ? "bg-primary animate-pulse" : "bg-success"}" />
            {/if}
            -->
		</div>
	</div>
</Header>
