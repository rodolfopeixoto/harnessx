import { StubPage } from "./Stub";

const ID = "plan";
const TITLE = "Plan";
const HINT = "Generated spec and plan land here for review before execution.";

export function PlanPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
