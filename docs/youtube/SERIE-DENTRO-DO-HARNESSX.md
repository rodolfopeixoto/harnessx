# Série "Dentro do harnessx" — 18 roteiros de code walkthrough

Cada episódio: 10 min. Formato fixo:
HOOK 15s / CONTEXTO 60s / CODE TOUR 6min / DEMO 90s / RECAP 30s / CTA 30s.
Voz Rodolfo, PT-BR técnico, direto, sem filler.
Repo real em `github.com/ropeixoto/harnessx`. Referências de arquivo
sempre com path absoluto de `internal/*.go` verificado.

---

## EP01 — Arquitetura em 10min: por que Clean Arch importa

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/HARNESSX-MASTER-PLAN.md`
- `/Users/ropeixoto/dev/projects/harnessx/ARCHITECTURE.md`
- `/Users/ropeixoto/dev/projects/harnessx/internal/domain/`
- `/Users/ropeixoto/dev/projects/harnessx/internal/app/`
- `/Users/ropeixoto/dev/projects/harnessx/internal/adapters/`
- `/Users/ropeixoto/dev/projects/harnessx/go.mod` (Go 1.25, sem CGO)

**Snippet em tela** (regra que ninguém quebra):
```go
// internal/domain — imports nothing além de stdlib.
// internal/app — importa domain + interfaces.
// internal/adapters — implementa interfaces.
// cmd/harness — só cola tudo.
```

**HOOK (15s):** "10 fases, 80k linhas de Go, zero CGO. Como um binário
único orquestra Claude, Codex, Gemini e Kimi sem virar espaguete?
Regra de ouro: `domain` não importa nada. Vou provar."

**CONTEXTO (60s):** Local-first. Sem daemon. Sem servidor. O CLI é o
harness. Clean Architecture aqui não é dogma — é o que permite trocar
adapter de agente sem tocar em `domain/`. Testes rodam em memória
porque `domain` não sabe o que é SQLite.

**CODE TOUR (6min):**
1. Abre `HARNESSX-MASTER-PLAN.md §9` — phase boundaries.
2. `ls internal/` — mostra 70+ pacotes, todos ancorados em 3 camadas.
3. `go list -deps ./internal/domain/...` — nenhum import de app/adapters.
4. Mostra `internal/adapters/sqlite` sendo consumido por `internal/memory`.
5. Referência: **Clean Architecture, Robert C. Martin (2017)**.

**DEMO (90s):** `make ci` — vet + race tests + build. Mostra binário
`harness` de 27MB single-file. `./harness --version` roda em Mac,
Linux, Windows sem dependência externa.

**RECAP (30s):** Regra do domínio puro + no-CGO = binário portável +
testes rápidos. Isso paga em Phase 10.

**CTA (30s):** Se curtir, deixa like. No próximo vídeo, monto o CLI
inteiro com Cobra em 10 minutos.

**Perguntas plateia:**
1. Você usa Clean Arch em Go ou acha overkill?
2. CGO vale a pena por SQLite nativo?
3. Qual seu limite pra dizer "chega, refatora"?

**Comando prático:**
```bash
go list -deps ./internal/domain/... | grep harnessx
# deve retornar apenas github.com/ropeixoto/harnessx/internal/domain
```

---

## EP02 — Cobra + lipgloss: como montar CLI em 60 arquivos

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/cmd/harness/main.go` (146 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/cmd/harness/cmd_agent.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/ui/theme.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/ui/workflowview.go`

**Snippet em tela** (`main.go` linhas 21-45, real):
```go
func main() {
    if err := newRoot().Execute(); err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
}

func newRoot() *cobra.Command {
    var plain, showVersion bool
    root := &cobra.Command{
        Use:   "harness",
        Short: "HarnessX — adaptive runtime for agentic software engineering",
        SilenceUsage:  true,
        SilenceErrors: true,
        PersistentPreRun: func(cmd *cobra.Command, _ []string) {
            ui.SetPlain(plain || os.Getenv("HARNESS_PLAIN") != "")
        },
    }
    root.PersistentFlags().BoolVar(&plain, "plain", false, "disable ANSI styling")
    root.Flags().BoolVarP(&showVersion, "version", "V", false, "print version and exit")
    root.AddCommand(newVersionCmd(), newInitCmd(), newDoctorCmd(), /* ...80 subcomandos */)
    return root
}
```

**HOOK (15s):** "80 subcomandos, um só `main.go` com 146 linhas. Truque?
Uma função `newRoot()` que apenas cola — cada arquivo `cmd_*.go`
constrói o próprio grupo Cobra."

**CONTEXTO (60s):** `main.go` é o ponto de entrada e nada mais. Toda
lógica de subcomando mora em `cmd_<nome>.go`. Isso permite testar
sem levantar o processo. `SilenceUsage:true` porque erros do runtime
não são erros de sintaxe do CLI.

**CODE TOUR (6min):**
1. `main.go` linha por linha (é curto).
2. `cmd_agent.go` — como `newAgentCmd()` devolve `*cobra.Command`.
3. `internal/ui/theme.go` — lipgloss styles centralizados.
4. `HARNESS_PLAIN=1` — modo grep-friendly.
5. Como testar `RunE` sem chamar processo.

**DEMO (90s):** `harness --help` mostra os 80 comandos agrupados.
`HARNESS_PLAIN=1 harness doctor` vs padrão colorido.

**RECAP (30s):** Arquivo por comando + `newX()` factory = testável e
navegável.

**CTA (30s):** Próximo: sensors. O coração determinístico do harness.

**Perguntas plateia:**
1. Você separa comandos por arquivo ou centraliza?
2. Cobra ou urfave/cli?
3. Vale a pena `PersistentPreRun` pra flags globais?

**Comando prático:**
```bash
harness --help | head -30
HARNESS_PLAIN=1 harness doctor
```

---

## EP03 — Sensors: o firewall determinístico

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/sensors/catalog.go` (~450 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/sensors/runner.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/sensors/types.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/sensors/scanners.go`
- `/Users/ropeixoto/dev/projects/harnessx/docs/sensors.md`

**Snippet em tela** (interface real de `types.go`):
```go
type Sensor interface {
    ID() string
    Category() Category
    Kind() Kind
    AppliesTo(p index.Profile) bool
    Run(rc RunCtx) Result
}

type Result struct {
    ID         string
    Status     Status // StatusPassed | StatusFailed | StatusSkipped
    Category   Category
    Kind       Kind
    Duration   time.Duration
    OutputPath string
    Detail     string
    ExitCode   int
    Confidence float64
    Scope      []string
    Verified   []string
    Unverified []string
    Risks      []string
}
```

**HOOK (15s):** "Antes do agente escrever uma linha, o harness passa
seus arquivos por uma bateria de sensors. Cada um determinístico,
cada um retorna passed / failed / skipped."

**CONTEXTO (60s):** LLM erra, muito. Sensors são a rede: rodam
localmente, sem custo, sem IA. Se falhar, o agente nem começa.
Categorias reais em `types.go`: `CatSpec`, `CatForbidden`, `CatSecrets`,
`CatChangedFile`, `CatFormat`, `CatLint`, `CatTypecheck`, `CatTest`,
`CatSecurity`, `CatPerf`, `CatDeps`, `CatLogs`, `CatImage`, `CatRuntime`,
`CatDocs`, `CatAPI`, `CatDesign`, `CatOther`. Kinds: `KindComputational`
vs `KindInferential`.

**CODE TOUR (6min):**
1. `types.go` — interface `Sensor` (5 métodos).
2. `catalog.go` — registro central. `AppliesTo` filtra por profile.
3. `runner.go` — executor paralelo com `RunCtx{Root, OutputDir}`.
4. `smell.go` + `shell.go` — sensors de code smell e shell scripts.
5. Referência: **"Constitutional AI: Harmlessness from AI Feedback",
   Bai et al., 2022 (arXiv:2212.08073)** — mesma ideia: rules
   determinísticas antes/depois do modelo.

**DEMO (90s):** `harness sensor list` mostra todos. `harness check`
roda a bateria. Mostra falha proposital (arquivo `.env` commitado).

**RECAP (30s):** Sensor = função pura sobre repo. Sem sensor, sem
merge.

**CTA (30s):** Amanhã: como o sensor de secrets acha AWS key com regex.

**Perguntas plateia:**
1. Você confiaria só em sensors ou combinaria com IA?
2. Falso positivo custa mais que falso negativo?
3. Que sensor você escreveria pro seu stack?

**Comando prático:**
```bash
harness sensor list
harness check
```

---

## EP04 — Secrets Sensor: regex vs .env commitado

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/sensors/scanners.go` linhas 22-58
- `/Users/ropeixoto/dev/projects/harnessx/internal/secrets/`
- `/Users/ropeixoto/dev/projects/harnessx/docs/security.md`

**Snippet em tela** (`scanners.go` linhas 22-58, real):
```go
var forbiddenPatterns = []string{
    ".env", ".envrc", "*.pem", "*.key",
    "id_rsa", "id_dsa", "id_ed25519",
    "secrets.yml", "secrets.yaml",
    "credentials.json", ".aws/credentials", ".npmrc.local",
}

var secretPatterns = []*regexp.Regexp{
    regexp.MustCompile(`AKIA[0-9A-Z]{16}`),                       // AWS
    regexp.MustCompile(`(?i)aws_secret_access_key\s*=\s*[A-Za-z0-9/+=]{40}`),
    regexp.MustCompile(`xox[abprs]-[A-Za-z0-9-]{20,}`),           // Slack
    regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
    regexp.MustCompile(`gh[opusr]_[A-Za-z0-9]{36,}`),             // GitHub
    regexp.MustCompile(`(?i)api[_-]?key\s*[:=]\s*['"][A-Za-z0-9_\-]{24,}['"]`),
}
```

**HOOK (15s):** "Um `AKIA` no commit derruba a startup. Mostro a regex
exata que o harness usa pra bloquear antes do push."

**CONTEXTO (60s):** `forbidden_files` mata pelo nome. `secretPatterns`
mata pelo conteúdo. Go RE2 não tem lookaround — então `git push
--force-with-lease` cai junto com `--force`. Falso positivo é
aceitável aqui: melhor um susto que uma chave vazada.

**CODE TOUR (6min):**
1. Linha 22-35: `forbiddenPatterns` (path-based glob).
2. Linha 50-58: `secretPatterns` (regex conservadora).
3. Linha 44-46: `forbiddenCommandRe` — `rm -rf /`, `curl | bash`.
4. `internal/secrets/keychain.go` — solução: usa Keychain no macOS.
5. Referência: **"Secrets in Source Code", Meli et al., NDSS 2019**
   — 100k+ chaves ativas em repos públicos.

**DEMO (90s):** Cria `.env` com `AKIA1234567890ABCDEF`. Roda
`harness check`. Sensor falha. Mostra `.harness/artifacts/` com o hit.

**RECAP (30s):** Regex + forbidden globs + Keychain = zero secret
commitado.

**CTA (30s):** Próximo: sensor de budget. Tokens custam dinheiro.

**Perguntas plateia:**
1. Você tem hook pre-commit pra secrets hoje?
2. Já vazou chave? Como recuperou?
3. Vale usar `gitleaks` junto?

**Comando prático:**
```bash
# Sensor ID real é `secrets_scan` (ver internal/sensors/scanners.go:222).
# `--root` aponta pra raiz do projeto; sem flag, usa cwd.
mkdir -p /tmp/scan-demo && echo "AKIAIOSFODNN7EXAMPLE" > /tmp/scan-demo/.env
harness sensor run secrets_scan --root /tmp/scan-demo
```

---

## EP05 — Budget Sensor: token counting sem surpresa

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/sensors/budget.go` (~254 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/router/cost.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/platform/tokens/` (estimator)
- `/Users/ropeixoto/dev/projects/harnessx/.harness/project/performance-budget.json`

**Snippet em tela** (`budget.go` linhas 30-73, real):
```go
func (s PerformanceBudgetSensor) Run(rc RunCtx) Result {
    var budget index.PerformanceBudget
    if err := index.ReadMap(rc.Root, index.MapPerformance, &budget); err != nil {
        return Result{Status: StatusSkipped, Detail: "no performance-budget.json"}
    }
    snapshot, snapPath := newestSnapshot(rc.Root)
    if snapshot == nil {
        return Result{Status: StatusSkipped, Detail: "no perf snapshot"}
    }
    var breaches []string
    for _, k := range sortedBudgetKeys(budget.Budgets) {
        limit, _ := numericBudget(budget.Budgets[k])
        actual, ok := snapshotValue(snapshot, k)
        if !ok { continue }
        if actual > limit {
            breaches = append(breaches, fmt.Sprintf("%s: %s > %s", k, fmtNum(actual), fmtNum(limit)))
        }
    }
    if len(breaches) == 0 {
        return Result{Status: StatusPassed, Detail: "within budget"}
    }
    return Result{Status: StatusFailed, Detail: strings.Join(breaches, "; ")}
}
```

**HOOK (15s):** "Um agente rodando fora de controle queima R$500 numa
madrugada. Mostro como o harness compara snapshot vs budget e trava."

**CONTEXTO (60s):** `perf-snapshot` grava métricas em
`.harness/artifacts/perf/`. `performance-budget.json` define limites
(RSS, tokens, deps). Sensor compara. Estoura, falha.

**CODE TOUR (6min):**
1. `snapshotResolvers` — mapa de chave → função (linha 117).
2. `container_memory_mb`, `process_rss_mb`, `noisy_log_call_sites`.
3. `internal/router/cost.go` — cost tracking por adapter.
4. `internal/platform/tokens/Heuristic4` — estimador (~4 chars/token).
5. Referência: **"Compute Optimal Language Model Training",
   Hoffmann et al., Chinchilla, arXiv:2203.15556** — porque medir
   token in/out muda decisão de modelo.

**DEMO (90s):** `harness perf-snapshot`. Edita budget pra 1MB. Rerroda
`harness check`. Falha com breach específico.

**RECAP (30s):** Snapshot + budget + comparador = orçamento
determinístico.

**CTA (30s):** Próximo: como o Context Builder monta o pack.

**Perguntas plateia:**
1. Você tem budget de token no CI?
2. Estimador heurístico serve ou precisa tokenizer real?
3. Que outra métrica você adicionaria?

**Comando prático:**
```bash
harness perf-snapshot
cat .harness/project/performance-budget.json
harness sensor run performance_budget
```

---

## EP06 — Context Engineering: git + rg + LSP + rerank

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/context/builder.go` (255 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/context/provider_git.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/context/provider_ripgrep.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/context/provider_lsp.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/context/provider_rerank.go`
- `/Users/ropeixoto/dev/projects/harnessx/docs/context-engineering.md`

**Snippet em tela** (`builder.go` linhas 35-49, real):
```go
func DefaultProviders() []Provider {
    return []Provider{
        MemoryProvider{},
        GitProvider{},
        RipgrepProvider{},
        TestMapProvider{},
        // Rerank runs last so it sees every entry contributed above.
        // Reorder still runs after the chain and owns Salience.
        RerankProvider{},
    }
}
```

E do `provider_ripgrep.go`:
```go
func (r RipgrepProvider) Apply(ctx context.Context, root string, pack *Pack) error {
    if !hasBinary("rg") {
        pack.Stats.ProvidersSkipped++
        return nil
    }
    // ... aplica extractKeywords + rgSearch por keyword, cada hit
    // vira FileEntry em pack.RelevantFiles com Reason "ripgrep:<kw>".
    return nil
}
```

**HOOK (15s):** "Mandar o repo inteiro pro LLM é queimar dinheiro.
Vou mostrar como o harness escolhe 20 arquivos entre 10 mil, com
cache SHA determinístico."

**CONTEXTO (60s):** Context Engineering (Phase 3) = pipeline de
providers. Cada um adiciona candidatos. Rerank ordena por salience
(BM25 + Jaccard dedup). Cache por SHA de (task + profile + git HEAD +
providers).

**CODE TOUR (6min):**
1. `builder.go:Build()` — orquestrador com cache SHA256.
2. `provider_git.go` — diff + status.
3. `provider_ripgrep.go` — keywords + `rg --json -m1`.
4. `provider_lsp.go` — `AutoLSP` detecta `go.mod` → wire gopls.
5. `provider_rerank.go` — BM25 + salience score.
6. Referência: **"Retrieval-Augmented Generation for
   Knowledge-Intensive NLP", Lewis et al., NeurIPS 2020
   (arXiv:2005.11401)** — a mesma ideia, mas com git como retrieval.

**DEMO (90s):** `harness context build --task "fix router fallback"`.
Mostra pack JSON com `RelevantFiles`, `EstimatedTokens`, `CacheHit`.
Roda de novo — cache hit em ms.

**RECAP (30s):** Provider chain + cache SHA = pack determinístico e
barato.

**CTA (30s):** Próximo: adapter do Claude Code.

**Perguntas plateia:**
1. Você usa RAG no seu tooling ou joga tudo no modelo?
2. Rerank BM25 ou embedding?
3. Cache por commit ou por diff?

**Comando prático:**
```bash
harness context build --task "add sensor for python coverage"
ls .harness/cache/context/
```

---

## EP07 — Adapter HTTP: falando com Claude/OpenAI/Gemini genérico

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/agents/http/adapter.go` (~278 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/agents/yaml/` (spec files)
- `/Users/ropeixoto/dev/projects/harnessx/internal/agents/types.go`

**Snippet em tela** (`adapter.go` linhas 50-99, real):
```go
func (a *Adapter) Run(ctx context.Context, req agents.AgentRequest) agents.AgentResult {
    start := time.Now()
    rctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    body, err := a.buildBody(req)
    httpReq, err := http.NewRequestWithContext(rctx, methodOrPost(a.Spec.API.Method),
        a.Spec.API.Endpoint, bytes.NewReader(body))
    httpReq.Header.Set("Content-Type", "application/json")
    a.applyAuth(httpReq)

    resp, err := a.Client.Do(httpReq)
    respBody, _ := io.ReadAll(resp.Body)

    out := agents.AgentOutput{Stdout: respBody}
    out.FinalMessage = extractJSONPath(respBody, a.Spec.API.Response.FinalMessage)

    res := agents.AgentResult{Output: out, ExitCode: resp.StatusCode, Duration: time.Since(start)}
    if resp.StatusCode >= 400 {
        res.Failure = classifyHTTP(resp.StatusCode)  // 401→Auth, 429→RateLimit, 5xx→Transient
    }
    res.Usage = a.ParseUsage(out)
    return res
}
```

**HOOK (15s):** "Anthropic, OpenAI, Gemini, Moonshot — mesmo adapter.
YAML define endpoint, template e JSONPath da resposta. Sem CLI."

**CONTEXTO (60s):** Nem sempre você quer subir CLI. O `http.Adapter`
lê spec YAML: endpoint, request template, response path, secret ref.
Falhas viram tipos (`Auth`, `RateLimit`, `Timeout`, `Transient`,
`Permanent`) para o Router decidir fallback.

**CODE TOUR (6min):**
1. `buildBody` — template `{{prompt}}` + `{{model}}` + extras.
2. `applyAuth` — resolve secret via Keychain/env.
3. `classifyHTTP` — mapa status → `FailureType`.
4. `ParseUsage` — JSONPath extrai `input_tokens` / `output_tokens`.
5. YAML spec em `internal/agents/yaml/`.

**DEMO (90s):** `harness agent list`. Mostra `anthropic.yaml`. Roda
`harness ask "hello" --agent anthropic`. Mostra `usage.input_tokens`.

**RECAP (30s):** Um adapter, N providers. Config declarativa.

**CTA (30s):** Próximo: o adapter fake que faz E2E rodar em 200ms.

**Perguntas plateia:**
1. Você prefere CLI adapter ou HTTP direto?
2. JSONPath ou schema tipado?
3. Onde guarda os secrets?

**Comando prático:**
```bash
harness agent list
cat internal/agents/yaml/anthropic.yaml  # ou similar
```

---

## EP08 — Fake Adapter: testes E2E sem gastar 1 centavo

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/agents/fake/fake.go` (100 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/agents/fake/extra_test.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/router/router_test.go`

**Snippet em tela** (`fake.go` linhas 15-51, real):
```go
type Adapter struct {
    IDValue      string
    NameValue    string
    CapsValue    agents.Capabilities
    HealthOK     bool
    HealthDetail string
    CLIVersion   string
    FinalMessage string
    StdoutBytes  []byte
    StderrBytes  []byte
    ExitCode     int
    RunDelay     time.Duration
    RunErr       error
    UsageValue   agents.Usage
    ForceFailure agents.FailureType
}

func New(id string) *Adapter {
    return &Adapter{
        IDValue:      id,
        NameValue:    "Fake " + id,
        HealthOK:     true,
        CLIVersion:   "fake-1.0.0",
        FinalMessage: "ok",
        CapsValue: agents.Capabilities{
            Text: true, Files: true, Folders: true, Diff: true,
            JSONOutput: true, StreamOutput: true,
            MaxContextTokens: 128000,
            Strengths:        []string{"implementation", "tests"},
            Models:           map[string]string{"default": "fake-default"},
        },
        UsageValue: agents.Usage{InputTokens: 100, OutputTokens: 50, Mode: "estimated"},
    }
}
```

**HOOK (15s):** "300+ testes E2E rodando em 4s, custo zero. Truque?
Um adapter fake que satisfaz a mesma interface do Claude."

**CONTEXTO (60s):** Interface segregation na prática. `AgentAdapter`
tem `ID`, `Run`, `Healthcheck`, `Capabilities`, `ParseUsage`,
`ClassifyFailure`. Fake implementa todos com campos exportados. Testes
setam `ForceFailure = FailureRateLimit` pra exercitar router.

**CODE TOUR (6min):**
1. Struct com campos exportados (test injection).
2. `Run()` respeita `RunDelay` + `ctx.Done()`.
3. `ForceFailure` — testa fallback do router.
4. Usado em `router_test.go` linha por linha.
5. Referência: **"Growing Object-Oriented Software, Guided by Tests",
   Freeman & Pryce (2009)** — test double clássico.

**DEMO (90s):** `go test -run TestRouter_Execute_FallbackChain
./internal/router/...` — 3ms, 5 cenários.

**RECAP (30s):** Interface + fake = confidence sem billing.

**CTA (30s):** Próximo: Router. Onde os fallbacks acontecem.

**Perguntas plateia:**
1. Você testa integração com mock ou VCR?
2. VCR do harness (`internal/agents/vcr`) — vale a pena?
3. Fake compromete cobertura real?

**Comando prático:**
```bash
go test -v -count=1 ./internal/agents/fake/...
go test -v -count=1 ./internal/router/...
```

---

## EP09 — Router: primary/fallback + explainable decision

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/router/router.go` (161 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/router/cost.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/router/metrics.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/router/strengths.go`
- `/Users/ropeixoto/dev/projects/harnessx/docs/orchestration.md`

**Snippet em tela** (`router.go` linhas 107-139, real):
```go
func (r *Router) Execute(ctx context.Context, task string, req agents.AgentRequest,
    clock func() time.Time) (ExecuteResult, error) {

    d, err := r.Select(task)
    out := ExecuteResult{Decision: d}
    for _, a := range d.Chain {
        res := a.Run(ctx, req)
        out.Selected, out.Result = a, res
        if res.Failure == agents.FailureNone && res.ExitCode == 0 && res.Err == nil {
            out.Succeeded = true
            return out, nil
        }
        out.Fallbacks = append(out.Fallbacks, FallbackEvent{
            From: a.ID(), Failure: res.Failure,
            Detail: trim(res.Output.Stderr, 240), At: clock(),
        })
        if !res.Failure.IsRecoverable() {
            return out, nil
        }
        req.Model = ""  // reset pro próximo escolher default
    }
    return out, nil
}
```

**HOOK (15s):** "Claude cai. O que acontece com sua feature? No harness,
o Router já pulou pro Codex antes de você notar."

**CONTEXTO (60s):** `RouteConfig{Primary, Fallback, BudgetUSD, Model}`.
Router monta Chain, tenta em ordem. `FailureType.IsRecoverable()`
decide: seguir ou parar. `Decision.Reasons` é auditável.

**CODE TOUR (6min):**
1. `RouteConfig` struct (linha 19).
2. `Select` — resolve chain, drop de adapters não-registrados.
3. `Execute` — loop com reset de model entre tentativas.
4. `FallbackEvent` — trilha auditável.
5. `strengths.go` — matching por `strengths` no `Capabilities`.
6. Referência: **"Cascade: Fast Cascaded LLM Serving", Chen et al.,
   arXiv:2305.05176** — cascade routing paga menos.

**DEMO (90s):** Config route `implementation` com primary=claude,
fallback=[codex, gemini]. Força 429 no fake. Ver decisão + fallback
event no output.

**RECAP (30s):** Chain + FailureType classificado + reasons = router
explicável.

**CTA (30s):** Próximo: memory com evidence gate.

**Perguntas plateia:**
1. Rate limit resolve com fallback ou backoff?
2. Chain fixa ou multi-armed bandit?
3. Custo x latência: qual manda?

**Comando prático:**
```bash
harness routes
harness explain route implementation
```

---

## EP10 — Memory: SQLite + evidence + confidence floor

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/memory/memory.go` (120 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/memory/kinds_test.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/adapters/sqlite/` (modernc.org/sqlite)
- `/Users/ropeixoto/dev/projects/harnessx/docs/memory.md`

**Snippet em tela** (`memory.go` linhas 60-100, real):
```go
const confidenceFloor = 0.4

var sensitiveRe = regexp.MustCompile(
    `(?i)(AKIA[0-9A-Z]{16}|aws_secret_access_key|xox[abprs]-[A-Za-z0-9-]{20,}|` +
    `-----BEGIN [A-Z ]*PRIVATE KEY-----|gh[opusr]_[A-Za-z0-9]{36,})`)

func Promote(ctx context.Context, repo *sqlite.Repo, c Candidate, db SQLExec) (domain.Memory, error) {
    if c.EvidenceRunID == "" {
        return domain.Memory{}, ErrMissingEvidence
    }
    if c.Confidence < confidenceFloor {
        return domain.Memory{}, ErrLowConfidence
    }
    if sensitiveRe.MatchString(c.Content) {
        return domain.Memory{}, ErrSensitive
    }
    if !validKind(c.Kind) {
        return domain.Memory{}, ErrUnknownKind
    }
    m := domain.Memory{ID: ids.New(), Scope: c.Scope, Kind: c.Kind, Content: c.Content,
        EvidenceRunID: c.EvidenceRunID, Confidence: c.Confidence, /* ... */}
    return m, writeMemory(ctx, db, m)
}
```

**HOOK (15s):** "LLM 'aprende' mentira o tempo todo. Memory só grava
o que tem evidence_run_id + confidence ≥ 0.4 + não bate no regex de
secret."

**CONTEXTO (60s):** SQLite via `modernc.org/sqlite` (puro Go, sem CGO).
Kinds taxonomia do paper: `working`, `semantic`, `experiential`,
`long_term`, `multi_agent`. Sensitive filter compartilhado com o
sensor de secrets.

**CODE TOUR (6min):**
1. Constantes `KindWorking`, etc.
2. `Promote` — 4 gates (evidence, confidence, sensitive, kind).
3. Schema em `internal/adapters/sqlite/migrations`.
4. Referência: **"Generative Agents: Interactive Simulacra of Human
   Behavior", Park et al., arXiv:2304.03442** — memory stream +
   reflection.

**DEMO (90s):** `harness memory promote` recebe candidato via stdin
(scope/kind/content/evidence_run_id/confidence) e roda o gate. `harness
memory list` mostra o que passou. Candidato com `confidence < 0.4` é
recusado com `ErrLowConfidence`; sem `EvidenceRunID`, `ErrMissingEvidence`.

**RECAP (30s):** Evidence + confidence + sensitive filter = memory que
não polui.

**CTA (30s):** Próximo: spec-driven dev.

**Perguntas plateia:**
1. Long-term memory ou só working?
2. Confidence 0.4 é agressivo?
3. Você deletaria memory antigo automaticamente?

**Comando prático:**
```bash
harness memory list
# `harness memory learn` analisa .harness/runs/* (não recebe --content);
# use `harness memory promote` para promover candidato pelo gate.
harness memory promote --help
```

---

## EP11 — Spec-driven: markdown template + gate

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/spec/spec.go` (~230 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/spec/fill.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/specflow/`
- `/Users/ropeixoto/dev/projects/harnessx/.harness/artifacts/specs/` (output)
- `/Users/ropeixoto/dev/projects/harnessx/docs/spec-driven-development.md`

**Snippet em tela** (`spec.go` struct real linhas 22-46):
```go
type Spec struct {
    ID                 string
    Title              string
    Mode               domain.Mode
    GeneratedAt        time.Time
    Prompt             string
    UserProblem        string
    ExpectedOutcome    string
    Scope              []string
    OutOfScope         []string
    BusinessRules      []string
    Security           []string
    Performance        []string
    TestPlan           []string
    AcceptanceCriteria []string
    RollbackPlan       []string
    DefinitionOfDone   []string
    // ... total 20 seções
}
```

**HOOK (15s):** "Sem spec, agente vira alucinação com CI verde. Vou
mostrar a struct de 20 seções que o harness gera antes de qualquer
código."

**CONTEXTO (60s):** Spec = markdown com seções obrigatórias. Sensor
`spec_gate` lê e recusa PR sem spec. Fluxo: `harness feature "..."`
gera arquivo em `.harness/artifacts/specs/<id>.md`, você preenche
gaps, IA fecha o resto.

**CODE TOUR (6min):**
1. Struct `Spec` — cada campo é seção obrigatória.
2. `NewFromPrompt` — seed com defaults honestos (vazios).
3. `internal/specflow/` — multi-round elicitation.
4. Sensor `spec_gate` em `internal/sensors/`.
5. Referência: **"Specification-Driven Development", Beck 2003 /
   TDD** — mesma ideia, escala pra IA.

**DEMO (90s):** `harness feature "add rate limit per user" --yes`.
Abre arquivo gerado. Mostra AcceptanceCriteria, TestPlan,
RollbackPlan seedados.

**RECAP (30s):** Spec markdown + gate = feature auditável.

**CTA (30s):** Próximo: plan contract.

**Perguntas plateia:**
1. 20 seções é overkill?
2. Você usa spec ou vai direto pro código?
3. IA que preenche spec compromete o valor?

**Comando prático:**
```bash
harness feature "add rate limit per user" --yes
cat .harness/artifacts/specs/*.md | head -60
```

---

## EP12 — Plan Contract: escopo declarado, escopo forçado

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/plancontract/plancontract.go` (~215 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/plan/plan.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/sensors/planscope_sensor.go`
- `/Users/ropeixoto/dev/projects/harnessx/docs/plan-contract.md`

**Snippet em tela** (`plancontract.go` linhas 139-165, real):
```go
func (c Contract) InScope(path string) bool {
    if len(c.Files) == 0 {
        return true
    }
    for _, f := range c.Files {
        if f == "_unconstrained_" { return true }
        if ok, _ := filepath.Match(f, path); ok { return true }
        if f == path { return true }
        if strings.HasSuffix(f, "/") && strings.HasPrefix(path, f) { return true }
        if rec, dir := recursiveGlob(f); rec && strings.HasPrefix(path, dir+"/") {
            return true
        }
    }
    return alwaysInScope(path)
}
```

**HOOK (15s):** "Agente adora tocar arquivo que não devia. Plan
Contract é a apólice: você declara os paths, sensor força."

**CONTEXTO (60s):** Plan = markdown em `.harness/artifacts/plans/`.
Seções: Intent, Files in scope, Invariants, Validation, Rollback,
Risk. Sensor `planscope` compara diff vs `InScope`.

**CODE TOUR (6min):**
1. Struct `Contract` — 8 campos (ID, Path, Intent, Files, Invariants, Validation, Rollback, Risk).
2. `Load` — parser markdown por seções.
3. `InScope` — glob + recursive + `alwaysInScope` (metadata files).
4. `alwaysInScope` lista `.gitignore`, `Makefile`, `go.mod`.
5. Sensor em `internal/sensors/planscope_sensor.go`.

**DEMO (90s):** Cria PLAN-42 com `Files in scope: internal/router/**`.
Edita `cmd/harness/main.go` (fora). Roda `harness check`. Falha com
"out-of-scope file".

**RECAP (30s):** Contrato + sensor = agente confinado.

**CTA (30s):** Próximo: skills versionadas.

**Perguntas plateia:**
1. Plan por PR ou por feature grande?
2. Recursive glob `/**` é intuitivo?
3. Você exclui config files do gate?

**Comando prático:**
```bash
ls .harness/artifacts/plans/
# intent = positional; escopo via --file (repeatable). --intent não existe.
harness plan write "refactor router fallback" --file "internal/router/**"
harness plan check --plan <id>
```

---

## EP13 — Skills: versionadas + gate por benchmark

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/skills/skills.go` (~190 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/skillpkg/`
- `/Users/ropeixoto/dev/projects/harnessx/.harness/skills/` (destino)
- `/Users/ropeixoto/dev/projects/harnessx/docs/skills.md`

**Snippet em tela** (`skills.go` linhas 48-89, real):
```go
func Promote(ctx context.Context, db *sql.DB, root, name, content string,
    scoreFn func(string) float64) (Skill, error) {

    score := scoreFn(content)
    best, err := BestScore(ctx, db, name)
    if score <= best {
        return Skill{}, fmt.Errorf("%w (new=%.3f best=%.3f)",
            ErrNoImprovement, score, best)
    }
    version, _ := nextVersion(ctx, db, name)
    hash := contentHash(content)
    path := filepath.Join(dir(root), fmt.Sprintf("%s.v%d.md", name, version))
    os.WriteFile(path, []byte(content), 0o644)
    db.ExecContext(ctx, `insert into skill_versions
        (id, skill_name, version, content_hash, score, accepted, created_at)
        values (?, ?, ?, ?, ?, ?, ?)`,
        ids.New(), name, version, hash, score, 1, now.Format(time.RFC3339Nano))
    return Skill{Name: name, Version: version, Score: score, Accepted: true}, nil
}
```

**HOOK (15s):** "Skill nova só ganha slot se bater o score da anterior.
Sem regressão. Zero prompt-hoarding."

**CONTEXTO (60s):** Skill = markdown com playbook. `skill_versions`
tabela SQLite guarda hash, version, score. `Promote` só aceita se
`score > best`. `scoreFn` é injetado — em produção, benchmark suite.

**CODE TOUR (6min):**
1. Struct `Skill`.
2. `Promote` — gate por score, sem exceção.
3. `BestScore` — max score aceito no histórico.
4. `staticScore` fallback (linhas não-comentário).
5. Referência: **"Voyager: Open-Ended Embodied Agent with LLMs",
   Wang et al., arXiv:2305.16291** — skill library com iterative
   promotion.

**DEMO (90s):** `harness skill list`. `harness skill promote
--name go-tests --file playbook.md`. Segundo `promote` com conteúdo
inferior falha.

**RECAP (30s):** Skill library com gate = evolução monotônica.

**CTA (30s):** Próximo: TUI Bubble Tea.

**Perguntas plateia:**
1. staticScore serve ou precisa benchmark suite real?
2. Skill markdown ou YAML estruturado?
3. Você aceita regressão pra ganhar simplicidade?

**Comando prático:**
```bash
harness skill list
ls .harness/skills/
```

---

## EP14 — TUI Bubble Tea: workflow view em tempo real

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/ui/workflowview.go` (~110 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/ui/theme.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/ui/doctor_view.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/repl/` (chat)

**Snippet em tela** (`workflowview.go` linhas 17-46, real):
```go
type Phase string
const (
    PhaseIntent  Phase = "INTENT"
    PhasePlan    Phase = "PLAN"
    PhaseRoute   Phase = "ROUTE"
    PhaseAgent   Phase = "AGENT"
    PhaseSensors Phase = "SENSORS"
    PhaseBudget  Phase = "BUDGET"
    PhaseReport  Phase = "REPORT"
    PhaseLoop    Phase = "LOOP"
)

type Presenter interface {
    Start(phase Phase, headline string)
    Detail(line string)
    End(phase Phase, status Status, summary string)
    Note(line string)
}
```

**HOOK (15s):** "Cada phase colorida em tempo real, com fallback grep-
friendly quando você joga pra pipe. Lipgloss + Bubble Tea, sem magia."

**CONTEXTO (60s):** `Presenter` interface = duas implementações:
`richPresenter` (lipgloss) e `plainPresenter` (grep). Auto-pick por
`ui.IsPlain()` que respeita `HARNESS_PLAIN` env.

**CODE TOUR (6min):**
1. Phase enum — mapeia ao pipeline real.
2. Interface `Presenter` — 4 métodos.
3. `theme.go` — paleta lipgloss centralizada.
4. `doctor_view.go` — checklist UI.
5. Referência: charmbracelet/bubbletea = Elm architecture.

**DEMO (90s):** `harness run "add readme"` mostra colorido.
`harness run ... | cat` mostra plain automaticamente.

**RECAP (30s):** Interface + auto-detect TTY = UX bonita sem quebrar
scripts.

**CTA (30s):** Próximo: dashboard React.

**Perguntas plateia:**
1. Bubble Tea vale a curva ou fica no fmt.Println?
2. Você respeita NO_COLOR?
3. TUI ou web dashboard?

**Comando prático:**
```bash
harness doctor
harness doctor | cat  # veja diferença
```

---

## EP15 — Dashboard React: métricas via /api

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/web/dashboard/src/App.tsx`
- `/Users/ropeixoto/dev/projects/harnessx/web/dashboard/src/api.ts`
- `/Users/ropeixoto/dev/projects/harnessx/web/dashboard/vite.config.ts`
- `/Users/ropeixoto/dev/projects/harnessx/web/dashboard/embed.go` (embed FS)
- `/Users/ropeixoto/dev/projects/harnessx/cmd/harness/cmd_dashboard.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/app/dashboardcmd/dashboardcmd.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/adapters/http/` (handlers `/api/*`)
- `/Users/ropeixoto/dev/projects/harnessx/docs/dashboard.md`

**Snippet em tela** (`api.ts` — fetch tipado da SPA):
```ts
// GET /api/runs, /api/sensors, /api/agents, /api/cost, /api/memory, ...
// endpoints reais em internal/adapters/http/handlers_*.go
```
E o embed em Go (`web/dashboard/embed.go`, real):
```go
//go:embed all:dist
var distFS embed.FS

func FS() fs.FS {
    sub, _ := fs.Sub(distFS, "dist")
    return sub
}
```

**HOOK (15s):** "Um binário único, mas quando você quer dashboard,
o React já vem embutido via `//go:embed`."

**CONTEXTO (60s):** Vite + React + TS. Build gera `dist/` empacotado
via `//go:embed` no CLI. `harness dashboard` sobe servidor local que
serve estático + `/api/*`.

**CODE TOUR (6min):**
1. `vite.config.ts` — output pra `web/dashboard/dist/`.
2. `App.tsx` — hooks + páginas em `web/dashboard/src/pages/`.
3. `api.ts` — fetch tipado dos endpoints `/api/*`.
4. `cmd_dashboard.go` → `internal/app/dashboardcmd.Run` → `internal/adapters/http.New`.
5. Endpoints reais: `/api/runs`, `/api/sensors`, `/api/agents`,
   `/api/cost`, `/api/memory`, `/api/autonomy`, `/api/health/score`,
   `/api/logs`, `/api/design`, `/api/roadmap`, `/api/features`.
6. Playwright em `web/playwright/`.

**DEMO (90s):** `harness dashboard --addr 127.0.0.1:7373` (default) ou
`--addr 127.0.0.1:0` (porta escolhida pelo kernel). Abre browser
(`--open`). Mostra memories, sensor runs, cost tracking.

**RECAP (30s):** Embed + Vite + net/http = single binary com UI web.

**CTA (30s):** Próximo: MCP protocol.

**Perguntas plateia:**
1. React embutido no CLI faz sentido?
2. Vite ou esbuild puro?
3. Vale usar tRPC ou fetch cru?

**Comando prático:**
```bash
harness dashboard --addr 127.0.0.1:0 --open
```

---

## EP16 — MCP: templates YAML pra 1-click install

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/mcppkg/templates.go` (71 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/mcppkg/templates/` (bundled YAML)
- `/Users/ropeixoto/dev/projects/harnessx/internal/mcpscan/`
- `/Users/ropeixoto/dev/projects/harnessx/cmd/harness/cmd_mcphook.go`

**Snippet em tela** (`templates.go` linhas 19-51, real):
```go
//go:embed templates/*.yaml
var bundled embed.FS

var bundledFS fs.FS = bundled

type Template struct {
    Name        string            `yaml:"name"`
    Description string            `yaml:"description"`
    Transport   string            `yaml:"transport"`
    Command     string            `yaml:"command"`
    Args        []string          `yaml:"args"`
    URL         string            `yaml:"url"`
    Env         map[string]string `yaml:"env"`
    Docs        string            `yaml:"docs"`
}

func Load(name string) (Template, error) {
    data, err := fs.ReadFile(bundledFS, "templates/"+name+".yaml")
    var t Template
    yaml.Unmarshal(data, &t)
    if t.Name == "" { t.Name = name }
    return t, nil
}
```

**HOOK (15s):** "MCP virou padrão, mas cada server tem command+args
diferente. Harness já vem com templates prontos, `//go:embed`."

**CONTEXTO (60s):** MCP = Model Context Protocol (Anthropic).
Server = stdio ou http. `mcppkg` carrega YAMLs bundled com command,
args, transport, env. `harness mcp install <name>` = 1 comando.

**CODE TOUR (6min):**
1. `//go:embed templates/*.yaml` — 100% ship-with-binary.
2. `Load`/`List` — API pública.
3. `bundledFS` swappable (test seam!).
4. `mcpscan` — audit dos servers configurados.
5. Referência: **Model Context Protocol Spec, Anthropic 2024**.

**DEMO (90s):** `harness mcp templates`. `harness mcp install filesystem`.
Mostra config gerada. Também: `harness mcp scan` (segurança) e
`harness mcp list` (servidores configurados).

**RECAP (30s):** Embed FS + YAML template = install sem manual.

**CTA (30s):** Próximo: design-to-product.

**Perguntas plateia:**
1. MCP vira padrão ou vai o caminho do LangChain?
2. YAML vs TOML vs JSON pra config?
3. Bundle no binário ou fetch remoto?

**Comando prático:**
```bash
harness mcp templates
harness mcp install <name>
harness mcp scan
```

---

## EP17 — Design-to-Product: ZIP → feature-map.json

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/design/build.go` (111 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/design/features.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/design/inventory.go`
- `/Users/ropeixoto/dev/projects/harnessx/internal/design/images.go`
- `/Users/ropeixoto/dev/projects/harnessx/docs/design-handoff-v2/`
- `/Users/ropeixoto/dev/projects/harnessx/docs/design-to-product.md`

**Snippet em tela** (`build.go` linhas 33-85, real):
```go
func Build(opts BuildOptions) (*BuildResult, error) {
    src, err := Resolve(opts.Source)
    defer src.Cleanup()

    manifest, err := Inventory(src)
    fm := BuildFeatureMap(manifest)
    tm := PromoteToggleMap(fm)
    roadmap := BuildRoadmap(fm)
    api := BuildAPIContracts(fm)
    flowMap := BuildFlowMap(manifest)

    cache := ImageCache{Root: opts.Root}
    images, _ := cache.AnalyseAll(src)

    dir := filepath.Join(paths.HarnessDir(opts.Root), "product")
    res.ManifestPath, _  = writeJSON(dir, "design-manifest.json", manifest)
    res.FeatureMapPath, _ = writeJSON(dir, "feature-map.json", fm)
    res.ToggleMapPath, _  = writeJSON(dir, "toggle-map.json", tm)
    res.RoadmapPath, _    = writeJSON(dir, "roadmap.json", roadmap)
    res.APIContractsPath, _ = writeJSON(dir, "api-contracts.json", api)
    res.FlowMapPath, _    = writeJSON(dir, "flow-map.json", flowMap)
    return res, nil
}
```

**HOOK (15s):** "Handoff Figma → dev sempre trava. Aqui você joga o
ZIP e sai feature-map, roadmap, API contracts, flow map — em JSON."

**CONTEXTO (60s):** Pipeline: Resolve → Inventory → BuildFeatureMap →
ToggleMap → Roadmap → APIContracts → FlowMap. Imagens passam por
`ImageCache.AnalyseAll` (heurísticas de OCR/vision opcional). Output
em `.harness/product/`.

**CODE TOUR (6min):**
1. `BuildOptions` — root + source.
2. `Inventory` — walk do source, mapeia artefatos.
3. `BuildFeatureMap` — dedução de features por artefato.
4. `PromoteToggleMap` — features → feature flags.
5. `BuildAPIContracts` — inferência REST heurística.

**DEMO (90s):** `harness design-to-product "port mockup" --source ./mocks.zip`.
Abre `.harness/product/feature-map.json`. Mostra features detectadas.

**RECAP (30s):** ZIP → 6 JSONs auditáveis = handoff acionável.

**CTA (30s):** Próximo (último): evolve.

**Perguntas plateia:**
1. Design-to-code é o futuro ou fantasia?
2. OCR nas imagens vale a pena?
3. Você commitaria `feature-map.json` no repo?

**Comando prático:**
```bash
harness design-to-product "extract product surface" --source ./design.zip
ls .harness/product/
```

---

## EP18 — Evolve: harness que se conserta sob HITL

**Duração:** 10min

**Arquivos reais:**
- `/Users/ropeixoto/dev/projects/harnessx/internal/evolve/evolve.go` (~223 linhas)
- `/Users/ropeixoto/dev/projects/harnessx/internal/evolve/sandbox.go`
- `/Users/ropeixoto/dev/projects/harnessx/.harness/logs/events.jsonl`
- `/Users/ropeixoto/dev/projects/harnessx/.harness/logs/mutations.jsonl`

**Snippet em tela** (`evolve.go` linhas 49-89, real):
```go
func Diagnose(root string) (Diagnosis, error) {
    f, err := os.Open(eventsPath(root))
    d := Diagnosis{Generated: time.Now().UTC()}
    clusters := map[string]*FailureCluster{}
    sc := bufio.NewScanner(f)
    for sc.Scan() {
        d.Events++
        var e Event
        json.Unmarshal(sc.Bytes(), &e)
        if e.Level != "error" && !isFailureFields(e.Fields) { continue }
        d.Failures++
        sig := signature(e.Fields)  // stage|sensor|agent|task|status
        c, ok := clusters[sig]
        if !ok {
            c = &FailureCluster{Signature: sig, Example: truncate(string(line), 256)}
            clusters[sig] = c
        }
        c.Count++
    }
    sort.Slice(d.Clusters, func(i, j int) bool { return d.Clusters[i].Count > d.Clusters[j].Count })
    return d, nil
}
```

E o gate HITL:
```go
func Promote(root string, opts PromoteOptions) error {
    if !opts.HITL {
        return errors.New("evolve: promote requires --hitl (paper §3.5.3 governed mutation)")
    }
    // ...
}
```

**HOOK (15s):** "Última do arco. Como o harness identifica cluster de
falha nos próprios logs, propõe mutação, e SÓ aplica com --hitl."

**CONTEXTO (60s):** Diagnose lê `events.jsonl`, agrupa por
`stage|sensor|agent|task|status`. Propose stage mutation. Replay
mede regressão contra held-out. Promote exige `--hitl`, senão erra.
`mutations.jsonl` auditável.

**CODE TOUR (6min):**
1. `Diagnose` — signature + cluster count.
2. `Propose` — grava "proposed" no log.
3. `Replay` — signature match em trace file.
4. `Promote` — gate HITL obrigatório.
5. Referência: **"Voyager" (arXiv:2305.16291)** + Anthropic's paper
   internamente citado no comment: "Code as Agent Harness §3.5.3
   governed mutation".

**DEMO (90s):** `harness evolve diagnose`. Mostra clusters. `harness
evolve propose --component sensor --description "..." --rationale "..."`.
`harness evolve promote <mutation-id>` falha sem `--hitl`. Com `--hitl
--reason "..."` grava em `mutations.jsonl`.

**RECAP (30s):** Diagnose → Propose → Replay → Promote(HITL) = harness
que evolui sob governança.

**CTA (30s):** Fim do arco. Se curtiu, playlist toda no canal. Próxima
temporada: benchmarks. Deixa nos comentários que módulo você quer
dissecado.

**Perguntas plateia:**
1. HITL escala ou vira gargalo?
2. Cluster por signature texto é frágil?
3. Você deixaria o harness se-modificar sozinho algum dia?

**Comando prático:**
```bash
harness evolve diagnose
cat .harness/logs/events.jsonl | tail -5
```

---

## Legenda de referências consolidadas

- Clean Architecture — Robert C. Martin, 2017
- Constitutional AI — Bai et al., arXiv:2212.08073
- Secrets in Source Code — Meli et al., NDSS 2019
- Chinchilla — Hoffmann et al., arXiv:2203.15556
- RAG — Lewis et al., arXiv:2005.11401
- Cascade LLM Serving — Chen et al., arXiv:2305.05176
- Generative Agents — Park et al., arXiv:2304.03442
- Voyager — Wang et al., arXiv:2305.16291
- GOOS (test doubles) — Freeman & Pryce, 2009
- Model Context Protocol Spec — Anthropic, 2024
- Paper interno citado no código: "Code as Agent Harness" §3.2, §3.5

## Convenção de gravação

- Terminal: `HARNESS_PLAIN=` (deixa cor); Font 18pt Berkeley Mono.
- Editor: Neovim + tema tokyonight, split vertical.
- Overlay: código lipgloss-highlighted à esquerda; câmera 320px canto.
- OBS Studio 1080p60 → export CRF 18.

## Bloco extra — Packages internal/

Cobertura de 10 pacotes `internal/` que ainda não tinham episódio dedicado.
Formato: HOOK 15s / CONTEXTO 60s / CODE TOUR 6min / DEMO 90s / RECAP 30s / CTA 30s.

---

### EP31 — Autonomy: 5 níveis, 6 operações, uma matriz de decisão

**Título:** EP31: Autonomy — a matriz que decide se o agente pergunta ou executa

**Pacote:** `internal/autonomy`

**HOOK (15s):** "Todo agente autônomo tem o mesmo dilema: perguntar demais irrita, executar demais destrói. Harness resolve isso com uma matriz 5×6 de 30 células. Cada célula responde: allow, require_approval, deny."

**CONTEXTO (60s):** Cinco níveis (`manual`, `plan_and_ask`, `safe_execute`, `full_project_loop`, `scheduled_maintenance`) cruzam com seis operações (`read`, `plan`, `execute_low_risk`, `execute_high_risk`, `clean`, `schedule`). O default é `plan_and_ask` — seguro pra onboarding sem sair do caminho.

**CODE TOUR (6min):**
1. `Level`, `Operation`, `Decision` — três tipos string tipados, não `int`, porque serializam direto pra JSON de audit.
2. `policy` — mapa aninhado `map[Level]map[Operation]Decision`. Sem reflection, sem YAML — a política é código.
3. `Gate(level, op)` — retorna a decisão + erro se o nível não existe (`ErrUnknownLevel`).
4. `DefaultSetting()` — sempre `PlanAndAsk`. Mostrar que o pacote não persiste estado — quem persiste é `store.go`.
5. Cruzar com `suggest.go` — a heurística que sobe o nível conforme o usuário aprova operações repetidas.

**Snippet:**
```go
d, err := autonomy.Gate(autonomy.SafeExecute, autonomy.OpExecuteHighRisk)
if err != nil { return err }
switch d {
case autonomy.DecisionAllow:
    return runner.Exec(cmd)
case autonomy.DecisionApproval:
    return approvals.Ask(cmd)
case autonomy.DecisionDeny:
    return fmt.Errorf("blocked at level=%s op=%s", autonomy.SafeExecute, autonomy.OpExecuteHighRisk)
}
```

**DEMO (90s):**
```bash
harness autonomy show
harness autonomy set safe_execute
harness autonomy simulate --op execute_high_risk
```

**RECAP (30s):** Matriz 5×6 explícita. Sem reflection, sem config runtime. Segurança que cabe em uma tabela.

**CTA (30s):** "Se seu nível preferido não existe, abra uma issue com o caso de uso — a matriz é feita pra crescer, não pra virar YAML."

**Perguntas plateia:**
1. Qual nível você usaria pra rodar em CI?
2. `execute_high_risk` deveria ter deny mesmo em `full_project_loop`?
3. Faz sentido um nível "chaos_monkey" pra testes?

---

### EP32 — Hookscan: descobrindo hooks fantasma no repo

**Título:** EP32: Hookscan — quem plantou esse pre-commit no meu repo?

**Pacote:** `internal/hookscan`

**HOOK (15s):** "Você abre um repo antigo e roda `git commit`. Falha. De onde veio esse hook? Claude? Codex? Um script esquecido de 2022? Hookscan responde em 200ms."

**CONTEXTO (60s):** Hookscan varre globs conhecidos (`.harness/hooks/**`, `.claude/hooks/**`, `.codex/hooks/**`, `.gemini/hooks/**`, `.kimi/hooks/**`, `scripts/git-hooks/*`) e classifica cada arquivo por source, event, scope, blocking e risk.

**CODE TOUR (6min):**
1. `Hook` struct — 8 campos JSON, todos derivados só do path e do `os.FileInfo`. Zero side effect, zero exec do hook.
2. `defaultGlobs` — lista fechada dos locais que hoje instalam hooks em projetos com múltiplos assistentes.
3. `Scan(root)` — pipeline: glob → `walkForHookNamed` (fallback semântico por nome) → dedup via `map[string]struct{}` → sort determinístico.
4. `classify` — deriva event de convenções Git (`pre-commit`, `pre-push`, `commit-msg`), scope global vs project, blocking (bloqueia commit se falhar), risk (medium para os que rodam antes de push).
5. `sourceOf` — atribui origem pelo prefixo do path relativo. Ordem importa: `.harness/` vem antes de wildcards.

**Snippet:**
```go
hooks, err := hookscan.Scan(root)
if err != nil { return err }
for _, h := range hooks {
    if h.Blocking && h.Risk == hookscan.RiskMedium {
        fmt.Printf("%s %s (%s) → %s\n", h.Source, h.Name, h.Event, h.ConfigPath)
    }
}
```

**DEMO (90s):**
```bash
harness hook scan
harness hook scan --json | jq '.[] | select(.blocking)'
```

**RECAP (30s):** Sem exec, sem parse — só classificação estática. Idempotente e barato o suficiente pra rodar em cada `harness check`.

**CTA (30s):** "Rode agora no seu repo mais antigo. Aposto que tem hook do Kimi que você esqueceu."

**Perguntas plateia:**
1. Devia haver comando `harness hook disable <name>`?
2. Como classificaria risk para hooks custom?
3. Faz sentido inspecionar o *conteúdo* do hook (parse de comando)?

---

### EP33 — Promptenh + Prompttpl: prompt engineering determinístico

**Título:** EP33: Enhance sem LLM — dois pacotes, zero token gasto

**Pacotes:** `internal/promptenh`, `internal/prompttpl`

**HOOK (15s):** "Todo mundo faz 'prompt engineering' chamando outro LLM. Harness enriquece prompt localmente, sem rede, sem custo — e o resultado é reproduzível bit-a-bit."

**CONTEXTO (60s):** `promptenh` combina skills bundle + summary do context pack + prompt original em uma ordem fixa. `prompttpl` guarda templates reutilizáveis em `.harness/prompts/<name>.md` com validação de nome.

**CODE TOUR (6min):**
1. `promptenh.Enhance(prompt, mode, pack, skills)` — retorna `Enhancement` com original, enhanced, skill_prefixes, context_summary, tokens_added.
2. Ordem fixa: skills (alfabetico por ID) → context summary (top 5 arquivos do pack) → task original. Ordem importa pro cache do LLM.
3. `SkillSource` interface — injeção pra evitar ciclo com `internal/skills`. Padrão clean architecture.
4. `pickSkillsForMode` — filtra por mode do skill (`*` casa com tudo), ordena por score desc + ID asc, corta em 5.
5. `estimateTokens` — heurística `len(s)/4`. Não é BPE de verdade, mas pra budget local basta.
6. `prompttpl.ValidName` — regex `^[a-z0-9][a-z0-9_-]{0,40}$`. Evita path traversal, permite naming humano.
7. `Save` / `Load` / `List` — CRUD em disco, `List` filtra por `.md` e ordena.

**Snippet:**
```go
enh := promptenh.Enhance(rawPrompt, domain.ModePlanning, pack, skills)
path, err := promptenh.Write(runDir, enh)
if err != nil { return err }
fmt.Printf("+%d tokens, artifact=%s\n", enh.TokensAdded, path)
```

**DEMO (90s):**
```bash
harness prompt save review-migration "Review the SQL migration for backwards compat"
harness prompt list
harness prompt enhance "add pagination to /users" --mode planning
```

**RECAP (30s):** Enhance é puro (mesmo input → mesmo output) e barato. Templates viram commodities do time.

**CTA (30s):** "Salve seus 3 prompts mais reutilizados agora e me manda os nomes."

**Perguntas plateia:**
1. Deveria existir `harness prompt share` pra sincronizar entre projetos?
2. Faz sentido versionar templates com git tag interno?
3. Top-5 skills é o limite certo? Por quê não 3?

---

### EP34 — Taskgraph: decompondo prompt em DAG por regex

**Título:** EP34: Taskgraph — como quebro "add auth and run tests" em 2 nós

**Pacote:** `internal/taskgraph`

**HOOK (15s):** "Um prompt, muitas tarefas. Harness decompõe em nós tipados por regex — sem LLM. Cada nó vai pra rota diferente: scaffold determinístico, sensor, ou modelo caro."

**CONTEXTO (60s):** 15 kinds (scaffold, lint, test, format, secrets, code, refactor, docs, review, image, vision, search, data, shell, generic). `Decompose` splita por conjunções ("and", "then", ",", ";") e classifica cada cláusula.

**CODE TOUR (6min):**
1. `Task` — kind + tags + prompt + lang (para scaffold) + `DependsOn` (índices) + confidence.
2. `clauseSplitter` regex `(?i)\s*(?:,| and then | then |, then |; |\sand\s)\s*` — captura conjunções sem confundir "understand" com "and".
3. `rules` — slice ordenado. Ordem importa: scaffold-python vem antes de scaffold-genérico. Regex `(?i)scaffold(?:s| a)? (?:go|golang)\b` captura variações naturais.
4. `classify` — first-match wins, fallback pra `KindGeneric` com confidence 0.3.
5. `Options.UseLLM` — comentado como opt-in futuro. Quando LLM entrar, quem tem confidence < 0.4 vai pro decompose LLM.
6. Como o router consome: cada `Task.Tags` é scoreada contra `adapter.Capabilities.Strengths`.

**Snippet:**
```go
tasks := taskgraph.Decompose("scaffold go api and run lint then write docs", taskgraph.Options{})
for i, t := range tasks {
    fmt.Printf("[%d] kind=%s lang=%s conf=%.2f tags=%v\n",
        i, t.Kind, t.Lang, t.Confidence, t.Tags)
}
```

**DEMO (90s):**
```bash
harness plan "refactor auth module and add tests and update readme"
harness plan "scaffold python fastapi and scan for secrets" --json
```

**RECAP (30s):** Regras > LLM pra decomposição. Rápido, testável, e o LLM só entra quando as regras falham.

**CTA (30s):** "Cole seu último prompt de agente nos comentários — eu falo se o taskgraph classificaria certo."

**Perguntas plateia:**
1. Regex vs LLM pra split — em que ponto vira problema?
2. Que kind está faltando? `KindDeploy`?
3. Confidence 0.3 pra generic é alto demais?

---

### EP35 — Configwiz: mutation com audit trail em JSONL

**Título:** EP35: Configwiz — wizard interativo com diff + audit em append-only

**Pacote:** `internal/configwiz`

**HOOK (15s):** "Toda config runtime vira pesadelo com o tempo. Harness resolve com wizard interativo, diff explícito e log append-only de cada mutação."

**CONTEXTO (60s):** Configwiz gerencia `.harness/config/routes.yaml` — mapeamento task → adapter primary/fallback/budget/model. Cada `Set`/`Delete` gera linha em `.harness/logs/config-mutations.jsonl`.

**CODE TOUR (6min):**
1. `Snapshot` — só um mapa `map[string]router.RouteConfig`. Simples, versionável em YAML.
2. `SetTaskPrimary` — carrega snapshot, aplica mutação, salva, **append audit**. Falha em qualquer passo aborta antes do audit.
3. `Mutation` — time UTC, action, before, after, notes. JSON encoder direto no `os.OpenFile(APPEND|CREATE|WRONLY)`.
4. `Diff(before, after)` — três linhas de diff: `-` remove, `+` adiciona, `~` altera. Formato ergonômico pra colar em PR.
5. `RunWizard` — prompt por task, default do valor atual, valida input não-vazio pra primary. `bufio.Reader` no `In`, `io.Writer` no `Out` — testável sem TTY.
6. `splitCSV` e `SplitCSVForCLI` — helper reexportado só pra CLI reaproveitar.

**Snippet:**
```go
err := configwiz.SetTaskPrimary(root, "planning",
    "anthropic-claude", []string{"openai-gpt4"}, 2.50, "claude-opus-4")
if err != nil { return err }
// audit line já foi escrita em .harness/logs/config-mutations.jsonl
```

**DEMO (90s):**
```bash
harness config wizard
harness config diff HEAD~5
tail -1 .harness/logs/config-mutations.jsonl | jq
```

**RECAP (30s):** Config como código + audit trail. Cada mudança tem timestamp, before, after. Reversível.

**CTA (30s):** "Compartilha nos comentários seu routes.yaml — quero ver como o pessoal balanceia budget entre tasks."

**Perguntas plateia:**
1. Devia existir `harness config revert <mutation-id>`?
2. Audit em JSONL vs SQLite — vale migrar?
3. Wizard interativo faz sentido em 2026 ou tudo devia ser flag?

---

### EP36 — Importwiz: 5 passos pra adotar projeto existente

**Título:** EP36: Importwiz — o wizard de 5 passos pra puxar um repo legado

**Pacote:** `internal/importwiz`

**HOOK (15s):** "Repo existente, ninguém quer refazer. Importwiz absorve em 5 passos verificáveis, cada um com status ok/failed."

**CONTEXTO (60s):** Cinco steps: detect (dir existe?) → stack (Go? Node? Python?) → register (adiciona no workspace registry) → index (stale fingerprints) → done. Cada step reporta status e detail — falha para o pipeline mas mantém contexto.

**CODE TOUR (6min):**
1. `Step` struct — ID, Title, Status, Detail. Pattern "state machine visível".
2. `Plan(opts)` — devolve slice de steps pending. Chamável antes de `Run` pra preview em UI.
3. `Run(ctx, registry, opts)` — sequencial, para no primeiro erro estrutural (folder inexistente, registry falha), mas *continua* se stale.Record falha (não-fatal).
4. `DetectStack` — 7 markers hard-coded (`go.mod`, `Cargo.toml`, `pyproject.toml`, etc). Dedup via `seen` map. Fallback pra `SlugFallbackName`.
5. Integração com `internal/workspace.Registry.Add` — o wizard não sabe SQLite, só chama o método.
6. Integração com `internal/stale.Record` — captura fingerprints pra que o próximo `harness check` já detecte drift.

**Snippet:**
```go
opts := importwiz.Options{Path: "/repos/legacy-api", DisplayName: "Legacy API", Confirm: true}
res, err := importwiz.Run(ctx, registry, opts)
for _, s := range res.Steps {
    fmt.Printf("[%s] %s → %s (%s)\n", s.Status, s.ID, s.Detail, s.Title)
}
```

**DEMO (90s):**
```bash
harness project import ~/code/legacy-api --name "Legacy API"
harness project current
```

**RECAP (30s):** Pipeline de 5 steps com status explícito. Testável, previsível, dá pra reexecutar sem side effect.

**CTA (30s):** "Importa 3 projetos hoje e me conta qual step falhou mais."

**Perguntas plateia:**
1. Que marker de stack está faltando? Deno? Bun?
2. Devia detectar monorepo (turborepo, nx) como stack próprio?
3. Faz sentido preview interativo dos steps antes do commit?

---

### EP37 — Backup: tar.gz reprodutível com manifest SHA256

**Título:** EP37: Backup — tar.gz com manifest SHA256 e restore idempotente

**Pacote:** `internal/backup`

**HOOK (15s):** "Backup sem manifest é loteria. Aqui cada arquivo vai com SHA256, versão do harness, OS, arch. Restore recusa entrada suspeita."

**CONTEXTO (60s):** `Pack` gera tar.gz com policy configurável (include/exclude/secrets), manifest JSON dentro do arquivo. `Unpack` valida path traversal, força target vazio (salvo `--force`), limita cópia em 500MB por arquivo pra evitar zip bomb.

**CODE TOUR (6min):**
1. `Manifest` — Tag, CreatedAt, HarnessVersion, OS, Arch, IncludedRoots, IncludedFiles (com SHA256), ExcludedReasons.
2. `Pack` — cria gzip → tar → walk cada root. `policy.Allow` filtra secrets. `writeTarEntry` calcula SHA256 em `io.MultiWriter(tw, h)` — passa uma vez pelo arquivo.
3. `PackName(tag)` — nome com timestamp UTC `20060102T150405Z` + tag sanitizado. Determinismo humano.
4. `Unpack` — preflight (target vazio), `captureManifest` intercepta o `.harness-backup-manifest.json`, `writeUnpackedEntry` recusa `..` e paths absolutos.
5. `io.CopyN(out, tr, 500*1024*1024)` — anti-zip-bomb explícito.
6. `safeTag` — sanitiza tag substituindo tudo que não é alnum/`-`/`_` por `-`.

**Snippet:**
```go
cfg := backup.Config{Include: []string{".harness", "docs"}}
manifest, err := backup.Pack(root, cfg, "pre-migrate", "backup.tar.gz", false)
if err != nil { return err }
fmt.Printf("packed %d files, harness=%s\n",
    len(manifest.IncludedFiles), manifest.HarnessVersion)
```

**DEMO (90s):**
```bash
harness backup pack --tag pre-refactor
harness backup ls
harness backup restore backup.tar.gz --to /tmp/restored
```

**RECAP (30s):** Manifest SHA256, path guard, size guard. Backup que aguenta auditoria.

**CTA (30s):** "Roda `harness backup pack` antes do próximo refactor grande. Depois vem me contar se salvou."

**Perguntas plateia:**
1. 500MB por arquivo é limite razoável? Deveria ser configurável?
2. Faz sentido split em múltiplos tar.gz por policy?
3. Devia haver `--verify` que compara SHA256 do disco vs manifest?

---

### EP38 — Analytics: agregando spend em stack/adapter/day

**Título:** EP38: Analytics — quanto seu prompt engineering está custando por stack

**Pacote:** `internal/analytics`

**HOOK (15s):** "Você sabe qual stack seu time queima mais token? Qual adapter tá abusando do budget? Analytics agrega spend de todos os projetos em segundos."

**CONTEXTO (60s):** Percorre `.harness/repl/sessions/` de N roots, agrega em 3 eixos: por stack, por adapter+task, por dia. Filtro `since` pra recortar janela.

**CODE TOUR (6min):**
1. `Row`, `AdapterRow`, `DayRow`, `Report` — tipos separados por eixo. Cada um tem CostUSD + counters.
2. `Walk(roots, since)` — 3 maps de agregação (`byStack`, `byAdapter`, `byDay`), depois ordena e converte pra slice.
3. `walkSessions` — abre cada sessão via `repl.Session`, chama callback. Reaproveita parser do REPL, não duplica.
4. `collectStackRow` / `collectAdapterRow` / `collectDayRow` — cada um cria row on-demand se ainda não existe (padrão upsert-in-map).
5. Cost só soma se `AdapterID != ""` ou `CostUSD > 0` — filtra turns locais que não custaram nada.
6. Ordenação final por CostUSD desc — top spenders no topo.

**Snippet:**
```go
report, err := analytics.Walk(roots, time.Now().AddDate(0, 0, -30))
if err != nil { return err }
fmt.Printf("$%.2f total across %d turns\n", report.TotalUSD, report.TotalTurns)
for _, s := range report.Stacks[:5] {
    fmt.Printf("  %-10s $%.2f (%d turns)\n", s.Stack, s.CostUSD, s.Turns)
}
```

**DEMO (90s):**
```bash
harness analytics report --since 30d
harness analytics report --by adapter --json
```

**RECAP (30s):** 3 eixos, 1 walk. Sem telemetria remota — tudo do `.harness/` local.

**CTA (30s):** "Roda `harness analytics report --since 7d` e me conta a stack mais cara. Aposto que é a que ninguém audita."

**Perguntas plateia:**
1. Que 4º eixo faz mais falta? Por hora do dia?
2. Faz sentido export pra Grafana?
3. Devia haver alerta quando spend passa X% da média histórica?

---

### EP39 — Multimodal: grounding text ↔ regions de imagem

**Título:** EP39: Multimodal — como confirmo que o LLM viu o botão certo

**Pacote:** `internal/multimodal`

**HOOK (15s):** "LLM alucina em imagem que é uma beleza. Multimodal.CheckGrounding compara texto gerado contra annotations manuais e retorna hits + missing."

**CONTEXTO (60s):** Sidecar JSON com `schema_version: 1` guarda regions (x,y,w,h) + label. Depois de gerar descrição/tarefa, `CheckGrounding` verifica se cada label aparece no texto.

**CODE TOUR (6min):**
1. `Region`, `Annotation`, `Sidecar` — 3 tipos simples, JSON-first. Sem lib externa de imagem.
2. `SchemaVersion = 1` const — versionamento explícito. Load recusa se não bater.
3. `LoadAnnotations(path)` — `os.ReadFile` + `json.Unmarshal` + check de schema. Erro descritivo com wrap.
4. `CheckGrounding(text, anns)` — normaliza tudo em lowercase, `strings.Contains` por label. Barato e determinístico.
5. `GroundingResult` — separa hits e missing. Missing é o valioso: o LLM não mencionou o "submit button" que estava anotado.
6. Como plugar num pipeline: gerar descrição via adapter vision → passar em `CheckGrounding` → se missing > 30% da lista, sinalizar low grounding e re-prompt.

**Snippet:**
```go
anns, err := multimodal.LoadAnnotations("login.sidecar.json")
if err != nil { return err }
desc := adapter.DescribeImage("login.png")
res := multimodal.CheckGrounding(desc, anns)
fmt.Printf("grounded: %d/%d, missing: %v\n",
    len(res.Hits), len(res.Hits)+len(res.Missing), res.Missing)
```

**DEMO (90s):**
```bash
harness multimodal check --image login.png --sidecar login.sidecar.json --text "$(cat desc.txt)"
```

**RECAP (30s):** Grounding local sem CV. Substring match não é IoU, mas serve pra detectar alucinação óbvia.

**CTA (30s):** "Cria uma sidecar pra sua última screenshot de PR e roda o check. Me manda o missing."

**Perguntas plateia:**
1. Substring é frágil demais? Devia usar embedding?
2. Vale suportar múltiplos labels por region?
3. Como estender pra vídeo (keyframe + timestamp)?

---

### EP40 — Stack: o tour end-to-end que testa a instalação

**Título:** EP40: Stack.Tour — o smoke test que exercita 8 subsistemas

**Pacote:** `internal/stack`

**HOOK (15s):** "Instalou o harness? O tour roda 8 subsistemas em sequência: workspace, catalog, cleanup, autonomy, health, dashboard probe. Se um falha, o instalação está quebrada."

**CONTEXTO (60s):** `stack.Tour` recebe root + templates source + registry path + dashboard probe URL. Retorna `[]StepResult` — cada step com name, detail, latency e err. Formato ideal pra smoke test em CI ou onboarding.

**CODE TOUR (6min):**
1. `Tour` struct — 5 campos configuráveis + `Now` injetável (testes determinísticos).
2. `Run(ctx, out)` — pipeline de 6-8 steps. Padrão: `start := clock(); ...; addStep(name, detail, start, err)`.
3. Steps: `ensure_root` → `copy_templates` → `workspace_open` → `workspace_add` → `catalog_plan` + `catalog_install` (MCP filesystem) → `cleanup_scan` → `autonomy_gate` → `health_score` → `dashboard_probe` (opcional).
4. `printStep` — formato tabular com status/name/latency/detail. Perfeito pra CLI e log de CI.
5. Erros abortam o pipeline (exceto autonomy_gate que é probe). Retorna slice parcial + err.
6. `copyTree` — walk + read + write recursivo. Sem symlink, sem chmod fancy.

**Snippet:**
```go
tour := &stack.Tour{
    Root: "/tmp/harness-tour",
    TemplatesSrc: "assets/templates",
    RegistryPath: "/tmp/registry.db",
    DashboardProbe: "http://localhost:8080/healthz",
}
results, err := tour.Run(ctx, os.Stdout)
fmt.Printf("%d steps, err=%v\n", len(results), err)
```

**DEMO (90s):**
```bash
harness stack tour --root /tmp/harness-demo
harness stack tour --root /tmp/harness-demo --dashboard-probe http://localhost:8080/healthz
```

**RECAP (30s):** Um comando exercita 8 pacotes. Se o tour passa, sua instalação está sã.

**CTA (30s):** "Roda o tour agora, na sua máquina limpa, e me manda o print da tabela final."

**Perguntas plateia:**
1. Que 9º step faria mais sentido? `sensor_run`?
2. Deveria haver `--parallel` pra steps independentes?
3. Vale exportar o resultado em JUnit XML pra CI?

---

## EP19 — harness ask vs harness chat: determinístico vs REPL

**Título (58 chars):** ask vs chat: quando o LLM entra no seu terminal

**HOOK (15s):** "Você digita uma pergunta. `ask` te devolve arquivos. `chat` te devolve um agente. Um custa 0 tokens. O outro custa o que você deixar."

**CONTEXTO (60s):** Duas superfícies distintas no `cmd/harness`. `ask` (`cmd_workflow.go:15`) monta pack de evidência via ripgrep + git status — sem LLM salvo se você passar `--agent`. `chat` (`cmd_chat.go:47`) entra num REPL scoped por goal (`dev|ads|research|ops`) com auto-pin do agente em `active.yaml`. Diferença arquitetural: `ask` é read-only + deterministic; `chat` mantém state em `.harness/sessions/` com resume/replay.

**CODE TOUR (6min):**
- `cmd/harness/cmd_workflow.go:15-46` — flags `--agent`, `--budget-usd 0.05`, `--evidence-only`.
- `internal/app/workflow` — função `Ask` monta contexto sem chamar router.
- `cmd/harness/cmd_chat.go:30-45` — `ChatOpts` (Goal, AdapterID, StepTimeout, ResumeID, ReplayID, AutoGate).
- `internal/repl` — loop iterativo + slash commands.
- Auto-pin: sem `--adapter`, lê `.harness/config/active.yaml`.

**Snippet Go real (cmd_workflow.go:29-45):**
```go
Args: cobra.MinimumNArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
    dir, err := cwd()
    if err != nil { return err }
    _, err = workflow.Ask(cmd.Context(), workflow.Options{
        StartDir: dir, Prompt: strings.Join(args, " "),
        AgentID: agentID, BudgetUSD: budgetUSD,
        EvidenceOnly: evidence,
    }, cmd.OutOrStdout())
    return err
},
```

**DEMO (90s):**
```bash
harness ask "onde está o router de adapters?" --evidence-only
harness chat --goal dev --step-timeout 3m
```

**RECAP (30s):** `ask` = context pack determinístico, LLM opt-in. `chat` = REPL goal-aware com auto-pin. Escolha por intent, não por hábito.

**CTA (30s):** Comenta qual dos dois entrou no seu workflow diário.

**Perguntas plateia:**
1. `--evidence-only` mata a UX ou salva conta no fim do mês?
2. REPL com resume/replay é overkill ou base pra reprodutibilidade?
3. Auto-pin via `active.yaml` deveria virar per-goal?

**Referência paper:** Code as Agent Harness §3.1.4 (REPL determinístico).

---

## EP20 — harness do: task graph + roteamento por adapter

**Título (55 chars):** harness do: um prompt vira grafo de tasks roteadas

**HOOK (15s):** "Você escreve `scaffold python and add /healthz`. O harness decompõe em duas tasks, roteia cada uma pro adapter que sabe fazer, e a determinística nem paga LLM."

**CONTEXTO (60s):** `harness do` (`cmd_do.go:42`) é a superfície de execução multi-adapter. Chama `docmd.Plan` → `taskgraph` decompõe → router escolhe adapter por match de tags. Tasks determinísticas (scaffold, lint, secrets-scan) skip LLM by default.

**CODE TOUR (6min):**
- `cmd_do.go:20-36` — `doOpts` (yes, det, budget, maxTasks, autonomy, image, asJSON, agentOverride).
- `docmd.Plan` — chama taskgraph + router.
- `internal/taskgraph` — decomposição por padrões.
- `docmd.DeterministicMatch` — decide se pula LLM.
- `newRouteCmd()` — `harness route show` faz dry-run.

**Snippet Go real (cmd_do.go:59-67):**
```go
c.Flags().BoolVar(&opts.yes, "yes", false, "skip plan confirmation prompt")
c.Flags().BoolVar(&opts.det, "deterministic", true, "prefer scaffold/sensor over LLM where possible")
c.Flags().Float64Var(&opts.budget, "budget-usd", 1.0, "max USD across all routed tasks")
c.Flags().IntVar(&opts.maxTasks, "max-tasks", 10, "hard cap on decomposed tasks")
c.Flags().StringVar(&opts.autonomy, "autonomy", "safe_execute", "autonomy level for LLM tasks")
c.Flags().StringVar(&opts.image, "image", "", "attach an image; auto-adds vision tag for routing")
c.Flags().BoolVar(&opts.asJSON, "json", false, "emit final result as JSON (implies --yes)")
c.Flags().StringVar(&opts.agentOverride, "agent", "", "force a specific adapter id (overrides router + active pin)")
```

**DEMO (90s):**
```bash
harness route show "scaffold python and add a /healthz endpoint"
harness do "scaffold python and add /healthz" --yes --budget-usd 0.30
```

**RECAP (30s):** Um prompt → grafo → adapters por task. `--deterministic true` (default) mata LLM onde dá.

**CTA (30s):** Diz aí qual task determinística você tira do seu adapter hoje.

**Perguntas plateia:**
1. Decomposição por regex/tag é robusta ou precisa LLM?
2. `max-tasks 10` é limite bom pra proteger budget?
3. `--json` implicando `--yes` é opinião perigosa?

**Referência paper:** Code as Agent Harness §3.3 (task routing).

---

## EP21 — harness ci vs check: mesmo sensor set, gate diferente

**Título (58 chars):** ci vs check: quando falha vira exit non-zero

**HOOK (15s):** "Mesmo conjunto de sensores. `check` te avisa. `ci` te bloqueia. E `--fast` decide se secrets_scan roda hoje ou não."

**CONTEXTO (60s):** Dois comandos gêmeos em `cmd_sensor.go`. Ambos chamam `sensorcmd.Run` mas com `FailOnError` diferente. `check` observa; `ci` enforce. `ci` ganha `--fast` (pula sensores lentos) e `--install-missing` (auto-instala bandit/mypy/pip-audit se ausente).

**CODE TOUR (6min):**
- `cmd_sensor.go:60-75` — `newCheckCmd()` com `FailOnError: false`.
- `cmd_sensor.go:77-101` — `newCICmd()` com `FailOnError: true` + flags `--fast` + `--install-missing`.
- Wire com `pre-push` hook em `CONTRIBUTING.md`.
- `internal/sensors` — set de sensores por linguagem/framework.

**Snippet Go real (cmd_sensor.go:85-100):**
```go
RunE: func(cmd *cobra.Command, _ []string) error {
    dir, err := cwd()
    if err != nil { return err }
    if _, err := sensorcmd.Run(cmd.Context(), sensorcmd.RunOptions{
        StartDir: dir, FailOnError: true, Fast: fast, InstallMissing: installMissing,
    }, cmd.OutOrStdout()); err != nil {
        return err
    }
    return nil
},
```

**DEMO (90s):**
```bash
harness check         # exit 0 sempre, só reporta
harness ci --fast     # exit non-zero se qualquer sensor falhar
```

**RECAP (30s):** `check` observa. `ci` enforça. Wire no pre-push do git.

**CTA (30s):** Diz aí se seu pre-push hoje é lento demais pra tolerar.

**Perguntas plateia:**
1. `--install-missing` deveria ser opt-in ou opt-out?
2. Sensor "lento" (secrets_scan) merece flag `--fast` global?
3. Dois comandos ou uma flag `--enforce`?

---

## EP22 — harness autonomy: ask, safe_execute, strict

**Título (59 chars):** autonomy: matriz de gates que decide se LLM pode agir

**HOOK (15s):** "`autonomy get` te mostra uma matriz. Cada linha é o quanto o harness deixa o LLM tocar seu filesystem. Da leitura pura ao full_project_loop."

**CONTEXTO (60s):** `harness autonomy` (`cmd_autonomy.go:14`) expõe a matriz de decisões: para cada operação (`Read`, `Plan`, `ExecuteLowRisk`, `ExecuteHighRisk`, `Clean`, `Schedule`) e cada nível, retorna a decisão do gate. `set` persiste em `.harness/config/autonomy`.

**CODE TOUR (6min):**
- `internal/autonomy` — enum `Operation` + `Level`.
- `cmd_autonomy.go:16-32` — `get` itera todos os pares e imprime tabela.
- `cmd_autonomy.go:33-49` — `set <level>` persiste.
- `cmd_autonomy.go:50-65` — `active` lê o pin (default `manual`).

**Snippet Go real (cmd_autonomy.go:19-32):**
```go
RunE: func(cmd *cobra.Command, _ []string) error {
    ops := []autonomy.Operation{autonomy.OpRead, autonomy.OpPlan, autonomy.OpExecuteLowRisk, autonomy.OpExecuteHighRisk, autonomy.OpClean, autonomy.OpSchedule}
    fmt.Fprintf(cmd.OutOrStdout(), "%-22s | %s\n", "level", "decisions")
    for _, lvl := range autonomy.AllLevels() {
        line := ""
        for _, op := range ops {
            dec, _ := autonomy.Gate(lvl, op)
            line += fmt.Sprintf("%s=%s ", op, dec)
        }
        fmt.Fprintf(cmd.OutOrStdout(), "%-22s | %s\n", lvl, line)
    }
    return nil
},
```

**DEMO (90s):**
```bash
harness autonomy get
harness autonomy set safe_execute
harness autonomy active
```

**RECAP (30s):** Autonomy = gate por operação. Matriz explícita, persistência em YAML, default paranoico.

**CTA (30s):** Qual level você usaria em produção? Comenta.

**Perguntas plateia:**
1. `manual` como default é excesso ou lucidez?
2. `full_project_loop` deveria exigir 2FA?
3. Faltam operações (Network, Secrets)?

**Referência paper:** Code as Agent Harness §3.5 (autonomy tiers).

---

## EP23 — harness plan write: plan-as-contract sem LLM

**Título (58 chars):** plan write: contrato de plano, zero LLM, escopo forçado

**HOOK (15s):** "Você declara os arquivos. Você declara os invariantes. Você declara o rollback. Se o agente sair da linha, sensor bloqueia. Sem LLM na hora de escrever."

**CONTEXTO (60s):** `harness plan write` (`cmd_plan_write.go:46`) materializa planejamento como contrato: intent + files + invariants + validation + rollback + risk tier. Sem chamada LLM. Salva em `.harness/artifacts/plans/PLAN-<ulid>.md`. `harness plan check --plan <id>` verifica que edits ficaram no escopo (sensor `planscope`).

**CODE TOUR (6min):**
- `cmd_plan_write.go:46-98` — flags `--file`, `--invariant`, `--validate`, `--rollback`, `--risk`.
- `renderPlanContract` (linhas 111-138) — template markdown determinístico.
- `internal/sensors/planscope` — `Check` diffa arquivos alterados vs escopo.
- `cmd_plan_write.go:18-44` — `plan check` roda planscope.

**Snippet Go real (cmd_plan_write.go:64-90):**
```go
RunE: func(cmd *cobra.Command, args []string) error {
    root, err := cwd()
    if err != nil { return err }
    prompt := strings.Join(args, " ")
    id := ids.New()
    body := renderPlanContract(planContract{
        ID: id, Intent: prompt, Files: files, Invariants: invariants,
        Validation: validation, Rollback: rollback, Risk: risk,
        CreatedAt: time.Now().UTC(),
    })
    dst := filepath.Join(root, ".harness", "artifacts", "plans", "PLAN-"+id+".md")
    if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
    if err := os.WriteFile(dst, []byte(body), 0o644); err != nil { return err }
    fmt.Fprintf(cmd.OutOrStdout(), "plan: wrote %s\n", dst)
    return nil
},
```

**DEMO (90s):**
```bash
harness plan write "add /healthz endpoint" \
  --file cmd/server/main.go --invariant "no new deps" \
  --validate "harness ci" --risk low
harness plan check --plan PLAN-01H...
```

**RECAP (30s):** Plano é contrato. Contrato é sensor. Sensor é gate. LLM só entra pra preencher, nunca pra decidir escopo.

**CTA (30s):** Comenta se seu time hoje planeja em markdown ou no ar.

**Perguntas plateia:**
1. Escopo declarado em YAML mata criatividade?
2. `risk` de 3 níveis é suficiente?
3. Rollback default `git revert HEAD` é frágil?

**Referência paper:** Code as Agent Harness §3.4.2 (PLAN contract).

---

## EP24 — harness cost-compare: delta USD honesto por modelo

**Título (57 chars):** cost-compare: quanto CADA modelo cobraria pelo prompt

**HOOK (15s):** "Você cola o prompt. O harness monta o context pack real. E imprime, ordenado, quanto Claude Opus, Sonnet, GPT-5, Gemini, Kimi, DeepSeek e Llama local cobrariam. Harness na coluna: $0."

**CONTEXTO (60s):** `harness cost-compare` (`cmd_cost_compare.go:11`) constrói o mesmo pack determinístico do harness, aplica tier de effort (low|medium|high — multiplicador output + overhead extended thinking), e compara custo por modelo bundled. `--criteria` imprime também a matriz task→(model, effort).

**CODE TOUR (6min):**
- `cmd_cost_compare.go:13-16` — flags `--out-tokens 4000`, `--effort medium`, `--criteria`.
- `internal/app/costcomparecmd` — `Estimate` + `Render`.
- Modelos bundled: Opus 4.7, Sonnet 4.6 1M, Haiku 4.5, GPT-5/mini/nano, Gemini 2.5, Kimi K2, DeepSeek V3, Llama 3.1 local.
- Harness column = $0 (0 tokens LLM, determinístico).

**Snippet Go real (cmd_cost_compare.go:34-51):**
```go
RunE: func(cmd *cobra.Command, args []string) error {
    root, err := cwd()
    if err != nil { return err }
    opts := costcomparecmd.Options{
        Root:         root,
        Prompt:       strings.Join(args, " "),
        OutputTokens: outBudgetTokens,
        Effort:       costcomparecmd.Effort(effort),
        ShowCriteria: showCriteria,
    }
    res, err := costcomparecmd.Estimate(cmd.Context(), opts)
    if err != nil { return err }
    return costcomparecmd.Render(cmd.OutOrStdout(), res, opts)
},
```

**DEMO (90s):**
```bash
harness cost-compare "refactor the router to add fallback tier" \
  --effort high --criteria
```

**RECAP (30s):** Honestidade tokenizada. Você vê o delta antes de gastar. Harness fica em $0 sempre.

**CTA (30s):** Diz qual modelo te surpreendeu no ranking.

**Perguntas plateia:**
1. Effort tier de 3 níveis captura reasoning models?
2. Llama local deveria contar custo elétrico?
3. Criteria table é o real killer-feature?

**Referência paper:** Code as Agent Harness §3.2 (cost transparency).

---

## EP25 — harness perf-snapshot + perf-compare: baseline versionado

**Título (58 chars):** perf-snapshot: baseline de resource antes/depois

**HOOK (15s):** "Antes da mudança: snapshot. Depois da mudança: snapshot. `perf-compare` diffa. Se piorou, você mostra a régua. Se melhorou, mostra a régua."

**CONTEXTO (60s):** `harness perf-snapshot` (`cmd_optimize.go:25`) captura consumo de recursos (Cycle A do resource-optimization playbook). `perf-compare` diffa dois snapshots — default os dois mais recentes. Reports em `.harness/artifacts/perf/`.

**CODE TOUR (6min):**
- `cmd_optimize.go:25-44` — `perf-snapshot` com `--label` + `--report`.
- `cmd_optimize.go:46-68` — `perf-compare [from] [to]` com flags também nomeadas.
- `internal/app/optimizecmd` — `PerfSnapshot`, `PerfCompare`.
- `harness optimize` roda A→G tudo.

**Snippet Go real (cmd_optimize.go:29-43):**
```go
c := &cobra.Command{
    Use:   "perf-snapshot",
    Short: "Capture a resource snapshot (Cycle A)",
    RunE: func(cmd *cobra.Command, _ []string) error {
        dir, err := cwd()
        if err != nil { return err }
        return optimizecmd.PerfSnapshot(optimizecmd.SnapshotOptions{
            StartDir: dir, Label: label, Report: report,
        }, cmd.OutOrStdout())
    },
}
c.Flags().StringVar(&label, "label", "", "human label for the snapshot")
c.Flags().BoolVar(&report, "report", false, "also write a markdown report")
```

**DEMO (90s):**
```bash
harness perf-snapshot --label pre-refactor --report
# ...faz a mudança...
harness perf-snapshot --label post-refactor --report
harness perf-compare
```

**RECAP (30s):** Snapshot antes. Snapshot depois. Diff explicito. Anexa no PR.

**CTA (30s):** Diz aí que métrica você mede hoje sem harness.

**Perguntas plateia:**
1. Snapshot deveria ser auto no pre-push?
2. `--label` livre é bom ou precisa taxonomy?
3. Compare default dos 2 mais recentes é intuitivo?

---

## EP26 — harness backup snapshot: rclone + snapshot cifrado remoto

**Título (60 chars):** backup snapshot: .harness portável em drive/s3/r2

**HOOK (15s):** "Um binário. Zero credenciais no config do harness. Só rclone. Snapshot de `.harness/` sobe em drive, s3, dropbox, onedrive, r2, webdav ou overlay crypt."

**CONTEXTO (60s):** `harness backup` (`cmd_backup.go:48`) wrapa rclone. Snapshot exclui secrets por default; incluir requer `--include-secrets` + env `HARNESS_BACKUP_I_UNDERSTAND_SECRETS=1`. Credenciais do provider ficam SÓ no rclone config — harness nunca toca.

**CODE TOUR (6min):**
- `cmd_backup.go:60-64` — subcomandos snapshot, restore, list, sync, remotes, remote-add, config, quickstart.
- `cmd_backup.go:113-135` — `snapshot [path]` com `--remote`, `--tag`, `--include-secrets`, `--dry-run`.
- `cmd_backup.go:179-198` — `sync push|pull` mirror config.
- `cmd_backup.go:211-227` — `remote add --provider drive|s3|dropbox|onedrive|r2|webdav|crypt`.

**Snippet Go real (cmd_backup.go:113-135):**
```go
func newBackupSnapshotCmd() *cobra.Command {
    var opts backupcmd.SnapshotOpts
    c := &cobra.Command{
        Use:   "snapshot [path]",
        Short: "Snapshot a project's .harness state to the configured remote",
        Args:  cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            root, err := cwd()
            if err != nil { return err }
            if len(args) == 1 { root = args[0] }
            return backupcmd.Snapshot(cmd.Context(), cmd.OutOrStdout(), root, opts, newRcloneFn)
        },
    }
    c.Flags().StringVar(&opts.Remote, "remote", "", "rclone remote name (overrides default_remote)")
    c.Flags().StringVar(&opts.Tag, "tag", "", "label appended to the snapshot filename")
    c.Flags().BoolVar(&opts.IncludeSecrets, "include-secrets", false, "include secrets.enc + secret-seed (DANGEROUS)")
    c.Flags().BoolVar(&opts.DryRun, "dry-run", false, "pack locally; skip upload")
    return c
}
```

**DEMO (90s):**
```bash
harness backup quickstart
harness backup remote add mydrive --provider drive
harness backup snapshot --tag pre-migration
harness backup list
```

**RECAP (30s):** Portabilidade sem lock-in. Credenciais no rclone, snapshot no bucket, restore em outra máquina em 30s.

**CTA (30s):** Comenta qual provider você usa. Suporta 7.

**Perguntas plateia:**
1. Wrappear rclone é abstração certa ou dependência?
2. `--include-secrets` deveria exigir mais um passphrase?
3. Sync push/pull vs snapshot: quando cada um?

---

## EP27 — harness audit tail: event log append-only cross-project

**Título (58 chars):** audit tail: timeline append-only de tudo que rolou

**HOOK (15s):** "Sensor rodou. Hook disparou. Agent respondeu. Cleanup limpou. Tudo em `events.jsonl`. `audit tail -30` imprime os últimos. Sem banco. Sem servidor."

**CONTEXTO (60s):** `harness audit` (`cmd_metrics.go:184`) lê `.harness/audit/events.jsonl` append-only. Filtra por `--kind sensor|hook|agent|cleanup|...`, limita com `--limit`, emite JSON com `--json`. Subcomando `tail` faz os últimos N. `harness audit replay --id <run-id>` (`internal/audit/replay.go`) filtra eventos daquele run pelos dois locais canônicos (`.harness/audit/events.jsonl` e `.harness/logs/events.jsonl`), ordena cronologicamente e imprime cada step com marcador `would replay` (default `--dry-run=true`) ou `replayed` (quando `--dry-run=false`, gravando `manifest.txt` só dentro do tmpdir). Nunca escreve fora do tmpdir; retorna erro `ErrRunNotFound` se run id não existir em nenhum log nem em `.harness/runs/<id>/`. Útil pra investigar regressão de sensor sem sujar tree.

**CODE TOUR (6min):**
- `cmd_metrics.go:184-230` — `newAuditCmd()` com filtros.
- `cmd_metrics.go:232-259` — `newAuditTailCmd()` — últimos N.
- `internal/audit` — writer append-only.
- JSONL: cada linha = `{OccurredAt, Kind, Source, Subject}`.

**Snippet Go real (cmd_metrics.go:232-258):**
```go
func newAuditTailCmd() *cobra.Command {
    var limit int
    c := &cobra.Command{
        Use:   "tail",
        Short: "Tail the last N events from the project event log",
        RunE: func(cmd *cobra.Command, _ []string) error {
            root, err := cwd()
            if err != nil { return err }
            events, err := loadAuditEvents(root)
            if err != nil { return err }
            if limit > 0 && len(events) > limit {
                events = events[len(events)-limit:]
            }
            w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
            fmt.Fprintln(w, "WHEN\tKIND\tSOURCE\tSUBJECT")
            for _, e := range events {
                fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.OccurredAt.Format("01-02 15:04:05"), e.Kind, e.Source, e.Subject)
            }
            return w.Flush()
        },
    }
    c.Flags().IntVar(&limit, "limit", 30, "max rows")
    return c
}
```

**DEMO (90s):**
```bash
harness audit --kind sensor --limit 20
harness audit tail --limit 10
harness audit --json | jq '.[] | select(.kind=="agent")'

# replay events de um run específico em tmpdir isolado (dry-run default)
harness audit replay --id 01KX3ANMQ33YY71GW97ZH9XBJD

# JSON estável pra pipeline
harness audit replay --id 01KX3ANMQ33YY71GW97ZH9XBJD --json | jq '.steps | length'

# materializa manifest dentro do tmpdir (--dry-run=false), sem tocar no repo
harness audit replay --id 01KX3ANMQ33YY71GW97ZH9XBJD --dry-run=false --tmp-dir /tmp/replay-01
```

**RECAP (30s):** JSONL append-only. Filesystem como source-of-truth. Replay sandbox em tmpdir. Post-mortem sem grafana.

**CTA (30s):** Marca esse video se você já debugou incidente com jq.

**Perguntas plateia:**
1. `audit replay` já existe em dry-run — próximo passo é re-executar hooks/sensors idempotentes de verdade dentro do tmpdir. Vale a pena?
2. JSONL escala pra 100k events?
3. `--kind` livre ou taxonomy fechado?

---

## EP28 — harness image-audit + dependency-audit: cycle B + C

**Título (59 chars):** image-audit + dependency-audit: matando peso morto

**HOOK (15s):** "Dockerfile com `apt-get install` sem `--no-install-recommends`. Deps sem uso. Duas linhas de comando pra achar tudo."

**CONTEXTO (60s):** Dois comandos do playbook resource-optimization. `image-audit` (`cmd_optimize.go:70`) faz auditoria estática de Dockerfile (Cycle B). `dependency-audit` (`cmd_optimize.go:84`) classifica deps e flag candidatos a remoção (Cycle C). Ambos delegam pra `internal/app/optimizecmd`.

**CODE TOUR (6min):**
- `cmd_optimize.go:70-82` — `image-audit`.
- `cmd_optimize.go:84-96` — `dependency-audit`.
- `internal/app/optimizecmd` — `ImageAudit`, `DependencyAudit`.
- Cycles A–G no `docs/resource-optimization.md`.

**Snippet Go real (cmd_optimize.go:70-96):**
```go
func newImageAuditCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "image-audit",
        Short: "Static Dockerfile audit (Cycle B)",
        RunE: func(cmd *cobra.Command, _ []string) error {
            dir, err := cwd()
            if err != nil { return err }
            return optimizecmd.ImageAudit(optimizecmd.AuditOptions{StartDir: dir}, cmd.OutOrStdout())
        },
    }
}

func newDependencyAuditCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "dependency-audit",
        Short: "Classify dependencies + flag removal candidates (Cycle C)",
        RunE: func(cmd *cobra.Command, _ []string) error {
            dir, err := cwd()
            if err != nil { return err }
            return optimizecmd.DependencyAudit(optimizecmd.AuditOptions{StartDir: dir}, cmd.OutOrStdout())
        },
    }
}
```

**DEMO (90s):**
```bash
harness image-audit
harness dependency-audit
harness optimize   # roda A→G tudo
```

**RECAP (30s):** Estática, offline, determinístico. Zero LLM. Cortou 400MB da minha imagem semana passada.

**CTA (30s):** Comenta qual dep zumbi você achou.

**Perguntas plateia:**
1. Auditoria estática de Dockerfile compete com hadolint?
2. `dependency-audit` deveria diffar `go.sum` histórico?
3. Cycle A–G é excesso de metodologia?

---

## EP29 — harness palette search: command palette no CLI

**Título (58 chars):** palette search: fuzzy nos projects, capabilities, commands

**HOOK (15s):** "VSCode tem cmd+shift+p. iOS tem spotlight. CLI ganha `harness palette search`. Três fontes: projects, capabilities, commands."

**CONTEXTO (60s):** `harness palette search` (`cmd_palette.go:15`) faz busca cross-source: workspace registry (projects), catalog (capabilities), builtin commands. Retorna hits com Source, Kind, Title, RouterPath.

**CODE TOUR (6min):**
- `cmd_palette.go:35-40` — instancia `palette.New` com 3 sources.
- `internal/palette` — `ProjectsSource`, `CapabilitiesSource`, `CommandsSource`.
- `palette.BuiltinCommands` — lista estática.
- `internal/workspace` — registry de projects.

**Snippet Go real (cmd_palette.go:26-52):**
```go
reg, _ := workspace.Open("")
if reg != nil { defer reg.Close() }
cat := catalogcmd.New()
root, err := cwd()
if err != nil { return err }
p := palette.New(
    palette.ProjectsSource{Registry: reg},
    palette.CapabilitiesSource{Catalog: cat, Root: root},
    palette.CommandsSource{Commands: palette.BuiltinCommands},
)
hits, err := p.Search(cmd.Context(), q)
if err != nil { return err }
if len(hits) == 0 { fmt.Fprintln(cmd.OutOrStdout(), "no hits"); return nil }
for _, h := range hits {
    fmt.Fprintf(cmd.OutOrStdout(), "%-12s %-14s %-40s %s\n", h.Source, h.Kind, h.Title, h.RouterPath)
}
return nil
```

**DEMO (90s):**
```bash
harness palette search backup
harness palette search sensor python
```

**RECAP (30s):** Discovery sem `--help` gigante. Três fontes, um search.

**CTA (30s):** Sugere fonte nova nos comentários.

**Perguntas plateia:**
1. Fuzzy no CLI escala melhor que autocomplete?
2. Faltam recent-history + memory como source?
3. Deveria ranquear por uso?

---

## EP30 — harness diagnose + fix-env: env quebrado vira checklist

**Título (58 chars):** diagnose + fix-env: quando seu GOROOT te trai

**HOOK (15s):** "GOROOT apontando pra pasta que não existe. PATH sem homebrew. Projeto Python sem `.venv`. `harness fix-env --apply` resolve. Sem LLM."

**CONTEXTO (60s):** Dois comandos complementares. `harness diagnose` (`cmd_diagnose.go:15`) roda diagnosers (tools ausentes, tree suja, plan não pinado) e salva em `.harness/artifacts/diagnoses/`. `harness fix-env` (`cmd_fix_env.go:9`) probes shell + project por falhas específicas (GOROOT, PATH, .venv, node_modules, git user.email, brew PATH, Gemfile.lock). Sem `--apply`: checklist. Com `--apply`: executa safe fixes.

**CODE TOUR (6min):**
- `cmd_diagnose.go:15-44` — `diagnose` com `--tool` (default `git`).
- `cmd_diagnose.go:46-86` — `fix [problem-id]` aplica remedies.
- `cmd_fix_env.go:9-40` — `fix-env` com `--apply`.
- `internal/twoagent` — `DefaultDiagnosers`, `DefaultFixers`.
- `internal/app/fixenvcmd` — probes específicos.

**Snippet Go real (cmd_fix_env.go:29-38):**
```go
RunE: func(cmd *cobra.Command, _ []string) error {
    root, err := cwd()
    if err != nil { return err }
    _, err = fixenvcmd.Run(cmd.OutOrStdout(), fixenvcmd.Options{Root: root, Apply: apply})
    return err
},
```

**DEMO (90s):**
```bash
harness diagnose --tool git --tool go --tool python
harness fix-env             # dry-run com checklist
harness fix-env --apply     # aplica safe fixes
harness fix --all           # aplica remedies do diagnose
```

**RECAP (30s):** Detecta → propõe → aplica. Zero LLM. Dois níveis: env (shell) e project (tree).

**CTA (30s):** Comenta qual env-bug engoliu sua manhã na última semana.

**Perguntas plateia:**
1. `--apply` deveria exigir confirmação por item?
2. Diagnosers custom via plugin: sim ou complexidade?
3. Fixer que faz `npm install` merece flag separada?

**Referência paper:** Code as Agent Harness §3.5.1 (two-agent diagnosis/fix).
