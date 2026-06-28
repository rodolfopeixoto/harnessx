package learncmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ropeixoto/harnessx/internal/execution"
)

type SkillScore struct {
	SkillID         string  `json:"skill_id"`
	Runs            int     `json:"runs"`
	Applied         int     `json:"applied"`
	Failed          int     `json:"failed"`
	WaitingApproval int     `json:"waiting_approval"`
	AvgRetries      float64 `json:"avg_retries"`
	SuccessRate     float64 `json:"success_rate"`
}

type SkillScoreReport struct {
	GeneratedAt time.Time    `json:"generated_at"`
	Scores      []SkillScore `json:"scores"`
}

func ScoreSkills(root string, out io.Writer) (SkillScoreReport, string, error) {
	runs, err := execution.ListRuns(root)
	if err != nil {
		return SkillScoreReport{}, "", err
	}
	stats := map[string]*SkillScore{}
	for _, r := range runs {
		skills := readSkillsForRun(root, r.RunID)
		if len(skills) == 0 {
			continue
		}
		retries := r.Recovery.Retries
		for _, sk := range skills {
			s, ok := stats[sk]
			if !ok {
				s = &SkillScore{SkillID: sk}
				stats[sk] = s
			}
			s.Runs++
			s.AvgRetries += float64(retries)
			switch r.Status {
			case execution.StatusApplied:
				s.Applied++
			case execution.StatusAgentFailed, execution.StatusSensorFailed, execution.StatusAutonomyDenied, execution.StatusConflict:
				s.Failed++
			case execution.StatusWaitingApproval:
				s.WaitingApproval++
			}
		}
	}
	rep := SkillScoreReport{GeneratedAt: time.Now().UTC()}
	for _, s := range stats {
		if s.Runs > 0 {
			s.AvgRetries /= float64(s.Runs)
			s.SuccessRate = float64(s.Applied) / float64(s.Runs)
		}
		rep.Scores = append(rep.Scores, *s)
	}
	sort.Slice(rep.Scores, func(i, j int) bool {
		if rep.Scores[i].SuccessRate != rep.Scores[j].SuccessRate {
			return rep.Scores[i].SuccessRate > rep.Scores[j].SuccessRate
		}
		return rep.Scores[i].Runs > rep.Scores[j].Runs
	})
	renderSkillScores(out, rep)
	path := filepath.Join(root, ".harness", "memory", "skill-scores.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return rep, "", err
	}
	body, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return rep, "", err
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return rep, "", err
	}
	return rep, path, nil
}

func readSkillsForRun(root, runID string) []string {
	path := filepath.Join(root, ".harness", "runs", runID, "enhancement.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var env struct {
		SkillPrefixes []string `json:"skill_prefixes"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil
	}
	return env.SkillPrefixes
}

func renderSkillScores(out io.Writer, rep SkillScoreReport) {
	if len(rep.Scores) == 0 {
		fmt.Fprintln(out, "no skill telemetry yet — run `harness auto` or `harness ship` with --agent so enhancement.json is written per run")
		return
	}
	fmt.Fprintf(out, "%-22s %6s %6s %6s %12s %12s\n", "SKILL", "RUNS", "OK", "FAIL", "AVG_RETRIES", "SUCCESS_RATE")
	for _, s := range rep.Scores {
		fmt.Fprintf(out, "%-22s %6d %6d %6d %12.2f %11.0f%%\n",
			s.SkillID, s.Runs, s.Applied, s.Failed, s.AvgRetries, s.SuccessRate*100)
	}
	winner := rep.Scores[0]
	if winner.Runs >= 3 && winner.SuccessRate > 0.7 {
		fmt.Fprintf(out, "\n→ %s is the highest-leverage skill (%d runs, %.0f%% success). Pin it via `harness skill install %s` if not already auto-applied.\n", winner.SkillID, winner.Runs, winner.SuccessRate*100, winner.SkillID)
	}
	if loser := rep.Scores[len(rep.Scores)-1]; loser.Runs >= 3 && loser.SuccessRate < 0.3 {
		fmt.Fprintf(out, "→ %s has %.0f%% success across %d runs — consider disabling via `.harness/config/skills.disabled` (one ID per line).\n", loser.SkillID, loser.SuccessRate*100, loser.Runs)
	}
	_ = strings.NewReader
}
