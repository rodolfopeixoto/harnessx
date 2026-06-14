import { StubPage } from "./Stub";

const ID = "command";
const TITLE = "Command — natural prompt";
const HINT = "Type a feature, bugfix or audit prompt. Mode detection wires the run.";

export function CommandPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
