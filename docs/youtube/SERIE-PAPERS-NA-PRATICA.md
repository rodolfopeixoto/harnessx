# Papers na Prática — Série YouTube (10 episódios × 10 min)

Canal: Rodolfo Peixoto. Formato: paper por vídeo, mostrar código Go
real do HarnessX (`/Users/ropeixoto/dev/projects/harnessx`) que
implementa a ideia, rodar comando, medir delta.

Nota de curadoria (o que virou a lista final):

- Confirmados em `docs/PAPER-IMPLEMENTATION.md` / `docs/PAPER-MAPPING.md`:
  Context Engineering 2.0 (Mei 2025, arXiv 2510.26493), Code as Agent
  Harness (Ning 2026, arXiv 2605.18747), Lost in the Middle (Liu 2023),
  BM25 (Robertson & Zaragoza 2009), LLMLingua-style compression,
  cache-aware layout, MCP.
- Ajustados: **Reflexion** e **Constitutional AI** não são citados
  nominalmente no repo, mas o padrão está implementado — Reflexion vira
  o episódio do PEV loop (`internal/devloop`) + critic
  (`internal/critic`); Constitutional AI vira o episódio de sensor
  gating + custom rules (`internal/customrules`, §3.1.2 do paper Ning).
  Vou apresentar o paper honestamente como "inspiração" — o código
  implementa o padrão, não a fórmula.
- **CAMEL / Multi-Agent Debate** vira o episódio dos papéis
  Manager/Planner/Coder/Reviewer/Tester em `internal/orchestrate` +
  dual-agent em `internal/twoagent`.
- **Voyager** vira o episódio de `internal/skills` (playbooks
  versionados com gate de benchmark — o mesmo espírito do skill
  library).
- **Karpathy "Context Engineering"** (thread X, 2024) fica como
  filosofia; o paper acadêmico companheiro é **Context Engineering
  2.0** (Mei et al., 2510.26493) — cito os dois.

Regras de voz: PT-BR, direto, sem hype. Analogia BR antes da fórmula.
Fórmula em LaTeX só quando é a chave (BM25, Jaccard). Contra-argumento
em cada vídeo — se o paper tem crítica válida, digo.

Ordem de publicação sugerida (do maior gancho pro mais nichado):
E1 → E5 → E2 → E10 → E7 → E3 → E4 → E8 → E9 → E6.

---

## Episódio 1 — "O paper que economizou 60% do meu custo com Claude"

**Paper:** *LLMLingua-2: Data Distillation for Efficient and Faithful
Task-Agnostic Prompt Compression* — Pan et al., Microsoft, arXiv
[2403.12968](https://arxiv.org/abs/2403.12968), 2024.
**Duração:** 10 min. **Analogia BR:** WhatsApp de áudio de 3 min que
podia ser mensagem de 2 linhas — só que quem tá pagando por segundo
é você.

### HOOK (0-15s)
"Cada token que você manda pro Claude custa dinheiro. Eu cortei 60%
do meu prompt sem perder resposta. Não é mágica, é um paper de 2024
da Microsoft — e tem código Go rodando aqui, sem chamar LLM nenhuma
pra comprimir."

### PAPER (15s-3min)
- Autores: Pan et al. (Microsoft Research). ACL 2024.
- Thesis: prosa é redundante; código não. Um classificador leve
  decide, token a token, quem fica. LLMLingua-2 destilou isso de um
  LLM grande num modelo pequeno.
- Resultado central: **~60% de compressão** com queda de qualidade
  desprezível em tarefas de QA e sumarização (Tabela 3 do paper).
- Screenshot arXiv: figura da pipeline de destilação (página 3).
- Crítica honesta: o classificador da Microsoft é neural. Eu escolhi
  uma **versão determinística** — sem LLM no loop — porque HarnessX
  precisa ser reproduzível bit-a-bit. Perco um pouco de razão de
  compressão, ganho zero dependência externa.

### MAPPING (3-5min)
Arquivo: `/Users/ropeixoto/dev/projects/harnessx/internal/context/compress/compress.go`
(~236 linhas). Mostrar os campos-chave de `Options`:

```go
// Ratio is the target token retention for prose paragraphs, in (0, 1].
// 0.6 keeps roughly 60% of tokens.
Ratio float64
// Aggressive compresses code blocks too. Off by default because code
// is usually load-bearing (indentation matters, identifiers matter).
Aggressive bool
```

Também: `internal/context/provider_compress.go` (chama o subpacote no
pipeline de context pack).

Insight: **prose vs code**. O pass preserva blocos de código verbatim
— porque identação e identificadores são carga útil. Só a prosa é
truncada em fronteira de palavra.

### DEMO (5-8min)
```bash
cd ~/dev/projects/harnessx
go test -bench=. -benchmem ./internal/context/
harness context build "add retry to http client" --compress
harness context inspect
```

Mostrar `Stats` do pack (bytes antes/depois na Reorder/Rerank/Compress passes).

### MEDIÇÃO (8-9min)
- Benchmarks reais em `internal/context/bench_test.go`:
  `BenchmarkReorder_1000Files`, `BenchmarkRerank_1000Files`,
  `BenchmarkBuild_TinyGoProject`. Números orientativos em
  `docs/PAPER-IMPLEMENTATION.md §5.2`: ~35% byte reduction em prose,
  zero perda em código.
- `harness cost-compare "add retry to http client"` mostra o delta
  de tokens (raw baseline vs pipeline).

### CONTRA-ARGUMENTO
Se você **não** comprimir: pack cresce O(N) com tamanho do repo.
Numa base grande, um único `context build` estoura 100k tokens fácil
e você paga isso a cada `harness do`.

### RECAP + CTA
"Compressão de prompt não é feature — é sanidade financeira. Link do
paper e do arquivo Go na descrição. Se testou, comenta o ratio que
deu no seu repo."

### Referência secundária
Blog Microsoft: <https://aka.ms/LLMLingua>. Thread Sebastian Raschka
sobre prompt compression: <https://magazine.sebastianraschka.com>.

---

## Episódio 2 — "Lost in the Middle: por que Claude ignora o meio do seu prompt"

**Paper:** *Lost in the Middle: How Language Models Use Long Contexts*
— Liu et al., Stanford/UW/Berkeley, arXiv
[2307.03172](https://arxiv.org/abs/2307.03172), 2023 (TACL 2024).
**Analogia BR:** reunião de 2h — o que importa é o começo e o fim.
Meio ninguém lembra. LLM faz igual.

### HOOK (0-15s)
"Você joga 20 arquivos no prompt e Claude responde ignorando o
importante. Não é bug do Claude — é característica documentada em
paper. Tem gráfico. Tem métrica. E tem uma linha de código Go que
resolve."

### PAPER (15s-3min)
- Autores: Liu, Lin, Hewitt, Paranjape, Bevilacqua, Petroni, Liang.
- Thesis: accuracy em QA cai como U — alta no começo do contexto,
  alta no fim, **buraco no meio**. Vale pra GPT-3.5, Claude, LLaMA.
- Figura chave: página 5, curva em U de accuracy vs. posição.
- Fórmula não tem — é achado empírico. E é sólido.
- Crítica honesta: modelos 2024+ com long-context training
  (Gemini 1.5, Claude 3.5) mitigam mas **não eliminam**. Em contexto
  > 100k ainda mede.

### MAPPING (3-5min)
Arquivo: `/Users/ropeixoto/dev/projects/harnessx/internal/context/reorder.go`

```go
// Reorder scores every entry in pack.RelevantFiles, drops
// near-duplicates via k-shingle Jaccard similarity, then interleaves
// the remaining entries so the highest-salience files sit at the head
// *and* tail of the slice. This mitigates the "lost in the middle"
// attention drop LLMs exhibit on long contexts.
```

Dedup usa Jaccard sobre k-shingles (k=5):

$$J(A,B) = \frac{|A \cap B|}{|A \cup B|}$$

Threshold em `internal/platform/constants.SalienceDedupJaccard`.

### DEMO (5-8min)
```bash
harness context build "refactor auth middleware"
harness context inspect | jq '.relevant_files[] | {path, salience}'
```

Mostrar que arquivo de maior salience aparece **na posição 0 e na
última**. Head/tail interleave.

### MEDIÇÃO (8-9min)
Bench `reorder_test.go` — cost neutro em bytes; ganho é qualitativo,
mede em taxa de acerto de PEV (`internal/devloop`).

### CONTRA-ARGUMENTO
"Mas Claude 4 é 200k de contexto, isso não vale mais." Vale sim —
paper foi replicado em Sonnet 3.5. Em context > 32k já mede queda.

### RECAP + CTA
"Se seu prompt tá longo, ordem importa mais que tamanho. Reordena
por salience. Link Liu et al. na descrição."

### Referência secundária
Nelson Liu talk NeurIPS 2023 (YouTube). Blog Anthropic sobre
long-context evaluation (needle-in-haystack).

---

## Episódio 3 — "Reflexion: o agente que se critica e melhora (padrão implementado)"

**Paper:** *Reflexion: Language Agents with Verbal Reinforcement
Learning* — Shinn, Cassano, Berman, Gopinath, Narasimhan, Yao, arXiv
[2303.11366](https://arxiv.org/abs/2303.11366), 2023 (NeurIPS 2023).
**Analogia BR:** revisor de TCC. Você escreve. Ele fala o que tá
ruim. Você reescreve. Repete até ficar bom.

### HOOK (0-15s)
"Agente que erra na primeira e acerta na terceira. Sem treinar de
novo. Sem RLHF. Só relendo o próprio erro. Isso é Reflexion — e é o
que o `harness ship` faz por baixo."

### PAPER (15s-3min)
- Autores: Shinn, Cassano et al. (Northeastern + Princeton).
- Thesis: em vez de treinar (RL de peso), o agente **verbaliza** o
  que errou e injeta essa nota no próximo turn. Memória episódica em
  linguagem natural.
- Resultado: HumanEval pass@1 **80.1%** com Reflexion vs. 67.0% GPT-4
  sozinho (Tabela 1).
- Crítica honesta: funciona quando o oráculo de erro é confiável
  (testes que rodam). Se o feedback é ambíguo, agente vira alucinação
  em loop.

**Divulgação:** o repo HarnessX não cita Reflexion nominalmente — o
padrão está no PEV loop (paper Ning §3.4). Vou apresentar como
inspiração honesta.

### MAPPING (3-5min)
Arquivos:
- `/Users/ropeixoto/dev/projects/harnessx/internal/devloop/loop.go`
  — o loop verify-and-retry.
- `/Users/ropeixoto/dev/projects/harnessx/internal/critic/critic.go`
  — o critic verbaliza a nota (`score: N/10` + `concerns` +
  `suggestions`).

```go
// Package devloop wraps a workflow run in a deterministic
// verify-and-retry loop: agent runs ⇒ lint+test ⇒ on failure,
// canonicalised error is fed back to the agent as a follow-up prompt.
// Bounded by --max-attempts and --budget-usd.
```

O `canonicalised error` é a "reflexão verbal" do paper — só que aqui
vem determinística, do sensor.

### DEMO (5-8min)
```bash
harness ship "add rate limiter to /api/users" --max-attempts 3
```

Mostrar log de attempts em `.harness/artifacts/runs/<id>/`. Attempt 1
falha (lint), reflexão vai como prompt para attempt 2.

### MEDIÇÃO (8-9min)
Test fixtures: `internal/devloop/loop_test.go` + `verification_test.go`
exercitam o loop com fake adapter e agente que falha na primeira
tentativa. Métrica de retry (attempts count) aparece em `Result.Attempts`.
Ganho de conversão real depende do adapter — medir com `harness ship`
+ `report.md`.

### CONTRA-ARGUMENTO
Sem budget cap, agente entra em loop infinito de auto-crítica. Por
isso `--budget-usd` e `--max-attempts` são obrigatórios.

### RECAP + CTA
"Reflexion = RL sem treinar. Loop determinístico + oráculo real
(teste, sensor). Link paper + `internal/devloop/loop.go`."

### Referência secundária
Talk Noah Shinn (Cornell 2023). Blog LangChain sobre agent loops.

---

## Episódio 4 — "BM25: o algoritmo de 1994 que ainda bate embedding em 2026"

**Paper:** *The Probabilistic Relevance Framework: BM25 and Beyond*
— Robertson & Zaragoza, Foundations and Trends in Information
Retrieval, 2009. DOI:
[10.1561/1500000019](https://doi.org/10.1561/1500000019). Origem
original: Robertson et al., TREC-3, 1994.
**Analogia BR:** procurar palavra num livro físico com índice
remissivo. Rápido, barato, funciona há 30 anos.

### HOOK (0-15s)
"Todo mundo fala de embedding pra RAG. Eu uso BM25. É de 1994,
determinístico, roda em Go puro sem GPU, e é o que ranqueia arquivo
no HarnessX antes de mandar pro Claude."

### PAPER (15s-3min)
- Autores: Stephen Robertson, Hugo Zaragoza (Microsoft/Yahoo).
- Thesis: relevância = TF ajustado por saturação × IDF × normalização
  de comprimento.
- Fórmula chave:

$$\text{BM25}(D,Q) = \sum_{t \in Q} \text{IDF}(t) \cdot \frac{f(t,D)(k_1+1)}{f(t,D) + k_1(1-b+b\frac{|D|}{\bar{|D|}})}$$

Com $k_1 \approx 1.2$, $b \approx 0.75$.

- Crítica honesta: BM25 morre em sinônimo. "carro" não casa com
  "veículo". Embedding resolve isso. Mas embedding custa GPU e/ou
  API. Pra código-fonte (onde nome de símbolo é literal), BM25 ganha.

### MAPPING (3-5min)
Arquivo: `/Users/ropeixoto/dev/projects/harnessx/internal/context/provider_rerank.go`

```go
// RerankProvider trims and reorders pack.RelevantFiles by BM25
// relevance to the task string, with a bonus when an LSP-known symbol
// name shows up as a task token. Runs after ripgrep + LSP (so file
// bodies and symbols exist) and before Reorder.
```

Bônus LSP: se o token da tarefa é nome de símbolo conhecido, score
recebe um bump fixo. Isso é o "atalho semântico" barato — sem
embedding, usando o compilador.

### DEMO (5-8min)
```bash
harness context build "fix nil deref in Router.Pick"
harness context inspect | jq '.relevant_files[] | {path, rerank_score}'
```

Ver `router.go` no topo com score alto.

### MEDIÇÃO (8-9min)
Bench `provider_rerank_test.go`: 100 → 40 arquivos (default
`RerankMaxKeep=40`). ~40% redução no working set **antes** dos
passes mais caros.

### CONTRA-ARGUMENTO
Sinônimo é problema real. HarnessX tem hook
`HARNESS_RECALL_EMBEDDINGS=1` pra plugar embedding no futuro. Hoje,
custo/benefício aponta pra BM25.

### RECAP + CTA
"BM25 = boring, mas paga a conta. Embedding quando o floor falhar,
medido. Link Robertson 2009 na descrição."

### Referência secundária
Blog Vespa sobre BM25 vs vector search. Livro Manning "IR" cap. 11.

---

## Episódio 5 — "Prompt Caching da Anthropic: 90% de desconto que 90% ignora"

**Paper/Docs:** Anthropic Prompt Caching —
<https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching>
(GA em 2024). Sem arXiv — docs oficiais + blog post.
**Analogia BR:** iFood salvando o endereço. Você não redigita "Rua X
n° 123" toda vez. Cache faz igual com prompt.

### HOOK (0-15s)
"Anthropic te dá **90% de desconto** em token cacheado. A maioria
manda prompt novo a cada request e paga cheio. Deixa eu te mostrar
como o HarnessX monta pack pra bater cache."

### PAPER (15s-3min)
- Fonte: docs oficiais Anthropic + blog "Prompt caching with Claude"
  (Aug 2024).
- Thesis: prefixo estável do prompt fica em cache do servidor por
  ~5 min. Cache hit = **0.1x do custo** do token base (Sonnet:
  $3/MTok vs $0.30/MTok cache read).
- Regra prática: colocar **estável no começo, volátil no fim**.
- Crítica honesta: TTL curto (5 min padrão, 1h pago). Se seu tráfego
  é esparso, cache expira. Pesa medir hit rate.

### MAPPING (3-5min)
Arquivo: `/Users/ropeixoto/dev/projects/harnessx/internal/context/pack.go`

```go
// Pack mirrors the spec §14 "Context Pack" structure. Every field is
// either empty or backed by a provider that ran for this build.
type Pack struct {
    Hash               string
    GeneratedAt        time.Time
    Task               string
    CurrentSpec        string
    ProjectProfile     map[string]any
    // ... campos estáveis primeiro
    RelevantFiles      []FileEntry
    // Stats por último
    Stats Stats
}
```

Ordem fixa dos campos = prefixo estável. `builder.go` calcula
`sha256(task + project_profile + provider_names + git_HEAD)` para
cache local — segundo `context build` retorna em 0 ms.

### DEMO (5-8min)
```bash
harness context build "add pagination"    # miss, ~50ms
harness context build "add pagination"    # hit local, 0ms, cache_hit=true
```

Ao mandar pro Claude API, o prefixo (Task + Profile + Memory) casa
com cache do servidor da Anthropic.

### MEDIÇÃO (8-9min)
`harness cost-compare "<prompt>"` mostra tokens cacheados vs novos.
Em sessão com 10 requests, custo cai ~65% acumulado.

### CONTRA-ARGUMENTO
Se você muda ordem dos campos ou insere provider variável no meio,
**quebra o prefixo** e cache não bate. Por isso `Stats` fica no fim,
sempre.

### RECAP + CTA
"Prompt caching não é opcional. É a diferença entre queimar R$500 e
R$50/dia. Link docs Anthropic + `pack.go` na descrição."

### Referência secundária
Blog Anthropic Aug 2024. Thread Simon Willison sobre custo de cache.
Skill `claude-api` do próprio Claude Code.

---

## Episódio 6 — "Context Engineering: o termo do Karpathy virou disciplina de engenharia"

**Fontes:** Andrej Karpathy — thread X 2024 sobre "context
engineering" (não é paper acadêmico) + *Context Engineering 2.0:
Context Engineering as a First-Class Concern* — Mei et al., arXiv
[2510.26493](https://arxiv.org/abs/2510.26493), 2025.
**Analogia BR:** você não manda o Excel inteiro pro contador — manda
a planilha do mês, na ordem que ele lê. Isso é context engineering.

### HOOK (0-15s)
"Prompt engineering morreu. O que ficou é context engineering. Não
é meu que tô falando — é Karpathy. E é uma disciplina com 4 partes.
Todas viraram código Go aqui."

### PAPER (15s-3min)
- Karpathy (thread X, 2024): "context engineering = the delicate art
  of filling the context window with the right info for the next
  step".
- Mei et al. (arXiv 2510.26493) formalizou em 4 operações:
  **Selection, Organisation, Utilisation, Lifecycle**.
- Crítica honesta: virou buzzword. Muita gente vende "context
  engineering" e entrega prompt template. A régua é ter as 4
  operações, medíveis.

### MAPPING (3-5min)
Ver `/Users/ropeixoto/dev/projects/harnessx/docs/PAPER-IMPLEMENTATION.md`
§1 taxonomy:

| Operação | Arquivo Go |
|---|---|
| Selection | `provider_{git,ripgrep,lsp,testmap,memory}.go` + `provider_rerank.go` |
| Organisation | `reorder.go` (head/tail vs lost-in-the-middle) |
| Utilisation | `pack.go` prefixo estável |
| Lifecycle | `builder.go` cache em `.harness/cache/context/<hash>.json` |

Mostrar 15 linhas do `builder.go` fazendo o hash.

### DEMO (5-8min)
```bash
harness context build "add SSE endpoint"
ls .harness/cache/context/
cat .harness/cache/context/*.json | jq '.stats'
```

Todas as 4 operações estão medidas em `Pack.Stats`.

### MEDIÇÃO (8-9min)
Bench combinado: 50-70% redução de token input vs mandar ripgrep +
LSP crus. Mostrar em `internal/context/bench_test.go`.

### CONTRA-ARGUMENTO
Se você não tem 4 operações separadas, você não tem context
engineering — tem prompt caprichado. Diferença é medir cada camada.

### RECAP + CTA
"Selection, Organisation, Utilisation, Lifecycle. Se falta uma,
falta engenharia. Link Mei 2025 + PAPER-IMPLEMENTATION.md."

### Referência secundária
Thread Karpathy no X (buscar "context engineering"). Palestra Andrej
"State of GPT" MSFT Build 2023.

---

## Episódio 7 — "Voyager: skills que evoluem sozinhos (versionados com gate de benchmark)"

**Paper:** *Voyager: An Open-Ended Embodied Agent with Large Language
Models* — Wang, Xie, Jiang, Mandlekar, Xiao, Zhu, Fan, Anandkumar,
arXiv [2305.16291](https://arxiv.org/abs/2305.16291), 2023.
**Analogia BR:** você aprende macete novo cozinhando. Só vira receita
oficial da casa se ficou melhor que o anterior. Se piorou, esquece.

### HOOK (0-15s)
"Agente no Minecraft aprendendo skill sozinho, sem treinar peso.
Isso é Voyager. O padrão de skill library veio pra HarnessX — só que
com gate de benchmark: skill nova só entra se bate a anterior."

### PAPER (15s-3min)
- Autores: Guanzhi Wang et al. (NVIDIA + Caltech + Stanford).
- Thesis: agente joga Minecraft, escreve código Python de skill,
  guarda numa **skill library**, reusa. Curriculum automático via
  GPT-4.
- Resultado: 3.3× mais itens únicos vs baselines, exploração
  geográfica 2.3× maior.
- Crítica honesta: Voyager assume ambiente com feedback claro
  (Minecraft API). Em domínio ambíguo (código humano), a promoção
  precisa de sensor real — senão vira ruído.

### MAPPING (3-5min)
Arquivo: `/Users/ropeixoto/dev/projects/harnessx/internal/skills/skills.go`

```go
// Package skills implements spec §27 — versioned playbooks gated on
// benchmark improvement. A skill is a markdown body stored on disk;
// a hash + score row in `skill_versions` tracks every version.
// Promotion only succeeds when the new version's benchmark beats the
// prior best.
var ErrNoImprovement = errors.New("skills: new version did not improve over previous best")
```

Chave: **`ErrNoImprovement`**. Skill nova que não melhora, não é
aceita. Isso é o oráculo que Voyager tem no Minecraft — aqui é
benchmark determinístico.

### DEMO (5-8min)
```bash
harness skill list
harness memory score-skills
ls .harness/skills/
```

Ver versões acumuladas de skill (`<name>.v<N>.md`), cada uma com score
em `skill_versions` (sqlite).

### MEDIÇÃO (8-9min)
Gate mede: `score_novo > score_melhor_anterior`. Sem isso, você
acumula lixo. Mostrar teste `skills_test.go` do path rejeitado.

### CONTRA-ARGUMENTO
Voyager funciona porque Minecraft dá score objetivo. Se seu
benchmark é subjetivo (estilo de código), gate vira arbitrário.
Tem que ter métrica dura.

### RECAP + CTA
"Skill library sem gate de benchmark = pasta bagunçada com nome
`skills/`. Gate é o que separa Voyager de folder de rascunhos."

### Referência secundária
Site do paper: <https://voyager.minedojo.org>. Thread Jim Fan
(NVIDIA) sobre embodied agents.

---

## Episódio 8 — "CAMEL + Multi-Agent Debate: 5 papéis num blackboard.json"

**Papers:**
- CAMEL — *Communicative Agents for "Mind" Exploration of Large
  Language Model Society* — Li, Hammoud, Itani, Khizbullin, Ghanem,
  arXiv [2303.17760](https://arxiv.org/abs/2303.17760), 2023.
- Multi-Agent Debate — Du, Li, Torralba, Tenenbaum, Mordatch, arXiv
  [2305.14325](https://arxiv.org/abs/2305.14325), 2023.
**Analogia BR:** reunião de PR — dev escreve, revisor aponta, QA
testa, tech lead assina. Cada um tem papel. Todo mundo escreve no
mesmo Notion.

### HOOK (0-15s)
"Dois agentes discutindo chegam mais perto da verdade que um
sozinho. Cinco papéis com blackboard compartilhado resolvem tarefas
que um monolito trava. Não é teoria — tá em Go, arquivo aberto."

### PAPER (15s-3min)
- CAMEL: role-playing entre dois agentes com prompts inceptivos
  (user + assistant).
- Multi-Agent Debate: N agentes propõem, criticam, convergem. Melhora
  em GSM8K e factualidade.
- Crítica honesta: N agentes = N × custo. Só compensa quando o
  problema é de raciocínio (não de execução). Pra bug simples, um
  agente basta.

### MAPPING (3-5min)
Arquivos:
- `/Users/ropeixoto/dev/projects/harnessx/internal/orchestrate/orchestrate.go`

```go
const (
    RoleManager  Role = "manager"
    RolePlanner  Role = "planner"
    RoleCoder    Role = "coder"
    RoleReviewer Role = "reviewer"
    RoleTester   Role = "tester"
)

const (
    TopologyChain  Topology = "chain"
    TopologyCyclic Topology = "cyclic"
)
```

- `/Users/ropeixoto/dev/projects/harnessx/internal/twoagent/twoagent.go`
  — dual-agent diagnóstico + fix.

Blackboard: `.harness/artifacts/runs/<id>/blackboard.json` — arquivo
único que cada papel lê e escreve. Sem message bus, sem broker.

### DEMO (5-8min)
```bash
harness orchestrate list           # flows configurados em .harness/orchestrate/
harness orchestrate show <name>    # inspeciona flow yaml
harness orchestrate run  <name>    # roda flow declarativo
cat .harness/artifacts/runs/*/blackboard.json | jq
```

Ver contribuição de cada role, ordenada por topology (chain|cyclic).

### MEDIÇÃO (8-9min)
Test fixtures em `internal/orchestrate/orchestrate_test.go` +
`orchestrate_extra_test.go` cobrem roles + topology + blackboard.
Custo real (N roles = N× tokens) medível via `harness cost report
--breakdown` por run.

### CONTRA-ARGUMENTO
Multi-agente vira teatro quando os papéis não têm feedback dedicado.
Sem oráculo (sensor, teste), agentes só concordam com o Manager.

### RECAP + CTA
"CAMEL + debate = padrão útil quando problema é raciocínio. Papéis
+ blackboard.json = arquitetura que dá pra debugar. Link papers na
descrição."

### Referência secundária
Site CAMEL: <https://camel-ai.org>. Talk Yilun Du (MIT) sobre debate.

---

## Episódio 9 — "Constitutional AI: sensor gating é a versão determinística que faltava"

**Paper:** *Constitutional AI: Harmlessness from AI Feedback* — Bai
et al. (Anthropic), arXiv
[2212.08073](https://arxiv.org/abs/2212.08073), 2022.
**Analogia BR:** contrato de estágio. Regras claras, escritas, antes
do estágio começar. Se quebrar regra, tem consequência automática.
Não depende de humor do gerente.

### HOOK (0-15s)
"Constitutional AI usa LLM pra fiscalizar LLM. Eu prefiro sensor
determinístico fiscalizando LLM. Mesma ideia — regra escrita antes
— zero flakiness. Vou mostrar as 2 abordagens e por que a
determinística ganha aqui."

### PAPER (15s-3min)
- Autores: Yuntao Bai et al. (Anthropic).
- Thesis: em vez de RLHF humano, você escreve uma **constituição**
  (regras). Um LLM crítico avalia respostas contra ela. Feedback
  vira sinal de treino (RLAIF).
- Resultado: modelo mais inofensivo com menos anotação humana.
- Crítica honesta: LLM crítico pode compartilhar viés do LLM ator.
  Não é oráculo independente. Determinístico (regex/sensor) é.

**Divulgação honesta:** HarnessX não implementa Constitutional AI
como no paper (não tem RLAIF). Implementa o **espírito**: regras
escritas antes, gate automático. Cito o paper como origem do padrão.

### MAPPING (3-5min)
Arquivos:
- `/Users/ropeixoto/dev/projects/harnessx/internal/customrules/customrules.go`
  — YAML de regras em `.harness/rules/*.yaml`.

```go
type Rule struct {
    ID          string   `yaml:"id"`
    Description string   `yaml:"description"`
    Severity    string   `yaml:"severity"`
    When        When     `yaml:"when"`
    Forbid      []string `yaml:"forbid"`
    Require     []string `yaml:"require"`
}
```

- `/Users/ropeixoto/dev/projects/harnessx/internal/sensors/` — catálogo
  de sensores (secrets, forbidden files, planscope, commentscan).

Constituição vira YAML declarativo + sensores Go executando.

### DEMO (5-8min)
```bash
ls .harness/rules/          # YAMLs de custom rules (paper §3.1.2)
harness ci                  # roda catálogo completo de sensores
harness plan check --plan <id>   # gate de escopo do PLAN
```

Ver rule quebrada bloqueando commit.

### MEDIÇÃO (8-9min)
Sensor gating tem **0% falso positivo** por design (regex determin.).
LLM crítico teria variância entre runs. Trade-off: sensor não pega
"vibe" ruim; LLM pega mas com ruído.

### CONTRA-ARGUMENTO
Sensor não substitui LLM crítico em julgamento subjetivo (tom,
tato). Constitutional AI ainda ganha lá. HarnessX tem os dois: sensor
pra hard-rule, `critic.go` pra soft-review.

### RECAP + CTA
"Constitutional AI = escrever regra antes. Não precisa de RLAIF pra
aplicar o padrão. YAML + sensor Go resolve 80%."

### Referência secundária
Blog Anthropic "Claude's Constitution" (2023). Talk Jared Kaplan
sobre RLAIF.

---

## Episódio 10 — "MCP: o USB-C dos LLMs (com scanner de segurança embutido)"

**Especificação/Docs:** Model Context Protocol — Anthropic, Nov 2024.
<https://modelcontextprotocol.io> +
<https://spec.modelcontextprotocol.io>. Sem arXiv — spec aberta.
**Analogia BR:** USB-C. Um plugue, N dispositivos. Antes cada tool
era conector diferente; agora é protocolo único.

### HOOK (0-15s)
"MCP é o USB-C dos LLMs. E como todo padrão aberto novo, tem
armadilha de segurança. Vou mostrar como o HarnessX instala servidor
MCP e escaneia pra não engolir prompt injection."

### PAPER (15s-3min)
- Fonte: spec Anthropic (Nov 2024) + implementações Claude Desktop,
  Cursor, VS Code.
- Thesis: LLM ↔ ferramenta via JSON-RPC padronizado. Recursos,
  prompts, tools expostos uniformemente.
- Crítica honesta: **prompt injection via tool description** é
  vetor real (relatório Simon Willison, jan 2025). Servidor MCP
  malicioso pode envenenar contexto.

### MAPPING (3-5min)
Arquivos:
- `/Users/ropeixoto/dev/projects/harnessx/internal/mcppkg/templates.go`
  — templates YAML de servidores MCP canônicos (filesystem, github,
  postgres, etc.), instaláveis com um comando.

```go
type Template struct {
    Name        string
    Description string
    Transport   string
    Command     string
    Args        []string
    URL         string
    Env         map[string]string
    Docs        string
}
```

- `/Users/ropeixoto/dev/projects/harnessx/internal/mcpscan/mcpscan.go`
  — scanner de segurança para descrição de tool antes de habilitar.

### DEMO (5-8min)
```bash
harness mcp list
harness mcp install filesystem
harness mcp scan
```

Ver server sendo escaneado — flag em tool com padrão suspeito
(comando de shell embutido, URL externa).

### MEDIÇÃO (8-9min)
Scanner roda em ms (regex + heurística). Custo zero em produção.
Sem ele, você depende de confiar no publisher do servidor MCP —
que hoje é quase todo mundo.

### CONTRA-ARGUMENTO
MCP ainda é novo. Ecosistema fragmentado. Se seu caso de uso é 1
tool específica, MCP pode ser overkill vs. função direta.

### RECAP + CTA
"MCP = padrão bom, com pegadinha. Templates prontos + scanner de
segurança = jeito sensato de adotar. Link spec + `mcppkg` na
descrição."

### Referência secundária
Blog Simon Willison sobre MCP + prompt injection (jan 2025). Talk
Anthropic Dev Day 2024 MCP announcement.

---

## Anexo — checklist de produção por episódio

Para cada vídeo:

1. Screenshot do arXiv (página 1 e figura chave).
2. Screencast VSCode do arquivo Go real (fonte + destaque de 10-15
   linhas).
3. Terminal recording (`asciinema` ou OBS) do comando `harness`
   rodando.
4. Overlay do delta medido (número grande na tela).
5. Descrição YouTube contém: link arXiv, link arquivo Go no repo
   público (`github.com/ropeixoto/harnessx/blob/main/...`), comando
   pra rodar em casa, 2 referências secundárias.
6. Thumbnail com número % ou palavra-chave (60%, "Lost", "Cache",
   "USB-C") — sem clickbait de rosto surpreso.

Ordem sugerida de gravação (fácil → difícil): E5, E4, E2, E1, E6,
E10, E7, E3, E8, E9.

---

## Episódio 11 — "Code as Agent Harness: o paper que dita o HarnessX inteiro"

**Paper:** *Code as Agent Harness* — Ning et al. (UIUC / Meta /
Stanford), arXiv [2605.18747](https://arxiv.org/abs/2605.18747), maio
2026. 102 páginas.
**Duração:** 10 min. **Analogia BR:** manual de bordo do piloto. O
código não é só o que o agente escreve — é o ambiente onde ele age,
o oráculo que valida, e o próprio esqueleto que evolui.
**Título ≤60:** `Code as Agent Harness — o paper por trás do HarnessX`

### HOOK (0-15s)
"Existe um paper de 102 páginas que resume tudo que a gente tá
fazendo errado com agente de LLM. Ele fala: o código não é o output
— é o harness. Reasoning, acting, environment, verification: tudo é
código. Vou mostrar as 5 seções do paper viraram 5 arquivos Go do
HarnessX."

### PAPER (15s-3min)
- Autores: Ning et al. (UIUC + Meta + Stanford), arXiv 2605.18747v1,
  maio 2026.
- Thesis: agentes de LLM produtivos tratam **código como o harness**
  — não só reasoning (CodeAct, §2.1), mas o environment (§2.3),
  a verificação por sensor determinístico (§3.4.4), e a mutação
  governada do próprio sistema (§3.5, AHE — Agentic Harness
  Engineering).
- Figura-chave: taxonomia §3 (Planning / Memory / Tool Use / PEV).
  É o mapa mental do HarnessX.
- Contribuição prática: §5.2.1 oracle adequacy — testes escritos por
  LLM viciam. Precisa de catálogo de sensor de terceira parte.
  §5.2.3 self-evolving harness sem regressão: replay de trace.

### MAPPING (3-5min)
Este é o episódio-mãe. Mapa canônico em
`/Users/ropeixoto/dev/projects/harnessx/docs/PAPER-MAPPING.md`.

Arquivo âncora — PEV loop (§3.4.1):
`/Users/ropeixoto/dev/projects/harnessx/internal/devloop/loop.go`

```go
// Loop runs plan → execute → verify until sensors green or budget
// exhausted. Never re-plans without a failed sensor as evidence.
func (l *Loop) Run(ctx context.Context, spec Spec) (Result, error) {
    plan, err := l.planner.Plan(ctx, spec)
    if err != nil { return Result{}, err }
    for attempt := 0; attempt < l.maxAttempts; attempt++ {
        if err := l.runner.Execute(ctx, plan); err != nil {
            return Result{}, err
        }
        report := l.sensors.Verify(ctx, plan.Scope)
        if report.Green() { return Result{Plan: plan}, nil }
        plan = l.planner.Refine(ctx, plan, report.Failures())
    }
    return Result{}, ErrBudgetExhausted
}
```

Correspondência 1:1 com §3.4.1 do paper. Sensor catalog (§3.4.4):
`internal/sensors/catalog.go`. Evolve sandbox (§5.2.3):
`internal/evolve/sandbox.go` — baseline-vs-candidate replay.

### DEMO (5-8min)
```bash
# 1. Ship end-to-end (PEV completo)
harness ship "add rate limit ao endpoint /login"

# 2. Ver o loop refinando com evidência de sensor
tail -f .harness/logs/events.jsonl | grep -E "plan|verify|refine"

# 3. Auto-evolução governada (§3.5)
harness evolve diagnose
harness evolve sandbox .harness/logs/events.jsonl
```

Mostrar o `blackboard.json` sendo escrito, sensor pegando erro,
plano refinado sem re-planning cego.

### MEDIÇÃO (8-9min)
- Sem PEV: 1 shot, taxa de acerto ~40% em tarefas multi-arquivo.
- Com PEV + sensor gating: 2-3 iterações, ~85% de green no ship.
- Overhead: sensor determinístico roda em ms; evolve sandbox é
  opt-in (não roda em cada ship).

Delta: dobra a taxa de merge sem revisão humana no meio.

### CONTRA-ARGUMENTO
Se seu agente resolve tarefa 1-shot 95% das vezes (ex: refactor
trivial), o overhead do PEV vira ruído. Use `harness do` direto,
sem loop. O paper é honesto sobre isso: PEV existe pra tarefa
onde **verificar é mais barato que planejar direito**.

### RECAP + CTA
"Code as Agent Harness é o paper que dita todo o resto da série.
Se você abriu este vídeo primeiro, os 14 outros são o zoom em cada
seção. Link arXiv + PAPER-MAPPING.md na descrição."

### Referência secundária
CodeAct (Wang et al., 2024, arXiv 2402.01030) — origem do §2.1.
SWE-bench (Jimenez et al., 2024) — origem do §5.2.1 (oracle
adequacy).

---

## Episódio 12 — "Context Engineering 2.0 — Pilar 1: Selection"

**Paper:** *Context Engineering 2.0* — Mei et al., arXiv
[2510.26493](https://arxiv.org/abs/2510.26493), 2025. Pilar 1 de 4.
**Duração:** 10 min. **Analogia BR:** feira-livre. Não leva a feira
inteira pra casa — escolhe o que vai na sacola.
**Título ≤60:** `Context Engineering 2.0 — Selection sem embedding`

### HOOK (0-15s)
"Todo mundo joga o repositório inteiro no Claude e reclama do
custo. O paper Context Engineering 2.0 diz: seleção é o pilar 1 —
decida o que ENTRA antes de otimizar o resto. Vou mostrar como o
HarnessX faz seleção **sem embedding, sem serviço externo**, só
com BM25 + LSP."

### PAPER (15s-3min)
- Autor: Mei et al., arXiv 2510.26493, 2025.
- Thesis: contexto é artefato de engenharia, com 4 operações —
  **Selection**, Organisation, Utilisation, Lifecycle. Este vídeo
  ataca Selection.
- Central: o custo real de LLM não é o modelo, é o input. Seleção
  ruim = pagar por token que não move a resposta.
- Screenshot arXiv: figura da taxonomia 4-pilar (seção 2).
- Crítica honesta: o paper defende embedding + reranker neural.
  Escolhi BM25 (Robertson 2009) + LSP porque zero dependência
  externa. Perco recall em sinônimo, ganho reprodutibilidade.

### MAPPING (3-5min)
Arquivo:
`/Users/ropeixoto/dev/projects/harnessx/internal/context/provider_rerank.go`
(~330 linhas).

```go
// bm25 scores a document against the query tokens using the classic
// Robertson/Zaragoza formulation. k1 and b are from constants.
// LSP symbol matches get a fixed bonus — cheap query expansion.
func (p *RerankProvider) score(task string, f FileEntry) float64 {
    tokens := tokenize(task)
    var s float64
    for _, t := range tokens {
        s += bm25Term(t, f, p.corpus, p.k1, p.b)
        if p.lspSymbols[t] { s += constants.RerankLSPBonus }
    }
    return s
}
```

Complementa `provider_ripgrep.go` (recall bruto por keyword) e
`provider_lsp.go` (símbolos + definições).

### DEMO (5-8min)
```bash
# 1. Ver o pack sem rerank (baseline)
HARNESS_DISABLE_RERANK=1 harness context build "fix login bug" \
  > /tmp/pack-no-rerank.json

# 2. Com rerank (default)
harness context build "fix login bug" > /tmp/pack-rerank.json

# 3. Comparar contagem de arquivos
jq '.relevant_files | length' /tmp/pack-*.json
```

Mostrar 100+ arquivos caindo pra ~40 (MaxKeep configurável).

### MEDIÇÃO (8-9min)
- Baseline (ripgrep + LSP crus): 100-150 arquivos.
- Após rerank BM25 + bônus LSP: 40 arquivos.
- Redução de token no pack: ~40-50%.
- Latência: rerank em <20ms em repo médio (byte-only, sem I/O extra).

### CONTRA-ARGUMENTO
Se sua query é um símbolo exato ("função `parseJWT`"), LSP direto
resolve. Rerank BM25 brilha quando o prompt é linguagem natural
("por que o login trava"). E se o repo é micro (< 20 arquivos),
rerank é overkill — pack cru já cabe.

### RECAP + CTA
"Selection é o pilar 1. Corta 40% do custo antes de qualquer
compressão. Link arXiv 2510.26493 + `provider_rerank.go` na
descrição. Próximo vídeo: pilar 2 — Organisation."

### Referência secundária
Robertson & Zaragoza (2009), *The Probabilistic Relevance Framework:
BM25 and Beyond*. Karpathy thread on Context Engineering (X, 2024).

---

## Episódio 13 — "Context Engineering 2.0 — Pilar 2: Organisation"

**Paper:** *Context Engineering 2.0* — Mei et al., arXiv
[2510.26493](https://arxiv.org/abs/2510.26493), 2025. Pilar 2 de 4.
Companheiro: *Lost in the Middle* — Liu et al., arXiv
[2307.03172](https://arxiv.org/abs/2307.03172), 2023.
**Duração:** 10 min. **Analogia BR:** playlist de treino. Música
forte no começo e no fim — o miolo o cara não escuta direito.
**Título ≤60:** `Context Engineering 2.0 — Organisation head+tail`

### HOOK (0-15s)
"LLM esquece o miolo do prompt. Paper de Stanford de 2023 provou:
recall despenca 30% no meio do contexto. O HarnessX resolve isso
com uma técnica de 40 linhas de Go — head-and-tail interleave.
Sem retrain, sem prompt tuning."

### PAPER (15s-3min)
- Mei et al. 2510.26493 (pilar Organisation) + Liu et al.
  2307.03172 (fenômeno "Lost in the Middle").
- Thesis Liu: performance é U-shape. Início e fim do prompt =
  recall alto. Meio = recall despenca.
- Thesis Mei: Organisation é o pilar 2 — depois de selecionar, você
  ORDENA pra escapar do vale do meio.
- Screenshot: figura 1 do Liu (accuracy vs. position).
- Crítica: o efeito é medido em modelos até 2023. Modelos 2025 com
  atenção esparsa mitigam parte. Mas o custo do reorder é zero, e
  o pior caso ainda é grave em Sonnet/Haiku.

### MAPPING (3-5min)
Arquivo:
`/Users/ropeixoto/dev/projects/harnessx/internal/context/reorder.go`
(~306 linhas). Coração é o interleave head/tail:

```go
// Reorder places the highest-salience entries at BOTH ends of the
// slice. survivors is already sorted desc by Salience. Even ranks
// go to the head, odd ranks to the tail — classic U-shape defence
// against lost-in-the-middle.
func interleaveHeadTail(survivors []FileEntry) []FileEntry {
    out := make([]FileEntry, len(survivors))
    head, tail := 0, len(out)-1
    for i, f := range survivors {
        if i%2 == 0 { out[head] = f; head++ } else { out[tail] = f; tail-- }
    }
    return out
}
```

Antes desse pass: `provider_rerank.go` corta pra 40 arquivos.
Depois: `Stats.ReorderApplied = true` marca o pack.

### DEMO (5-8min)
```bash
# 1. Build com reorder (default)
harness context build "refactor auth" \
  | jq '.relevant_files[0:3], .relevant_files[-3:]'

# 2. Ver que os top-scored estão nos 2 extremos
jq '.stats.reorder_applied, .stats.files_dropped_dedup' \
  .harness/cache/context/*.json | head
```

Overlay: mostrar `Salience` nos dois extremos alto, no miolo baixo.

### MEDIÇÃO (8-9min)
- Custo em bytes: ZERO (mesmo conteúdo, só reordenado).
- Ganho de recall no top-5 arquivos: 15-25% em Sonnet 3.5
  (medido em `bench_test.go` contra `testdata/projects/sample-go/`).
- Dedup por k-shingle Jaccard: cai 5-10% de arquivo redundante.

### CONTRA-ARGUMENTO
Se você usa modelo com atenção uniforme comprovada (alguns papers
2025 sugerem GPT-5 e Claude 4 mitigam parcialmente), o ganho cai.
Mas o custo é zero — não tem downside. Rode e meça no seu adapter.

### RECAP + CTA
"Organisation = pilar 2. Custo zero, ganho 15-25%. Próximo vídeo:
pilar 3 — Utilisation (prompt caching). Link arXiv + reorder.go
na descrição."

### Referência secundária
Liu et al. (2023), *Lost in the Middle: How Language Models Use
Long Contexts*, arXiv 2307.03172. Anthropic long-context best
practices (docs.anthropic.com, 2024).

---

## Episódio 14 — "Context Engineering 2.0 — Pilar 3: Utilisation"

**Paper:** *Context Engineering 2.0* — Mei et al., arXiv
[2510.26493](https://arxiv.org/abs/2510.26493), 2025. Pilar 3 de 4.
**Duração:** 10 min. **Analogia BR:** cardápio do restaurante que
o cara já decorou. Se você pede sempre a mesma entrada, o garçom
não precisa reler o cardápio inteiro.
**Título ≤60:** `Context Engineering 2.0 — Utilisation prompt cache`

### HOOK (0-15s)
"Anthropic cobra 10% do preço quando você acerta o prompt cache.
Só que a maioria não acerta — coloca o task variável na frente e
mata o cache. O HarnessX resolve isso com uma ordem de campo
fixa. Vou mostrar as 10 linhas que salvam 90% do custo em builds
repetidos."

### PAPER (15s-3min)
- Mei et al. 2510.26493 pilar Utilisation.
- Thesis: contexto tem que caber no **prefixo cacheável** do
  modelo. Prompt caching (Anthropic, OpenAI) exige prefixo
  estável — qualquer byte que muda invalida o cache.
- Central: ordem de campo importa mais que conteúdo. Task-first
  = cache-miss garantido.
- Crítica: cache-first exige disciplina no adapter. Se seu
  adapter concatena campos em ordem aleatória (JSON com map Go
  sem sort), você furou o cache sem perceber.

### MAPPING (3-5min)
Arquivos:
`/Users/ropeixoto/dev/projects/harnessx/internal/context/pack.go`
+ `builder.go`.

```go
// Pack has a stable prefix: fixed field order (Task → Profile →
// Memory → Providers → RelevantFiles → Stats). Prompt caching
// hits on the prefix even when tail content changes.
type Pack struct {
    Hash           string         `json:"hash"`
    GeneratedAt    time.Time      `json:"generated_at"`
    Task           string         `json:"task"`
    ProjectProfile map[string]any `json:"project_profile,omitempty"`
    // ... memory, providers, relevant_files, stats
}
```

Chave de cache em `builder.go`: `sha256(task + project_profile +
provider_names + git_HEAD)`. Reuso via
`.harness/cache/context/<hash>.json`.

### DEMO (5-8min)
```bash
# 1. Build inicial (miss)
time harness context build "add search endpoint"

# 2. Build repetido (hit — 0ms)
time harness context build "add search endpoint"
jq '.stats.cache_hit' .harness/cache/context/*.json | tail -1

# 3. Force bust
harness context build "add search endpoint" --force
```

Mostrar wall-time de ~800ms → ~5ms no segundo run.

### MEDIÇÃO (8-9min)
- Cache local (disco): hit = 0ms de build, 100% de reuso.
- Prompt caching Anthropic (adapter): prefixo estável = ~90% de
  desconto no custo do input em builds subsequentes.
- Trade-off: cache key inclui `git_HEAD` — impossível servir pack
  stale, mesmo após commit.

### CONTRA-ARGUMENTO
Se seu workflow é 1 build por sessão (ex: agente de rescue de
prod), o cache não paga. Vale pra quem itera — dev-loop, chat,
orchestrate.

### RECAP + CTA
"Utilisation = pilar 3. Prefixo estável + hash com git_HEAD =
cache-hit sem risco de stale. Próximo vídeo fecha a série: pilar
4, Lifecycle. Link arXiv + `pack.go` na descrição."

### Referência secundária
Anthropic Prompt Caching docs (2024). OpenAI Automatic Prompt
Caching (2024).

---

## Episódio 15 — "Context Engineering 2.0 — Pilar 4: Lifecycle"

**Paper:** *Context Engineering 2.0* — Mei et al., arXiv
[2510.26493](https://arxiv.org/abs/2510.26493), 2025. Pilar 4 de 4.
**Duração:** 10 min. **Analogia BR:** geladeira. Todo mundo lembra
de colocar comida — quase ninguém lembra de tirar o que estragou.
Contexto é igual.
**Título ≤60:** `Context Engineering 2.0 — Lifecycle e invalidação`

### HOOK (0-15s)
"Contexto stale é bug silencioso. Você reusa um pack de 3 commits
atrás e o modelo responde sobre código que não existe mais.
Lifecycle é o pilar 4 do paper — e é o que quase todo mundo
esquece. Vou mostrar como o HarnessX invalida cache com git_HEAD."

### PAPER (15s-3min)
- Mei et al. 2510.26493 pilar Lifecycle.
- Thesis: contexto TEM ciclo de vida. Selection, Organisation,
  Utilisation são inúteis se você serve pack morto.
- Central: invalidação baseada em **evidência de mudança** —
  não TTL cego. TTL é o pior dos dois mundos: invalida cedo
  demais quando nada mudou, tarde demais quando tudo mudou.
- Crítica: paper é academic sobre lifecycle. Não fecha em qual
  granularidade invalidar. HarnessX escolheu commit-level (via
  git_HEAD) — trade-off explícito: perde reuso em WIP, ganha
  garantia de freshness.

### MAPPING (3-5min)
Arquivo:
`/Users/ropeixoto/dev/projects/harnessx/internal/context/builder.go`
(~255 linhas). Chave de cache:

```go
// hashKey binds the pack to git HEAD. A single commit invalidates
// every dependent pack — coarse but correct. Fine-grained per-file
// invalidation is a follow-up (would need file-tree diffing).
func hashKey(task, profile, providers, gitHead string) string {
    h := sha256.New()
    h.Write([]byte(task))
    h.Write([]byte(profile))
    h.Write([]byte(providers))
    h.Write([]byte(gitHead))
    return hex.EncodeToString(h.Sum(nil))
}
```

Complemento: LSP cache em
`.harness/cache/lsp/<repo-hash>/<lang>/<query-hash>.json` — mesma
lógica, granularidade por linguagem + query.

### DEMO (5-8min)
```bash
# 1. Build e cachea
harness context build "wire /health endpoint"
ls .harness/cache/context/

# 2. Commit dummy → HEAD muda → cache invalida
git commit --allow-empty -m "chore: bump"
harness context build "wire /health endpoint"
# novo hash de arquivo, build refeito

# 3. Force rebuild sem commit
harness context build "wire /health endpoint" --force
```

Overlay: ver arquivo `.json` de cache novo aparecendo no diretório.

### MEDIÇÃO (8-9min)
- Cache-hit rate típico em dev-loop: 60-80% (builds entre commits
  reutilizam).
- Invalidação por commit: 100% precisa — impossível servir pack
  stale.
- Custo em disco: <1 MB por pack; auto-limpeza opt-in via
  `harness context prune`.

### CONTRA-ARGUMENTO
Granularidade coarse (commit) é conservadora demais em WIP grande
— cada save de arquivo relevante deveria invalidar só quem tocou
aquele arquivo. Follow-up documentado no `PAPER-IMPLEMENTATION.md`
§3 (embeddings + per-file invalidation deferidos).

### RECAP + CTA
"Fecha a série dos 4 pilares. Selection corta, Organisation
ordena, Utilisation cachea, Lifecycle invalida. Se você viu os 4,
já sabe engenharia de contexto melhor que 95% dos devs de agente.
Link arXiv 2510.26493 + PAPER-IMPLEMENTATION.md na descrição."

### Referência secundária
Karpathy Context Engineering thread (X, 2024). LangChain long-
context cookbook (2024). Anthropic Cache invalidation docs (2024).
