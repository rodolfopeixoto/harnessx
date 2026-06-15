import { Panel } from "../components/Panel";

export function BackupPage() {
  return (
    <div data-testid="page-backup" style={{ display: "grid", gap: 16 }}>
      <Panel title="Portable backup">
        <p>
          Backup is operator-driven via the CLI to avoid uploading from the dashboard process.
          Use <code>harness backup</code> on the host.
        </p>
        <h4>Cheatsheet</h4>
        <pre style={{ background: "#F1F5F9", padding: 12, borderRadius: 6, fontSize: 13 }}>{`harness install rclone
rclone config                                # one-time provider auth
harness backup remote add gdrive --provider drive --interactive
harness backup snapshot --remote gdrive --tag pre-experiment
harness backup list --remote gdrive
harness backup restore <snapshot> --remote gdrive --target /tmp/restored
harness backup sync push --remote gdrive
harness backup sync pull --remote gdrive --dry-run`}</pre>
        <p style={{ fontSize: 12, color: "#64748B" }}>
          Default config excludes secrets. Include them only with{" "}
          <code>--include-secrets</code> AND <code>HARNESS_BACKUP_I_UNDERSTAND_SECRETS=1</code>.
        </p>
      </Panel>
    </div>
  );
}
