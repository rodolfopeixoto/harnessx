import { StubPage } from "./Stub";

const ID = "cleanup";
const TITLE = "Cleanup";
const HINT = "Worktrees · caches · abandoned harness dirs · VM leftovers. Policy required.";

export function CleanupPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
