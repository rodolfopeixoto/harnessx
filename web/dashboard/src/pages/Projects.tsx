import { StubPage } from "./Stub";

const ID = "projects";
const TITLE = "Projects";
const HINT = "Registered projects, last seen, stack, health, switch + import wizard.";

export function ProjectsPage() {
  return <StubPage id={ID} title={TITLE} description={HINT} />;
}
