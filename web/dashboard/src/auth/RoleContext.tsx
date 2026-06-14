import { createContext, useContext, type ReactNode } from "react";
import { ROLE_OPERATOR, type Role } from "./roles";

type RoleContextValue = {
  role: Role;
};

const Ctx = createContext<RoleContextValue>({ role: ROLE_OPERATOR });

type ProviderProps = {
  role?: Role;
  children: ReactNode;
};

export function RoleProvider({ role = ROLE_OPERATOR, children }: ProviderProps) {
  return <Ctx.Provider value={{ role }}>{children}</Ctx.Provider>;
}

export function useRole(): Role {
  return useContext(Ctx).role;
}
