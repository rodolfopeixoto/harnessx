import { StubPage } from "./Stub";

const ID = "onboarding";
const TITLE = "Onboarding";
const HINT = "First-run wizard: pick project, detect stack, certify agents, run doctor.";

export function OnboardingPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
