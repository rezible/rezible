import { describe, expect, mock, test } from "bun:test";
import { compileModule } from "svelte/compiler";

import type { Situation, SituationNotice } from "$lib/api";
import { noticeLines, noticeRatingsText } from "$features/situations/lib/notices";

import type { SituationNoticesController } from "./situationNotices.svelte.ts";

// Bun does not compile runes. Compile the real controller and supply its external dependencies locally,
// avoiding module mocks that would affect other tests.
const compileRunes = async (relativePath: string) => {
	const source = await Bun.file(new URL(relativePath, import.meta.url)).text();
	const javascript = new Bun.Transpiler({ loader: "ts" }).transformSync(source);
	const options = { filename: relativePath, generate: "client", dev: false } as const;
	const compiled = compileModule(javascript, options).js.code;
	return compiled.replace(/^import [^;]*;$/gm, "").replace(/^export /gm, "");
};

const runtime = await import(import.meta.resolve("svelte/internal/client"));
const controllerCode = await compileRunes("./situationNotices.svelte.ts");

type MutationOptions = {
	onSuccess: (response: { data: Situation }) => Promise<void>;
};

const failedNotice: SituationNotice = {
	channelId: "C0AAA1",
	channelName: "incidents",
	status: "failed",
	error: "not_in_channel",
};

function setup(situation: Situation | undefined) {
	const retried = { ...situation, attributes: { notices: [] } } as unknown as Situation;
	const mutate = mock((variables: unknown) => {
		void variables;
		return options.onSuccess({ data: retried });
	});
	let options: MutationOptions;
	const createMutation = (optionsFn: () => MutationOptions) => {
		options = optionsFn();
		return { isPending: false, mutate };
	};
	const retrySituationNoticeMutation = () => ({});
	const toast = { error: () => {} };

	const Controller = new Function(
		"$",
		"createMutation",
		"toast",
		"retrySituationNoticeMutation",
		"noticeLines",
		"noticeRatingsText",
		`${controllerCode}\nreturn SituationNoticesController;`
	)(runtime, createMutation, toast, retrySituationNoticeMutation, noticeLines, noticeRatingsText) as new (
		...args: ConstructorParameters<typeof SituationNoticesController>
	) => SituationNoticesController;

	const applySituation = mock(async (situation: Situation) => {
		void situation;
	});
	const controller = new Controller(() => situation, applySituation);
	return { controller, mutate, applySituation, retried };
}

describe("situation notices", () => {
	test("retry asks the situation's notices to be delivered again and shows the result", async () => {
		const situation = {
			id: "situation-1",
			attributes: { notices: [failedNotice], noticeRatings: { useful: 0, notUseful: 0 } },
		} as unknown as Situation;
		const { controller, mutate, applySituation, retried } = setup(situation);

		expect(controller.lines.map((line) => line.retryable)).toEqual([true]);
		controller.retry();

		expect(mutate).toHaveBeenCalledWith({ path: { id: "situation-1" } });
		expect(applySituation).toHaveBeenCalledWith(retried);
	});

	test("there is nothing to retry before the situation loads", () => {
		const { controller, mutate } = setup(undefined);

		controller.retry();

		expect(controller.lines).toEqual([]);
		expect(mutate).not.toHaveBeenCalled();
	});
});
