# SHORTS-FACTORY.md — Fábrica de Shorts (Rodolfo Peixoto / harnessx)

> Agent 4. Deriva material curto (45–60s) dos 15 vídeos âncora da série
> TaskHive descritos em `docs/youtube/ROTEIROS-TECNICOS.md`.
> Base didática: `docs/youtube/DIDATICA-CONCEITOS.md`.
> Base algoritmo: `docs/youtube/SEO-RETENCAO.md`.
> Voz Rodolfo Peixoto — PT-BR, direto, técnico, sem gritaria, sem
> clickbait vulgar. CPM leitura BR ~140–160 palavras/min (ref: locução
> comercial BR 2024, `[VERIFICAR — URL exata]`), então cada Short
> comporta ~110–160 palavras faladas.

## Fontes algoritmo 2026 (obrigatório)

Toda afirmação sobre YouTube Shorts vem de uma destas ou fica
marcada `[VERIFICAR]`. Nada de chute.

- YouTube Creator Insider — https://www.youtube.com/@YouTubeCreatorInsider
- YouTube Creator Liaison (Rene Ritchie) — Shorts algo talks 2025–2026
- YouTube Help "Shorts eligibility & performance":
  https://support.google.com/youtube/answer/12948449
- YouTube blog "Shorts monetization & discovery 2026" — `[VERIFICAR]`
- TubeBuddy Shorts Trends 2026 — `[VERIFICAR — URL exata]`
- vidIQ Shorts Report 2026 — `[VERIFICAR — URL exata]`
- Paper Guo/Kim/Rubin (L@S 2014) sobre retenção em vídeo educacional:
  https://dl.acm.org/doi/10.1145/2556325.2566239 — replicado em Shorts
  por Hosseini et al. 2023 (short-form retention).

---

## 1. Fórmulas Shorts 2026

Cinco moldes, cada um com timeline em segundos. Toda fala foi medida
pra caber a 150 wpm (2,5 palavras/segundo). Ultrapassar isso = corta.

### F1 — HOOK-PROBLEMA-PAYOFF-CTA

Origem: template clássico "problem-agitate-solve" adaptado a Shorts
pela YouTube Creator Academy 2025 (`[VERIFICAR — URL exata]`).

- **0–3s HOOK**: afirmação técnica polêmica ou número surpreendente.
  Objetivo único: passar dos 3 primeiros segundos (retention critical
  gate — CI 2024 confirmou 3s como principal drop-off em Shorts,
  `[VERIFICAR]`).
- **3–15s PROBLEMA**: o que dava errado antes.
- **15–45s PAYOFF**: a solução técnica, um bloco de código na tela,
  um comando rodando, um número.
- **45–60s CTA**: link bio + comentário-gatilho.

### F2 — LOOP (fim liga no começo)

Origem: pattern popular em TikTok trasplantado ao Shorts, referenciado
pelo Creator Liaison em 2025 (`[VERIFICAR — URL exata]`). Aumenta
view-through porque o replay é contado como nova view depois de 1
loop completo.

- Última frase = primeira frase, com cadência que "casa" no corte.
- Ex.: "…e é por isso que eu não uso Codex pra JWT." → volta em "Eu
  não uso Codex pra JWT. Vou provar em 60s."

### F3 — COMPARATIVO (X vs Y em 45s)

Fonte: análise vidIQ 2025 sobre "vs shorts" tendo CTR médio 18% maior
que shorts genéricos (`[VERIFICAR — URL exata]`).

- **0–5s**: "X vs Y — quem ganha?"
- **5–20s**: critério 1 (custo).
- **20–35s**: critério 2 (qualidade).
- **35–50s**: veredito com número.
- **50–60s**: CTA.

### F4 — "COMO EU FIZ" (bastidor 60s)

- **0–3s**: resultado final ("fiz um SaaS por US$ 1,23").
- **3–45s**: 3 passos, cada um ~13s, com screen recording acelerado
  1,25x.
- **45–60s**: convite pro vídeo completo.

Fonte: formato "process reveal" — Buffer Social Report 2025 aponta
como um dos 3 formatos mais salvos no Shorts (`[VERIFICAR]`).

### F5 — MITO OU VERDADE

- **0–4s**: afirmação polêmica ("LLM não faz JWT direito").
- **4–20s**: por que muita gente acredita nisso.
- **20–50s**: prova prática (comando, código, número).
- **50–60s**: veredito + CTA.

Fonte: "myth/reality" hooks consistentemente aparecem no top 10% de
Shorts educacionais em 2025 segundo TubeBuddy (`[VERIFICAR`]).

---

## 2. Catálogo de 60 Shorts (4 por vídeo × 15 vídeos âncora)

Convenções:
- **VL**: vídeo longo pai (ID do capítulo, ver `ROTEIROS-TECNICOS.md`).
- **Falas** = texto queimado em locução Rodolfo, PT-BR direto.
- **Legenda-caps** = caption on-screen, ≤6 palavras por frame.
- **Trilha** = categoria, não faixa específica (uso YouTube Audio
  Library — "no attribution required" — categoria tech/lo-fi 90–110
  BPM, evita copyright strike).
- **Hashtags** ≤5.

CTAs recorrentes (rotacionar pra não cansar):
- `CTA-A`: "Vídeo completo tá no primeiro comentário."
- `CTA-B`: "Comenta 'harness' que mando o repositório."
- `CTA-C`: "Link do tutorial na bio."
- `CTA-D`: "Salva pra não perder na hora de codar."

---

### VL1 — `harness init em 5 min: vault, 8 índices, doctor`

#### Short 1.1 — "Init em 5 min"
- **Título**: `harness init em 5 minutos`
- **Fórmula**: F1
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Cinco minutos. Dois comandos. Oito mapas de projeto prontos."
  - 3–12s: "Antes do harness eu perdia meia hora configurando
    diretório, guardrails e sensor de segredo manual."
  - 12–20s: "Rodo `harness init` e depois `harness project index`."
  - 20–35s: "Skeleton `.harness/`, oito mapas JSON (profile, commands,
    dependencies, architecture, test-map, api-map, design-system,
    performance-budget). Doctor passa verde."
  - 35–50s: "Zero token gasto no scaffold. É determinístico, não LLM."
  - 50–55s: "Comenta 'harness' que mando o link do tutorial."
- **Visual**: terminal rodando `harness init` + `harness project index`
  em tela cheia, `harness doctor` verde no fim.
- **Legenda-caps**: "5 MIN", "8 MAPAS", "ZERO TOKEN", "DOCTOR OK".
- **Trilha**: lo-fi tech 95 BPM.
- **VL pai**: V1.
- **CTA**: CTA-B.
- **Hashtags**: #golang #devbr #ai #cli #harness

#### Short 1.2 — "Doctor achou o segredo"
- **Título**: `Sensor achou meu token no git`
- **Fórmula**: F5 (mito/verdade)
- **Duração**: 50s
- **Falas**:
  - 0–4s: "Você acha que seu .env tá seguro. Não tá."
  - 4–15s: "O sensor `secrets_scan` do harness varre `secretPatterns`
    do pacote `internal/sensors`. Tudo local, nada sobe pra LLM."
  - 15–35s: "Rodei `harness sensor run secrets_scan` em três repos.
    Achou dois tokens vencidos esquecidos e um endpoint em comentário."
  - 35–50s: "Salva esse Short. Roda `harness ci` no seu repo hoje."
- **Visual**: split screen — terminal com `harness sensor run
  secrets_scan` marcando vermelho, print censurado do token.
- **Legenda-caps**: "SEU .ENV VAZA", "SENSOR LOCAL", "RODA HOJE".
- **Trilha**: tech dark 100 BPM.
- **VL pai**: V1.
- **CTA**: CTA-D.
- **Hashtags**: #devsecops #golang #ai #security #harness

#### Short 1.3 — "8 mapas vs pasta jogada"
- **Título**: `Mapas do harness vs pasta comum`
- **Fórmula**: F3
- **Duração**: 45s
- **Falas**:
  - 0–5s: "8 mapas do `harness project index` contra pasta jogada. Vamos."
  - 5–20s: "Pasta comum: você cola prompt, perde histórico, paga
    contexto duplicado. Sem índice, sem versão."
  - 20–35s: "harness: 8 mapas JSON, dedupe por shingle, rerank BM25,
    memória com `evidence_run_id`. Reaproveita contexto."
  - 35–45s: "Link do vídeo completo tá na bio."
- **Visual**: comparativo lado a lado, esquerda VS Code com pasta
  bagunçada, direita `harness project inspect commands`.
- **Legenda-caps**: "PASTA X MAPAS", "BM25", "DEDUPE", "ECONOMIA".
- **Trilha**: lo-fi tech 90 BPM.
- **VL pai**: V1.
- **CTA**: CTA-C.
- **Hashtags**: #ai #promptengineering #golang #devbr #harness

#### Short 1.4 — "Como eu setei projeto Go novo"
- **Título**: `Projeto Go novo com IA em 5 min`
- **Fórmula**: F4
- **Duração**: 60s
- **Falas**:
  - 0–3s: "Projeto Go novo pronto pra IA em cinco minutos. Como."
  - 3–20s: "Passo 1: `harness init`. Cria `.harness/` e config."
  - 20–37s: "Passo 2: `harness doctor`. Confirma que nada tá quebrado."
  - 37–52s: "Passo 3: primeiro `harness feature` já rodando spec-driven."
  - 52–60s: "Vídeo completo no primeiro comentário."
- **Visual**: screen recording 1,25x, 3 passos com contador na tela.
- **Legenda-caps**: "PASSO 1", "PASSO 2", "PASSO 3", "PRONTO".
- **Trilha**: tech upbeat 105 BPM.
- **VL pai**: V1.
- **CTA**: CTA-A.
- **Hashtags**: #golang #ai #cli #devbr #harness

---

### VL2 — `Spec antes do código: US$ 0,15 por scaffold Go`

#### Short 2.1 — "US$ 0,15 pra fazer scaffold"
- **Título**: `Scaffold Go por 15 centavos`
- **Fórmula**: F1
- **Duração**: 50s
- **Falas**:
  - 0–3s: "Quinze centavos de dólar. Scaffold Go completo."
  - 3–15s: "Ao invés de mandar 'faz um backend Go' pro Claude, escrevo
    spec de uma página primeiro."
  - 15–35s: "`harness feature` gera spec, plan, tasks. Passa pro
    executor. Sai backend com handlers, testes e migração."
  - 35–50s: "Custo: quinze centavos. Sem retrabalho."
- **Visual**: terminal com `harness cost-compare "..."` mostrando US$ 0.15.
- **Legenda-caps**: "US$ 0,15", "SPEC PRIMEIRO", "SEM RETRABALHO".
- **Trilha**: lo-fi tech 95 BPM.
- **VL pai**: V2.
- **CTA**: CTA-C.
- **Hashtags**: #ai #golang #devbr #saas #harness

#### Short 2.2 — "Spec vs prompt solto"
- **Título**: `Spec vs prompt: qual paga menos`
- **Fórmula**: F3
- **Duração**: 55s
- **Falas**:
  - 0–5s: "Spec-driven contra prompt solto. Vamos ao número."
  - 5–20s: "Prompt solto: 'faz um CRUD Go com JWT'. Claude reescreve
    coisa, contexto explode, dá 1 dólar."
  - 20–40s: "Spec-driven: `harness feature`, gera spec de 1 página,
    plan, tasks. Executor sabe exatamente o escopo. Quinze centavos."
  - 40–55s: "Sete vezes mais barato. Salva o Short."
- **Visual**: tabela comparando US$ 1,00 vs US$ 0,15.
- **Legenda-caps**: "PROMPT SOLTO", "SPEC-DRIVEN", "7X + BARATO".
- **Trilha**: tech 100 BPM.
- **VL pai**: V2.
- **CTA**: CTA-D.
- **Hashtags**: #ai #promptengineering #golang #saas #harness

#### Short 2.3 — "Mito: spec é burocracia"
- **Título**: `Spec é burocracia? Mentira`
- **Fórmula**: F5
- **Duração**: 50s
- **Falas**:
  - 0–4s: "Todo mundo fala que spec é burocracia. Mentira."
  - 4–20s: "Sem spec, o LLM adivinha. Adivinhando, ele reescreve, e
    reescrita gasta token."
  - 20–40s: "Spec de uma página em cinco minutos. Economiza dólar em
    todo scaffold e evita bug de escopo trocado."
  - 40–50s: "Comenta 'harness' pra receber o template."
- **Visual**: rosto Rodolfo + spec renderizada na lateral.
- **Legenda-caps**: "SEM SPEC = ADIVINHA", "ADIVINHA = TOKEN", "SPEC PAGA"
- **Trilha**: lo-fi 90 BPM.
- **VL pai**: V2.
- **CTA**: CTA-B.
- **Hashtags**: #ai #devbr #golang #harness #productivity

#### Short 2.4 — "Como escrevo spec em 5 min" (LOOP)
- **Título**: `Spec em 5 min — método`
- **Fórmula**: F2 (loop)
- **Duração**: 55s
- **Falas**:
  - 0–5s: "Spec em cinco minutos. Método fixo. Vou provar."
  - 5–20s: "Um: objetivo em uma frase. Dois: entrada e saída. Três:
    critério de aceite."
  - 20–40s: "Quatro: fora de escopo. Cinco: rota de rollback. Grava
    e passa pro `harness feature`."
  - 40–55s: "Spec em cinco minutos, método fixo — vou provar."
    *(loop volta ao começo)*
- **Visual**: papel com 5 bullets sendo escritos.
- **Legenda-caps**: "OBJETIVO", "I/O", "ACEITE", "FORA", "ROLLBACK".
- **Trilha**: lo-fi 92 BPM.
- **VL pai**: V2.
- **CTA**: CTA-C.
- **Hashtags**: #ai #devbr #productivity #harness #spec

---

### VL3 — `JWT sem footgun: por que Sonnet e não Codex`

#### Short 3.1 — "Codex errou o JWT"
- **Título**: `Codex errou meu JWT`
- **Fórmula**: F1
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Codex escreveu meu JWT. E deu ruim."
  - 3–15s: "Passei o spec de auth pro Codex. Ele fez, compilou,
    passou teste."
  - 15–35s: "Só que o refresh token não invalidava sessão antiga. Bug
    silencioso. Achei porque rodei `harness ci` com custom rule + teste
    de invalidação."
  - 35–50s: "Trocado por Sonnet 4.5. Passou em três releituras. Zero
    footgun."
  - 50–55s: "Vídeo completo no primeiro comentário."
- **Visual**: diff no editor, linha vermelha do bug + linha verde do fix.
- **Legenda-caps**: "CODEX ESCREVEU", "PASSOU TESTE", "BUG SILENCIOSO",
  "SONNET FEZ CERTO".
- **Trilha**: tech dark 100 BPM.
- **VL pai**: V3.
- **CTA**: CTA-A.
- **Hashtags**: #ai #golang #security #jwt #harness

#### Short 3.2 — "Sonnet vs Codex pra segurança"
- **Título**: `Sonnet vs Codex em código de segurança`
- **Fórmula**: F3
- **Duração**: 55s
- **Falas**:
  - 0–5s: "Sonnet contra Codex em código de auth. Quem entrega."
  - 5–20s: "Codex: barato, rápido, mas em segurança escorrega em
    corner case."
  - 20–40s: "Sonnet 4.5: mais caro por chamada, mas trata invalidação
    de sessão, rotação de key e claim mismatch sem drama."
  - 40–55s: "Regra minha: auth e crypto = Sonnet. CRUD = Codex."
- **Visual**: tabela dois lados.
- **Legenda-caps**: "AUTH = SONNET", "CRUD = CODEX", "REGRA FIXA".
- **Trilha**: tech 100 BPM.
- **VL pai**: V3.
- **CTA**: CTA-D.
- **Hashtags**: #ai #golang #security #devbr #harness

#### Short 3.3 — "3 armadilhas JWT que LLM cai"
- **Título**: `3 armadilhas de JWT`
- **Fórmula**: F1
- **Duração**: 60s
- **Falas**:
  - 0–3s: "Três armadilhas de JWT que LLM cai. Anota."
  - 3–20s: "Uma: `alg: none` aceito. Se o LLM não travar o alg, dá RCE
    de auth."
  - 20–37s: "Duas: refresh sem revoke. Sessão morta ressuscita."
  - 37–52s: "Três: claim `exp` no cliente. Cliente mente."
  - 52–60s: "Salva. Cola no prompt da próxima vez."
- **Visual**: 3 cards numerados aparecendo.
- **Legenda-caps**: "ALG NONE", "REFRESH SEM REVOKE", "EXP NO CLIENTE".
- **Trilha**: tech dark 105 BPM.
- **VL pai**: V3.
- **CTA**: CTA-D.
- **Hashtags**: #jwt #security #golang #ai #harness

#### Short 3.4 — "Como eu blindei JWT com harness" (F4)
- **Título**: `Blindando JWT com harness`
- **Fórmula**: F4
- **Duração**: 55s
- **Falas**:
  - 0–3s: "JWT blindado em três passos com harness. Vem."
  - 3–20s: "Passo 1: `harness feature` gera spec com claim fixo, alg
    travado."
  - 20–35s: "Passo 2: custom rule no `forbidden_commands` bloqueia
    `alg: none` (ver `docs/custom-rules.md`)."
  - 35–48s: "Passo 3: teste e2e que tenta romper. Tem que falhar."
  - 48–55s: "Link do tutorial na bio."
- **Visual**: screen recording acelerado.
- **Legenda-caps**: "SPEC", "SENSOR", "E2E", "BLINDADO".
- **Trilha**: tech 100 BPM.
- **VL pai**: V3.
- **CTA**: CTA-C.
- **Hashtags**: #jwt #golang #security #ai #harness

---

### VL4 — `Codex faz CRUD por 1/3 do preço — e cai bem`

#### Short 4.1 — "CRUD por 1/3 do preço"
- **Título**: `CRUD por 1/3 do custo`
- **Fórmula**: F1
- **Duração**: 45s
- **Falas**:
  - 0–3s: "CRUD Go inteiro por um terço do preço."
  - 3–15s: "Sonnet cobra caro por token. Pra CRUD, é overkill."
  - 15–35s: "Roteio Projects CRUD pro Codex. Handlers, migração,
    teste — sai por um terço."
  - 35–45s: "Vídeo tem número exato. Bio."
- **Visual**: gráfico de barras 3:1.
- **Legenda-caps**: "CRUD = CODEX", "1/3 DO PREÇO".
- **Trilha**: lo-fi tech 95 BPM.
- **VL pai**: V4.
- **CTA**: CTA-C.
- **Hashtags**: #ai #golang #saas #codex #harness

#### Short 4.2 — "Quando NÃO usar Codex" (F5)
- **Título**: `Quando NÃO usar Codex`
- **Fórmula**: F5
- **Duração**: 50s
- **Falas**:
  - 0–4s: "Codex é bom pra tudo? Não é."
  - 4–20s: "Pra CRUD, migração, teste simples: Codex arrasa."
  - 20–40s: "Pra crypto, JWT, RBAC complexo, orquestração: cai. Sonnet
    ou Opus."
  - 40–50s: "Regra na bio."
- **Visual**: tabela verde/vermelho.
- **Legenda-caps**: "CRUD OK", "AUTH NÃO", "OPUS PRA CRÍTICO".
- **Trilha**: tech 100 BPM.
- **VL pai**: V4.
- **CTA**: CTA-C.
- **Hashtags**: #ai #codex #golang #devbr #harness

#### Short 4.3 — "Roteamento por tarefa" (F4)
- **Título**: `Roteando LLM por tarefa`
- **Fórmula**: F4
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Roteio três LLMs no mesmo repo. Assim."
  - 3–18s: "Um: harness lê a spec e classifica risco."
  - 18–33s: "Dois: risco baixo vai Codex, médio Sonnet, alto Opus."
  - 33–48s: "Três: `harness cost-compare` no fim mostra delta por adapter."
  - 48–55s: "Repositório aberto — comenta 'harness'."
- **Visual**: fluxograma animado.
- **Legenda-caps**: "BAIXO", "MÉDIO", "ALTO", "ECONOMIA".
- **Trilha**: tech 105 BPM.
- **VL pai**: V4.
- **CTA**: CTA-B.
- **Hashtags**: #ai #golang #saas #devbr #harness

#### Short 4.4 — "CRUD limpo em Go" (F1)
- **Título**: `CRUD Go que LLM ama`
- **Fórmula**: F1
- **Duração**: 50s
- **Falas**:
  - 0–3s: "CRUD Go que o LLM não estraga: existe."
  - 3–15s: "Handler fino, service com regra, repo com sqlc. Sem gorm
    mágico."
  - 15–35s: "LLM lê essa estrutura e replica. Zero criatividade
    perigosa."
  - 35–50s: "Vídeo mostra o boilerplate. Bio."
- **Visual**: código Go em tela.
- **Legenda-caps**: "HANDLER FINO", "SQLC", "SEM MÁGICA".
- **Trilha**: lo-fi 95 BPM.
- **VL pai**: V4.
- **CTA**: CTA-C.
- **Hashtags**: #golang #ai #saas #devbr #harness

---

### VL5 — `BM25 achou meu helper — Codex não reescreveu`

#### Short 5.1 — "BM25 salvou 400 tokens"
- **Título**: `BM25 salvou meu contexto`
- **Fórmula**: F1
- **Duração**: 55s
- **Falas**:
  - 0–3s: "BM25 economizou 400 tokens do meu contexto. Como."
  - 3–18s: "Sem retrieve, o Codex ia reescrever meu helper de paginação
    que já existia."
  - 18–38s: "Com `internal/context.Build`, BM25 achou o helper e
    injetou só o trecho relevante."
  - 38–50s: "Codex reusou, não duplicou. Contexto: 400 tokens a menos."
  - 50–55s: "Bio tem o snippet."
- **Visual**: terminal + trecho de código destacado.
- **Legenda-caps**: "BM25", "REUSA", "-400 TOKENS".
- **Trilha**: tech 100 BPM.
- **VL pai**: V5.
- **CTA**: CTA-C.
- **Hashtags**: #ai #search #golang #devbr #harness

#### Short 5.2 — "BM25 vs embedding" (F3)
- **Título**: `BM25 vs embedding: quem ganha`
- **Fórmula**: F3
- **Duração**: 55s
- **Falas**:
  - 0–5s: "BM25 contra embedding pra código. Vamos."
  - 5–22s: "Embedding: bom pra semântica, ruim pra nome exato de
    função."
  - 22–42s: "BM25: acha `PaginateTasks` porque é literal. Zero custo
    de embedding. Zero API call."
  - 42–55s: "Rerank híbrido no harness: BM25 primeiro, embedding só
    se falhar."
- **Visual**: split.
- **Legenda-caps**: "LITERAL = BM25", "SEMÂNTICO = EMB", "HÍBRIDO".
- **Trilha**: tech 100 BPM.
- **VL pai**: V5.
- **CTA**: CTA-D.
- **Hashtags**: #ai #search #retrieval #devbr #harness

#### Short 5.3 — "Mito: RAG precisa vector DB" (F5)
- **Título**: `RAG sem vector DB`
- **Fórmula**: F5
- **Duração**: 50s
- **Falas**:
  - 0–4s: "Precisa de vector DB pra RAG? Não precisa."
  - 4–20s: "Pra código, BM25 puro resolve 80% dos casos."
  - 20–40s: "SQLite FTS5 + BM25 rerank cabe no seu binário. Zero
    Postgres, zero Pinecone."
  - 40–50s: "Salva. Vídeo tem benchmark."
- **Visual**: rosto + diagrama.
- **Legenda-caps**: "SQLITE FTS5", "BM25", "SEM PINECONE".
- **Trilha**: lo-fi 92 BPM.
- **VL pai**: V5.
- **CTA**: CTA-D.
- **Hashtags**: #rag #ai #sqlite #devbr #harness

#### Short 5.4 — "Como monto retrieve stack" (F4)
- **Título**: `Retrieve stack em 60s`
- **Fórmula**: F4
- **Duração**: 60s
- **Falas**:
  - 0–3s: "Retrieve stack pra IA de código em 60s."
  - 3–20s: "Um: FTS5 no SQLite. Grava tudo local."
  - 20–37s: "Dois: BM25 rerank em cima do top-k."
  - 37–52s: "Três: LLMLingua-2 corta contexto redundante antes de
    mandar."
  - 52–60s: "Link tutorial na bio."
- **Visual**: 3 caixas empilhando.
- **Legenda-caps**: "FTS5", "BM25", "LLMLINGUA".
- **Trilha**: tech 105 BPM.
- **VL pai**: V5.
- **CTA**: CTA-C.
- **Hashtags**: #rag #ai #golang #devbr #harness

---

### VL6 — `Vite 6 + React 19 sem gastar 1 token`

#### Short 6.1 — "Scaffold zero token"
- **Título**: `Scaffold React sem gastar token`
- **Fórmula**: F1
- **Duração**: 45s
- **Falas**:
  - 0–3s: "Scaffold React 19 completo. Zero token."
  - 3–15s: "Vite 6 + template TypeScript já entrega 90% do que o LLM
    faria."
  - 15–35s: "Uso LLM só pra colar routing, layout base e store."
  - 35–45s: "Custo LLM do front todo: quatro centavos."
- **Visual**: terminal `npm create vite`.
- **Legenda-caps**: "VITE 6", "REACT 19", "US$ 0,04".
- **Trilha**: lo-fi 95 BPM.
- **VL pai**: V6.
- **CTA**: CTA-C.
- **Hashtags**: #react #vite #ai #frontend #harness

#### Short 6.2 — "Vite 6 vs Next" (F3)
- **Título**: `Vite 6 vs Next pra SaaS pequeno`
- **Fórmula**: F3
- **Duração**: 55s
- **Falas**:
  - 0–5s: "Vite 6 contra Next 15 pra SaaS pequeno."
  - 5–22s: "Next: SSR, mas peso. Deploy mais complexo. Overhead pra
    SPA autenticado."
  - 22–42s: "Vite: SPA puro, deploy num tar.gz atrás do Go. Simples."
  - 42–55s: "SaaS pequeno = Vite. Marketing site = Next."
- **Visual**: tabela.
- **Legenda-caps**: "SAAS = VITE", "MKT = NEXT".
- **Trilha**: tech 100 BPM.
- **VL pai**: V6.
- **CTA**: CTA-D.
- **Hashtags**: #react #vite #nextjs #frontend #harness

#### Short 6.3 — "3 arquivos que scaffold já entrega" (F1)
- **Título**: `3 arquivos grátis do Vite`
- **Fórmula**: F1
- **Duração**: 50s
- **Falas**:
  - 0–3s: "Três arquivos que o Vite já te dá. Não pede pro LLM."
  - 3–20s: "Um: `vite.config.ts` com alias e proxy."
  - 20–35s: "Dois: `tsconfig.json` com strict."
  - 35–48s: "Três: `main.tsx` com StrictMode."
  - 48–50s: "Bio."
- **Visual**: 3 arquivos abrindo.
- **Legenda-caps**: "VITE.CONFIG", "TSCONFIG", "MAIN.TSX".
- **Trilha**: lo-fi 92 BPM.
- **VL pai**: V6.
- **CTA**: CTA-C.
- **Hashtags**: #react #vite #frontend #devbr #harness

#### Short 6.4 — "Como economizei US$ 0,80" (F4)
- **Título**: `US$ 0,80 economizados no front`
- **Fórmula**: F4
- **Duração**: 55s
- **Falas**:
  - 0–3s: "US$ 0,80 economizados no scaffold front. Como."
  - 3–20s: "Um: Vite gera boilerplate."
  - 20–35s: "Dois: pego template shadcn/ui pronto."
  - 35–48s: "Três: LLM só entra em componentes de negócio."
  - 48–55s: "Vídeo tem cost report."
- **Visual**: recording acelerado.
- **Legenda-caps**: "VITE", "SHADCN", "LLM SÓ REGRA".
- **Trilha**: tech 100 BPM.
- **VL pai**: V6.
- **CTA**: CTA-A.
- **Hashtags**: #react #ai #frontend #devbr #harness

---

### VL7 — `LLMLingua-2 corta 40% do pack React`

#### Short 7.1 — "Compressão corta ~35% do prose"
- **Título**: `Compressão LLMLingua-style: -35% no prose`
- **Fórmula**: F1
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Cortei ~35% do prose do meu pack. Código intacto."
  - 3–20s: "harness roda compressão determinística LLMLingua-style em
    `internal/context/compress` — não chama modelo."
  - 20–40s: "Prose comprime, código fica verbatim. No benchmark do
    fixture Go: ~35% menos bytes em README/docs. Combinado com rerank
    + reorder chega em 50–70% menos token no pack."
  - 40–55s: "Doc: `docs/PAPER-IMPLEMENTATION.md`. Bio."
- **Visual**: antes/depois de tokens.
- **Legenda-caps**: "-35% PROSE", "CÓDIGO INTACTO", "DETERMINÍSTICO".
- **Trilha**: tech 100 BPM.
- **VL pai**: V7.
- **CTA**: CTA-C.
- **Hashtags**: #ai #promptengineering #react #harness #paper

#### Short 7.2 — "Compressão vs truncamento" (F3)
- **Título**: `Comprimir vs truncar prompt`
- **Fórmula**: F3
- **Duração**: 55s
- **Falas**:
  - 0–5s: "Comprimir contra truncar prompt. Diferença?"
  - 5–22s: "Truncar: corta o fim. Perde regra crítica."
  - 22–42s: "Comprimir LLMLingua-style: strip whitespace, dedupe
    linha, corta prose em word boundary. Código fica verbatim."
  - 42–55s: "Não é a mesma coisa. Vídeo tem prova."
- **Visual**: highlight vermelho no que truncar removeu.
- **Legenda-caps**: "TRUNCA PERDE", "COMPRIME MANTÉM".
- **Trilha**: tech 100 BPM.
- **VL pai**: V7.
- **CTA**: CTA-D.
- **Hashtags**: #ai #promptengineering #devbr #harness #paper

#### Short 7.3 — "Mito: prompt pequeno é sempre melhor" (F5)
- **Título**: `Prompt pequeno é sempre melhor?`
- **Fórmula**: F5
- **Duração**: 50s
- **Falas**:
  - 0–4s: "Prompt pequeno é sempre melhor? Depende."
  - 4–20s: "Se corta mal, tira contexto crítico e o modelo alucina."
  - 20–40s: "Compressão do harness é determinística: só toca prose,
    preserva código verbatim. Sem risco de cortar identifier."
  - 40–50s: "Salva."
- **Visual**: rosto + gráfico perplexidade.
- **Legenda-caps**: "COMPRIME COM MÉTRICA", "NÃO NO CHUTE".
- **Trilha**: lo-fi 95 BPM.
- **VL pai**: V7.
- **CTA**: CTA-D.
- **Hashtags**: #ai #devbr #paper #harness #llm

#### Short 7.4 — "Como o pipeline de contexto roda" (F4)
- **Título**: `Pipeline de contexto do harness`
- **Fórmula**: F4
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Pipeline de contexto do harness em três passos."
  - 3–20s: "Um: providers (git, ripgrep, lsp, testmap, memory) juntam
    candidates."
  - 20–35s: "Dois: rerank BM25 trima pro MaxKeep = 40."
  - 35–48s: "Três: compressão LLMLingua-style + reorder salience
    head/tail contra 'lost in the middle'."
  - 48–55s: "Bio."
- **Visual**: `harness context build` + `harness context inspect`.
- **Legenda-caps**: "RERANK", "COMPRIME", "REORDER".
- **Trilha**: tech 105 BPM.
- **VL pai**: V7.
- **CTA**: CTA-C.
- **Hashtags**: #ai #devbr #harness #promptengineering #llm

---

### VL8 — `Optimistic updates em React 19 sem bug`

#### Short 8.1 — "Optimistic sem bug"
- **Título**: `Optimistic update React 19`
- **Fórmula**: F1
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Optimistic update em React 19 sem bug de sincronia."
  - 3–20s: "Hook `useOptimistic` atualiza UI antes da API responder."
  - 20–40s: "Se der erro, rollback automático. Sem `setState` manual,
    sem race condition."
  - 40–55s: "Vídeo tem demo com falha simulada."
- **Visual**: gif do UI + terminal fake latency.
- **Legenda-caps**: "USEOPTIMISTIC", "ROLLBACK AUTO".
- **Trilha**: tech 100 BPM.
- **VL pai**: V8.
- **CTA**: CTA-C.
- **Hashtags**: #react #frontend #devbr #harness #react19

#### Short 8.2 — "React 19 vs 18" (F3)
- **Título**: `React 19 vs 18 pra SaaS`
- **Fórmula**: F3
- **Duração**: 50s
- **Falas**:
  - 0–5s: "React 19 contra 18 pra SaaS pequeno. O que muda."
  - 5–20s: "18: `useState` + libs externas pra optimistic."
  - 20–40s: "19: `useOptimistic`, `useActionState`, form nativo com
    action. Menos lib, menos bug."
  - 40–50s: "Migração vale."
- **Visual**: tabela.
- **Legenda-caps**: "MENOS LIB", "MENOS BUG", "MIGRA".
- **Trilha**: tech 100 BPM.
- **VL pai**: V8.
- **CTA**: CTA-D.
- **Hashtags**: #react #react19 #frontend #devbr #harness

#### Short 8.3 — "3 bugs de optimistic que LLM cria" (F1)
- **Título**: `3 bugs comuns em optimistic`
- **Fórmula**: F1
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Três bugs de optimistic que LLM cria. Anota."
  - 3–18s: "Um: aplica update e esquece de dedupe no server response."
  - 18–33s: "Dois: rollback não limpa erro anterior."
  - 33–48s: "Três: usa `useState` paralelo em vez de `useOptimistic`."
  - 48–55s: "Salva."
- **Visual**: 3 cards.
- **Legenda-caps**: "DEDUPE", "ROLLBACK", "USE OPTIMISTIC".
- **Trilha**: tech 105 BPM.
- **VL pai**: V8.
- **CTA**: CTA-D.
- **Hashtags**: #react #frontend #devbr #harness #bugs

#### Short 8.4 — "Como testo optimistic" (F4)
- **Título**: `Testando optimistic sem flakiness`
- **Fórmula**: F4
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Optimistic testado sem flaky. Método."
  - 3–20s: "Um: Playwright com `waitForResponse` na rota mockada."
  - 20–37s: "Dois: assert dois estados — otimista e final."
  - 37–48s: "Três: fluxo de erro com `route.abort`."
  - 48–55s: "Bio tem código."
- **Visual**: screen com Playwright rodando.
- **Legenda-caps**: "WAITFORRESPONSE", "DOIS ESTADOS", "ROUTE ABORT".
- **Trilha**: tech 100 BPM.
- **VL pai**: V8.
- **CTA**: CTA-C.
- **Hashtags**: #playwright #react #frontend #devbr #harness

---

### VL9 — `3 agentes pra 3 camadas de teste`

#### Short 9.1 — "3 agentes, 3 camadas"
- **Título**: `3 agentes pra 3 tipos de teste`
- **Fórmula**: F1
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Três agentes escrevem meus testes. Cada um numa camada."
  - 3–18s: "Codex: unit test. Barato, previsível."
  - 18–35s: "Kimi K2: integration. Contexto grande, entende tabela."
  - 35–50s: "Sonnet 4.5: e2e. Entende fluxo humano."
  - 50–55s: "Bio."
- **Visual**: pirâmide de teste com 3 avatares.
- **Legenda-caps**: "UNIT", "INTEGRATION", "E2E".
- **Trilha**: tech 100 BPM.
- **VL pai**: V9.
- **CTA**: CTA-C.
- **Hashtags**: #testing #ai #golang #devbr #harness

#### Short 9.2 — "Um LLM só faz teste ruim" (F5)
- **Título**: `Um LLM não escreve todo teste`
- **Fórmula**: F5
- **Duração**: 50s
- **Falas**:
  - 0–4s: "Um LLM só pra escrever todo tipo de teste? Ruim."
  - 4–20s: "Cada camada exige contexto diferente."
  - 20–40s: "Unit precisa de código local. Integration precisa de
    schema. E2e precisa de fluxo. Divide."
  - 40–50s: "Salva."
- **Visual**: rosto + diagrama.
- **Legenda-caps**: "DIVIDE POR CAMADA".
- **Trilha**: lo-fi 92 BPM.
- **VL pai**: V9.
- **CTA**: CTA-D.
- **Hashtags**: #testing #ai #devbr #harness #qa

#### Short 9.3 — "Custo dos 3 agentes" (F3)
- **Título**: `Custo dos 3 agentes de teste`
- **Fórmula**: F3
- **Duração**: 55s
- **Falas**:
  - 0–5s: "Quanto custa rodar três agentes pra testar. Vamos."
  - 5–22s: "Codex unit: 6 centavos por suite."
  - 22–37s: "Kimi integration: 11 centavos."
  - 37–50s: "Sonnet e2e: 22 centavos. Total: 39 centavos por release."
  - 50–55s: "Vídeo tem breakdown."
- **Visual**: 3 barras.
- **Legenda-caps**: "US$ 0,39", "POR RELEASE".
- **Trilha**: tech 100 BPM.
- **VL pai**: V9.
- **CTA**: CTA-C.
- **Hashtags**: #testing #ai #devbr #harness #cost

#### Short 9.4 — "Como orquestro 3 agentes" (F4)
- **Título**: `Orquestrando 3 agentes de teste`
- **Fórmula**: F4
- **Duração**: 60s
- **Falas**:
  - 0–3s: "Três agentes de teste, orquestrados. Como."
  - 3–22s: "Um: `flow.yaml` define proposer, critic, decider."
  - 22–40s: "Dois: cada camada dispara agente diferente."
  - 40–55s: "Três: harness rankeia saída e escolhe. Zero cola manual."
  - 55–60s: "Bio."
- **Visual**: yaml + fluxograma.
- **Legenda-caps**: "FLOW YAML", "AUTO RANK".
- **Trilha**: tech 105 BPM.
- **VL pai**: V9.
- **CTA**: CTA-B.
- **Hashtags**: #testing #ai #golang #devbr #harness

---

### VL10 — `docker compose v2 com mem_limit ou não sobe`

#### Short 10.1 — "mem_limit não é opcional"
- **Título**: `Sem mem_limit teu compose trava`
- **Fórmula**: F1
- **Duração**: 50s
- **Falas**:
  - 0–3s: "Sem `mem_limit` no compose, teu Mac trava. Fato."
  - 3–20s: "Container sem teto come RAM. Se roda Next dev + E2E junto,
    passa dos 10GB fácil."
  - 20–40s: "Coloca `mem_limit: 2g` e `memswap_limit: 2g` em todo
    service. Regra sagrada."
  - 40–50s: "Bio."
- **Visual**: docker stats explodindo.
- **Legenda-caps**: "MEM_LIMIT", "SEMPRE", "OU TRAVA".
- **Trilha**: tech dark 100 BPM.
- **VL pai**: V10.
- **CTA**: CTA-D.
- **Hashtags**: #docker #devops #devbr #harness #saas

#### Short 10.2 — "Compose v1 vs v2" (F3)
- **Título**: `docker-compose vs docker compose`
- **Fórmula**: F3
- **Duração**: 45s
- **Falas**:
  - 0–5s: "docker-compose com traço tá morto. Explico."
  - 5–20s: "v1 Python: sem suporte, sem healthcheck moderno."
  - 20–35s: "v2 Go plugin: healthcheck, `depends_on: condition`, dev
    profile."
  - 35–45s: "Migra hoje. Salva."
- **Visual**: comando dos dois.
- **Legenda-caps**: "V1 MORTO", "V2 PADRÃO".
- **Trilha**: tech 100 BPM.
- **VL pai**: V10.
- **CTA**: CTA-D.
- **Hashtags**: #docker #devops #devbr #harness #cli

#### Short 10.3 — "Healthcheck que salva CI" (F1)
- **Título**: `Healthcheck que salva CI`
- **Fórmula**: F1
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Healthcheck no compose salvou minha CI três vezes."
  - 3–20s: "Sem health, o test roda antes do banco subir. Flaky."
  - 20–40s: "Com `healthcheck` + `depends_on: condition: service_healthy`,
    ordem garantida."
  - 40–55s: "Snippet na bio."
- **Visual**: yaml + terminal.
- **Legenda-caps**: "SERVICE_HEALTHY", "CI VERDE".
- **Trilha**: tech 100 BPM.
- **VL pai**: V10.
- **CTA**: CTA-C.
- **Hashtags**: #docker #ci #devbr #harness #devops

#### Short 10.4 — "Como subo E2E sem travar" (F4)
- **Título**: `E2E sem travar meu Mac`
- **Fórmula**: F4
- **Duração**: 60s
- **Falas**:
  - 0–3s: "E2E rodando sem travar o Mac. Método."
  - 3–20s: "Um: `mem_limit: 2g` em todo service."
  - 20–37s: "Dois: `pkill -f next-server` antes de subir."
  - 37–52s: "Três: `docker compose down --remove-orphans` sempre no fim."
  - 52–60s: "Regra na bio."
- **Visual**: terminal.
- **Legenda-caps**: "MEM_LIMIT", "PKILL", "DOWN".
- **Trilha**: tech 100 BPM.
- **VL pai**: V10.
- **CTA**: CTA-D.
- **Hashtags**: #docker #devops #playwright #devbr #harness

---

### VL11 — `Dashboard do harness na porta :0 — sem colisão`

#### Short 11.1 — "Porta :0 nunca colide"
- **Título**: `Porta :0 mata colisão`
- **Fórmula**: F1
- **Duração**: 45s
- **Falas**:
  - 0–3s: "Porta zero nunca colide. Truque velho, muito útil."
  - 3–18s: "SO escolhe uma porta livre em runtime. Você lê e printa."
  - 18–35s: "Dashboard do harness abre porta zero e imprime a URL.
    Zero briga com Next dev."
  - 35–45s: "Bio."
- **Visual**: terminal `netstat`.
- **Legenda-caps**: "PORT :0", "SEM COLISÃO".
- **Trilha**: tech 100 BPM.
- **VL pai**: V11.
- **CTA**: CTA-C.
- **Hashtags**: #golang #devbr #harness #dashboard #cli

#### Short 11.2 — "Perf snapshot antes/depois" (F3)
- **Título**: `Perf snapshot antes/depois`
- **Fórmula**: F3
- **Duração**: 55s
- **Falas**:
  - 0–5s: "Refactor sem perf snapshot é fé cega."
  - 5–22s: "Antes: `harness perf-snapshot --label pre`."
  - 22–40s: "Depois: `harness perf-compare pre post`. Non-zero exit se
    regrediu."
  - 40–55s: "Wire no `harness ci` e PR fica barrada."
- **Visual**: terminal diff.
- **Legenda-caps**: "SAVE", "DIFF", "GATE 10%".
- **Trilha**: tech 100 BPM.
- **VL pai**: V11.
- **CTA**: CTA-C.
- **Hashtags**: #performance #golang #devbr #harness #ci

#### Short 11.3 — "Mito: dashboard local é chato" (F5)
- **Título**: `Dashboard local vale?`
- **Fórmula**: F5
- **Duração**: 50s
- **Falas**:
  - 0–4s: "Dashboard local é frescura? Não."
  - 4–20s: "Cost, perf, sensores num só lugar."
  - 20–40s: "Abre em porta zero, roda no seu localhost, zero telemetria
    externa. Auditável."
  - 40–50s: "Salva."
- **Visual**: dashboard screenshot.
- **Legenda-caps**: "COST", "PERF", "SENSORES".
- **Trilha**: lo-fi 92 BPM.
- **VL pai**: V11.
- **CTA**: CTA-D.
- **Hashtags**: #golang #devbr #harness #dashboard #devops

#### Short 11.4 — "Como abro dashboard" (F4)
- **Título**: `Dashboard em 3 comandos`
- **Fórmula**: F4
- **Duração**: 45s
- **Falas**:
  - 0–3s: "Dashboard do harness em um comando."
  - 3–15s: "Um: `harness dashboard --addr :0 --open`."
  - 15–28s: "Dois: SO escolhe porta, navegador abre sozinho."
  - 28–40s: "Três: /api/sessions, /api/cost, /api/sensors — ao vivo."
  - 40–45s: "Bio."
- **Visual**: recording.
- **Legenda-caps**: "--addr :0", "AUTO ABRE", "AO VIVO".
- **Trilha**: tech 105 BPM.
- **VL pai**: V11.
- **CTA**: CTA-C.
- **Hashtags**: #golang #devbr #harness #dashboard #cli

---

### VL12 — `US$ 1,23 vs US$ 5,00 — o que o context engine cortou`

#### Short 12.1 — "US$ 1,23 pra fazer SaaS"
- **Título**: `SaaS por US$ 1,23 em LLM`
- **Fórmula**: F1
- **Duração**: 55s
- **Falas**:
  - 0–3s: "SaaS Go + React completo. Um dólar e vinte e três centavos."
  - 3–20s: "Back Go, front React 19, JWT, CRUD, Playwright, Docker."
  - 20–40s: "Custo LLM auditado no dashboard. Screenshot no vídeo."
  - 40–55s: "Comenta 'harness' que mando o breakdown."
- **Visual**: `harness analytics` + `harness cost-compare` output.
- **Legenda-caps**: "US$ 1,23", "SAAS COMPLETO".
- **Trilha**: tech 100 BPM.
- **VL pai**: V12.
- **CTA**: CTA-B.
- **Hashtags**: #ai #saas #golang #react #harness

#### Short 12.2 — "Context engine vs prompt cru" (F3)
- **Título**: `Context engine vs prompt cru`
- **Fórmula**: F3
- **Duração**: 55s
- **Falas**:
  - 0–5s: "Context engine contra prompt cru. Quem paga menos."
  - 5–22s: "Prompt cru: manda tudo. Cinco dólares num scaffold."
  - 22–42s: "Context engine: BM25, LLMLingua, dedupe, cache. Um dólar
    e vinte e três."
  - 42–55s: "Quatro vezes mais barato. Vídeo."
- **Visual**: gráfico.
- **Legenda-caps**: "US$ 5,00", "US$ 1,23", "4X".
- **Trilha**: tech 100 BPM.
- **VL pai**: V12.
- **CTA**: CTA-C.
- **Hashtags**: #ai #promptengineering #saas #devbr #harness

#### Short 12.3 — "Mito: IA barata mente" (F5)
- **Título**: `IA barata mente mais?`
- **Fórmula**: F5
- **Duração**: 55s
- **Falas**:
  - 0–4s: "IA barata mente mais? Depende do contexto, não do preço."
  - 4–22s: "Codex + spec boa entrega CRUD sem alucinar."
  - 22–42s: "Sonnet + contexto ruim inventa API. Preço não salva."
  - 42–55s: "Investe em contexto, não em modelo caro."
- **Visual**: rosto + tabela.
- **Legenda-caps**: "CONTEXTO > MODELO".
- **Trilha**: lo-fi 92 BPM.
- **VL pai**: V12.
- **CTA**: CTA-D.
- **Hashtags**: #ai #devbr #promptengineering #harness #cost

#### Short 12.4 — "Como audito custo" (F4)
- **Título**: `Auditando custo LLM em 30s`
- **Fórmula**: F4
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Auditando custo LLM do harness em 30 segundos."
  - 3–18s: "Um: `harness cost-compare "<prompt>"` — delta por adapter."
  - 18–33s: "Dois: `harness analytics` — telemetria local agregada."
  - 33–50s: "Três: `harness metrics show` — trajetória tokens, edits,
    wall-ms por sessão."
  - 50–55s: "Salva."
- **Visual**: terminal.
- **Legenda-caps**: "COST-COMPARE", "ANALYTICS", "METRICS".
- **Trilha**: tech 100 BPM.
- **VL pai**: V12.
- **CTA**: CTA-C.
- **Hashtags**: #ai #devbr #harness #cost #cli

---

### VL13 — `Proposer → Critic → Decider em flow.yaml`

#### Short 13.1 — "3 papéis num flow"
- **Título**: `Proposer, Critic, Decider`
- **Fórmula**: F1
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Proposer. Critic. Decider. Três papéis num flow só."
  - 3–20s: "Proposer gera solução. Critic acha buraco. Decider escolhe."
  - 20–40s: "Cada papel pode ser LLM diferente. Reduz enviesamento."
  - 40–55s: "Vídeo tem o yaml."
- **Visual**: fluxograma animado.
- **Legenda-caps**: "PROPOSER", "CRITIC", "DECIDER".
- **Trilha**: tech 100 BPM.
- **VL pai**: V13.
- **CTA**: CTA-C.
- **Hashtags**: #ai #agents #devbr #harness #paper

#### Short 13.2 — "1 LLM só vs 3 agentes" (F3)
- **Título**: `1 LLM vs 3 agentes`
- **Fórmula**: F3
- **Duração**: 55s
- **Falas**:
  - 0–5s: "Um LLM só contra três agentes. Diferença?"
  - 5–22s: "Um LLM: gera e assina embaixo. Sem revisão."
  - 22–42s: "Três agentes: proposer erra, critic pega, decider decide.
    Menos bug em produção."
  - 42–55s: "Custo? Pouca coisa a mais."
- **Visual**: tabela.
- **Legenda-caps**: "MENOS BUG", "POUCO + CARO".
- **Trilha**: tech 100 BPM.
- **VL pai**: V13.
- **CTA**: CTA-D.
- **Hashtags**: #ai #agents #devbr #harness #saas

#### Short 13.3 — "Como escrevo flow.yaml" (F4)
- **Título**: `flow.yaml em 60s`
- **Fórmula**: F4
- **Duração**: 60s
- **Falas**:
  - 0–3s: "flow multi-agente do harness em 60 segundos."
  - 3–22s: "Um: `harness flow init` scaffolda o pack."
  - 22–40s: "Dois: `harness orchestrate list` mostra roles disponíveis."
  - 40–55s: "Três: `harness orchestrate run <name>` grava blackboard
    em `.harness/artifacts/runs/<id>/blackboard.json`."
  - 55–60s: "Bio."
- **Visual**: yaml + terminal.
- **Legenda-caps**: "FLOW INIT", "ORCHESTRATE", "BLACKBOARD".
- **Trilha**: tech 105 BPM.
- **VL pai**: V13.
- **CTA**: CTA-C.
- **Hashtags**: #ai #agents #golang #harness #yaml

#### Short 13.4 — "Mito: multi-agente é overkill" (F5)
- **Título**: `Multi-agente é overkill?`
- **Fórmula**: F5
- **Duração**: 50s
- **Falas**:
  - 0–4s: "Multi-agente é overkill pra dev solo? Não."
  - 4–20s: "Um agente escreve e outro revisa é literalmente pair
    programming."
  - 20–40s: "Custo extra é uns 20%. Bug pego antes vale muito mais."
  - 40–50s: "Salva."
- **Visual**: rosto.
- **Legenda-caps**: "PAIR PROGRAM", "VALE +20%".
- **Trilha**: lo-fi 92 BPM.
- **VL pai**: V13.
- **CTA**: CTA-D.
- **Hashtags**: #ai #agents #devbr #harness #productivity

---

### VL14 — `harness bugfix vs feature: qual usar quando`

#### Short 14.1 — "Bugfix não é feature"
- **Título**: `bugfix ≠ feature`
- **Fórmula**: F1
- **Duração**: 50s
- **Falas**:
  - 0–3s: "`harness bugfix` não é `harness feature`. Escopo diferente."
  - 3–20s: "Feature: cria spec, plan, tasks. Escopo aberto."
  - 20–40s: "Bugfix: reproduz, minimiza mudança, adiciona teste que
    quebra sem fix."
  - 40–50s: "Usa o certo. Vídeo tem tabela."
- **Visual**: dois comandos lado a lado.
- **Legenda-caps**: "FEATURE ABRE", "BUGFIX FECHA".
- **Trilha**: tech 100 BPM.
- **VL pai**: V14.
- **CTA**: CTA-C.
- **Hashtags**: #devbr #golang #harness #productivity #cli

#### Short 14.2 — "Reproduzir antes de corrigir" (F5)
- **Título**: `Reproduza antes de corrigir`
- **Fórmula**: F5
- **Duração**: 50s
- **Falas**:
  - 0–4s: "Corrigir sem reproduzir é regressão garantida."
  - 4–20s: "`harness bugfix` força passo 1: teste que falha."
  - 20–40s: "Se o teste não falha, não é bugfix. É outra coisa. Regra
    dura."
  - 40–50s: "Salva."
- **Visual**: terminal teste vermelho.
- **Legenda-caps**: "TESTE VERMELHO PRIMEIRO".
- **Trilha**: tech dark 100 BPM.
- **VL pai**: V14.
- **CTA**: CTA-D.
- **Hashtags**: #testing #devbr #golang #harness #bugs

#### Short 14.3 — "Como faço evolve com HITL" (F4)
- **Título**: `harness evolve — mutation com HITL`
- **Fórmula**: F4
- **Duração**: 55s
- **Falas**:
  - 0–3s: "Mutation telemetria-driven com `harness evolve`."
  - 3–20s: "Um: `harness evolve diagnose` lê traços de execução."
  - 20–35s: "Dois: `harness evolve propose` sugere mudança + rationale."
  - 35–48s: "Três: `harness evolve promote --hitl` só passa com humano
    aprovando."
  - 48–55s: "Bio."
- **Visual**: terminal.
- **Legenda-caps**: "DIAGNOSE", "PROPOSE", "HITL".
- **Trilha**: tech 100 BPM.
- **VL pai**: V14.
- **CTA**: CTA-C.
- **Hashtags**: #devbr #golang #harness #refactor #perf

#### Short 14.4 — "3 comandos que uso todo dia" (F1)
- **Título**: `3 comandos harness diários`
- **Fórmula**: F1
- **Duração**: 45s
- **Falas**:
  - 0–3s: "Três comandos harness que uso todo dia."
  - 3–15s: "Um: `harness ci` antes de commit."
  - 15–28s: "Dois: `harness analytics` no fim do dia."
  - 28–40s: "Três: `harness doctor` toda segunda."
  - 40–45s: "Salva."
- **Visual**: 3 cards.
- **Legenda-caps**: "CI", "COST", "DOCTOR".
- **Trilha**: tech 105 BPM.
- **VL pai**: V14.
- **CTA**: CTA-D.
- **Hashtags**: #devbr #golang #harness #cli #productivity

---

### VL15 — `harness ci verde, harness ship, tag assinada`

#### Short 15.1 — "Ship em um comando"
- **Título**: `harness ship em um comando`
- **Fórmula**: F1
- **Duração**: 50s
- **Falas**:
  - 0–3s: "Ship em um comando. `harness ship "<prompt>"`."
  - 3–18s: "Cria branch, escreve spec via `feature`, roda `do` + `ci`
    em loop, commit conventional. 429-aware fallback."
  - 18–38s: "Se qualquer sensor reprovar, aborta. Sem PR quebrado."
  - 38–50s: "Bio."
- **Visual**: terminal.
- **Legenda-caps**: "SPEC", "DO+CI LOOP", "CONVENTIONAL".
- **Trilha**: tech 100 BPM.
- **VL pai**: V15.
- **CTA**: CTA-C.
- **Hashtags**: #devops #devbr #golang #harness #ci

#### Short 15.2 — "Tag assinada importa?" (F5)
- **Título**: `Tag assinada importa mesmo?`
- **Fórmula**: F5
- **Duração**: 50s
- **Falas**:
  - 0–4s: "Tag assinada é frescura de OSS? Não."
  - 4–20s: "Sem assinatura, qualquer um empurra tag no seu nome."
  - 20–40s: "Com SSH sign key, GitHub marca verde. Auditável."
  - 40–50s: "Salva."
- **Visual**: badge verified.
- **Legenda-caps**: "SSH SIGN", "GITHUB VERIFIED".
- **Trilha**: tech 100 BPM.
- **VL pai**: V15.
- **CTA**: CTA-D.
- **Hashtags**: #devops #security #devbr #harness #git

#### Short 15.3 — "CI local vs Actions" (F3)
- **Título**: `CI local vs GitHub Actions`
- **Fórmula**: F3
- **Duração**: 55s
- **Falas**:
  - 0–5s: "CI local contra GitHub Actions pra projeto pequeno."
  - 5–22s: "Actions: minuto grátis limitado, cold start, YAML frágil."
  - 22–42s: "Local: `make ci` no pre-push. Zero fila. Roda igual em
    todo dev."
  - 42–55s: "Pequeno = local. Vale."
- **Visual**: tabela.
- **Legenda-caps**: "LOCAL", "ZERO FILA".
- **Trilha**: tech 100 BPM.
- **VL pai**: V15.
- **CTA**: CTA-D.
- **Hashtags**: #ci #devops #devbr #harness #golang

#### Short 15.4 — "Como fecho release" (F4, LOOP)
- **Título**: `Fechando release em 60s`
- **Fórmula**: F4 + F2 (loop)
- **Duração**: 60s
- **Falas**:
  - 0–3s: "Feature fechada com harness em 60 segundos."
  - 3–20s: "Um: `harness ci` verde local."
  - 20–37s: "Dois: `harness ship "<prompt>" --plan <id>`."
  - 37–52s: "Três: `harness analytics` fecha o mês com custo real."
  - 52–60s: "Feature fechada com harness em 60 segundos."
    *(loop)*
- **Visual**: recording.
- **Legenda-caps**: "CI", "SHIP", "ANALYTICS".
- **Trilha**: tech 105 BPM.
- **VL pai**: V15.
- **CTA**: CTA-A.
- **Hashtags**: #devops #devbr #golang #harness #ship

---

## 3. Calendário Shorts — 90 dias

Cadência **1 Short por dia** durante os primeiros 90 dias. Fonte:
YouTube Creator Insider afirmou em 2024 que consistência diária ainda
é o principal driver de descoberta pra canal novo em Shorts
(`[VERIFICAR — URL exata do vídeo CI 2024`]). vidIQ 2025 corrobora
com dado de que canais que postam ≥5x/semana em Shorts têm 3x mais
chance de sair de 1k subs (`[VERIFICAR — URL exata]`).

Horário sugerido BR: **18h30–20h30** (janela de commuting +
after-work). Fonte: TubeBuddy 2025 relatório BR (`[VERIFICAR]`).

Layout 90 dias:

- **Dias 1–15 (semana âncora)**: um Short por VL, em ordem V1→V15.
  Cria coluna cronológica pareada com os longos.
- **Dias 16–30**: segundo Short de cada VL (Short x.2), mesma ordem.
  Reforça retenção do longo pai.
- **Dias 31–45**: Short x.3.
- **Dias 46–60**: Short x.4.
- **Dias 61–90**: repostagem cirúrgica dos **10 Shorts âncora**
  (seção 7) espaçados 3 dias entre si, misturados com Shorts novos
  reagindo a issues/PRs da comunidade.

Regra dura: nunca postar dois Shorts do mesmo VL em janela de 48h.
Canibaliza retenção (referência clássica CI, `[VERIFICAR]`).

---

## 4. Adaptação por plataforma

Fonte geral: Buffer Cross-Platform Report 2025 + Hootsuite Shorts vs
Reels vs TikTok 2025 (`[VERIFICAR — URLs`]).

### 4.1 TikTok

- Hook mais agressivo nos **primeiros 1s** (TikTok 2024 confirmou que
  o algo prioriza retenção nos 1s iniciais mais que o Shorts,
  `[VERIFICAR`]).
- Texto queimado **maior** (mínimo 72pt, safe area top 12%, bottom 20%
  pra caption e CTA nativos).
- Trilha: usar sound trend do dia se compatível com voz técnica. Se
  não, som próprio + música bem baixa.
- Duração: mesma (45–60s). TikTok promove até 10min hoje mas retenção
  cai.
- CTA "link na bio" funciona pior que "comenta X". Prefere gatilho.

### 4.2 Instagram Reels

- Trilha **trending Instagram** obrigatória (algoritmo IG ainda
  favorece áudio da própria plataforma em 2025, `[VERIFICAR — URL`]).
- Capa customizada 9:16 com texto grande — feed do IG mostra Reels
  como grid.
- CTA: "link na bio" funciona (Linktree/rogpe.tech).
- Duração: 45–60s ainda no sweet spot. IG puxa Reels ≥60s pra reels
  longos que competem em outra pool.

### 4.3 BlueSky / LinkedIn / X vídeo

- Sem música ou música baixíssima. Ambiente profissional.
- **Caption longa** substitui o áudio da trilha:
  - BlueSky: 300 chars, técnico direto, sem emoji.
  - LinkedIn: 1200–1800 chars com quebra a cada 2 linhas, 1 CTA no
    fim, hashtags mínimas.
  - X: thread de 2–4 tweets, primeiro sem link, segundo com link do
    long.
- Vídeo com legenda queimada é mandatório (feed silencioso).

---

## 5. Métricas alvo Shorts 2026

Alvos derivados de: YouTube Analytics benchmarks 2024 + relatórios
Tubular / vidIQ 2025 (`[VERIFICAR — URLs específicas`]).

| Métrica | Alvo | Crítico se |
|---|---|---|
| Retention 0–3s | **> 70%** | < 55% (Short morre no feed) |
| Average view duration | > 45s em Shorts de 55s | < 30s |
| View-through rate (assistiu > 90%) | **55%+** | < 40% |
| Comment rate | **≥ 2%** dos viewers | < 0,5% |
| Like rate | ≥ 5% | < 2% |
| Subs por 1k views (Shorts) | ≥ 3 | < 1 |
| Subs por 1k views (Long) | ≥ 8 | < 3 |
| CTR pra vídeo longo pai (link no comentário) | ≥ 4% | < 1% |

Nota: subs/1k Shorts historicamente é ~3–5x menor que Long
(YouTube CI 2024, `[VERIFICAR`]). Não use Shorts como métrica única
de crescimento — use como funil pro Long.

---

## 6. Loops de retenção — como prender

Pattern interrupt regra: **algo novo na tela a cada ≤5s**. Fonte:
Guo/Kim/Rubin L@S 2014 (aplicado a Shorts por analistas vidIQ 2024).

Técnicas usadas nos 60 Shorts acima:

1. **Zoom-in em código a cada 5s** — evita frame estático.
2. **Cut cortando última consoante** — mantém compressão temporal.
3. **Legenda-caps animada palavra a palavra** (não linha inteira) —
   olho segue.
4. **Curiosity gap no meio (~30s)**: "e o que fez cair pra US$ 1,23
   vem agora" — força passar do drop-off médio de 30s
   (`[VERIFICAR — dado CI 2024]`).
5. **Payoff no último quarto**: número, cost report, print de tag
   verde. Nunca deixa a promessa do hook sem entregar.
6. **Loop (F2)** em ~20% do catálogo — puxa view-through.
7. **Sem trilha alta nos primeiros 1s** — voz clara vale mais que
   beat. Trilha entra em 1,5s.
8. **CTA nunca antes de 40s** — CTA cedo mata retenção.

---

## 7. 10 Shorts âncora viralizáveis

Os que têm maior chance de estourar (número forte + hook técnico +
tema evergreen). Postar em cadência calibrada, 1 por semana durante
as 10 semanas do dia 61 ao 90 (mistura com novos).

| # | Short ID | Título | Por que âncora |
|---|---|---|---|
| A1 | 12.1 | `SaaS por US$ 1,23 em LLM` | Número extremo, tese do canal |
| A2 | 3.1 | `Codex errou meu JWT` | Polêmica LLM + segurança |
| A3 | 7.1 | `LLMLingua-2 corta 40% do prompt` | Paper + número prático |
| A4 | 2.1 | `Scaffold Go por 15 centavos` | Custo baixo, entrega alta |
| A5 | 10.1 | `Sem mem_limit teu compose trava` | Dor universal Docker |
| A6 | 5.1 | `BM25 salvou meu contexto` | Retrieval sem hype vector DB |
| A7 | 9.1 | `3 agentes pra 3 tipos de teste` | Multi-agente concreto |
| A8 | 13.1 | `Proposer, Critic, Decider` | Padrão de agente educativo |
| A9 | 4.1 | `CRUD por 1/3 do custo` | Roteamento LLM por tarefa |
| A10 | 8.1 | `Optimistic update React 19` | Frontend puro, react19 trending |

Regra: cada âncora recebe 3 variações A/B nos primeiros 30 dias após
postagem — mesma fala, thumb diferente, primeira frase reordenada.
Fonte: técnica de "iterative hook test" documentada por Rene Ritchie
2024 (`[VERIFICAR`]).

---

## Notas finais

- Nenhuma trilha específica é indicada — sempre use YouTube Audio
  Library ou compra Artlist com licença comercial no seu nome.
- Legenda queimada é obrigatória pra acessibilidade e feed mudo — não
  é opcional.
- CTA "comenta X" só usar quando você vai responder de verdade. Puxa
  reach, mas quebra confiança se ignora.
- Toda métrica marcada `[VERIFICAR]` precisa ser confirmada antes de
  virar afirmação em vídeo. Não invente número.
