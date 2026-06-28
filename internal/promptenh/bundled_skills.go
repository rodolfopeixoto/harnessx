package promptenh

import "github.com/ropeixoto/harnessx/internal/skillpkg"

type BundledSkills struct{}

func (BundledSkills) List() ([]SkillSnippet, error) {
	all, err := skillpkg.List()
	if err != nil {
		return nil, err
	}
	out := make([]SkillSnippet, 0, len(all))
	for _, t := range all {
		out = append(out, SkillSnippet{
			ID:    t.Name,
			Mode:  t.Mode,
			Body:  t.Body,
			Score: 0,
		})
	}
	return out, nil
}
