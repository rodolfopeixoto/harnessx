import { StubPage } from "./Stub";

const ID = "stakeholder";
const TITLE = "Stakeholder view";
const HINT = "Non-technical readout: ready · blocked · mocked · approval pending · cost.";

export function StakeholderPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
