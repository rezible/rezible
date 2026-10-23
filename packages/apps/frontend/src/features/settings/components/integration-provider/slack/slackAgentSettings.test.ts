import { describe, expect, test } from "bun:test";

import { slackAgentSettings } from "./slackAgentSettings";

describe("Slack agent settings", () => {
	test("a missing notice channel reads as none", () => {
		expect(slackAgentSettings.read({})).toEqual({ situationNoticeChannelId: "" });
		expect(slackAgentSettings.read({ situation_notice_channel_id: null })).toEqual({
			situationNoticeChannelId: "",
		});
	});

	test("a saved notice channel reads back", () => {
		const settings = slackAgentSettings.read({ situation_notice_channel_id: "C0AAA1" });

		expect(settings).toEqual({ situationNoticeChannelId: "C0AAA1" });
	});

	test("writing the notice channel keeps other settings", () => {
		const written = slackAgentSettings.write({ situationNoticeChannelId: "C0AAA1" }, { other: true });

		expect(written).toEqual({ other: true, situation_notice_channel_id: "C0AAA1" });
	});

	test("clearing the notice channel writes it empty", () => {
		const written = slackAgentSettings.write(
			{ situationNoticeChannelId: "" },
			{ situation_notice_channel_id: "C0AAA1" }
		);

		expect(written).toEqual({ situation_notice_channel_id: "" });
	});
});
