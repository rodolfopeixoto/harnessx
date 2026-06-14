// SPDX-License-Identifier: MIT

package context

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/ropeixoto/harnessx/internal/platform/config"
	// modernc.org/sqlite registers the "sqlite" driver via init().
	_ "modernc.org/sqlite"
)

// MemoryProvider reads project memory entries from the local SQLite DB.
// It runs before any LLM provider so the agent sees confirmed facts first.
type MemoryProvider struct {
	MaxEntries int
}

func (MemoryProvider) Name() string { return "memory" }

func (m MemoryProvider) Apply(ctx context.Context, root string, pack *Pack) error {
	cfg, _ := config.Load(filepath.Join(root, ".harness", "config", "harness.yaml"), root)
	dbPath := config.Resolve(root, cfg.Database.Path)
	if _, err := os.Stat(dbPath); err != nil {
		pack.Stats.ProvidersSkipped++
		return nil
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		pack.Stats.ProvidersSkipped++
		return nil
	}
	defer db.Close()

	limit := m.MaxEntries
	if limit <= 0 {
		limit = 25
	}
	rows, err := db.QueryContext(ctx, `
		select id, scope, kind, content, confidence
		from memories
		order by confidence desc, updated_at desc
		limit ?`, limit)
	if err != nil {
		pack.Stats.ProvidersSkipped++
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var mem Memory
		if err := rows.Scan(&mem.ID, &mem.Scope, &mem.Kind, &mem.Content, &mem.Confidence); err != nil {
			continue
		}
		pack.Memories = append(pack.Memories, mem)
	}
	pack.Stats.ProvidersRan++
	return nil
}
