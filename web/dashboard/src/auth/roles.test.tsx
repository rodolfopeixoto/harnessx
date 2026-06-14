import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { ALL_ROLES, ROLE_ADMIN, ROLE_ANONYMOUS, ROLE_OPERATOR, roleCanAdminister, roleCanMutate } from "./roles";
import { RoleProvider, useRole } from "./RoleContext";

function Probe() {
  const role = useRole();
  return <span data-testid="role">{role}</span>;
}

describe("roles", () => {
  it("hierarchy: operator can mutate, anonymous cannot", () => {
    expect(roleCanMutate(ROLE_ANONYMOUS)).toBe(false);
    expect(roleCanMutate(ROLE_OPERATOR)).toBe(true);
    expect(roleCanMutate(ROLE_ADMIN)).toBe(true);
  });

  it("hierarchy: only admin can administer", () => {
    expect(roleCanAdminister(ROLE_ANONYMOUS)).toBe(false);
    expect(roleCanAdminister(ROLE_OPERATOR)).toBe(false);
    expect(roleCanAdminister(ROLE_ADMIN)).toBe(true);
  });

  it("ALL_ROLES enumerates three", () => {
    expect(ALL_ROLES).toEqual([ROLE_ANONYMOUS, ROLE_OPERATOR, ROLE_ADMIN]);
  });
});

describe("RoleProvider", () => {
  it.each(ALL_ROLES)("propagates %s via useRole", (role) => {
    render(
      <RoleProvider role={role}>
        <Probe />
      </RoleProvider>,
    );
    expect(screen.getByTestId("role").textContent).toBe(role);
  });
});
