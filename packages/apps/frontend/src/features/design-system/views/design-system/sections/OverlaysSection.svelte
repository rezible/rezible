<script lang="ts">
	import RiAlertLine from "remixicon-svelte/icons/alert-line";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import * as Dialog from "$components/ui/dialog";
	import * as DropdownMenu from "$components/ui/dropdown-menu";
	import { Input } from "$components/ui/input";
	import * as Popover from "$components/ui/popover";
	import { Separator } from "$components/ui/separator";
	import * as Sheet from "$components/ui/sheet";
	import { Skeleton } from "$components/ui/skeleton";
	import * as Tooltip from "$components/ui/tooltip";

	let dialogOpen = $state(false);
	let sheetOpen = $state(false);
</script>

<section class="flex flex-col gap-6">
	<Card.Root>
		<Card.Header><Card.Title>Overlays, feedback, and loading</Card.Title></Card.Header>
		<Card.Content class="flex flex-wrap items-center gap-2">
			<Tooltip.Root>
				<Tooltip.Trigger>
					{#snippet child({ props })}<Button variant="outline" {...props}>Hover tooltip</Button
						>{/snippet}
				</Tooltip.Trigger>
				<Tooltip.Content>Keyboard-accessible context.</Tooltip.Content>
			</Tooltip.Root>
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}<Button variant="outline" {...props}>Open menu</Button
						>{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content>
					<DropdownMenu.Group>
						<DropdownMenu.Label>Saved views</DropdownMenu.Label>
						<DropdownMenu.Item>First symptoms</DropdownMenu.Item>
						<DropdownMenu.Item>Before deployment</DropdownMenu.Item>
					</DropdownMenu.Group>
				</DropdownMenu.Content>
			</DropdownMenu.Root>
			<Popover.Root>
				<Popover.Trigger>
					{#snippet child({ props })}<Button variant="outline" {...props}>Open popover</Button
						>{/snippet}
				</Popover.Trigger>
				<Popover.Content>
					<Popover.Header>
						<Popover.Title>Evidence filter</Popover.Title>
						<Popover.Description>Compact contextual controls.</Popover.Description>
					</Popover.Header>
					<Input placeholder="Filter sources" />
				</Popover.Content>
			</Popover.Root>
			<Button variant="outline" onclick={() => (dialogOpen = true)}>Open dialog</Button>
			<Button variant="outline" onclick={() => (sheetOpen = true)}>Open sheet</Button>
		</Card.Content>
	</Card.Root>

	<Dialog.Root bind:open={dialogOpen}>
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Review evidence</Dialog.Title>
				<Dialog.Description
					>A restrained overlay with preserved focus and dismissal behavior.</Dialog.Description
				>
			</Dialog.Header>
			<div class="flex items-center gap-2 text-sm text-muted-foreground">
				<RiAlertLine class="size-4 shrink-0" aria-hidden="true" />
				<span>Illustrative content only.</span>
			</div>
			<Dialog.Footer>
				<Button variant="outline" onclick={() => (dialogOpen = false)}>Close</Button>
				<Button onclick={() => (dialogOpen = false)}>Continue</Button>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>

	<Sheet.Root bind:open={sheetOpen}>
		<Sheet.Content side="right">
			<Sheet.Header>
				<Sheet.Title>Context panel</Sheet.Title>
				<Sheet.Description>Portal surface for secondary evidence.</Sheet.Description>
			</Sheet.Header>
			<Separator />
			<Skeleton class="h-24 w-full" />
		</Sheet.Content>
	</Sheet.Root>
</section>
