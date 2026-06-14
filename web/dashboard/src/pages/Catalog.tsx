import { StubPage } from "./Stub";

const ID = "catalog";
const TITLE = "Capabilities";
const HINT = "Agents · MCPs · Hooks · Sensors · Skills · Context Providers · Plugins.";

export function CatalogPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
