<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import * as Card from "$components/ui/card";
	import * as Command from "$components/ui/command";
	import * as Table from "$components/ui/table";

	let selectedPattern = $state("None selected");
</script>

<section class="flex flex-col gap-6">
	<Card.Root>
		<Card.Header>
			<Card.Title>Command search</Card.Title>
			<Card.Description
				>Interactive filtering and selection using the shared Command primitive.</Card.Description
			>
		</Card.Header>
		<Card.Content class="flex flex-col gap-3">
			<Command.Root class="rounded-lg border">
				<Command.Input placeholder="Search patterns…" />
				<Command.List>
					<Command.Empty>No patterns found.</Command.Empty>
					<Command.Group heading="Suggested">
						<Command.Item value="latency" onclick={() => (selectedPattern = "Latency increase")}
							>Latency increase</Command.Item
						>
						<Command.Item value="deploy" onclick={() => (selectedPattern = "Recent deployment")}
							>Recent deployment</Command.Item
						>
					</Command.Group>
				</Command.List>
			</Command.Root>
			<p class="text-xs text-muted-foreground" aria-live="polite">Selected: {selectedPattern}</p>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header><Card.Title>Tables and evidence</Card.Title></Card.Header>
		<Card.Content class="overflow-x-auto rounded-lg border p-0">
			<Table.Root>
				<Table.Header>
					<Table.Row
						><Table.Head>Evidence</Table.Head><Table.Head>Status</Table.Head><Table.Head
							class="text-right">Updated</Table.Head
						></Table.Row
					>
				</Table.Header>
				<Table.Body>
					<Table.Row data-state="selected">
						<Table.Cell class="font-medium">Selected deployment correlation</Table.Cell>
						<Table.Cell><Badge variant="success">Supported</Badge></Table.Cell>
						<Table.Cell class="tabular-nums text-right">14:07</Table.Cell>
					</Table.Row>
					<Table.Row>
						<Table.Cell>Standard one-line event</Table.Cell>
						<Table.Cell><Badge variant="neutral">Unreviewed</Badge></Table.Cell>
						<Table.Cell class="tabular-nums text-right">14:12</Table.Cell>
					</Table.Row>
					<Table.Row class="h-auto min-h-[60px]">
						<Table.Cell
							><p class="font-medium">Redis pool pressure rose after deployment</p>
							<p class="mt-1 text-xs text-muted-foreground">
								Metric correlation · 12 sources · cause unconfirmed
							</p></Table.Cell
						>
						<Table.Cell><Badge variant="warning">Needs evidence</Badge></Table.Cell>
						<Table.Cell class="tabular-nums text-right">14:16</Table.Cell>
					</Table.Row>
				</Table.Body>
			</Table.Root>
		</Card.Content>
	</Card.Root>
</section>
