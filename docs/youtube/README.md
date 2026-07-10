# Canal YouTube — Universo harnessx (índice mestre v2)

8.697 linhas de material pronto pra gravação. **65 vídeos longos**
(17 + 10 + 18 + 10 + 10) + 60 Shorts + brand + funil. PT-BR.
Contagens reconciliadas com `PLAYLISTS-GROWTH-MASTER.md §1` e séries
canônicas (`SERIE-PAPERS-NA-PRATICA.md`, `SERIE-DENTRO-DO-HARNESSX.md`).

## Documentos (9 camadas)

| # | Arquivo | Papel | Linhas |
|---|---|---|---|
| 1 | [PLAYLISTS-GROWTH-MASTER.md](PLAYLISTS-GROWTH-MASTER.md) | 5 playlists, funil TOF→BOF, roadmap 12m, brand system, cadência | 675 |
| 2 | [ROTEIROS-TECNICOS.md](ROTEIROS-TECNICOS.md) | 17 roteiros PL1 (TaskHive) — falas + comandos + timecodes | 896 |
| 3 | [DIDATICA-CONCEITOS.md](DIDATICA-CONCEITOS.md) | 40 termos + analogias BR + erros + checklist aluno | 774 |
| 4 | [SEO-RETENCAO.md](SEO-RETENCAO.md) | 17 vídeos: título A/B, descrição, tags, thumb brief, retention loops | 1237 |
| 5 | [SERIE-DENTRO-DO-HARNESSX.md](SERIE-DENTRO-DO-HARNESSX.md) | 18 code walkthrough Go, 55 arquivos reais, 18 snippets | 1323 |
| 6 | [SERIE-PAPERS-NA-PRATICA.md](SERIE-PAPERS-NA-PRATICA.md) | 10 papers, cada um mapeado ao arquivo Go que implementa | 782 |
| 7 | [SHORTS-FACTORY.md](SHORTS-FACTORY.md) | 60 Shorts, 5 fórmulas, calendário 90d, TikTok/Reels/BlueSky | 1366 |
| 8 | [ENGAGEMENT-GROWTH-2026.md](ENGAGEMENT-GROWTH-2026.md) | 100 CTAs, community, comment seeding, KPIs, 30 dias dia-a-dia | 564 |
| 9 | [BRAND-THUMB-TITLE-MATRIX.md](BRAND-THUMB-TITLE-MATRIX.md) | Brand book + 195 títulos + 130 thumb briefs + A/B protocol | 1018 |

## As 5 playlists

| PL | Nome | Vídeos | Fase funil |
|---|---|---|---|
| **PL1** | Construindo SaaS com Agentes IA (TaskHive) | 17 | MOF — retenção |
| **PL2** | Papers na Prática | 10 | BOF — fideliza super-fã |
| **PL3** | Dentro do harnessx (code walkthrough) | 18 | BOF — hardcore |
| **PL4** | Agentes IA em 2026 (conceitual) | 10 | TOF — viralização |
| **PL5** | Do Zero ao SaaS Deployado | 10 | TOF+MOF — bootstrap |
| **Total** |   | **65** vídeos longos + **60** Shorts |   |

## Como usar (fluxo gravação)

1. Escolhe vídeo alvo na tabela PL1-05 / PL2-03 / etc
2. Abre **4 arquivos** lado a lado filtrando pelo ID do vídeo:
   - Roteiro técnico ([2] ou [5] ou [6] conforme PL)
   - Didática ([3])
   - SEO ([4] ou trecho da matriz [9])
   - Brand/thumb ([9])
3. Grava seguindo Roteiro, adaptando hook com analogia da Didática
4. Publica com pacote SEO + thumb A do brand matrix
5. Depois 48h: A/B test título (protocolo em [9])
6. Deriva 4 Shorts do arquivo [7]
7. Cross-post seguindo [8] cross-platform

## Top 10 vídeos ROI (gravar primeiro)

Consolidado dos 6 agentes. Todos os IDs abaixo mapeiam a séries reais
(`PLAYLISTS-GROWTH-MASTER.md`, `SERIE-PAPERS-NA-PRATICA.md`,
`SERIE-DENTRO-DO-HARNESSX.md`).

1. **PL4-A00** — "Vibe coding é mentira. Provei em 8 min de terminal"
2. **PL5-S00** — "Fiz um SaaS por US$ 12 no primeiro mês" `[VERIFICAR — número exige nota fiscal antes de publicar]`
3. **PL1-v00** — "TaskHive: SaaS Go+React por US$ 1,50 em LLM" (banda 1.20–1.50 comprovada em `docs/TUTORIAL-BUILD-TASKHIVE.md`)
4. **PL1-v12** — "SaaS por US$ 1,23 em LLM — breakdown real" (fonte: TUTORIAL-BUILD-TASKHIVE.md §custo real, Sonnet+Codex+Kimi mix)
5. **PL2-E1** — "O paper que corta 60% do custo com Claude" (LLMLingua-2 — banda 50–70% documentada em `docs/PAPER-IMPLEMENTATION.md §5`)
6. **PL2-E5** — "90% de desconto que 90% ignora" (Prompt Caching Anthropic — `[VERIFICAR banda oficial 2026]`)
7. **PL2-E2** — "Lost in the Middle — por que sua ordem quebra o agente" (implementado em `internal/context/reorder.go`)
8. **PL3-EP04** — "Secrets Sensor: regex `AKIA…` na tela" (real em `internal/sensors/scanners.go`)
9. **PL3-EP09** — "Router com fallback: Claude caiu, e agora?" (real em `internal/router/router.go`, config `fallback: [...]`)
10. **PL3-EP18** — "Evolve: harness que se conserta" — `harness evolve diagnose|propose|replay|sandbox|promote --hitl`

## Brand system (resumo)

- **Nome série**: [3 propostas em PLAYLISTS-GROWTH-MASTER.md]
- **Paleta**: `#0B0B0F` (base) + `#00E38A` (acento verde CRT) + amarelo destaque + vermelho alerta
- **Tipografia**: Inter Bold display + JetBrains Mono terminal ([VERIFICAR licenças OFL 2026])
- **Grid thumb**: 1280×720 — rosto canto inf. direito, número gigante centro, texto ≤4 palavras top
- **Intro 5s**: cursor `$_` piscando → logo → hook
- **Outro 10s**: 3 endscreens (próximo vídeo / playlist / inscrever)

## Fórmulas título top 3

1. **F5** — "O paper que [resultado inesperado]" — assinatura PL2
2. **F6** — "Eu quebrei [X] usando [Y]" — drama pessoal, CTR mobile alto
3. **F3** — "[A] vs [B] em [tarefa]" — dominante search tech BR

## Cadência publicação

- **Quinta 19h BRT** — vídeo longo (rotação forçada: S1 TOF / S2 MOF / S3 BOF / S4 ponte)
- **Todo dia** — 1 Short (primeiros 90 dias)
- **Seg / Qua / Sex** — 1 Community post cada
- **Dia 21** — primeira live

## Métricas alvo

| Fase | CTR | AVD | Subs 30d |
|---|---|---|---|
| Cold start | ≥4% | ≥45% | 0→100 |
| 30-90d | ≥6% | ≥55% | 100→1k |
| 90-180d | ≥7% | ≥60% | 1k→10k |
| 180-365d | ≥8% | ≥65% | 10k→50k |

## Monetização (12 meses)

- **T1**: bootstrap subs, sem monetização
- **T2**: YPP + affiliate (Anthropic Console, hosting)
- **T3**: Membership R$ 9,90 / 29,90 (skill.dev acesso antecipado)
- **T4**: Cursos skill.dev R$ 297 / 497 / 897 + patrocínio

## Referências principais

- **Papers**: [LLMLingua-2](https://arxiv.org/abs/2403.12968), [Lost in the Middle](https://arxiv.org/abs/2307.03172), [Voyager](https://arxiv.org/abs/2305.16291), [Constitutional AI](https://arxiv.org/abs/2212.08073), [RAG](https://arxiv.org/abs/2005.11401), [Chinchilla](https://arxiv.org/abs/2203.15556), [Context Engineering 2.0](https://arxiv.org/abs/2510.26493), [Generative Agents](https://arxiv.org/abs/2304.03442)
- **Docs oficiais**: [Anthropic Prompt Caching](https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching), [MCP](https://modelcontextprotocol.io), [Chi](https://github.com/go-chi/chi), [modernc/sqlite](https://pkg.go.dev/modernc.org/sqlite)
- **Growth 2026**: YouTube Creator Insider (canal oficial), MrBeast deck público, Fireship style analysis, Guo/Kim/Rubin L@S

## Marcadores `[VERIFICAR]`

Distribuídos nos 9 arquivos. Concentrados em:
- Licenças fonts 2026
- WCAG contraste 2026 revision
- YouTube Test & Compare docs (feature em rollout)
- Preços APIs Anthropic/OpenAI T4 2026
- Handles canais BR mencionados

Revisar antes de gravar qualquer vídeo público.

## Próximos passos sugeridos

1. Confirmar nome oficial série (3 propostas em [1])
2. Validar `[VERIFICAR]` marcadores críticos
3. Gravar **PL4-A00** como piloto (menor risco produção, maior hook)
4. Após 3 vídeos: revisar analytics e recalibrar tópicos T2
