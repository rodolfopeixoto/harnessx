import { StubPage } from "./Stub";

const ID = "context";
const TITLE = "Context pack";
const HINT = "Files selected/rejected by the engineer, with token cost and cache hit rate.";

export function ContextPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
