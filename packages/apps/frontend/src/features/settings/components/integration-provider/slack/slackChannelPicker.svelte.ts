import { createInfiniteQuery } from "@tanstack/svelte-query";
import { watch, type Getter } from "runed";

import { type ChatChannel, type ApiError, listIntegrationChatChannels } from "$lib/api";

// Loads public Slack channels page by page through the installation, only while the picker is open.
export class SlackChannelPickerController {
	open = $state(false);
	private installationId = $state("");

	constructor(installationId: Getter<string>) {
		watch(installationId, (id) => {
			this.installationId = id;
		});
	}

	channelsQuery = createInfiniteQuery(() => {
		const id = this.installationId;
		return {
			queryKey: ["slack-channel-picker", id],
			queryFn: async ({ pageParam, signal }) => {
				const { data } = await listIntegrationChatChannels({
					path: { id },
					query: { cursor: pageParam },
					signal,
					throwOnError: true,
				});
				return data.data;
			},
			initialPageParam: "",
			getNextPageParam: (page) => page.nextCursor || undefined,
			enabled: this.open && !!id,
		};
	});

	channels: ChatChannel[] = $derived(this.channelsQuery.data?.pages.flatMap((page) => page.channels) ?? []);
	loadError = $derived(this.channelsQuery.error as ApiError | null);
	loaded = $derived(!!this.channelsQuery.data);
	hasMore = $derived(this.channelsQuery.hasNextPage);
	loadingMore = $derived(this.channelsQuery.isFetchingNextPage);

	loadMore = () => {
		this.channelsQuery.fetchNextPage();
	};

	// A saved channel that is not among the loaded channels (including private channels) shows its raw ID.
	labelFor(channelId: string) {
		if (!channelId) {
			return "No announcement channel";
		}

		const channel = this.channels.find((ch) => ch.id === channelId);
		if (!channel) {
			return channelId;
		}

		return `#${channel.name}`;
	}
}
