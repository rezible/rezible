import { z } from "zod";

const invalidKeyMessage = "Choose a service account key file downloaded from Google Cloud.";

const serviceAccountKeySchema = z.looseObject(
	{
		type: z.literal("service_account", { error: invalidKeyMessage }),
		client_email: z.string({ error: "The key file is missing client_email." }).min(1, {
			error: "The key file is missing client_email.",
		}),
		private_key: z.string({ error: "The key file is missing private_key." }).min(1, {
			error: "The key file is missing private_key.",
		}),
	},
	{ error: invalidKeyMessage }
);

type ServiceAccountKey = z.infer<typeof serviceAccountKeySchema>;

// Service account credentials stay in this form only, and are cleared after installation.
// Replacing a key installs again with the connection's customer ID, which keeps its settings.
export class GoogleInstallForm {
	customerId = $state("");
	customerIdTouched = $state(false);
	readonly replacing: boolean;

	fileName = $state<string>();
	fileError = $state<string>();
	// Changing the key recreates the file input so it forgets the previous selection.
	fileInputKey = $state(0);
	private credentials = $state.raw<ServiceAccountKey>();
	private fileReadCount = 0;

	customerIdError = $derived(
		this.customerId.trim() ? undefined : "Enter your Google Workspace customer ID."
	);
	hasCredentials = $derived(!!this.credentials);
	canSubmit = $derived(!this.customerIdError && this.hasCredentials);

	constructor(replaceCustomerId?: string) {
		this.replacing = !!replaceCustomerId;
		this.customerId = replaceCustomerId ?? "";
	}

	selectFile = async (file: File | undefined) => {
		this.fileReadCount += 1;
		const readCount = this.fileReadCount;

		this.credentials = undefined;
		this.fileError = undefined;
		this.fileName = file?.name;
		if (!file) return;

		let text: string;
		try {
			text = await file.text();
		} catch {
			text = "";
		}

		// A newer selection replaced this file while it was being read.
		if (readCount !== this.fileReadCount) return;

		let parsed: unknown;
		try {
			parsed = JSON.parse(text);
		} catch {
			this.fileError = "This file is not valid JSON.";
			return;
		}

		const result = serviceAccountKeySchema.safeParse(parsed);
		if (!result.success) {
			this.fileError = result.error.issues.at(0)?.message ?? invalidKeyMessage;
			return;
		}
		this.credentials = result.data;
	};

	clearFile = () => {
		this.selectFile(undefined);
		this.fileInputKey += 1;
	};

	reset() {
		if (!this.replacing) {
			this.customerId = "";
		}
		this.customerIdTouched = false;
		this.clearFile();
	}

	installConfig() {
		if (!this.canSubmit) {
			return undefined;
		}
		return {
			CustomerID: this.customerId.trim(),
			ServiceAccountCredentials: this.credentials,
		};
	}
}
