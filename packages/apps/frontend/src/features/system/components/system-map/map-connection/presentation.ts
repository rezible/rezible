import { relationshipPredicateLabel } from "$features/system/lib/system-map/labels";
import type { MapConnection } from "$features/system/lib/system-map/connections";

export function connectionLabel(connection: MapConnection): string {
	const predicate = relationshipPredicateLabel(connection.predicate);
	const count = connection.relationshipIds.length;
	if (count > 1) {
		return `${predicate} · ${count} relationships`;
	}
	return predicate;
}
