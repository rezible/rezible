
export type User = {
    attributes: UserAttributes;
    id: string;
};

export type UserAttributes = {
    email: string;
    name: string;
};

export type OncallRoster = {
	attributes: OncallRosterAttributes;
	id: string;
};

export type OncallRosterAttributes = {
	handoverTemplateId: string;
	name: string;
	schedules: Array<OncallSchedule>;
	slug: string;
};

export type OncallSchedule = {
	attributes: OncallScheduleAttributes;
	id: string;
};

export type OncallScheduleAttributes = {
	description: string;
	participants: Array<OncallScheduleParticipant>;
	roster: OncallRoster;
	timezone: string;
};

export type OncallScheduleParticipant = {
	order: number;
	user: User;
};

export type OncallShift = {
	attributes: OncallShiftAttributes;
	id: string;
};

export type OncallShiftAttributes = {
	endAt: string;
	role: string;
	roster: OncallRoster;
	startAt: string;
	user: User;
};

export type BacklogItem = {
	id: string;
	attributes: BacklogItemAttributes;
};

export type BacklogItemAttributes = {
	title: string;
	priority: number;
	createdAt: Date;
};