# ROTEIROS TÉCNICOS — Série TaskHive com harnessx

> Agent 1 de 3 — Roteirista técnico BR.
> Voz: Rodolfo Peixoto. Português BR, direto, técnico. Sem "vamos",
> "galera", "basicamente", "hoje eu vou". Cada afirmação técnica tem
> referência inline. Base: `docs/TUTORIAL-BUILD-TASKHIVE.md`.
> Duração alvo: ≤ 10 min por vídeo. Formato 16:9, 1080p mínimo.

## Convenções

- **HOOK** 0–15s | **CONTEXTO** 15–60s | **DEMO** 60s–8min |
  **RECAP** 8–9min | **CTA** 9–10min.
- Comandos em terminal: fonte JetBrains Mono 18pt, tema dark, zoom
  automático no comando.
- Highlight = retângulo amarelo animado 300ms sobre o trecho.
- B-roll neutro em cortes de fala: terminal parado, diagrama, browser.
- Referências ao tutorial usam âncora Markdown do arquivo original.

---

## Vídeo 0 — INTRO da série

**Título:** `TaskHive: SaaS Go+React por US$ 1,50 em LLM`
**Duração alvo:** 6 min

### HOOK (0–15s)
> "Um SaaS completo. Go no back, React 19 no front, JWT, CRUD,
> Playwright, Docker. Custo total em LLM: um dólar e vinte e três
> centavos. Vou provar."

B-roll: `harness cost report --breakdown` rodando, tabela final aparece.

### CONTEXTO (15–60s)
> "Quinze episódios. Um por capítulo do tutorial oficial do harnessx.
> Cada vídeo entrega uma feature que compila e passa nos testes. No
> fim da série você tem o repositório, o tag de release, e o
> dashboard de custo aberto no navegador."
>
> "Pré-requisitos: Go 1.25 ou mais novo, Node 20, harness v0.170.1,
> uma chave de API — Anthropic é o default, mas roda com Codex, Kimi
> ou o adapter fake se você quiser fazer offline."

Referência: [Prólogo do tutorial](../TUTORIAL-BUILD-TASKHIVE.md#prologue).
Docs oficiais: <https://go.dev/dl/>, <https://nodejs.org/>.

### DEMO (60s–4min)
Tour visual do que a série entrega:
1. Terminal: `harness --version` → v0.170.1.
2. Browser: dashboard React em `127.0.0.1:<porta>`, aba `/api/runs`.
3. VSCode split: `.harness/artifacts/specs/<id>.md` de um lado, diff
   Go do outro.
4. Terminal: `docker compose ps` mostrando backend + nginx healthy.

Zoom: nome de cada arquivo tocado — `active.yaml`, `routes.yaml`,
`spec.md`, `flow.yaml`.

### RECAP (4–5min)
> "Regra da série: cada vídeo corresponde a um capítulo do
> `docs/TUTORIAL-BUILD-TASKHIVE.md`. Cada comando exibido está
> literal no repositório. Nada de pseudocódigo."

### CTA (5–6min)
> "Inscreve. Ativa o sino. O próximo vídeo é setup: init, index,
> doctor, sensor list — cinco minutos. Link do repositório e do
> tutorial na descrição."

Links descrição: repo harnessx, `docs/TUTORIAL-BUILD-TASKHIVE.md`,
`docs/GETTING-STARTED.md`.

---

## Vídeo 1 — Capítulo 1: Setup + primeiro sensor pass

**Título:** `harness init em 5 min: vault, 8 índices, doctor`
**Duração alvo:** 7 min

### HOOK (0–15s)
> "Oito arquivos JSON. É o que separa uma IDE que grita 'tudo o que
> abri' do context engine que manda só o que interessa. Vou construir
> os oito agora."

B-roll: `ls .harness/project/` mostrando os oito `.json` reais.

### CONTEXTO (15–60s)
> "Objetivo: repositório vazio, `harness init` roda, `harness project
> index` gera os mapas, `harness doctor` confirma toolchain, e
> `routes.yaml` fica pronto com budget de cinco dólares."
>
> "Custo desta etapa: zero. Nenhuma chamada de LLM sai do PATH."

Referência: [Capítulo 1](../TUTORIAL-BUILD-TASKHIVE.md#chapter-1--setup-and-the-first-sensor-pass).

### DEMO (60s–5min)
Comandos exatos:
```bash
mkdir -p ~/dev/taskhive && cd ~/dev/taskhive
git init -q
harness init
harness project index
harness doctor
harness agent list
harness use claude
```

Zoom/highlight:
- Linha por linha dos mapas reais em `.harness/project/`
  (`profile.json`, `commands.json`, `dependencies.json`,
  `architecture.json`, `test-map.json`, `api-map.json`,
  `design-system.json`, `performance-budget.json`) — ver
  `internal/index/index.go` `MapName` enum.
- Saída de `harness doctor`: destacar linha "Go 1.25+ ok".
- Lista de adapters registrados (`harness agent list`): cursor sobre
  `claude`, `codex`, `kimi`, `fake`.

Depois, criar o `routes.yaml`:
```yaml
# .harness/config/routes.yaml
budget_usd: 5.00
default: claude
routes:
  - match: { intent: spec }
    agent: claude
    fallback: [codex, kimi]
  - match: { intent: implement }
    agent: codex
    fallback: [claude]
  - match: { intent: test }
    agent: kimi
    fallback: [codex]
  - match: { intent: docs }
    agent: claude-haiku
    fallback: [kimi]
```

Highlight: `budget_usd: 5.00` — comentar que é hard cap com fonte
`internal/costtrack/`.

B-roll: diagrama simples mostrando os oito índices alimentando
`internal/context.Build`.

### RECAP (5–6min)
> "Vault criado. Oito índices. Doctor verde. Router com budget
> travado em cinco dólares. Zero LLM."

Referência lib: [Chi router docs](https://github.com/go-chi/chi),
[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite).

### CTA (6–7min)
> "Próximo vídeo: escrever spec, gerar plan, aplicar patch. É onde o
> primeiro dólar aparece — e onde o cache de prompt começa a pagar."

---

## Vídeo 2 — Capítulo 2: Spec-driven backend init

**Título:** `Spec antes do código: US$ 0,15 por scaffold Go`
**Duração alvo:** 8 min

### HOOK (0–15s)
> "Não abrir o editor. Escrever a spec. O diff que o LLM aplica
> depois é derivado dela. Se a spec está errada, o resto está
> errado em silêncio."

B-roll: split — spec.md à esquerda, patch.diff à direita.

### CONTEXTO (15–60s)
> "Stack alvo do backend: Chi router, modernc.org/sqlite — sem CGO —,
> JWT HS256, slog estruturado. Tudo declarado na spec antes de rodar
> `harness run`."

Referência: [Capítulo 2](../TUTORIAL-BUILD-TASKHIVE.md#chapter-2--spec-driven-backend-init).
Docs: [Chi](https://github.com/go-chi/chi), [modernc sqlite](https://pkg.go.dev/modernc.org/sqlite),
[log/slog](https://pkg.go.dev/log/slog).

### DEMO (60s–6:30)
```bash
harness feature "Go HTTP API scaffold with Chi router, \
modernc.org/sqlite persistence, JWT HS256 auth, structured slog" \
  --agent claude --yes
```

Cortar para o arquivo gerado:
```
.harness/artifacts/specs/<id>.md
```
Highlight: seções "acceptance criteria", "non-goals", "test seams",
"ADR sqlite vs Postgres".

```bash
harness plan
```
Highlight: budget de tokens estimado, adapter escolhido, fallback.

```bash
harness run
```
B-roll: sensores rodando na staging area — secrets, forbidden files,
embed `..`. Referência: `internal/sensors/`.

```bash
harness check
```
Zoom: saída de `go vet`, `go test ./... -race`, `staticcheck`.

### RECAP (6:30–8min)
> "Spec, plan, run, check. Delta de custo: quinze centavos. O
> segundo call cai no cache de prefixo do Anthropic — [prompt caching
> docs](https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching)."
>
> "Se check falhar, próximo vídeo mostra `harness bugfix`. Não editar
> na mão."

CTA: "Capítulo 3: JWT. É onde o modelo escolhido importa mais."

---

## Vídeo 3 — Capítulo 3: Autenticação JWT com Claude

**Título:** `JWT sem footgun: por que Sonnet e não Codex`
**Duração alvo:** 9 min

### HOOK (0–15s)
> "JWT `alg=none`. Bcrypt com cost 4. Compare de senha vazando timing.
> Três formas clássicas de vazar dados de auth. Sonnet pega. Codex
> às vezes não."

B-roll: trecho do [RFC 7519](https://datatracker.ietf.org/doc/html/rfc7519)
piscando na tela.

### CONTEXTO (15–60s)
> "Rota: POST `/auth/register`, POST `/auth/login`. Bcrypt cost 12,
> HS256, expiry 24h. Duplicata de email retorna 409. Senha fraca
> retorna 400."

Referência: [Capítulo 3](../TUTORIAL-BUILD-TASKHIVE.md#chapter-3--feature-authentication-jwt).
Libs: [`golang.org/x/crypto/bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt),
[`github.com/golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt).

### DEMO (60s–7min)
```bash
harness feature "add POST /auth/register and POST /auth/login \
using bcrypt cost 12 and JWT HS256 with a 24h expiry; \
refuse duplicate emails; return 400 on weak passwords" \
  --agent claude --yes
```

B-roll: log do context builder:
```
context: pack=18 files reordered by salience (router.go +3, main.go +2)
saved: ~210 tokens vs. lexicographic order
```

Highlight: linha "salience reorder". Referência real:
`internal/context/reorder.go` (função `Reorder` + `SalienceScorer`) —
não existe arquivo `salience.go` separado.

Zoom explicativo: por que o fundo do pack fica byte-idêntico — cache
prefix-sensitive da Anthropic.

Verificação:
```bash
harness check
go test ./internal/auth -run TestJWT -v
```

### RECAP (7–8:30)
> "Escolha do adapter é rubric, não lei. Auth = Sonnet. CRUD = Codex.
> Rubric completa está no Appendix A do tutorial."

Referência inline: [Apêndice A](../TUTORIAL-BUILD-TASKHIVE.md#appendix-a--agent-choice-rubric).

### CTA (8:30–9min)
> "Próximo: `/projects` CRUD com Codex. Onde o desconto por token
> aparece."

---

## Vídeo 4 — Capítulo 4: Projects CRUD com Codex

**Título:** `Codex faz CRUD por 1/3 do preço — e cai bem`
**Duração alvo:** 7 min

### HOOK (0–15s)
> "Codex custa cerca de um terço por token de output vs Sonnet
> [VERIFICAR: tabela de preços 2026 OpenAI]. Em CRUD com auth já
> pronta, ele acerta o diff no primeiro shot."

### CONTEXTO (15–60s)
> "CRUD escopado por usuário autenticado. POST, GET list, GET one,
> PATCH, DELETE. SQLite com created_at/updated_at. 404 em acesso
> cruzado — nunca 403 pra não vazar existência."

Referência: [Capítulo 4](../TUTORIAL-BUILD-TASKHIVE.md#chapter-4--feature-projects-crud).

### DEMO (60s–5:30)
```bash
harness feature "add /projects CRUD (POST, GET list, GET one, \
PATCH, DELETE) scoped by the authenticated user; sqlite schema \
with created_at, updated_at; 404 on cross-user access" \
  --agent codex --yes
```

B-roll: log de fallback simulado:
```
adapter=codex status=timeout after=42s
fallback -> claude
```

Highlight: `routes.yaml` sendo consultado. Referência real:
`internal/router/router.go` (`Router.Execute` + `FallbackEvent`).

```bash
harness cost report --since 1h
```

Zoom: tabela `agent × intent × USD`. Fontes reais:
`internal/costtrack/tracker.go` + `reader.go`.

### RECAP (5:30–6:30)
> "Delta: oito centavos. Auth (vinte centavos) + CRUD (oito) = vinte
> e oito no total do backend até aqui."

### CTA (6:30–7min)
> "Próximo: tasks CRUD com BM25 rerank achando o helper de
> autorização que já existe."

---

## Vídeo 5 — Capítulo 5: Tasks CRUD + BM25 rerank

**Título:** `BM25 achou meu helper — Codex não reescreveu`
**Duração alvo:** 7 min

### HOOK (0–15s)
> "O helper de autorização de projeto já existe. BM25 rerank do
> harness acha ele por keyword match no índice de símbolos e promove
> pro pack. O LLM lê em vez de reescrever."

B-roll: [BM25 na Wikipedia](https://en.wikipedia.org/wiki/Okapi_BM25).

### CONTEXTO (15–60s)
> "Tasks aninhadas em `/projects/{id}/tasks`. Enum de status
> `todo|doing|done`. `due_at` nullable. Ownership vem via middleware
> do capítulo anterior."

Referência: [Capítulo 5](../TUTORIAL-BUILD-TASKHIVE.md#chapter-5--feature-tasks-crud-with-permissions).

### DEMO (60s–5min)
```bash
harness feature "add /tasks CRUD nested under /projects/{id}/tasks; \
task has title, description, status enum (todo|doing|done), \
due_at nullable; enforce project ownership via existing middleware" \
  --agent codex --yes
```

Highlight no log:
```
bm25 rerank: promoted internal/auth/project_authorization.go (+0.71)
skipped duplicate middleware generation
```

Referência real: `internal/context/provider_rerank.go` (funções
`bm25Score`/`bm25ScoreTF`, constantes `RerankBM25K1`/`RerankBM25B` em
`internal/platform/constants`) — não existe arquivo `bm25.go` separado.
Diagrama: pack antes / pack depois com o helper subindo pro topo.

### RECAP (5–6min)
> "Esse é o padrão mais comum de economia em brownfield. Sem BM25 o
> Codex reescreveria middleware idêntico e você pagaria os tokens
> duas vezes."

### CTA (6–7min)
> "Backend fechado. Próximo: frontend scaffold — zero LLM, cem por
> cento template."

---

## Vídeo 6 — Capítulo 6: Frontend scaffold determinístico

**Título:** `Vite 6 + React 19 sem gastar 1 token`
**Duração alvo:** 5 min

### HOOK (0–15s)
> "Scaffold não é feature. Template Vite+React+TS é copiado, não
> gerado. Zero LLM. Zero token. Zero não-determinismo."

### CONTEXTO (15–60s)
> "Regra: se a resposta é óbvia, use scaffold. Se é ambígua, use
> feature. Documentação oficial do harness deixa isso explícito."

Referência: [Capítulo 6](../TUTORIAL-BUILD-TASKHIVE.md#chapter-6--frontend-scaffold).
Docs: [Vite 6](https://vite.dev/), [React 19](https://react.dev/blog/2024/12/05/react-19).

### DEMO (60s–3:30)
```bash
cd ~/dev
harness new taskhive-web --stack react --yes
```

B-roll: árvore de arquivos gerada. Zoom: `package.json` com
versões do template real. [VERIFICAR: o CLI real usa
`--stack <name>` (ver `cmd/harness/cmd_new.go`); template disponível
é `react` (React 18) em `internal/scaffoldpkg/templates/react/`. Se o
tutorial usa `--template vite-react-ts`, alinhar tutorial ↔ CLI antes
de gravar.]

### RECAP (3:30–4:30)
> "Custo: zero. Tempo: dois segundos. Próximo call de LLM vai ser
> auth do front."

### CTA (4:30–5min)
> "Amanhã: LLMLingua comprimindo o pack React em 40 por cento."

---

## Vídeo 7 — Capítulo 7: Auth flow no front + LLMLingua

**Título:** `LLMLingua-2 corta 40% do pack React`
**Duração alvo:** 8 min

### HOOK (0–15s)
> "Código React é verboso. Import blocks repetidos, comments mortos.
> LLMLingua-2 corta em torno de quarenta por cento sem perder
> qualidade acima de 0.9."

B-roll: [paper LLMLingua-2](https://arxiv.org/abs/2403.12968) — abstract
piscando.

### CONTEXTO (15–60s)
> "Formulário de login, wrapper de rota protegida lendo token do
> localStorage, axios interceptor que redireciona pra `/login` em
> 401."

Referência: [Capítulo 7](../TUTORIAL-BUILD-TASKHIVE.md#chapter-7--frontend-feature-auth-flow).
Libs: [axios](https://axios-http.com/), [react-router](https://reactrouter.com/).

### DEMO (60s–6min)
```bash
cd taskhive-web
harness feature "login form with email+password, \
protected route wrapper reading token from localStorage, \
axios interceptor that refreshes on 401 by redirecting to /login" \
  --agent claude --yes
```

Highlight no log:
```
llmlingua: compressed pack 42.3KB -> 25.1KB (-40.6%) at quality=0.92
```

Referência real: `internal/context/compress/compress.go` (sub-pacote) e
`internal/context/provider_compress.go` (provider que chama o sub-pacote).

Zoom nos arquivos gerados: `src/hooks/useAuth.ts`,
`src/components/RequireAuth.tsx`, `src/lib/api.ts`.

Alerta técnico:
> "Token em localStorage é padrão do tutorial, não recomendação de
> segurança. Pra produção, HttpOnly cookie + SameSite Strict."

Referência: [OWASP JWT best practices](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html).

### RECAP (6–7min)
> "Delta: dezoito centavos. Quarenta por cento de compressão vira
> uns oito centavos economizados nesse call específico."

### CTA (7–8min)
> "Próximo: listas, modais, updates otimistas."

---

## Vídeo 8 — Capítulo 8: Projects + Tasks views

**Título:** `Optimistic updates em React 19 sem bug`
**Duração alvo:** 7 min

### HOOK (0–15s)
> "Update otimista errado = UI mente. Update otimista certo = UI
> responde em zero ms e converge com o servidor. Codex escreve os
> dois — só que um passa nos testes."

### CONTEXTO (15–60s)
> "Página de lista de projetos, detalhe com tasks agrupadas por
> status, modais de criar/editar, e mudança de status com update
> otimista."

Referência: [Capítulo 8](../TUTORIAL-BUILD-TASKHIVE.md#chapter-8--frontend-feature-project-and-task-views).
Lib: [TanStack Query v5](https://tanstack.com/query/latest) — hook
`useMutation` com `onMutate` + `onError` rollback.

### DEMO (60s–5:30)
```bash
harness feature "project list page, project detail page showing \
its tasks grouped by status, modal for create/edit project, \
modal for create/edit task, optimistic updates on status change" \
  --agent codex --yes
```

B-roll: browser com drag de card mudando status, DevTools Network
tab mostrando PATCH antes da confirmação visual.

### RECAP (5:30–6:30)
> "Delta: doze centavos. Codex acerta CRUD-shape porque já viu mil.
> Se falhar, fallback pra Claude."

### CTA (6:30–7min)
> "Próximo: três agentes, três tipos de teste, um vídeo."

---

## Vídeo 9 — Capítulo 9: Testes (Codex + Kimi + Claude)

**Título:** `3 agentes pra 3 camadas de teste`
**Duração alvo:** 9 min

### HOOK (0–15s)
> "Backend com Codex. Front com Kimi. E2E com Claude. Cada um pelo
> motivo certo. Vinte e cinco centavos no total."

### CONTEXTO (15–60s)
> "Regra: teste de handler = padrão, Codex serve. Vitest de hook =
> patterned, Kimi serve. Playwright = timing async e seletor, aí
> precisa de Sonnet."

Referência: [Capítulo 9](../TUTORIAL-BUILD-TASKHIVE.md#chapter-9--tests).
Docs: [`net/http/httptest`](https://pkg.go.dev/net/http/httptest),
[Vitest](https://vitest.dev/), [Playwright 1.5x](https://playwright.dev/).

### DEMO (60s–7min)

Backend:
```bash
cd ~/dev/taskhive
harness feature "add table-driven tests for every HTTP handler \
using net/http/httptest; cover happy path, auth failure, \
cross-user 404, validation 400" \
  --agent codex --yes
```

Frontend:
```bash
cd ~/dev/taskhive-web
harness feature "add Vitest tests for useAuth hook, \
project list rendering, and the axios 401 interceptor" \
  --agent kimi --yes
```

E2E:
```bash
cd ~/dev
harness feature "add Playwright test at ./e2e: \
register -> login -> create project -> add three tasks -> \
mark one done -> assert count on dashboard" \
  --agent claude --yes
```

B-roll: `npx playwright test --reporter=list` rodando headless.

Alerta CLAUDE.md:
> "Headless sempre. Nunca browser aberto em task longa. Regra do
> projeto — memória do Mac é finita."

### RECAP (7–8min)
> "Três chamadas. Vinte e cinco centavos. Cobertura das três camadas."

### CTA (8–9min)
> "Próximo: Docker Compose com mem_limit obrigatório."

---

## Vídeo 10 — Capítulo 10: Docker Compose + healthcheck

**Título:** `docker compose v2 com mem_limit ou não sobe`
**Duração alvo:** 9 min

### HOOK (0–15s)
> "Container sem `mem_limit` come RAM até travar o host. Neste
> projeto é regra dura do CLAUDE.md. Sem limite, sem merge."

### CONTEXTO (15–60s)
> "Multi-stage Dockerfile: builder Go + distroless final. Nginx
> alpine servindo o build do Vite. `mem_limit` 512m no backend,
> 128m no nginx. Healthcheck nos dois."

Referência: [Capítulo 10](../TUTORIAL-BUILD-TASKHIVE.md#chapter-10--docker-compose-and-ci).
Docs: [Docker Compose v2](https://docs.docker.com/compose/),
[distroless](https://github.com/GoogleContainerTools/distroless).

### DEMO (60s–7:30)
```bash
harness feature "multi-stage Dockerfile for the Go backend \
(builder + distroless final), nginx-alpine for the built frontend, \
docker-compose.yml with mem_limit 512m on backend + 128m on nginx, \
healthchecks on both, shared network, sqlite volume" \
  --agent claude --yes
```

Highlight: bloco YAML do `mem_limit` no `docker-compose.yml` gerado.

Sequência obrigatória do CLAUDE.md:
```bash
docker compose build --no-cache
docker compose up -d
timeout 60s bash -c \
  'until curl -sf http://localhost:8080/healthz > /dev/null; \
   do sleep 2; done' && echo ok
docker compose down --remove-orphans
```

B-roll: `docker stats --no-stream` mostrando ambos os containers
abaixo do teto.

### RECAP (7:30–8:30)
> "Compose sobe. Healthz responde. Compose desce. Regra do projeto
> cumprida."

### CTA (8:30–9min)
> "Próximo: dashboard React nativo do harness."

---

## Vídeo 11 — Capítulo 11: Dashboard + perf snapshot

**Título:** `Dashboard do harness na porta :0 — sem colisão`
**Duração alvo:** 6 min

### HOOK (0–15s)
> "`--addr 127.0.0.1:0`. Dois-pontos-zero. O kernel escolhe a porta,
> imprime, e você nunca mais colide com o dev server de outro
> projeto."

### CONTEXTO (15–60s)
> "Endpoints reais em `internal/adapters/http/handlers_*.go`:
> `/api/runs`, `/api/agents`, `/api/sensors`, `/api/cost`,
> `/api/memory`, `/api/logs`, `/api/autonomy`, `/api/health/score`,
> `/api/design`, `/api/roadmap`, `/api/features`, `/api/toggles`,
> `/api/profile`. Todos em JSON. SPA React lê e renderiza."

Referência: [Capítulo 11](../TUTORIAL-BUILD-TASKHIVE.md#chapter-11--dashboard-and-observability).
Fontes reais: `cmd/harness/cmd_dashboard.go`,
`internal/app/dashboardcmd/dashboardcmd.go`,
`internal/adapters/http/` (handlers). O pacote `internal/dashboardapi/`
NÃO existe.

### DEMO (60s–4:30)
```bash
harness dashboard --addr 127.0.0.1:0
```

Zoom: linha impressa `listening on 127.0.0.1:54721` (porta variável).

Browser: tour rápido pelas cinco abas.

```bash
harness perf-snapshot --label chapter-11-baseline
```

B-roll: JSON do snapshot no filesystem.

### RECAP (4:30–5:30)
> "Custo: zero. Observabilidade: cinco endpoints, um SPA, um
> baseline salvo."

### CTA (5:30–6min)
> "Próximo: análise de custo. Quanto o context engine economizou."

---

## Vídeo 12 — Capítulo 12: Cost analysis + benchmark

**Título:** `US$ 1,23 vs US$ 5,00 — o que o context engine cortou`
**Duração alvo:** 9 min

### HOOK (0–15s)
> "Um dólar e vinte e três, com context engine. Três a cinco dólares,
> sem. O `benchmark cost` reroda os mesmos prompts com pack
> lexicográfico e mostra a diferença."

### CONTEXTO (15–60s)
> "Quatro alavancas: cache-aware layout, salience reorder, BM25
> rerank, LLMLingua compression. Cada uma tem uma linha no
> `harness cost report --breakdown`."

Referência: [Capítulo 12](../TUTORIAL-BUILD-TASKHIVE.md#chapter-12--cost-analysis).

### DEMO (60s–7min)
```bash
harness cost report --since 24h
harness benchmark cost --scenario taskhive-replay --agent claude-raw
harness cost report --breakdown
```

Highlight na saída sample:
```
cache-aware layout      $0.42 saved   (34%)
salience reorder        $0.18 saved   (15%)
bm25 rerank             $0.29 saved   (24%)
llmlingua compression   $0.33 saved   (27%)
```

Referências inline:
- Prompt caching: <https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching>
- BM25: <https://en.wikipedia.org/wiki/Okapi_BM25>
- LLMLingua-2: <https://arxiv.org/abs/2403.12968>

### RECAP (7–8min)
> "Quando o harness ganha o pão: repositório grande, ciclo iterativo,
> workflow multi-agente. Quando não ganha: greenfield minúsculo. Isso
> está no fim do capítulo do tutorial."

### CTA (8–9min)
> "Próximo: orquestração multi-agente. Debate proposer-critic-decider
> por cinco centavos."

---

## Vídeo 13 — Capítulo 13: Orquestração (flow.yaml)

**Título:** `Proposer → Critic → Decider em flow.yaml`
**Duração alvo:** 8 min

### HOOK (0–15s)
> "Claude propõe. Kimi critica. Claude decide. ADR sai pronto pra
> commit em `docs/adr/`. Cinco centavos."

### CONTEXTO (15–60s)
> "Padrão proposer-critic-decider. Cada step depende do output do
> anterior via `@step.output`. Budget do flow trava em vinte
> centavos."

Referência: [Capítulo 13](../TUTORIAL-BUILD-TASKHIVE.md#chapter-13--multi-agent-orchestration).

### DEMO (60s–6:30)
```yaml
# flow.yaml
name: deploy-strategy-debate
budget_usd: 0.20
steps:
  - id: propose
    agent: claude
    role: proposer
    prompt: "Propose a deploy strategy for TaskHive: \
             fly.io single region vs. render.com vs. self-hosted docker."
  - id: critique
    agent: kimi
    role: critic
    depends_on: [propose]
    prompt: "Given @propose.output, list the two weakest points \
             and one blocker for a two-person team."
  - id: decide
    agent: claude
    role: decider
    depends_on: [propose, critique]
    prompt: "Given @propose.output and @critique.output, \
             pick one option and write the ADR."
```

```bash
harness orchestrate run flow.yaml
```

B-roll: pasta `.harness/artifacts/orchestrate/<run-id>/` com os três
outputs em ordem.

Alerta:
> "ADR entra no repositório. Não confiar cegamente. Ler antes de
> commitar."

### RECAP (6:30–7:30)
> "Três agentes. Um artefato. Cinco centavos. Padrão reutilizável
> pra qualquer decisão arquitetural."

### CTA (7:30–8min)
> "Próximo: bugfix e evolve — quando o código já existe."

---

## Vídeo 14 — Capítulo 14: Bugfix + evolve

**Título:** `harness bugfix vs feature: qual usar quando`
**Duração alvo:** 7 min

### HOOK (0–15s)
> "Feature escreve spec. Bugfix não. Bugfix restringe o plano a um
> fix lógico. Usar o comando errado custa dinheiro e tempo."

### CONTEXTO (15–60s)
> "Simular regressão: deletar `queryClient.invalidateQueries` depois
> de editar task. Lista fica stale. Rodar `harness bugfix` com Codex."

Referência: [Capítulo 14](../TUTORIAL-BUILD-TASKHIVE.md#chapter-14--iterating-bugfix-and-evolve).
Fonte real: `internal/app/workflow/workflow.go` (`Bugfix` mode) +
`internal/app/workflow/runner.go`. Não existe `internal/intent/bugfix.go`.

### DEMO (60s–5:30)
Editar manualmente `web/src/pages/ProjectDetail.tsx`, remover a linha
de invalidate.

```bash
harness bugfix "task list on project detail page goes stale \
after editing a task status; suspect missing invalidation" \
  --agent codex --yes
```

Highlight no diff: a linha volta.

```bash
harness evolve diagnose --json
```

Zoom na saída: propostas de teste preventivo baseadas em churn +
histórico de sensores.

### RECAP (5:30–6:30)
> "Bugfix = agora. Evolve = periódico. Não confundir."

### CTA (6:30–7min)
> "Próximo: `harness ship`. Fecha o loop."

---

## Vídeo 15 — Capítulo 15: Ship

**Título:** `harness ci verde, harness ship, tag assinada`
**Duração alvo:** 6 min

### HOOK (0–15s)
> "`harness ci` verde. `harness ship`. Rebase, PR, merge, tag. Sem
> força bruta em `main`. Sem pular hook."

### CONTEXTO (15–60s)
> "GitFlow: branch de `develop`. `main` só recebe release. Pre-push
> hook roda `make ci` — vet, race, build, e2e de cada fase."

Referência: [Capítulo 15](../TUTORIAL-BUILD-TASKHIVE.md#chapter-15--ship).
Docs: `docs/WORKFLOW.md`, `CONTRIBUTING.md`.

### DEMO (60s–4:30)
```bash
harness ci
harness ship
```

B-roll: `git log --oneline` mostrando o merge commit e a tag.

Alerta CLAUDE.md:
> "Nunca `--no-verify`. Nunca force-push em `main` ou `develop`."

### RECAP (4:30–5:30)
> "Custo: zero. Repositório fechado. Release tag no ar. TaskHive
> shipped."

### CTA (5:30–6min)
> "Último vídeo: outro. Encerra a série com o custo total e a
> rubric de escolha de agente."

---

## Vídeo 16 — OUTRO da série

**Título:** `US$ 1,23. SaaS pronto. 15 vídeos. Fim.`
**Duração alvo:** 5 min

### HOOK (0–15s)
> "TaskHive: rodando, testado, containerizado, com dashboard. Custo
> total em LLM: um dólar e vinte e três centavos. Recibo aberto na
> tela."

B-roll: `harness cost report --breakdown` final.

### CONTEXTO (15–60s)
> "Quinze capítulos. Quinze vídeos. Um repositório. Uma tag. Uma
> confiança nova sobre quando usar qual modelo."

### DEMO (60s–3min)
Tour rápido:
1. Repositório final no GitHub.
2. Tag `v0.1.0` [VERIFICAR: tag real do repositório final].
3. Dashboard aberto no browser.
4. Tabela do Apêndice A (rubric de escolha de agente).

### RECAP (3–4min)
> "Regra que sobra: se um erro vaza dado, acorda alguém, ou
> compõe silencioso — gasta o token do Sonnet ou do Opus. Caso
> contrário, desconto do Codex ou Kimi."

Referência: [Apêndice A](../TUTORIAL-BUILD-TASKHIVE.md#appendix-a--agent-choice-rubric),
[Apêndice B](../TUTORIAL-BUILD-TASKHIVE.md#appendix-b--when-to-skip-agents),
[Apêndice C](../TUTORIAL-BUILD-TASKHIVE.md#appendix-c--debugging-harnessx-itself).

### CTA (4–5min)
> "Próxima série: deploy do TaskHive em produção com observabilidade
> real. Inscreve. Comenta qual capítulo você quer que eu detalhe
> mais. Link do repositório na descrição."

---

## Notas gerais de produção

- Todos os comandos exibidos são copiados literalmente de
  `docs/TUTORIAL-BUILD-TASKHIVE.md`. Nenhuma alteração de sintaxe.
- Versões alvo (tutorial): Go 1.25+, Node 20 LTS, template `react`
  em `internal/scaffoldpkg/templates/react/` usa React 18 / Vite (ver
  `package.json.tmpl`); Playwright 1.5x, Docker Compose v2. Se o
  tutorial cita React 19 / Vite 6, marcar `[VERIFICAR]` — o template
  atual do repo é React 18.
- Marcadores `[VERIFICAR: ...]`:
  1. Preço por token do Codex vs Sonnet (Vídeo 4).
  2. Versões `react`/`vite` no template default do v0.170.1 (Vídeo 6).
  3. Tag de release final do repositório TaskHive (Vídeo 16).
- Voz Rodolfo: frases curtas, verbos no infinitivo ou imperativo,
  zero enchimento. Se uma frase pode ser cortada sem perder sentido
  técnico, cortar.
- Legenda queimada opcional em português BR — recomendo por retenção
  em mudo.
