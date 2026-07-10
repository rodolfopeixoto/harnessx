# DIDATICA-CONCEITOS.md — Camada pedagógica da série TaskHive/harnessx

> Agent 2 (Pedagogo). Companion de `docs/TUTORIAL-BUILD-TASKHIVE.md`
> e do `HARNESSX-MASTER-PLAN.md`. Serve pra roteiro dos 15 vídeos da
> série YouTube (PT-BR, ≤10min cada). Aqui não tem código pronto —
> tem *como explicar* cada peça pra quem nunca ouviu falar.

Público-alvo: dev júnior/pleno brasileiro que já mexeu com Git e um
framework web, mas nunca usou LLM em fluxo de trabalho sério.
Registro: informal, sem gringuês desnecessário, analogia de padaria
sempre que possível.

---

## 1. Glossário de conceitos-chave (40 termos)

Cada termo segue o mesmo esqueleto:
**O QUE É → ANALOGIA BR → POR QUE USAR → COMO SERIA SEM →
VANTAGENS/DESVANTAGENS → REFERÊNCIA → COMANDO/SNIPPET DEMO**.

### 1.1 — harness (harnessx)

- **O que é.** CLI Go que orquestra LLMs, sensores, contexto e
  workflow de dev num binário só.
- **Analogia BR.** O *gerente da padaria*. Ele não faz pão, mas
  decide qual padeiro pega qual encomenda, cronometra o forno e
  garante que ninguém coloque sal no lugar do açúcar.
- **Por que usar.** Padroniza spec→plan→run→check, corta token,
  audita custo, prende regra de arquitetura no CI.
- **Como seria sem.** Você abre o ChatGPT no browser, copia o
  repositório inteiro pra dentro do prompt e reza. Sem cache, sem
  auditoria, sem gate.
- **Vantagens.** Single binary, offline-friendly (`fake` adapter),
  multi-agent, memória vault local.
- **Desvantagens.** Curva de aprendizado, precisa spec-first,
  ainda é v0.170 (API pode mudar).
- **Referência.** Original harnessx — `HARNESSX-MASTER-PLAN.md §1-4`.
- **Demo.** `harness init && harness doctor`

### 1.2 — Agent (adapter de LLM)

- **O que é.** Plugin que fala com um provedor de LLM específico
  (Claude, Codex, Kimi, Gemini, Qwen, etc.).
- **Analogia BR.** Cada agent é *um padeiro diferente*: um é bom
  em sovar (Codex, implementação), outro é bom em decorar bolo
  (Claude, spec), outro faz pão de queijo por atacado barato
  (Kimi, testes).
- **Por que usar.** Modelo certo pra tarefa certa reduz custo 3-5x.
- **Como seria sem.** Rodar tudo no GPT-4o do browser — caro e
  lento.
- **Vantagens.** Fallback automático, custo auditável, health
  check.
- **Desvantagens.** Precisa API key de cada provedor.
- **Referência.** `internal/orchestrate/router.go`; conceito
  parecido em LangChain agents (https://python.langchain.com/docs/modules/agents/).
- **Demo.** `harness agent list && harness use claude`

### 1.3 — Sensor

- **O que é.** Regra estática que roda em cima do patch antes de
  aplicar (secret scan, forbidden files, embed `..`, license
  drift).
- **Analogia BR.** *Detector de metal na entrada do show*.
  Ninguém entra com faca. Sensor não pergunta pro LLM se pode —
  ele bloqueia.
- **Por que usar.** LLM às vezes vaza `.env` ou apaga teste;
  sensor barra antes do disco.
- **Como seria sem.** Você descobre a chave AWS commitada no
  GitHub 2h depois, quando a fatura já explodiu.
- **Vantagens.** Determinístico, custo zero, roda em CI.
- **Desvantagens.** Falso positivo em código legado.
- **Referência.** `internal/sensors/`; inspiração em
  `git-secrets` (https://github.com/awslabs/git-secrets) e
  Semgrep (https://semgrep.dev/docs/).
- **Demo.** `harness sensor list`

### 1.4 — Skill

- **O que é.** Pacote reusável de conhecimento + prompt +
  ferramentas que o agent carrega sob demanda (Phase 6+).
- **Analogia BR.** *Receita plastificada* que o padeiro pega da
  parede quando o cliente pede sonho de creme. Não precisa
  decorar 200 receitas — pega quando precisa.
- **Por que usar.** Contexto sob demanda; não polui o prompt
  base.
- **Como seria sem.** Prompt de sistema gigante de 20k tokens
  com "instruções pra tudo".
- **Vantagens.** Modular, versionável, testável.
- **Desvantagens.** Descoberta (qual skill invocar quando?) ainda
  é aberta.
- **Referência.** `HARNESSX-MASTER-PLAN.md §13`; Anthropic Skills
  https://docs.claude.com/en/docs/agents-and-tools/agent-skills/overview.
- **Demo.** `ls ~/.claude/skills/`

### 1.5 — Spec (specification)

- **O que é.** Documento Markdown gerado pelo agent com user
  story, critérios de aceite, non-goals, layout de arquivos.
- **Analogia BR.** *Comanda do garçom* antes do chef cozinhar.
  Sem comanda, o chef inventa e sai errado.
- **Por que usar.** Contrato revisável antes de queimar token
  em código.
- **Como seria sem.** Você prompta "faz auth" e recebe 400
  linhas que você nem consegue revisar.
- **Vantagens.** Trava escopo, vira ADR grátis.
- **Desvantagens.** 2 min a mais por feature.
- **Referência.** `docs/spec-driven-development.md`; método
  Spec-First (https://github.com/github/spec-kit).
- **Demo.** `harness feature "add login" --yes`

### 1.6 — Plan

- **O que é.** Arquivo derivado da spec com: arquivos a criar,
  arquivos a modificar, budget de token, adapter escolhido,
  fallback chain.
- **Analogia BR.** *Planta baixa do pedreiro* antes de subir a
  parede. Você vê onde vai a porta antes de furar.
- **Por que usar.** Você reprova ANTES do custo LLM.
- **Como seria sem.** Roda direto, gasta USD 0.30, descobre que
  ele tocou 12 arquivos que você não queria.
- **Vantagens.** Preview de custo e escopo.
- **Desvantagens.** Uma etapa extra pra iniciante.
- **Referência.** `internal/plan/`; conceito de "plan mode" tipo
  Claude Code (https://docs.claude.com/en/docs/claude-code/overview).
- **Demo.** `harness plan`

### 1.7 — Run (workflow router)

- **O que é.** `harness run "<prompt>"` (ou a forma natural
  `harness "<prompt>"`) classifica o intent e roteia pro workflow
  correto (ask / plan / feature / bugfix). Por baixo, o adapter
  gera patch e os sensores rodam via `harness ci`.
- **Analogia BR.** *Recepcionista.* Você diz o que precisa e ele
  te manda pra fila certa.
- **Por que usar.** Um comando cobre a maioria dos casos sem você
  memorizar o verbo exato.
- **Como seria sem.** Escolher manualmente entre `ask`, `plan`,
  `feature`, `bugfix` sempre.
- **Vantagens.** Ergonomia; classifier prende no workflow certo.
- **Desvantagens.** Menos controle explícito.
- **Referência.** `cmd/harness/cmd_workflow.go`.
- **Demo.** `harness run "add /health endpoint"`

### 1.8 — Check e CI

- **O que é.** `harness check` roda todos os sensores aplicáveis
  em modo **informativo** (exit 0 mesmo com red).
  `harness ci` roda os mesmos sensores em modo **enforcing**
  (exit non-zero em qualquer red) — é o que o hook `pre-push`
  dispara.
- **Analogia BR.** *Provador de sopa* (check) vs *fiscal
  sanitário* (ci) antes de servir o cliente.
- **Por que usar.** Iteração durante dev com `check`, gate duro
  antes do push com `ci`.
- **Como seria sem.** Você descobre o bug em produção.
- **Referência.** `cmd/harness/cmd_check.go`,
  `cmd/harness/cmd_ci.go`; `docs/FEATURES.md §Phase 4`.
- **Demo.** `harness check` (dev) / `harness ci` (push gate)

### 1.9 — Salience (reorder por relevância)

- **O que é.** Reordenar arquivos no pack de contexto de forma
  que os mais relevantes fiquem NO TOPO e o rodapé fique
  byte-idêntico entre chamadas (pra preservar prefix cache do
  provedor).
- **Analogia BR.** *Vitrine da padaria de manhã.* Pão quente na
  frente (foco), café e açucareiro no fundo, sempre no mesmo
  lugar (cache).
- **Por que usar.** LLM presta mais atenção no topo E o
  provedor cobra menos porque o rodapé bateu no cache.
- **Como seria sem.** Ordem alfabética — a atenção do modelo
  vai pro que começa com `a`.
- **Vantagens.** 10-20% menos token, +qualidade.
- **Desvantagens.** Precisa symbol index atualizado.
- **Referência.** `internal/context/reorder.go` (head/tail
  interleave contra "lost in the middle"); Liu et al. 2023
  https://arxiv.org/abs/2307.03172; `docs/PAPER-IMPLEMENTATION.md
  §2.1`.
- **Demo.** Log do run: `context: pack=18 reorder applied`

### 1.10 — BM25 rerank

- **O que é.** Algoritmo clássico de recuperação de informação
  (não-neural) que dá score a documentos por keyword
  matching. Usado pra promover arquivos existentes ao pack.
- **Analogia BR.** *Bibliotecária que conhece a estante de cor.*
  Você pede "livro sobre auth" e ela puxa o volume certo sem
  precisar de IA.
- **Por que usar.** Evita LLM reimplementar helper que já
  existe.
- **Como seria sem.** Codex escreve `authorizeProject` do zero
  porque não viu o `project_authorization.go` já pronto.
- **Vantagens.** Custo zero, provado desde 1994.
- **Desvantagens.** Só keyword — perde sinônimo semântico
  (por isso vai combinado com salience).
- **Referência.** `internal/context/provider_rerank.go` (BM25 +
  LSP symbol bonus, trim a `MaxKeep=40`); Robertson &
  Zaragoza 2009 "The Probabilistic Relevance Framework"
  (https://www.staff.city.ac.uk/~sbrp622/papers/foundations_bm25_review.pdf);
  `docs/PAPER-IMPLEMENTATION.md §2.2`.
- **Demo.** Stat no pack: `FilesReranked` > 0.

### 1.11 — LLMLingua / compressão de contexto

- **O que é.** Compressor **determinístico** (não é o modelo
  neural do paper) que remove linhas vazias, colapsa whitespace
  e trunca **prosa** por um `Ratio` (~0.6), **preservando código
  verbatim** (fenced e indent-based). Inspirado em LLMLingua,
  não uma reimplementação dela.
- **Analogia BR.** *Resumo do capítulo do vestibulando.* Não
  precisa reler o livro inteiro — só o que cai na prova.
- **Por que usar.** Ganho típico ~35% em README/docs prosa-heavy;
  não mexe em código.
- **Como seria sem.** Manda os 42KB crus e paga por eles.
- **Vantagens.** Zero LLM no loop → determinístico e cacheable.
- **Desvantagens.** Sem ganho em pack all-code; ratio agressivo
  demais corta contexto.
- **Referência.** `internal/context/compress/` +
  `internal/context/provider_compress.go`; paper LLMLingua-2
  Pan et al. 2024 (https://arxiv.org/abs/2403.12968);
  `docs/PAPER-IMPLEMENTATION.md §2.3`.
- **Demo.** Bench: `go test -bench=. ./internal/context/`

### 1.12 — Context engineering

- **O que é.** Disciplina de montar o pack de contexto (o que
  entra, em que ordem, comprimido como) pra maximizar
  qualidade/custo.
- **Analogia BR.** *Mala pra viagem de 3 dias.* Você não leva o
  guarda-roupa inteiro — leva o que combina com o roteiro.
- **Por que usar.** Diferença entre USD 1 e USD 5 no mesmo
  projeto.
- **Como seria sem.** Contexto lexicográfico, sem cache, sem
  compressão.
- **Vantagens.** Ganho composto (cache + salience + BM25 +
  compressão).
- **Desvantagens.** Precisa índice atualizado.
- **Referência.** `docs/context-engineering.md`; termo
  popularizado por Andrej Karpathy 2024
  (https://x.com/karpathy/status/1937902205765607626).
- **Demo.** `harness context build "auth"`

### 1.13 — Cache-aware layout

- **O que é.** Ordenar o pack pra que o PREFIXO (topo) seja
  byte-idêntico entre chamadas — assim o provedor (Anthropic,
  OpenAI) reusa o cache dele e cobra até 90% menos naqueles
  tokens.
- **Analogia BR.** *Cardápio fixo do rodízio.* Se o garçom
  sempre traz na mesma ordem, você já sabe. O cérebro (cache)
  não precisa reprocessar.
- **Por que usar.** Redução direta na fatura da API.
- **Como seria sem.** Cada call é prompt novo — zero cache
  hit.
- **Vantagens.** Ganho maior nos providers com prompt cache
  (Anthropic, Gemini).
- **Desvantagens.** Precisa disciplina — 1 byte muda no topo
  e o cache já era.
- **Referência.** Anthropic prompt caching
  (https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching).
- **Demo.** `harness cost report --breakdown` mostra "cache-aware
  layout $X saved".

### 1.14 — Clean Architecture

- **O que é.** Camadas em cebola: `domain` puro no meio, `app`
  em volta, `adapters` só na borda.
- **Analogia BR.** *Coxinha.* Frango (domain, valioso), massa
  (app), fritura por fora (adapter). A fritura pode trocar por
  ar quente sem mexer no frango.
- **Por que usar.** Trocar SQLite por Postgres não vaza pra
  regra de negócio.
- **Como seria sem.** `sql.DB` no meio do handler HTTP; migrar
  vira reescrever.
- **Vantagens.** Testável, portável.
- **Desvantagens.** Boilerplate de interface no começo.
- **Referência.** Robert C. Martin, "Clean Architecture" 2017
  (https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html);
  `HARNESSX-MASTER-PLAN.md §5`.
- **Demo.** `tree internal/{domain,app,adapters} -L 1`

### 1.15 — GitFlow

- **O que é.** Modelo de branching: `main` só release,
  `develop` integração, `feature/*` e `fix/*` a partir de
  `develop`.
- **Analogia BR.** *Fila de banco.* Guichê (`main`) só
  atende quem passou pela senha (`develop`) — não tem fura-fila.
- **Por que usar.** Release estável, feature isoladas.
- **Como seria sem.** Todo mundo commita no `main`, quebra
  produção.
- **Vantagens.** Previsível, hook-friendly.
- **Desvantagens.** Overhead pra time de 1 pessoa.
- **Referência.** Vincent Driessen 2010
  (https://nvie.com/posts/a-successful-git-branching-model/);
  `CONTRIBUTING.md`.
- **Demo.** `git checkout -b feature/auth develop`

### 1.16 — Embed (`//go:embed`)

- **O que é.** Diretiva do Go 1.16+ que embute arquivos no
  binário em tempo de compile.
- **Analogia BR.** *Kit pronto do supermercado.* Vem tudo na
  caixa — não precisa comprar farinha separado.
- **Por que usar.** Single-binary distribution real; sem `assets/`
  ao lado.
- **Como seria sem.** Precisa distribuir `.tar.gz` com
  `templates/` do lado.
- **Vantagens.** Um `go build` só.
- **Desvantagens.** Binário maior; sem `..` no path.
- **Referência.** https://pkg.go.dev/embed;
  `HARNESSX-MASTER-PLAN.md §3`.
- **Demo.** `//go:embed templates/*` no topo do arquivo Go.

### 1.17 — Router (multi-agent)

- **O que é.** `routes.yaml` mapeando intent (spec, implement,
  test, docs) pra agent + fallback chain + budget.
- **Analogia BR.** *Central de atendimento.* SAC 1 vai pra
  vendas, SAC 2 pra cancelamento — cada um pro atendente certo.
- **Por que usar.** Modelo caro só onde precisa.
- **Como seria sem.** Tudo no Sonnet, fatura 3x.
- **Vantagens.** Auditável, previsível.
- **Desvantagens.** YAML pra manter.
- **Referência.** `internal/orchestrate/router.go`.
- **Demo.** `cat .harness/config/routes.yaml`

### 1.18 — Budget (`budget_usd`)

- **O que é.** Cap duro de USD por sessão; ultrapassou, plan é
  recusado.
- **Analogia BR.** *Limite do vale-refeição.* Passou de R$40 no
  almoço — máquina nega.
- **Por que usar.** Evita fatura surpresa.
- **Como seria sem.** Loop bugado gasta USD 200 madrugada
  adentro.
- **Vantagens.** Sono tranquilo.
- **Desvantagens.** Precisa ajustar por sessão longa.
- **Referência.** `internal/costtrack/`.
- **Demo.** `budget_usd: 5.00` no routes.yaml.

### 1.19 — Sensor: secret scan

- **O que é.** Regex pra detectar chave AWS, `aws_secret_access_key`,
  token Slack/GitHub (`ghp_/ghs_/gho_/ghu_/ghr_`), PEM privada.
  ID canônico do sensor: `secrets_scan`.
- **Analogia BR.** *Cão farejador na alfândega.*
- **Por que usar.** Chave commitada = fatura ou vazamento.
- **Como seria sem.** Descobre no `git log` público.
- **Vantagens.** Barato, roda em CI (`harness ci`).
- **Desvantagens.** Falso positivo em fixture de teste.
- **Referência.** `internal/sensors/scanners.go` (`SecretsScanSensor`);
  vocabulário sincronizado com `internal/memory.sensitiveRe`;
  TruffleHog (https://github.com/trufflesecurity/trufflehog).
- **Demo.** `harness sensor run secrets_scan`

### 1.20 — Sensor: forbidden files

- **O que é.** Bloqueia LLM de deletar teste, `go.mod`,
  `CLAUDE.md`. ID canônico: `forbidden_files`.
- **Analogia BR.** *Placa "NÃO PISE NA GRAMA".*
- **Por que usar.** Evita "vou apagar o teste que tá falhando"
  criativo do LLM.
- **Referência.** `internal/sensors/` (ver `runner_test.go` pra
  lista universal: `forbidden_files`, `forbidden_commands`,
  `secrets_scan`, `changed_files`).
- **Demo.** `harness sensor run forbidden_files`

### 1.21 — Cost tracking

- **O que é.** Store SQLite que grava tokens in/out × agent ×
  intent × USD por run.
- **Analogia BR.** *Extrato do cartão.* Você vê onde queimou
  grana.
- **Por que usar.** Sem número, não tem otimização.
- **Referência.** `internal/costtrack/tracker.go` +
  `internal/costtrack/reader.go`.
- **Demo.** `harness cost report --breakdown --since 7d`
  (flag `--since` aceita `1d|7d|30d|all`).

### 1.22 — Memory vault

- **O que é.** Store local (SQLite) de fatos aprendidos com
  `evidence_run_id` + confidence score.
- **Analogia BR.** *Caderninho do gerente.* "Cliente João não
  gosta de gergelim" — anota, usa depois.
- **Por que usar.** LLM não precisa redescobrir a mesma coisa.
- **Como seria sem.** Toda sessão começa do zero.
- **Vantagens.** Custo cai em iteração longa.
- **Desvantagens.** Precisa política de invalidação.
- **Referência.** `HARNESSX-MASTER-PLAN.md §12`.
- **Demo.** `ls .harness/memory/`

### 1.23 — Project index (8 mapas)

- **O que é.** `profile`, `commands`, `dependencies`,
  `architecture`, `test-map`, `api-map`, `design-system`,
  `performance-budget` — 8 JSONs sob `.harness/project/`.
- **Analogia BR.** *Ficha catalográfica da biblioteca.* Sem
  ficha, ninguém acha nada.
- **Por que usar.** Contexto rápido sem re-walk do tree.
- **Referência.** `docs/FEATURES.md §Phase 2`;
  `harness project inspect <map>` pretty-printa cada um.
- **Demo.** `harness project index && ls .harness/project/`

### 1.24 — `harness feature`

- **O que é.** Comando que gera spec (`.harness/artifacts/
  specs/<id>.md`). Sem adapter ativo cai em **stub spec mode**
  (heurística ripgrep + git status).
- **Analogia BR.** *Combo Big Mac.* Vem a spec pronta.
- **Referência.** `cmd/harness/cmd_workflow.go` (`newFeatureCmd`);
  `docs/FEATURES.md §Phase 6`.
- **Demo.** `harness feature "add login" --yes`

### 1.25 — `harness bugfix`

- **O que é.** Workflow spec-first constrangido a um fix.
  Aceita `--mode ask|safe_execute|strict`. Sem `--reason` —
  ponha o motivo no prompt.
- **Analogia BR.** *Ponto na meia.* Costura o buraco, não faz
  meia nova.
- **Referência.** `cmd/harness/cmd_workflow.go` (`newBugfixCmd`).
- **Demo.** `harness bugfix "task list stale after edit"`

### 1.26 — `harness evolve`

- **O que é.** Sistema de mutação governada com telemetria:
  `diagnose` (clustering de falhas em `events.jsonl`),
  `replay`, `sandbox` (A/B real), `propose`, `promote --hitl`.
  **Não** aceita `--area` — não é scanner de risco por diretório.
- **Analogia BR.** *Revisão do carro com base no odômetro.*
- **Referência.** `cmd/harness/cmd_evolve.go`;
  `docs/FEATURES.md §Self-evolution`; paper §3.5.
- **Demo.** `harness evolve diagnose`

### 1.27 — `harness ship`

- **O que é.** SDLC driver: branch `feature/<slug>` de `--base`,
  escreve spec, roda até `--max-attempts` de `harness do +
  harness ci`, backoff em HTTP 429, plan-scope check se
  `--plan <id>` for passado, **conventional commit** no sucesso.
  Não abre PR nem tageia release automaticamente.
- **Analogia BR.** *Vistoria antes de entregar chave do apê.*
- **Referência.** `cmd/harness/cmd_ship.go`;
  `docs/COMMANDS.md`.
- **Demo.** `harness ship "add rate limit" --plan PLAN-01J...`

### 1.28 — `harness doctor`

- **O que é.** Diagnóstico do host: ferramentas requeridas,
  LSP, helpers de supply-chain, agents registrados, root do
  projeto, estado do DB. Cita o comando que corrige cada red.
- **Analogia BR.** *Check-list do piloto antes de decolar.*
- **Referência.** `docs/COMMANDS.md §Bootstrap`.
- **Demo.** `harness doctor --plain`

### 1.29 — Orchestrate multi-agente

- **O que é.** DSL YAML sob `.harness/orchestrations/<name>.yaml`
  com steps que declaram `role` (Manager/Planner/Coder/
  Reviewer/Tester), `command` e/ou `adapter`. Blackboard JSON
  fica em `.harness/artifacts/runs/<id>/blackboard.json`.
- **Analogia BR.** *Reunião de projeto.* Manager delega,
  Planner planeja, Coder implementa, Reviewer revisa, Tester
  testa — cada mensagem vai pro blackboard compartilhado.
- **Referência.** `internal/orchestrate/orchestrate.go`;
  `docs/orchestration.md`; paper §4.1.
- **Demo.** `harness orchestrate list && harness orchestrate run <name>`

### 1.30 — Fallback chain

- **O que é.** Lista ordenada de agents tentados em caso de
  falha de adapter (timeout, 5xx).
- **Analogia BR.** *Plano B, C, D.* Padeiro faltou? Chama o
  auxiliar. Auxiliar faltou? Compra pronto.
- **Referência.** `routes.yaml`.
- **Demo.** `fallback: [codex, kimi]`

### 1.31 — Reviewer role (quality gate)

- **O que é.** Um dos roles do orchestrate (Manager/Planner/
  Coder/**Reviewer**/Tester). O step com `role: Reviewer` julga
  a saída do Coder antes do próximo step ler o blackboard.
- **Analogia BR.** *Provador oficial da churrascaria.*
- **Referência.** `internal/orchestrate/orchestrate.go`;
  `docs/orchestration.md`.
- **Demo.** No YAML do flow: `- role: Reviewer`.

### 1.32 — Perf snapshot

- **O que é.** Baseline de RSS/CPU/latência marcado com label.
- **Analogia BR.** *Foto do velocímetro antes da viagem.*
- **Referência.** `harness perf-snapshot`.
- **Demo.** `harness perf-snapshot --label baseline`

### 1.33 — Dashboard

- **O que é.** SPA React local (`web/dashboard/dist/`) + API JSON.
  Endpoints: `/api/{health, sessions, sessions/{id},
  runs/{id}, sensors, agents, memory, cost, logs, profile,
  design, roadmap, features, toggles}`. Unknown `/api/*`
  devolve envelope 404 JSON.
- **Analogia BR.** *Painel do Uber-motorista* — quanto rodou,
  quanto rendeu.
- **Referência.** `cmd/harness/cmd_dashboard.go`;
  `docs/FEATURES.md §Phase 8`. Passar `--addr :0` pega porta
  efêmera (o comando **imprime** a porta resolvida).
- **Demo.** `harness dashboard --addr :0`

### 1.34 — modernc.org/sqlite (no-CGO)

- **O que é.** SQLite pure-Go, dispensa toolchain C.
- **Analogia BR.** *Ovo já quebrado no potinho.* Sem sujeira de
  compilar C.
- **Por que usar.** Cross-compile simples, binário único.
- **Referência.** https://pkg.go.dev/modernc.org/sqlite;
  `HARNESSX-MASTER-PLAN.md §3`.
- **Demo.** `import _ "modernc.org/sqlite"`

### 1.35 — Conventional Commits

- **O que é.** Convenção `type(scope): subject` com 50 chars.
- **Analogia BR.** *Etiqueta de mercadoria.* Todo mundo lê
  igual.
- **Referência.** https://www.conventionalcommits.org/;
  hook `commit-msg` no repo.
- **Demo.** `feat(auth): add JWT HS256`

### 1.36 — i18n (`i18n.T(key)`)

- **O que é.** Toda string user-facing vai via key + bundle
  JSON.
- **Analogia BR.** *Cardápio traduzido.* Não colar tradução no
  código-fonte.
- **Referência.** `internal/platform/i18n/`; `CLAUDE.md` §Always.
- **Demo.** `i18n.T("cmd.doctor.ok")`

### 1.37 — Constants centralizados

- **O que é.** Valor compartilhado em 2 pacotes → move pra
  `internal/platform/constants`.
- **Analogia BR.** *Preço do pão na lousa* — um lugar só.
- **Referência.** `CLAUDE.md` §Always.
- **Demo.** `grep -r "512m" internal/` → tem que dar 1 lugar.

### 1.38 — Local CI/CD (sem GitHub Actions)

- **O que é.** `make install-hooks` planta pre-push que roda
  `make ci`.
- **Analogia BR.** *Sindico da portaria.* Não deixa passar
  antes de conferir.
- **Por que usar.** Feedback rápido, sem esperar minions
  remotos.
- **Referência.** `CLAUDE.md` §Always.
- **Demo.** `make install-hooks && make ci`

### 1.39 — Playwright (E2E)

- **O que é.** Framework de teste end-to-end (browser real).
- **Analogia BR.** *Cliente-fantasma da rede.* Faz o percurso
  completo, do login à compra.
- **Referência.** https://playwright.dev/.
- **Demo.** `npx playwright test --reporter=list`

### 1.40 — `harness use fake` (offline mode)

- **O que é.** Adapter que retorna resposta canned — pra rodar
  tutorial sem gastar token.
- **Analogia BR.** *Simulador de direção* antes da autoescola
  real.
- **Referência.** `internal/adapters/fake/`.
- **Demo.** `harness use fake`

---

## 2. Progressão didática (15 vídeos)

Regra de ouro: nunca explicar X antes de Y. Ordem sugerida
**pra iniciante** (linear). Ordem **pra experiente** (pula
setup, vai direto no ganho).

### 2.1 Mapa de dependências pedagógicas

```
Ep 1 (Setup)      → introduz: harness, agent, sensor, doctor, index
Ep 2 (Spec init)  → introduz: spec, plan, run, check, budget
Ep 3 (Auth JWT)   → introduz: salience, cache-aware, agent choice
Ep 4 (Projects)   → introduz: router, fallback, cost report
Ep 5 (Tasks)      → introduz: BM25 rerank, memory vault
Ep 6 (Frontend)   → introduz: scaffold, template, no-LLM path
Ep 7 (Auth SPA)   → introduz: LLMLingua compression
Ep 8 (Views)      → aprofunda: multi-file diff, salience
Ep 9 (Tests)      → introduz: agent-per-intent, Kimi barato
Ep 10 (Docker)    → introduz: mem_limit, healthcheck, CLAUDE.md
Ep 11 (Dashboard) → introduz: perf-snapshot, observability
Ep 12 (Cost)      → aprofunda: breakdown, benchmark, when-not-to
Ep 13 (Orchestrate)→ introduz: flow.yaml, proposer/critic/decider
Ep 14 (Bugfix)    → introduz: bugfix vs feature, evolve
Ep 15 (Ship)      → introduz: GitFlow, ci gate, ship
```

### 2.2 Regras de precedência

- **BM25 (Ep 5) só depois de salience (Ep 3)** — precisa
  entender "por que reordenar" antes de "como escolher o que
  entra".
- **LLMLingua (Ep 7) só depois de cache-aware (Ep 3)** —
  compressão sem cache é otimização parcial.
- **Cost report (Ep 4) antes de breakdown (Ep 12)** — número
  bruto antes da categoria.
- **Orchestrate (Ep 13) depois de router (Ep 4)** — flow é
  router com estado.
- **Ship (Ep 15) depois de GitFlow mencionado no Ep 1** —
  branch já explicada.

### 2.3 Trilha do EXPERIENTE (skip mode)

Assiste: **Ep 1 (10 min), Ep 3, Ep 5, Ep 12, Ep 13**. Total
~50 min, pega os 5 conceitos que dão o ganho real: salience,
BM25, LLMLingua, cost breakdown, orchestrate.

### 2.4 Trilha do INICIANTE (linear)

Ep 1 → 15 na ordem, com pausas. Cada vídeo tem seção "cola no
próximo" nos últimos 30s (gancho pedagógico).

---

## 3. Analogias mundo real por episódio (hook de abertura)

Cada hook cabe em 15 segundos, abre o vídeo antes do vinheta.

- **Ep 1 — Setup.** "Imagina abrir uma padaria sem contratar
  gerente, sem forno testado, sem ficha de estoque. É isso que
  é começar projeto de dev com LLM sem harness. Hoje a gente
  contrata o gerente."
- **Ep 2 — Spec-first.** "Você já pediu pastel de queijo e
  veio pastel de carne? A comanda faltou. Hoje a gente escreve
  a comanda antes do LLM ir pra chapa."
- **Ep 3 — Auth JWT.** "JWT sem cuidado é fechadura de
  papelão. Vamos ver por que aqui a gente chama o chaveiro
  bom (Claude), não o mais barato."
- **Ep 4 — Projects CRUD.** "CRUD é o feijão-com-arroz. Não
  precisa contratar chef michelin — o cozinheiro do bandejão
  (Codex) resolve. Hoje: como o roteador decide."
- **Ep 5 — Tasks + permissões.** "Bibliotecária boa não deixa
  você escrever livro que já existe na estante. Hoje: BM25,
  a bibliotecária do harness."
- **Ep 6 — Frontend scaffold.** "Pra que chamar arquiteto pra
  desenhar tijolo padrão? Template resolve. Hoje: quando NÃO
  usar LLM."
- **Ep 7 — Auth SPA.** "React é verboso feito advogado. Hoje
  a gente comprime o processo pra caber no envelope —
  LLMLingua."
- **Ep 8 — Views.** "Tela de projeto e tarefa. Repare como o
  contexto se acumula sem inchar a fatura."
- **Ep 9 — Testes.** "Três provadores de sopa: Codex,
  Kimi e Claude. Cada um prova um caldo diferente pelo preço
  certo."
- **Ep 10 — Docker.** "Container sem `mem_limit` é gaveta sem
  fundo — cai tudo. Hoje a regra do CLAUDE.md que impede
  Mac travar."
- **Ep 11 — Dashboard.** "Uber-motorista olha o painel.
  Você também vai olhar — quanto rodou, quanto gastou."
- **Ep 12 — Cost analysis.** "Extrato do cartão do harness.
  Aqui a gente prova o número: cache, salience, BM25,
  LLMLingua — cada um segurando um pedaço da conta."
- **Ep 13 — Orchestrate.** "Reunião de projeto em YAML:
  Manager, Planner, Coder, Reviewer e Tester dividem tarefa
  num blackboard compartilhado. Multi-agent sem virar salada."
- **Ep 14 — Bugfix.** "Ponto na meia, não meia nova. Diferença
  entre `bugfix` e `feature`."
- **Ep 15 — Ship.** "Vistoria da chave do apê. `ci` verde,
  `ship` orquestra branch → spec → do → ci-loop →
  conventional commit. PR e tag fica pra você (ou pro CI
  remoto)."

---

## 4. Erros comuns por episódio (prevenir na fala)

- **Ep 1.** Rodar `harness` sem `harness init` → falta vault.
  *Fala:* "Antes de qualquer comando, `harness init`. Não
  pula essa etapa."
- **Ep 2.** Aceitar spec sem ler. *Fala:* "Pausa aqui, lê a
  spec. 2 minutos aqui economizam 1 hora depois."
- **Ep 3.** Rodar auth no adapter `fake` e achar que é
  produção. *Fala:* "Se você usa `fake`, o JWT NÃO é real —
  é dublê."
- **Ep 4.** Esquecer `budget_usd` no routes.yaml. *Fala:*
  "Sem budget, sua fatura pode explodir num loop bugado."
- **Ep 5.** Ignorar log de BM25 e reimplementar helper na
  mão. *Fala:* "Se o log diz 'promoted X', o LLM viu X — não
  reimplementa."
- **Ep 6.** Usar `feature` pra criar boilerplate React que já
  tem template. *Fala:* "Scaffold quando a resposta é óbvia;
  feature quando é ambígua."
- **Ep 7.** Threshold de LLMLingua agressivo (< 0.85) → perde
  contexto. *Fala:* "Mantém em 0.92 até você medir o
  impacto."
- **Ep 8.** Editar arquivo no meio do run (staged). *Fala:*
  "Espera o run acabar; depois edita."
- **Ep 9.** Rodar E2E com `next dev` ligado paralelamente.
  *Fala:* "CLAUDE.md §2: um processo por vez. Docker OR
  dev — nunca os dois."
- **Ep 10.** `docker-compose.yml` sem `mem_limit`. *Fala:*
  "Se o gerador esqueceu, VOCÊ adiciona antes de subir."
- **Ep 11.** Dashboard com `--addr :8090` fixo → collision.
  *Fala:* "Sempre `:0`. Kernel escolhe. Ponto."
- **Ep 12.** Achar que ganho é uniforme em todo projeto.
  *Fala:* "Em greenfield pequeno, harness não paga. Mede,
  não assume."
- **Ep 13.** Flow sem `budget_usd` → gasta tudo no debate.
  *Fala:* "Budget no flow, sempre."
- **Ep 14.** Usar `bugfix` pra feature nova. *Fala:* "Bugfix
  é fix DE COISA QUE EXISTIA. Feature é novo."
- **Ep 15.** `git push --force` no `main`. *Fala:* "Nunca.
  Ponto final."

---

## 5. Checklist "aluno entendeu?" (3 perguntas por vídeo)

Cola no fim do vídeo, ou pin no comentário. Se responde as 3,
tá redondo.

### Ep 1 — Setup
1. Pra que serve o diretório `.harness/`?
2. Qual a diferença entre `harness init` e `harness project
   index`?
3. O que `harness doctor` checa?

### Ep 2 — Spec
1. Por que spec antes de code?
2. Qual arquivo o `harness plan` gera?
3. O que o gate `harness check` roda?

### Ep 3 — Auth JWT
1. Por que Claude (Sonnet) e não Codex pra auth?
2. O que é salience reorder em uma frase?
3. Qual é o ganho médio de token que ele traz?

### Ep 4 — Projects CRUD
1. O que faz o `budget_usd` no routes.yaml?
2. Quando o fallback chain é acionado?
3. O que aparece no `harness cost report --since 1d --breakdown`?

### Ep 5 — Tasks
1. BM25 é IA? (Não — é keyword ranking clássico.)
2. Qual sinal no log indica que BM25 promoveu um arquivo?
3. Por que BM25 economiza mais que salience em brownfield?

### Ep 6 — Frontend scaffold
1. Quando usar scaffold vs feature?
2. Qual o custo LLM de `harness new`?
3. Onde ficam os templates?

### Ep 7 — Auth SPA
1. LLMLingua remove o quê?
2. Qual threshold default de qualidade?
3. Em que tipo de código o ganho é maior?

### Ep 8 — Views
1. Como o pack cresce entre chapters?
2. Por que o rodapé fica igual entre calls?
3. O que "optimistic update" tem a ver com custo?

### Ep 9 — Testes
1. Por que Kimi pra teste unitário de front?
2. Por que Claude pra Playwright?
3. Qual sinal indica que o teste cobre cross-user 404?

### Ep 10 — Docker
1. Por que `mem_limit` é regra dura?
2. Qual healthcheck endpoint o tutorial usa?
3. Sequência correta: build → up → E2E → ?

### Ep 11 — Dashboard
1. Por que `--addr :0` (porta efêmera)?
2. Cite 3 endpoints do dashboard (ex: `/api/cost`, `/api/runs/{id}`, `/api/sensors`).
3. Pra que serve o `perf-snapshot`?

### Ep 12 — Cost
1. Qual das 4 técnicas normalmente salva mais? (LLMLingua ou
   cache, depende do repo)
2. Quando harness NÃO paga?
3. O que `harness benchmark cost --agent <id> --iterations N` compara?

### Ep 13 — Orchestrate
1. Cite 3 dos 5 roles suportados (Manager, Planner, Coder,
   Reviewer, Tester).
2. Diferença entre `role: Reviewer` e um sensor?
3. Onde fica o blackboard? (`.harness/artifacts/runs/<id>/blackboard.json`)

### Ep 14 — Bugfix
1. Bugfix ainda escreve spec? (Sim — constrangida a um fix.)
2. `harness evolve diagnose` faz o quê? (Clustering de falhas em
   `events.jsonl`.)
3. Como se promove uma mutação sugerida? (`evolve promote --hitl`.)

### Ep 15 — Ship
1. O que `harness ci` executa por baixo? (todos sensores
   aplicáveis, exit non-zero em red)
2. `harness ship` abre PR e tageia release? (Não — cria branch,
   spec, roda do+ci, faz conventional commit.)
3. Como o `pre-push` hook é instalado?
   (`harness install-git-hooks`)

---

## 6. Notas finais pra Agent 3 (SEO/retenção)

- Cada episódio tem 1 hook (§3), 1 analogia mãe, 3 perguntas
  (§5) — dá subtítulo, dá thumbnail, dá pinned comment.
- Termos gringos que **não devem ser traduzidos** (buscador
  cai): `harness`, `agent`, `sensor`, `spec`, `plan`, `run`,
  `BM25`, `LLMLingua`, `JWT`, `E2E`, `CRUD`, `CI/CD`,
  `GitFlow`, `Playwright`, `Docker`, `React`, `Go`.
- Termos que **devem ser traduzidos** na fala pra abrir
  público: `context engineering` → "engenharia de contexto",
  `cache-aware layout` → "layout com cache em mente",
  `salience` → "relevância", `fallback` → "plano B".

Fim.
