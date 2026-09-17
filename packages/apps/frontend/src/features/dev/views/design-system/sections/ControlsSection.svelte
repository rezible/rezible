<script lang="ts">
	import RiCheckLine from "remixicon-svelte/icons/check-line";
	import RiLoader4Line from "remixicon-svelte/icons/loader-4-line";
	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";
	import { Checkbox } from "$components/ui/checkbox";
	import * as Field from "$components/ui/field";
	import { Input } from "$components/ui/input";
	import * as RadioGroup from "$components/ui/radio-group";
	import * as Select from "$components/ui/select";
	import { Switch } from "$components/ui/switch";
	import * as Tabs from "$components/ui/tabs";
	import { Textarea } from "$components/ui/textarea";
	import { Toggle } from "$components/ui/toggle";

	let checked = $state(true);
	let switched = $state(true);
	let selectedEvidence = $state("metrics");
</script>

<section class="flex flex-col gap-6">
	<div class="grid gap-4 lg:grid-cols-2">
		<Card.Root>
			<Card.Header><Card.Title>Actions and states</Card.Title></Card.Header>
			<Card.Content class="flex flex-wrap gap-2">
				<Button><RiCheckLine data-icon="inline-start" />Open report</Button>
				<Button variant="outline">View evidence</Button>
				<Button variant="secondary">Secondary</Button>
				<Button variant="ghost">More</Button>
				<Button variant="destructive">Delete</Button>
				<Button disabled>Disabled</Button>
				<Button disabled class="min-w-28">
					<RiLoader4Line class="animate-spin motion-reduce:animate-none" data-icon="inline-start" />
					Loading
				</Button>
				<Button aria-invalid="true" variant="outline">Invalid</Button>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header><Card.Title>Selection controls</Card.Title></Card.Header>
			<Card.Content class="flex flex-col gap-4">
				<div class="flex flex-wrap items-center gap-4">
					<label class="flex items-center gap-2"><Checkbox bind:checked />Checkbox</label>
					<RadioGroup.Root value="evidence" class="flex gap-3">
						<label class="flex items-center gap-2">
							<RadioGroup.Item value="evidence" />Evidence
						</label>
						<label class="flex items-center gap-2">
							<RadioGroup.Item value="timeline" />Timeline
						</label>
					</RadioGroup.Root>
					<label class="flex items-center gap-2"><Switch bind:checked={switched} />Switch</label>
					<Toggle pressed>Toggle</Toggle>
				</div>
				<Tabs.Root value="graph">
					<Tabs.List>
						<Tabs.Trigger value="graph">Filled tab</Tabs.Trigger>
						<Tabs.Trigger value="list">List</Tabs.Trigger>
					</Tabs.List>
				</Tabs.Root>
				<Tabs.Root value="findings">
					<Tabs.List variant="line">
						<Tabs.Trigger value="findings">Line tab</Tabs.Trigger>
						<Tabs.Trigger value="sources">Sources</Tabs.Trigger>
					</Tabs.List>
				</Tabs.Root>
			</Card.Content>
		</Card.Root>
	</div>

	<div class="grid gap-4 lg:grid-cols-2">
		<Card.Root>
			<Card.Header>
				<Card.Title>Fields</Card.Title>
				<Card.Description>Persistent labels, help, errors, and disabled values.</Card.Description>
			</Card.Header>
			<Card.Content>
				<Field.FieldGroup>
					<Field.Field>
						<Field.FieldLabel for="title">Investigation title</Field.FieldLabel>
						<Input id="title" value="Redis pressure and search timeouts" />
						<Field.FieldDescription>Use a concise human-readable summary.</Field.FieldDescription>
					</Field.Field>
					<Field.Field data-invalid>
						<Field.FieldLabel for="owner">Owner</Field.FieldLabel>
						<Input id="owner" aria-invalid="true" placeholder="Select an owner" />
						<Field.FieldError>Owner is required.</Field.FieldError>
					</Field.Field>
					<Field.Field>
						<Field.FieldLabel for="notes">Notes</Field.FieldLabel>
						<Textarea id="notes" value="Capture evidence and uncertainty here." />
					</Field.Field>
					<Field.Field data-disabled>
						<Field.FieldLabel for="source">Immutable source</Field.FieldLabel>
						<Input id="source" disabled value="Telemetry import" />
					</Field.Field>
				</Field.FieldGroup>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header><Card.Title>Select and command</Card.Title></Card.Header>
			<Card.Content class="flex flex-col gap-4">
				<Field.Field>
					<Field.FieldLabel>Evidence source</Field.FieldLabel>
					<Select.Root
						type="single"
						value={selectedEvidence}
						onValueChange={(value) => (selectedEvidence = value)}
					>
						<Select.Trigger class="w-full">
							{selectedEvidence === "logs"
								? "Logs"
								: selectedEvidence === "tickets"
									? "Support tickets"
									: "Metrics"}
						</Select.Trigger>
						<Select.Content>
							<Select.Group>
								<Select.Label>Sources</Select.Label>
								<Select.Item value="metrics">Metrics</Select.Item>
								<Select.Item value="logs">Logs</Select.Item>
								<Select.Item value="tickets">Support tickets</Select.Item>
							</Select.Group>
						</Select.Content>
					</Select.Root>
				</Field.Field>
			</Card.Content>
		</Card.Root>
	</div>
</section>
