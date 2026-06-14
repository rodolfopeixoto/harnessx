import { StubPage } from "./Stub";

const ID = "activerun";
const TITLE = "Active run";
const HINT = "Stage-by-stage telemetry, agent IO, sensors and cost will stream here.";

export function ActiveRunPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
