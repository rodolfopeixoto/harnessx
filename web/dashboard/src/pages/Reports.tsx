import { StubPage } from "./Stub";

const ID = "reports";
const TITLE = "Reports";
const HINT = "Archive of every run report, filtered by project, period, agent or mode.";

export function ReportsPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
