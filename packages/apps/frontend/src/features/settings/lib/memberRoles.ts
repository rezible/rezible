import type { UserAttributes } from "$lib/api";

type MemberRole = UserAttributes["organizationRole"];

type MemberRolePresentation = {
	label: string;
	badgeVariant: "default" | "secondary";
};

const memberRoles: Record<MemberRole, MemberRolePresentation> = {
	admin: { label: "Admin", badgeVariant: "default" },
	member: { label: "Member", badgeVariant: "secondary" },
};

export function memberRolePresentation(role: MemberRole) {
	return memberRoles[role];
}
