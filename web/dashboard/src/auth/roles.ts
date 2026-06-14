export const ROLE_ANONYMOUS = "anonymous";
export const ROLE_OPERATOR = "operator";
export const ROLE_ADMIN = "admin";

export type Role = typeof ROLE_ANONYMOUS | typeof ROLE_OPERATOR | typeof ROLE_ADMIN;

export const ALL_ROLES: Role[] = [ROLE_ANONYMOUS, ROLE_OPERATOR, ROLE_ADMIN];

const HIERARCHY: Record<Role, number> = {
  [ROLE_ANONYMOUS]: 0,
  [ROLE_OPERATOR]: 1,
  [ROLE_ADMIN]: 2,
};

export function roleCanMutate(role: Role): boolean {
  return HIERARCHY[role] >= HIERARCHY[ROLE_OPERATOR];
}

export function roleCanAdminister(role: Role): boolean {
  return HIERARCHY[role] >= HIERARCHY[ROLE_ADMIN];
}
