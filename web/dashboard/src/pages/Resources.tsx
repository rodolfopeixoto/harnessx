import { StubPage } from "./Stub";

const ID = "resources";
const TITLE = "Resources";
const HINT = "Worktrees · caches · containers · large files. Cleanup goes through policy.";

export function ResourcesPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
