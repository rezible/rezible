import type {
	Extension,
	onAuthenticatePayload,
} from "@hocuspocus/server";

// TODO: codegen this
type EditorSessionUser = {
	id: string;
	username: string;
}
type VerifyAuthSessionResponseData = {
	user: EditorSessionUser;
	readOnly: boolean;
}

export interface SessionAuthConfig {
	verifySessionUrl: string;
}

export type AuthContext = {
	user: EditorSessionUser;
}

export class SessionAuthentication implements Extension {
	extensionName: string;

	config: SessionAuthConfig = {
		verifySessionUrl: ""
	};

	constructor(config: Partial<SessionAuthConfig>) {
		this.extensionName = "Rezible Authentication";
		this.config = { ...this.config, ...config };
	}

	async onAuthenticate(data: onAuthenticatePayload): Promise<AuthContext> {
		const auth = await this.verifyDocumentAuth(data.documentName, data.token);

		data.connection.readOnly = !!auth.readOnly;

		const authCtx: AuthContext = {
			user: auth.user,
		};

		return authCtx;
	}

	async verifyDocumentAuth(documentName: string, token: string) {
		const reqBody = {attributes: {documentName}};
		const res = await fetch(this.config.verifySessionUrl, {
			method: "POST",
			headers: [["Authorization", `Bearer ${token}`]],
			body: JSON.stringify(reqBody),
		});

		if (!res.ok || res.status != 200) {
            console.log("err", res);
			console.log("error status", res.status);
			throw new Error("Authentication Failed");
		}

		const resp = await res.json();
		if (!("data" in resp)) {
			console.error("unknown auth response body: %s", resp);
			throw new Error("Invalid Auth Response");
		}
		return resp.data as VerifyAuthSessionResponseData;
	}
}