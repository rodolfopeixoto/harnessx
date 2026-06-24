# HarnessX — Tutorial Canônico (End-to-End)

> Este é **o único** tutorial. Substitui as versões por stack
> (ecommerce, todoist, multi-stack, mobile-desktop), agora removidas
> — git log preserva o histórico se você precisar reler. Aqui você
> sai do zero até uma aplicação **completa em produção** usando todo o
> potencial da ferramenta: spec-driven, TDD, multi-agente, gate
> determinístico, observabilidade de custo, backup.
>
> Tempo total: ~90 min lendo + ~3h aplicando.

---

## 0. Filosofia (por que isso existe)

HarnessX implementa o paper **"Code as Agent Harness"** (`docs/paper/`).
Três ideias centrais:

1. **Plan-Execute-Verify (PEV, §3.1.4 + §3.4):** todo passo passa por
   plano → execução → verificação determinística. Um LLM nunca faz
   merge direto; um sensor checa antes.
2. **Plan-as-contract (§3.4.2):** a spec é a única fonte de verdade.
   Código que diverge da spec é regredido, não promovido.
3. **Multi-agent routing (§3.5):** modelo caro só na implementação;
   planejamento e revisão usam chains baratas. Custo cai 10–40× sem
   queda mensurável de qualidade.

Tradução prática para você: **não precisa colar prompt de "siga clean
code, SOLID, separa constantes"** em toda mensagem. A ferramenta força
isso por:

- sensores deterministícos no `harness ci` (lint + format + security +
  type-check + dead-code + dependency-audit + log-audit) que bloqueiam
  commits que violem,
- regras embarcadas em `internal/platform/constants` que rejeitam
  "magic numbers" duplicados via `harness audit`,
- templates de scaffold (`harness new`) que já vêm com layout SOLID,
  testes, lint configurado.

---

## 1. Setup (uma vez por máquina)

### 1.1 Instalação

```bash
# binário pré-compilado
curl -L https://github.com/rodolfopeixoto/harnessx/releases/latest/download/harness-darwin-arm64.tar.gz \
  | tar xz -C /usr/local/bin

harness version           # confirma v0.151.0+
```

### 1.2 Wizard de onboarding

```bash
harness onboarding --interactive
```

O que cada prompt faz e **por quê**:

| Prompt | O que faz | Por que aceitar |
|---|---|---|
| `install missing tools? (git/uv/rg/jq/rclone)` | `brew install` | `rg` e `jq` são usados pelos sensores; `rclone` por backup |
| `pick the default adapter` | grava `.harness/config/active.yaml` | sem pin, todo `harness chat` precisa de `--adapter` |
| `configure multi-adapter routing?` | grava `.harness/config/routes.yaml` | habilita routing por tag — planning vai para modelo barato, implementação para o forte |
| `show backup recipe?` | imprime `harness backup quickstart` | desligar = sessões só vivem localmente |
| `pick a stack to scaffold` | gera `~/dev/my-api` | esqueleto já SOLID + testes + lint configurado |
| `install python sensor tools?` | `pip install bandit mypy pip-audit` | sem isso `harness ci` pula 3 gates de segurança |
| `configure MCP servers?` | escreve `.harness/mcp/<n>.json` | filesystem/github/postgres/sqlite — agentes ganham ferramentas reais |
| `scaffold Dockerfile + compose?` | gera ambos com `mem_limit: 2g` | enforce-cap garantido do CLAUDE.md |

### 1.3 Verificar pin de adapter

```bash
harness use claude                # pin manual se quiser sobrescrever
cat .harness/config/active.yaml
```

### 1.4 Configurar routes (multi-agente)

`.harness/config/routes.yaml` gerado pelo wizard fica assim:

```yaml
routes:
  planning:
    primary: gemini       # ~$0.0002/1k in, ótimo para perguntas curtas
    fallback: [claude, kimi]
  implementation:
    primary: claude       # opus-4-7, $0.015/1k in, ganha em refactor
    fallback: [codex, gemini]
  cheap_review:
    primary: kimi         # ~$0.0001/1k in, perfeito para "isso passou?"
    fallback: [gemini, claude]
```

Por que importa: dentro de `harness chat`, `/plan` → `planning`, texto
puro → `implementation`, `/recap` → `cheap_review`. **Você não escolhe
por mensagem; o router resolve.**

---

## 2. Scaffold do projeto

Vamos construir **uma aplicação Todoist completa** com FastAPI (API) +
React (web) + Tauri (mobile/desktop). Backend primeiro.

```bash
harness new python-ecommerce ~/dev/todoist-api --yes --with-deps
cd ~/dev/todoist-api
```

O que `harness new python-ecommerce` traz **embarcado** (não precisa
pedir):

- `pyproject.toml` com `ruff` (lint+format), `pytest`, `bandit`,
  `mypy`, `pip-audit` configurados,
- `app/` separando `domain/`, `service/`, `adapter/` (clean arch),
- `tests/` com fixtures + conftest,
- `internal/constants.py` para constantes compartilhadas (rejeita
  "magic numbers" duplicados),
- `Dockerfile` distroless + `docker-compose.yml` com mem cap,
- `.harness/config/project.yaml` apontando os sensores.

```bash
tree -L 2 app tests
```

Veja a separação. **Isso é o anti-prompt: você não precisa pedir clean
arch; ela já está.**

---

## 3. Primeira spec (interativa, editável)

Em vez de pedir "implementa /tasks endpoint" e rezar, você
**escreve a spec primeiro** com o agente fazendo perguntas.

```bash
harness spec author "API REST de tarefas: CRUD com filtro por status, paginação, validação de campos obrigatórios"
```

### 3.1 Fase Q&A (clarifying questions)

Aparecem 5 perguntas-base + 1–3 contextuais geradas pelo LLM:

```
clarifying questions — empty line skips optional ones
* Who uses this feature? (role / persona): mobile/web clients via JSON
* What does success look like in observable terms?: 200 com lista paginada, 201 ao criar, 404 quando id não existe
  What is explicitly OUT of scope?: auth (próxima spec), websockets
  Any known risks, edge cases, or constraints?: SQL injection, validação de title obrigatório
  How will we test it (unit / e2e / manual)?: pytest unit + httpx async client
  How big the result page can be?: max 50 itens, default 20
```

**Por que perguntas baseline + contextuais:** as 5 baseline (users,
acceptance, scope_out, risks, tests) cobrem tudo que uma spec
profissional precisa. As contextuais (1-3) vêm da chain `planning`
(modelo barato) e pegam contexto do prompt — neste caso "quanto vem
por página?".

### 3.2 Draft gerado

Sai um markdown com seções:

```
## Summary
## Users
## Acceptance Criteria
## Out of Scope
## Risks
## Test Plan
## Implementation Notes
```

### 3.3 Edit loop (você verifica TUDO)

```
spec> /show                                  # imprime o draft
spec> /sections                              # lista H2 headings
spec> /edit                                  # abre $EDITOR (vim/nano)
spec> /refine Implementation Notes: detalhe os endpoints com paths e bodies
spec> /diff                                  # mostra delta da última revisão
spec> /undo                                  # reverte 1 revisão
spec> /expand Risks                          # LLM adiciona detalhe
spec> /shrink Summary                        # LLM enxuga
spec> /save                                  # persiste + history.jsonl
```

**Verificação humana a cada passo:** depois de qualquer `/refine`,
`/expand` ou `/shrink`, sempre `/diff` + decisão `/undo` ou seguir.
A spec só é gravada com `/save` — você é o gate final.

**Sob o capô:**

- `/refine` chama o adapter com task tag `planning` → router pega
  chain barata (`gemini`/`kimi`), gasta ~$0.001 por refine,
- toda revisão é registrada em
  `.harness/artifacts/specs/<id>.history.jsonl` (estrutura
  `{time, source, section, body}`),
- `/save` escreve `.harness/artifacts/specs/<id>.md` com header
  `<!-- harness-spec-id: ... -->` que outros comandos consomem.

### 3.4 Adicionar / alterar spec depois

```bash
# nova spec adjacente
harness spec author "adicionar autenticação JWT às rotas de /tasks"

# editar uma existente
$EDITOR .harness/artifacts/specs/01kvvd...md

# spec por nome curto (versão antiga, integra com harness feature)
harness spec init "auth" --name auth --mode feature
```

---

## 4. Do plan ao código (TDD via `harness drive`)

Spec na mão, agora vem o ciclo PEV completo:

```bash
harness drive --features .harness/artifacts/specs/01kvvd...md
```

Por baixo:

1. **Plano (`intentplan.Plan`)** — chain `planning` lê a spec e emite
   JSON com steps (`{action, args, sensor, gate}`),
2. **Test scaffolding** — chain `cheap_review` gera arquivo de teste
   sob `tests/` que **falha** (pq impl não existe),
3. **Verificação 1:** `harness test` confirma 1+ falha esperada,
4. **Implementação** — chain `implementation` (modelo caro, claude)
   edita `app/` para passar os testes; proibido editar `tests/`,
5. **Gate (`harness ci`)** — todos os sensores; se vermelho retorna
   ao passo 4 (até `--max-attempts=3`),
6. **Commit conventional** na branch atual; **sem auto-merge**, você
   abre PR.

### 4.1 Sensores que rodam no gate

Cada um é determinístico (zero LLM):

| Sensor | O que checa | Bloqueia se |
|---|---|---|
| `py_ruff` | lint + estilo PEP8 + import order | warning ou erro |
| `py_ruff_format` | formatação | drift |
| `py_pytest` | testes unitários | falha ou cobertura < 90% |
| `py_mypy` | type-check estrito | erro de tipo |
| `py_bandit` | OWASP top 10 | severity ≥ low |
| `py_pip_audit` | CVE em deps | qualquer CVE conhecido |
| `secrets_scan` | regex de tokens | match em arquivos tracked |
| `dead_code` | imports/funções não-usados | qualquer |
| `dep_audit` | versões pinadas | unpinned |
| `log_audit` | PII em logs | match de email/cpf/cartão |

**Por isso você não precisa pedir "siga boas práticas":** se o agente
escrever código que viola, o gate barra. Ele aprende ou volta.

### 4.2 Custo deste ciclo

```bash
harness analytics --since 1h
```

Saída típica:

```
by stack
  STACK       SESS  TURN  CHAT     IN       OUT       COST
  python      1     12    8        14200    3100      $0.1842

by adapter / task
  ADAPTER     TASK              TURNS    COST
  claude      implementation    4        $0.1700
  kimi        cheap_review      6        $0.0089
  gemini      planning          2        $0.0053
```

**Leitura:** 95% do gasto foi nas 4 turns de claude/implementation —
exatamente onde precisa de força. Se você não tivesse routing, todas
as 12 turns seriam claude/$0.45+.

#### Onde aparece o custo

| Lugar | Quando |
|---|---|
| **inline em cada turn** do `harness chat`: `✓ claude done in 38.4s · in=101124 out=1158 · ~$0.2644` | sempre, após cada chamada de adapter |
| **`/cost`** dentro do chat | soma da sessão atual por adapter/task |
| **`harness analytics --since <d>`** | cross-sessão, cross-projeto |
| **`.harness/sessions/<id>.jsonl`** | persistência crua (cada turn é um JSON) |
| **`/btw <pergunta>`** | resposta na chain mais barata, ~$0.0001 por turn |

---

## 5. `harness chat` — para quando não dá pra escrever spec antes

Spec é o ideal. Mas exploração rápida ("será que isso compila?")
roda em chat:

```bash
harness chat --auto-gate
```

`--auto-gate` faz `harness ci` rodar **depois de cada turn do
agente**. Verde, segue. Vermelho, ele revisa.

Slashes essenciais (`/help` lista todos):

| Slash | Uso | Chain |
|---|---|---|
| texto puro | conversar com agente pinado | implementation |
| `/exec <p>` | plano determinístico + exec | planner local + impl |
| `/do <p>` | alias para `/exec` | — |
| `/plan <p>` | só planejar, não executar | planning (cheap) |
| `/drive <p>` | spec → teste → impl → ci | mix |
| `/spec <p>` | spec author loop dentro do chat | planning |
| `/ship <p>` | branch + spec + impl + commit | mix |
| `/ci`, `/test`, `/lint` | gates manuais | nenhum (determ) |
| `/cost` | gasto cumulativo da sessão | nenhum |
| `/recap` | resumo da sessão | cheap_review |
| `/btw <q>` | pergunta lateral curta sem mexer no fluxo | cheap_review |
| `/cycle` | rotaciona para o próximo adapter registrado | — |
| `/login` | dispara o comando de login do adapter pinado | — |
| `/auto-gate` | liga/desliga gate automático | nenhum |
| `/use <id>` | troca adapter ad-hoc | — |
| `/budget <usd>` | corta sessão ao estourar | nenhum |
| `/save <label>` | rotula sessão p/ resume | nenhum |
| `!<cmd>` | shell direto | nenhum |

### 5.1 Determinístico vs não-determinístico — quando usar qual

| Cenário | Use | Por quê |
|---|---|---|
| "lint passa?" | `harness ci` | sensor já existe, zero tokens |
| "atualiza versão" | `!sed -i ...` ou `harness new --upgrade` | reproduzível |
| "refactor 200 linhas" | chat com claude | LLM lê contexto, propõe |
| "explica esse trecho" | `/plan` (cheap) | resposta curta, modelo barato |
| "quais MCP servers carregam?" | `harness mcp scan` | comando determinístico |
| "gera teste de regressão para esse bug" | `/drive` | TDD, gate determinístico no fim |

**Regra de bolso:** se um sensor faz, sensor faz. LLM só quando há
ambiguidade ou geração nova.

---

## 6. Frontend (React) com a mesma loop

```bash
harness new react ~/dev/todoist-web --yes --with-deps
cd ~/dev/todoist-web
harness chat --auto-gate
```

```
/spec listagem de tarefas: GET /tasks, render <li> por item, vitest com mock fetch
```

Q&A → draft → /edit ajustes → /save → `/drive` para implementar.

O scaffold `react` já vem com:

- vite + vitest + testing-library configurado,
- eslint + prettier,
- `src/api/` separado de `src/components/`,
- `tsconfig.json` strict.

---

## 7. Mobile + desktop (Tauri)

```bash
mkdir ~/dev/todoist-mobile && cd ~/dev/todoist-mobile
cargo create-tauri-app --manager npm --template vanilla-ts --tauri-version 2
cd todoist
npm install && harness init --yes
harness chat --auto-gate
```

```
/spec view de tarefas em src/main.ts, GET http://localhost:8000/tasks, vitest com fetch mockado
```

`npm run tauri ios dev` / `tauri android dev` / `tauri dev` (desktop)
abrem cada plataforma. **Mesmo backend.**

---

## 8. Observabilidade contínua

### 8.1 Analytics cross-project

```bash
harness analytics \
  --root ~/dev/todoist-api \
  --root ~/dev/todoist-web \
  --root ~/dev/todoist-mobile \
  --since 168h
```

Reporta por stack, adapter+task, dia. Use semanalmente para detectar:

- task tag dominando custo → ajustar `routes.yaml` (ex: mover
  `implementation` para codex se claude saiu do orçamento),
- stack mais cara → talvez precisa de melhor scaffold ou MCP server.

### 8.2 JSON stream para CI

```bash
echo "/help" | harness chat --output-json > turns.jsonl
```

Cada turn vira `{session, time, input, action, adapter_id, task_tag,
in_tokens, out_tokens, cost_usd, ok}`. Pipe para Grafana / S3 /
Datadog.

### 8.3 Sessões

```bash
harness chat list
harness chat --resume <id-or-label>
harness chat --replay <id> # read-only, só /history /agents /cost /diff
```

---

## 9. Backup

```bash
harness backup quickstart      # imprime recipe interativa
harness backup remote add gdrive --provider drive --interactive
harness backup config set-default-remote gdrive
harness backup snapshot        # cron-friendly
```

O que vai pro snapshot: `.harness/sessions/`, `.harness/artifacts/`,
`.harness/config/`. Nada de código (assumimos git).

---

## 10. Workflow diário (TDD + multi-agente — receita oficial)

```bash
# 1. Sincroniza
cd ~/dev/todoist-api
git fetch && git checkout develop && git pull --rebase

# 2. Branch
git checkout -b feature/F<N>-<slug> develop

# 3. Spec da feature (editar até concordar)
harness spec author "<descrição>"

# 4. Drive — TDD + impl + gate
harness drive --features .harness/artifacts/specs/<id>.md

# 5. (opcional) iterar via chat para ajustes finos
harness chat --auto-gate --resume <session-label>

# 6. Snapshot performance se mexeu em hot path
harness perf-snapshot

# 7. Push (pre-push hook roda make ci)
git push -u origin feature/F<N>-<slug>

# 8. PR via gh
gh pr create --base develop --fill
```

**Regras do projeto que o gate enforce (sem você lembrar):**

- Conventional Commits ≤ 50 char (hook `commit-msg`)
- GitFlow: feature/ → develop → release/ → main (hook bloqueia push
  direto)
- `make ci` antes de push (pre-push hook; bypass só com
  `HARNESS_SKIP_CI=1` em hotfix documentado)
- Cobertura ≥ 90% nos arquivos novos (sensor `coverage`)
- Sem CGO; só `modernc.org/sqlite`
- Constantes compartilhadas em `internal/platform/constants` (Go) /
  `internal/constants.py` (Python) — sensor reclama se duplicar

---

## 11. Conceitos sob o capô (para entender o porquê)

### 11.1 PEV loop (paper §3.4)

Cada step de plan = `Plan → Execute → Verify`. Verify é
determinístico (sensor). Se vermelho, replan ou abort. **Nunca**
"execute + commit + reza".

### 11.2 Plan-as-contract (paper §3.4.2)

A spec é o contrato. O agente pode propor revisão da spec, mas a
mudança precisa passar pelo edit loop (você aprova). Isso evita
"scope creep" silencioso onde o LLM resolve ao mesmo tempo o que
pediu + 5 coisas a mais.

### 11.3 Sensor catalogue (paper §3.6)

Cada linguagem expõe um vetor de sensores nomeados em
`internal/sensors/catalog.go`. Eles têm contrato `Detect → Run →
Result{OK,Failures[]}` e o agente lê só `Result` (não a stderr crua).
Reduz contexto + permite paralelizar.

### 11.4 Multi-agent router (paper §3.5)

`internal/router/router.go` mapeia task tag → chain ordenada de
adapters com fallback em erro recoverable. Routes vêm de
`routes.yaml` (do wizard) sobre `Defaults(registry)`.

### 11.5 Context engineering (paper §4)

`internal/context.Build()` é o único caminho aprovado para enviar
arquivos ao LLM. Ele:

- pega só arquivos relevantes (via `harness context`),
- recorta ao orçamento de tokens do adapter,
- nunca envia binários, `.env`, `node_modules/`, etc.

Por isso `harness ask "como funciona X"` é barato.

### 11.6 VCR para testes E2E (`internal/agents/vcr`)

Grava chamadas reais do agente em fixture; reproduz sem token. Use
quando quiser teste integrado mas determinístico.

### 11.7 Skills & Hooks

`harness skill list` / `harness hook list` — atalhos reutilizáveis
(ex: `git-pre-push`, `pre-commit-format`). Você instala via
`harness hook install <name>`.

---

## 12. Checklist de adoção (mark every box antes de chamar de "em produção")

- [ ] `harness onboarding --interactive` rodou e gravou `active.yaml`
      + `routes.yaml`
- [ ] `harness ci --install-missing` instalou bandit/mypy/pip-audit
- [ ] `make install-hooks` instalou pre-push gate
- [ ] Backup config-urado com `harness backup remote add`
- [ ] MCP servers `filesystem` + `github` instalados via
      `harness mcp install`
- [ ] `Dockerfile` + `docker-compose.yml` com `mem_limit` (auto pelo
      wizard)
- [ ] Pelo menos 1 spec gravada em `.harness/artifacts/specs/`
- [ ] Pelo menos 1 ciclo `harness drive` completo com gate verde
- [ ] `harness analytics --since 24h` mostra distribuição
      planning/implementation/cheap_review (não 100% no caro)
- [ ] `git log` mostra Conventional Commits + nenhum push direto a
      main/develop

---

## 13. Onde ler mais

- `HARNESSX-MASTER-PLAN.md` — roadmap completo
- `docs/paper/` — original "Code as Agent Harness"
- `docs/agents.md` — adapter contract
- `docs/sensors.md` — como adicionar sensor novo
- `docs/skills.md` — skills + hooks
- `docs/context-engineering.md` — paper §4 aplicado
- `docs/resource-optimization.md` — CGO-free, limites de mem
- `docs/security.md` — política de segredos, log audit, secret scan
- `CHANGELOG.md` — o que mudou em cada release
- `CONTRIBUTING.md` — GitFlow, hooks, padrão de PR

---

Fim. Pronto para construir aplicação **completa**, sem precisar colar
prompt de "siga boas práticas" — a ferramenta força.
