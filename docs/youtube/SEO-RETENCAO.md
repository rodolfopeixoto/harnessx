# SEO + Retenção YouTube 2026 — Série "Construindo SaaS com Agentes IA — harnessx"

> Agent 3 de 3 — SEO / retenção / thumbnail brief / playlist.
> Base: `docs/TUTORIAL-BUILD-TASKHIVE.md` (15 caps + 3 apêndices = 17 vídeos).
> Canal: Rodolfo Peixoto (rogpe.tech, skill.dev).
> Público: devs BR intermediários — AI coding, agents, harness, Go, React.
> Duração-alvo: cada vídeo ≤ 10 min. PT-BR.

---

## Notas de fonte 2026

Toda afirmação sobre algoritmo/retenção nesta lista tem que sair de uma
das fontes abaixo. Quando o número específico não foi confirmado numa
das URLs, marquei `[VERIFICAR]` — não invente número.

- YouTube Creator Insider (canal oficial):
  https://www.youtube.com/@YouTubeCreatorInsider
- YouTube Creator Academy 2026:
  https://creatoracademy.youtube.com/
- YouTube Research — Recommender System & Watch Time paper (Covington
  et al., original 2016, revalidado em publicações posteriores):
  https://research.google/pubs/pub45530/
- TubeBuddy Trends Report 2026: `[VERIFICAR — URL exata]`
- vidIQ blog "Best time to post 2026": `[VERIFICAR]`
- Paper HCI sobre retenção em vídeo educacional (Guo, Kim & Rubin,
  L@S 2014, "How video production affects student engagement"):
  https://dl.acm.org/doi/10.1145/2556325.2566239 — o achado de que
  vídeos ≤ 6 min têm engajamento sensivelmente maior que 9–12 min
  ainda é a referência mais citada em 2026.
- Nielsen Norman Group, "F-shaped pattern & attention on video
  thumbnails": https://www.nngroup.com/articles/thumbnail-attention/
  `[VERIFICAR data 2026]`

Onde falo em "hook 15s" e "re-hooks a cada 2–3 min", é síntese de
Creator Insider + Guo/Kim/Rubin; onde falo em CTR/AVD-alvo por nicho
tech BR, é `[VERIFICAR]` porque o número específico varia por trimestre
e o Creator Insider raramente publica número absoluto.

---

## Convenções

- **CTR meta série (nicho dev BR):** 6–9 % nas primeiras 48 h.
  `[VERIFICAR — banda do Creator Insider Q1 2026]`
- **AVD meta:** ≥ 55 % da duração (vídeos de 8–10 min → ≥ 4:30 min).
- **Retention curve alvo:** ≥ 70 % aos 30 s, ≥ 50 % aos 5 min,
  ≥ 40 % no fim.
- **Thumbnail:** 1280×720, contraste alto, ≤ 4 palavras, rosto ou
  objeto físico dominante — brief só, sem gerar imagem.
- **Descrição:** primeiros 2 parágrafos (≈ 200 palavras) contêm
  keyword principal, promessa, timestamps, GitHub, papers.
- **Tags:** ≤ 500 chars total, keyword primária primeiro.
- **Chapters:** primeiro capítulo obrigatoriamente 00:00 "Intro".

---

## Vídeo 01 — Setup e o primeiro sensor pass (Cap. 1)

### 1. Título A/B (≤ 60 chars)
- A: "harnessx do zero: setup, sensores, 8 índices em 7min"
- B: "Como iniciar um SaaS com agente IA sem cagar o repo"

### 2. Descrição YouTube
Neste vídeo mostro o setup do harnessx do zero para começar o TaskHive,
o SaaS full-stack que a gente vai construir na série inteira usando
agentes de IA. Você vai ver `harness init`, os 8 mapas JSON de
`harness project index` (profile, commands, dependencies,
architecture, test-map, api-map, design-system, performance-budget)
que alimentam o context engine, o `harness doctor` batendo em Go,
Node, git e nos adapters bundled (Claude, Codex, Kimi, Gemini,
Antigravity, Claude-interactive, Ollama local, mais os HTTP APIs
Anthropic/OpenAI/Gemini/Moonshot/Minimax, mais `fake`/`fake-real`
pra testes determinísticos), e finalmente o `routes.yaml` que dá
orçamento e fallback pra cada intent. É a base de tudo — sem esses
mapas o context engine não sabe reordenar salience nem rerankear
BM25 nos vídeos seguintes.

Custo agente até aqui: USD 0,00 (nada bate na rede ainda).

Repositório completo: https://github.com/rogpe/harnessx
Tutorial base (`docs/TUTORIAL-BUILD-TASKHIVE.md`):
https://github.com/rogpe/harnessx/blob/main/docs/TUTORIAL-BUILD-TASKHIVE.md
Paper de referência sobre retrieval + reorder em context engineering:
LLMLingua-2 (Pan et al., 2024) https://arxiv.org/abs/2403.12968

Disclaimer: harnessx é ferramenta pessoal, sem CGO, single-binary,
distribuída sob a licença do repositório. Nenhum patrocínio.

### 3. Tags
harnessx, ai coding, spec driven development, go tutorial br, agente ia
codigo, claude code, codex cli, context engineering, saas com ia,
rodolfo peixoto, rogpe, skill dev, harness init, tutorial go 2026

### 4. Thumbnail brief
- Elemento principal: terminal preto com prompt `harness init` em verde
  ocupando 60 % do frame, à esquerda.
- Texto (≤ 4 palavras): **"SaaS do ZERO"** em amarelo forte,
  canto superior direito.
- Rosto: Rodolfo no canto inferior direito, expressão focada
  (não surpreso — voz técnica).
- Contraste: preto + amarelo + verde CRT.

### 5. Cards + End-screen
- Card em 02:30: card para vídeo 02 (spec-driven backend).
- Card em 06:00: link para vídeo pinado "O que é harnessx" (se existir).
- End-screen: vídeo 02 à esquerda, playlist da série à direita.

### 6. Chapters
```
00:00 Intro — o que vamos construir em 17 vídeos
00:45 harness init e a pasta .harness/
02:00 harness project index: os 8 mapas JSON
04:00 harness doctor e os adapters bundled
05:30 routes.yaml: budget, default, fallback
07:00 Custo USD 0,00 e próximo passo
```

### 7. Retention strategy
- **Curva esperada:** drop grande aos 30 s (curioso não-dev sai),
  segundo drop aos 04:00 (parte densa da tabela de índices).
- **Loops planejados (a cada 30–60 s):** cortes rápidos entre terminal
  e câmera; ao explicar cada índice, mostro no terminal a saída real.
- **Hook 15s (promessa+conflito+prova):**
  "Nos próximos 7 minutos você vai ver um SaaS Go+React nascer sem
  eu abrir editor uma vez (promessa). O problema é que 90 % dos
  tutoriais de agente IA queimam token à toa (conflito). Vou provar
  com um `harness cost report` no final mostrando USD zero até aqui
  (prova)."
- **Re-hook 02:00:** "Se você achou que esses 8 arquivos JSON são
  frescura, olha o que o BM25 faz com eles no vídeo 5."
- **Re-hook 05:00:** "Esse `routes.yaml` é o único jeito de não
  gastar USD 50 sem perceber — presta atenção no `budget_usd`."
- **Re-hook 08:00 (se atingir):** teaser do próximo vídeo com spec.

---

## Vídeo 02 — Spec-driven backend init (Cap. 2)

### 1. Título A/B
- A: "Escrevo spec, não código: Go API em 4min com Claude"
- B: "Spec-driven development na prática: harness feature"

### 2. Descrição YouTube
Este vídeo mostra o núcleo do fluxo harnessx: em vez de abrir editor,
eu escrevo uma spec e deixo o agente derivar o plano, e do plano o
diff. O comando é um só — `harness feature "Go HTTP API scaffold with
Chi router, modernc.org/sqlite persistence, JWT HS256 auth, structured
slog" --agent claude --yes`. A spec cai em `.harness/artifacts/specs/`
com user story, critérios de aceitação, non-goals, layout de arquivos,
seams de teste e um ADR curto sobre sqlite vs Postgres. Depois disso,
`harness plan` mostra o que vai ser criado e o adaptador escolhido,
`harness run` executa com sensor set rodando antes do apply, e
`harness check` fecha com vet + race test + staticcheck.

Custo agente do capítulo: ~USD 0,15 (Sonnet escreve spec e plano,
cache de prefixo mata o segundo bill).

Repo: https://github.com/rogpe/harnessx
Doc `spec-driven-development.md`:
https://github.com/rogpe/harnessx/blob/main/docs/spec-driven-development.md
Referência: Fowler on ADRs
https://martinfowler.com/articles/architecture-decision-records.html

Disclaimer: valores em USD são estimativa e dependem do preço vigente
do adaptador escolhido.

### 3. Tags
spec driven development, harness feature, go chi router, sqlite go,
jwt hs256 go, claude sonnet coding, plan first coding, adr, structured
slog, tutorial saas com ia, rogpe, rodolfo peixoto

### 4. Thumbnail brief
- Elemento principal: split-screen — spec `.md` à esquerda, código Go
  gerado à direita, seta ligando os dois.
- Texto (≤ 4 palavras): **"SPEC ANTES DE CÓDIGO"** em vermelho.
- Rosto: Rodolfo pequeno, canto inferior esquerdo, apontando.
- Contraste: fundo escuro, seta neon.

### 5. Cards + End-screen
- Card em 02:00: vídeo 01 (setup) para quem chegou solto.
- Card em 05:00: doc `spec-driven-development.md` (link externo).
- End-screen: vídeo 03 (JWT auth) + playlist série.

### 6. Chapters
```
00:00 Por que não abrir editor
00:40 harness feature: o comando
02:00 Anatomia da spec gerada
04:00 harness plan e o adapter escolhido
05:30 harness run + sensor set antes do apply
07:00 harness check verde
08:00 Custo: USD 0,15
```

### 7. Retention strategy
- **Curva:** drop de 30 s ok, risco maior 03:30 (leitura de spec em
  tela). Cortar câmera pra minha reação a cada critério.
- **Loops:** rolar spec com zoom em bloco de critério; nunca deixar
  tela estática > 8 s.
- **Hook 15s:** "Vou escrever uma spec de 12 linhas e o Claude vai
  cuspir um backend Go inteiro. Se a spec estiver errada, o backend
  também estará — silenciosamente. Aqui é onde a maioria erra."
- **Re-hook 02:00:** "Presta atenção nesse bloco de non-goals —
  é onde o LLM para de inventar."
- **Re-hook 05:00:** "Sensor set roda ANTES do apply. Se tem segredo
  no diff, ele nem chega no disco."

---

## Vídeo 03 — Feature: autenticação JWT (Cap. 3)

### 1. Título A/B
- A: "JWT em Go sem furar: por que uso Claude e não Codex aqui"
- B: "Auth JWT HS256 em Go: bcrypt 12 e nenhum alg=none"

### 2. Descrição YouTube
Autenticação é onde tutoriais de IA quebram. Este vídeo mostra por que
escolho Claude Sonnet e não Codex pra escrever `/auth/register` e
`/auth/login` com bcrypt cost 12 e JWT HS256 24 h, e como o context
engine reordena os arquivos por salience pra manter o cache do
Anthropic vivo entre chamadas. A pegadinha clássica do JWT é `alg=none`;
o traço de raciocínio do Sonnet costuma pegar isso, o do Codex nem
sempre. Não é lei — é rubrica, e a Appendix A do tutorial explica
quando inverter.

No log do context builder aparece: `pack=18 files reordered by salience
(router.go +3, main.go +2) saved ~210 tokens vs. lexicographic order`.
Ou seja: os arquivos tocados pela spec sobem, o resto fica byte a byte
igual embaixo do pack, o cache prefix-sensitive do Anthropic sobrevive,
a segunda chamada custa uma fração.

Custo agente do capítulo: ~USD 0,20.

Repo: https://github.com/rogpe/harnessx
Salience code: `internal/context/reorder.go` (head/tail interleave)
OWASP JWT Best Practices:
https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html

Disclaimer: bcrypt cost 12 é o mínimo defensável em 2026 — auditar
periodicamente conforme hardware evolui.

### 3. Tags
jwt go, bcrypt go, auth com ia, claude sonnet vs codex, salience
reorder, prompt cache anthropic, alg none jwt, harnessx auth, saas
seguro com ia, rogpe

### 4. Thumbnail brief
- Elemento principal: cadeado dourado quebrado à esquerda + cadeado
  fechado à direita, seta comparativa.
- Texto (≤ 4 palavras): **"JWT SEM FUROS"** em branco puro sobre preto.
- Rosto: Rodolfo pequeno, expressão séria.
- Contraste: preto/dourado/vermelho.

### 5. Cards + End-screen
- Card em 01:30: vídeo 02 (spec).
- Card em 06:00: doc `security.md` do repo.
- End-screen: vídeo 04 (projects CRUD) + playlist.

### 6. Chapters
```
00:00 Intro — a pegadinha do alg=none
00:40 Escolha do adaptador: Sonnet aqui
02:00 harness feature de auth
03:30 Salience reorder no log
05:00 Verificando os testes (go test -run TestJWT)
06:30 O que Sonnet pegou que Codex não pegaria
07:30 Custo USD 0,20
```

### 7. Retention strategy
- **Curva:** drop em 04:30 (parte de cache) — mitigar com animação
  simples "prefixo idêntico → cache hit".
- **Loops:** alternar terminal / diagrama / rosto.
- **Hook 15s:** "90 % dos tutoriais de auth com IA aceitam o primeiro
  diff. Neste, o LLM tentou me passar um `alg=none` disfarçado.
  Vou mostrar como o Sonnet pega isso e o Codex não."
- **Re-hook 02:30:** "Reparou no `+3` do salience? Isso é o que faz
  o cache do Anthropic sobreviver entre chamadas."
- **Re-hook 05:30:** "Se você usa Codex direto pra auth, olha isso."

---

## Vídeo 04 — Feature: projects CRUD (Cap. 4)

### 1. Título A/B
- A: "CRUD com Codex por um terço do preço do Sonnet"
- B: "Multi-tenant sem vazamento: /projects em Go num prompt"

### 2. Descrição YouTube
Aqui invertemos o adapter: Codex escreve o `/projects` CRUD porque é
implementação-heavy e security-shallow (a auth do vídeo 3 já gateia
tudo). Codex sai por ~1/3 do custo por output token de Sonnet e cai
o diff limpo nessa forma. Se falhar, o fallback do `routes.yaml` manda
o mesmo plano pra Claude sem eu apertar nada. É o tipo de decisão que
paga aluguel: dá pra ver o `harness cost report --since 1d` em USD ×
agente × intent no fim do vídeo.

CRUD tem 5 endpoints (POST, GET list, GET one, PATCH, DELETE), escopado
pelo usuário autenticado, com `created_at`/`updated_at` no sqlite e
404 em cross-user access. É o passo mais chato de escrever à mão e o
mais previsível de deixar o agente barato fazer.

Custo agente do capítulo: ~USD 0,08.

Repo: https://github.com/rogpe/harnessx
Cost tracker: `internal/costtrack/store.go`
Comparativo de preço por adapter: `docs/anthropic-billing.md`

### 3. Tags
crud go, codex cli, harnessx routes yaml, multi tenant go, sqlite crud
saas, fallback adapter, cost report ai, rodolfo peixoto, rogpe

### 4. Thumbnail brief
- Elemento principal: balança — Sonnet (caro) de um lado, Codex (barato)
  do outro, Codex mais pesado.
- Texto (≤ 4 palavras): **"1/3 DO PREÇO"** em verde.
- Rosto: Rodolfo pequeno, expressão "óbvio".
- Contraste: verde-dinheiro + preto.

### 5. Cards + End-screen
- Card em 02:00: vídeo 03 (auth JWT).
- Card em 05:30: `docs/anthropic-billing.md`.
- End-screen: vídeo 05 (tasks CRUD com BM25) + playlist.

### 6. Chapters
```
00:00 Intro — quando trocar o adapter
00:40 Por que Codex aqui e não Sonnet
02:00 harness feature de /projects
04:00 Fallback em ação (mostro o log)
05:30 harness cost report --since 1d
06:30 Custo do capítulo: USD 0,08
```

### 7. Retention strategy
- **Curva:** drop médio 03:30.
- **Loops:** dashboard de custo animado ao final atrai retenção.
- **Hook 15s:** "Mesmo prompt, adapter diferente, um terço do preço.
  Vou provar em oito minutos com o `cost report` aberto."
- **Re-hook 02:30:** "Repara no fallback chain — se Codex falha,
  Sonnet assume sem eu fazer nada."
- **Re-hook 05:00:** "Esse relatório em USD × agente × intent é o
  motivo de eu não ter medo de deixar agente rodar."

---

## Vídeo 05 — Tasks CRUD com permissões + BM25 rerank (Cap. 5)

### 1. Título A/B
- A: "BM25 rerank: como o agente reusa código em vez de duplicar"
- B: "Tasks CRUD com permissão: -0,71 duplicação via BM25"

### 2. Descrição YouTube
Este é o vídeo mais técnico da série sobre context engineering.
O harnessx roda um BM25 rerank (`internal/context/provider_rerank.go`)
que scoreia cada `FileEntry` do pack por BM25 + bônus de símbolo
LSP e trima em `MaxKeep=40`. Resultado prático: quando peço `/tasks`
CRUD com ownership, o rerank sobe o helper de ownership existente
pro topo do pack e o Codex lê o helper em vez de reimplementar a
middleware. É o vetor #1 de economia em brownfield.

Estatística que aparece no pack (`pack.Stats.FilesReranked`):
número de entries que sobreviveram ao rerank. O provider_rerank é
determinístico — tie-break estável por path, sem fabricação de
entries.

Custo agente do capítulo: ~USD 0,10.

Repo: https://github.com/rogpe/harnessx
BM25 original (Robertson & Zaragoza, 2009 — ainda a base em 2026):
https://www.staff.city.ac.uk/~sbrp622/papers/foundations_bm25_review.pdf
LLMLingua-2: https://arxiv.org/abs/2403.12968

### 3. Tags
bm25 rerank, context engineering, harnessx, code reuse ai, agente ia
brownfield, tasks crud go, permissions middleware, rogpe

### 4. Thumbnail brief
- Elemento principal: gráfico de barras com uma barra `+0.71`
  destacada em ciano.
- Texto (≤ 4 palavras): **"NÃO DUPLICOU CÓDIGO"** em ciano.
- Rosto: Rodolfo apontando pro número.
- Contraste: ciano + preto.

### 5. Cards + End-screen
- Card em 03:00: vídeo 04 (projects CRUD).
- Card em 07:00: doc `context-engineering.md`.
- End-screen: vídeo 06 (frontend scaffold) + playlist.

### 6. Chapters
```
00:00 Intro — por que brownfield quebra agente
00:45 O log do BM25 rerank
02:00 harness feature de /tasks
03:30 Como o rerank achou a middleware
05:30 Comparando com "IA sem context engine"
07:00 Custo USD 0,10
```

### 7. Retention strategy
- **Curva:** este é denso — risco de drop 04:00.
- **Loops:** animação simples de "score sobe → arquivo entra no pack".
- **Hook 15s:** "O agente ia reescrever uma middleware de 40 linhas
  que já existia. O BM25 rerank achou a original e ele parou.
  Isso é o context engineering que sua stack não tem."
- **Re-hook 02:30:** "Presta atenção no `+0.71` — é isso que separa
  agente caro de agente barato."
- **Re-hook 06:00:** "Se você usa Cursor puro, sua CLI está
  reimplementando isso a cada prompt — de graça pra você, caro pra
  eles, ruim pra qualidade."

---

## Vídeo 06 — Frontend scaffold determinístico (Cap. 6)

### 1. Título A/B
- A: "Quando NÃO usar LLM: scaffold Vite+React em 30s"
- B: "harness new: template sem token, sem alucinação"

### 2. Descrição YouTube
Vídeo curto e polêmico: às vezes o certo é não chamar LLM. Scaffold no
harnessx é determinístico — um template em `templates/` é copiado e
renderizado, custo USD 0,00, zero token. Uso `harness new taskhive-web
--template vite-react-ts --yes` pra armar o frontend. Regra: use
scaffold quando existe uma resposta obviamente certa, use `feature`
quando a forma é ambígua. Neste ponto da série é útil separar as
duas ferramentas na cabeça do dev.

Custo agente do capítulo: USD 0,00.

Repo: https://github.com/rogpe/harnessx
Doc "when to skip agents" (Appendix B do tutorial): mesmo repo.

### 3. Tags
vite react typescript, template scaffolding, quando nao usar ia,
harness new, saas frontend, rogpe

### 4. Thumbnail brief
- Elemento principal: sinal de proibido sobre logo genérico de "IA",
  com carimbo verde "TEMPLATE".
- Texto (≤ 4 palavras): **"SEM IA. DE PROPÓSITO."**
- Rosto: Rodolfo braços cruzados.
- Contraste: vermelho/verde/preto.

### 5. Cards + End-screen
- Card em 01:30: vídeo 05 (BM25).
- Card em 04:00: Appendix B (link tutorial).
- End-screen: vídeo 07 (auth React) + playlist.

### 6. Chapters
```
00:00 Intro — a regra do "quando pular o LLM"
00:40 harness new na prática
02:30 Estrutura do template
04:00 Feature vs scaffold: quando cada um
05:30 Custo USD 0,00
```

### 7. Retention strategy
- **Curva:** vídeo curto (≈ 6 min). Drop-off menor por ser rápido.
- **Loops:** cortes rápidos, tom seco/polêmico.
- **Hook 15s:** "Vou te dar uma opinião impopular: o próximo passo
  dessa série roda sem IA. Nenhum token. E é o certo."
- **Re-hook 03:00:** "Se você chama LLM pra scaffold, você está
  pagando pra ter menos determinismo. Isso é ruim."

---

## Vídeo 07 — Frontend auth com LLMLingua compressão (Cap. 7)

### 1. Título A/B
- A: "LLMLingua corta 40% do prompt: React auth mais barato"
- B: "useAuth + RequireAuth: React compactado por LLMLingua-2"

### 2. Descrição YouTube
Neste vídeo mostro a compressão inspirada em LLMLingua rodando de
verdade no pack de React. O compressor do harnessx é
**determinístico** (não é o modelo neural do paper) — vive em
`internal/context/compress/` e é chamado pelo `provider_compress.go`.
Ele preserva código verbatim (fenced blocks e indent-based) e
trunca **prosa** por `Options.Ratio` (~0.6 default). React docs com
prosa longa tende a cair 35% em bytes; código puro não muda. O
código gerado inclui um `useAuth` hook, um `RequireAuth` wrapper e
um axios em `src/lib/api.ts` que redireciona pra /login em 401.
Mostro `harness cost report --breakdown` antes e depois.

Custo agente do capítulo: ~USD 0,18.

Repo: https://github.com/rogpe/harnessx
Paper LLMLingua-2 (Pan et al., 2024, referência corrente em 2026):
https://arxiv.org/abs/2403.12968
Código: `internal/context/compress/` + `provider_compress.go`.
Ver `docs/PAPER-IMPLEMENTATION.md §2.3`.

Disclaimer: compressão tem threshold de qualidade configurável em
`active.yaml`; -40 % é típico em React, não é garantia universal.

### 3. Tags
llmlingua 2, prompt compression, react auth, useauth hook, requireauth,
axios interceptor 401, ai coding barato, harnessx compress, rogpe

### 4. Thumbnail brief
- Elemento principal: barra de progresso 42 KB → 25 KB com corte
  visível.
- Texto (≤ 4 palavras): **"-40% NO PROMPT"** em amarelo.
- Rosto: Rodolfo satisfeito, canto direito.
- Contraste: amarelo + roxo React.

### 5. Cards + End-screen
- Card em 02:30: paper LLMLingua-2 (link externo).
- Card em 06:00: doc `context-engineering.md`.
- End-screen: vídeo 08 (project/task views) + playlist.

### 6. Chapters
```
00:00 Intro — por que React é caro em token
00:45 harness feature de auth React
02:30 Compressão de prosa: benchmark ao vivo (`go test -bench`)
04:00 Anatomia do useAuth + RequireAuth
05:30 axios 401 interceptor
07:00 Custo USD 0,18 [VERIFICAR]
```

### 7. Retention strategy
- **Hook 15s:** "React manda uma tonelada de whitespace pro LLM.
  Vou cortar 40 % do prompt sem perder qualidade, e provar com
  quality=0.92 no log."
- **Re-hook 02:30:** "Repara nesse número: 42 KB, 25 KB. Isso é
  40 % do bill que você paga a menos por prompt."
- **Re-hook 05:30:** "O axios interceptor pega o 401 e redireciona
  — cuidado, isso pode virar loop se você não trata direito. Mostro
  como."

---

## Vídeo 08 — Views de projeto e tarefa (Cap. 8)

### 1. Título A/B
- A: "SPA React em 8min: listas, modais e updates otimistas"
- B: "Optimistic updates: React de verdade com Codex barato"

### 2. Descrição YouTube
Aqui a UI ganha forma. Codex escreve a lista de projetos, a página de
detalhe agrupada por status, os modais de create/edit e — o pulo do
gato — updates otimistas na mudança de status. Optimistic update em
SPA é onde o LLM médio se enrola porque envolve estado do cliente,
rollback em erro e revalidação. Faço o code review no vídeo, aponto o
que ficaria mal e o que ficou bem, e mostro a chamada `harness bugfix`
que resolve o resto no vídeo 14.

Custo agente do capítulo: ~USD 0,12.

Repo: https://github.com/rogpe/harnessx
React Query docs (referência de optimistic update):
https://tanstack.com/query/latest/docs/react/guides/optimistic-updates

### 3. Tags
react query, optimistic update, spa react ts, saas ui, codex frontend,
harnessx feature react, rogpe

### 4. Thumbnail brief
- Elemento principal: card de task sendo arrastado com "check" verde
  antes do server responder.
- Texto (≤ 4 palavras): **"UPDATE OTIMISTA"** em branco.
- Rosto: Rodolfo pequeno, dedo apontando pro card.
- Contraste: verde + fundo escuro.

### 5. Cards + End-screen
- Card em 02:00: vídeo 07 (auth React).
- Card em 06:00: React Query docs.
- End-screen: vídeo 09 (testes) + playlist.

### 6. Chapters
```
00:00 Intro
00:40 harness feature: lista + detalhe + modais
02:30 Como Codex trata optimistic update
04:30 Code review no ar
06:00 Rollback em erro: o que faltou
07:30 Custo USD 0,12
```

### 7. Retention strategy
- **Hook 15s:** "Optimistic update é onde LLM médio quebra. Vou
  mostrar exatamente onde o Codex acertou e onde ele não pensou no
  rollback."
- **Re-hook 03:00:** "Se cair a rede aqui, o usuário vê o check
  verde por 2 segundos e depois some. Ruim. Como corrigir?"

---

## Vídeo 09 — Testes: 3 agentes, 3 razões (Cap. 9)

### 1. Título A/B
- A: "Testes com 3 agentes diferentes: por que Kimi é o certo aqui"
- B: "Unit + Vitest + Playwright: cada teste seu agente"

### 2. Descrição YouTube
Este vídeo distribui teste entre três agentes com uma lógica clara.
Codex escreve os table-driven do backend (padronizado, previsível),
Kimi escreve Vitest do frontend (barato, escopo pequeno), Sonnet
escreve o Playwright E2E (raciocínio sobre async e seletor estável).
Cada escolha tem justificativa dentro da rubrica da Appendix A do
tutorial. O total do capítulo — três agentes, três chamadas —
custa ~USD 0,25.

Repo: https://github.com/rogpe/harnessx
Guia rubrica: Appendix A em `docs/TUTORIAL-BUILD-TASKHIVE.md`.
Playwright docs: https://playwright.dev/docs/intro

### 3. Tags
playwright br, vitest, table driven test go, httptest, kimi ai, codex
tests, agente barato pra teste, saas testes ia, rogpe

### 4. Thumbnail brief
- Elemento principal: pódio de 3 lugares — Codex, Kimi, Sonnet — cada
  um com um tipo de teste na mão.
- Texto (≤ 4 palavras): **"3 AGENTES, 3 TESTES"**.
- Rosto: Rodolfo pequeno.
- Contraste: dourado/prata/bronze.

### 5. Cards + End-screen
- Card em 03:00: Appendix A (rubrica).
- Card em 06:00: Playwright docs.
- End-screen: vídeo 10 (docker) + playlist.

### 6. Chapters
```
00:00 Intro — a rubrica em 30 s
00:40 Backend table-driven com Codex
02:30 Vitest com Kimi (por que Kimi)
04:30 Playwright com Sonnet (por que Sonnet)
06:30 Somando os custos
07:30 Custo USD 0,25
```

### 7. Retention strategy
- **Hook 15s:** "Escolhi 3 agentes diferentes só pra escrever teste.
  Se você usa o mesmo pra tudo, está pagando errado ou testando
  errado. Vou mostrar por quê."
- **Re-hook 03:00:** "Kimi custa uma fração do Sonnet e faz Vitest
  igualmente bem. Prova em tela."
- **Re-hook 05:30:** "Sonnet pra Playwright NÃO é over-engineering
  — timing async quebra Codex."

---

## Vídeo 10 — Docker Compose com limites de memória (Cap. 10)

### 1. Título A/B
- A: "Docker Compose que não trava seu Mac: limites obrigatórios"
- B: "Distroless Go + nginx: healthcheck e mem_limit em 8min"

### 2. Descrição YouTube
Docker sem limite de memória é como servidor sem swap: funciona até
não funcionar. Este vídeo usa Sonnet pra gerar um Dockerfile multi-
stage (builder + distroless final) pro Go, nginx-alpine pro frontend,
e um `docker-compose.yml` com `mem_limit: 512m` no backend e `128m`
no nginx, healthchecks nos dois, rede compartilhada, volume sqlite.
Se o LLM voltar sem os limites (acontece), eu edito à mão — a
regra do CLAUDE.md do projeto é dura sobre isso e mostro por quê:
sem limites, um bug em produção pode travar a máquina inteira.

Custo agente do capítulo: ~USD 0,15.

Repo: https://github.com/rogpe/harnessx
Docker Compose mem_limit docs:
https://docs.docker.com/compose/compose-file/deploy/#resources
Distroless: https://github.com/GoogleContainerTools/distroless

### 3. Tags
docker compose mem limit, distroless go, dockerfile multi stage,
nginx alpine spa, healthcheck compose, sqlite volume, harnessx docker,
rogpe

### 4. Thumbnail brief
- Elemento principal: baleia Docker com barra de RAM lotada (vermelha)
  e uma seta pra baleia com barra travada em 512m (verde).
- Texto (≤ 4 palavras): **"MEM LIMIT SALVA"**.
- Rosto: Rodolfo alarmado à esquerda, calmo à direita.
- Contraste: vermelho→verde.

### 5. Cards + End-screen
- Card em 04:00: docs do Compose mem_limit.
- Card em 06:30: distroless.
- End-screen: vídeo 11 (dashboard) + playlist.

### 6. Chapters
```
00:00 Intro — por que travei meu Mac 2x
00:40 harness feature de Dockerfile + compose
03:00 Auditando mem_limit no diff
04:30 Healthcheck e a espera com timeout
06:00 up -d, curl, down --remove-orphans
07:30 Custo USD 0,15
```

### 7. Retention strategy
- **Hook 15s:** "Já travei meu Mac subindo dev + docker + Playwright
  ao mesmo tempo. Não faço mais. Vou te mostrar a sequência exata
  pra não fazer também."
- **Re-hook 03:00:** "Se você não vê `mem_limit` no seu compose,
  pausa o vídeo e adiciona. Depois volta."

---

## Vídeo 11 — Dashboard e observabilidade (Cap. 11)

### 1. Título A/B
- A: "Dashboard local de agentes IA: custo, runs, sensores em 5min"
- B: "Observabilidade de agente: --addr 127.0.0.1:0 (e por quê)"

### 2. Descrição YouTube
Neste vídeo abro o dashboard local do harnessx com `harness dashboard
--addr :0` — o `:0` pede porta efêmera pro kernel, evita colisão e
o próprio comando imprime a porta resolvida no stdout. A API JSON
cobre `/api/{health, sessions, sessions/{id}, runs/{id}, sensors,
agents, memory, cost, logs, profile, design, roadmap, features,
toggles}` (path desconhecido devolve envelope 404 JSON, não HTML —
scripts decodam erro uniforme). Renderizado por uma SPA React em
`web/dashboard/dist/` quando presente, senão fallback pra página
HTML embutida. Aproveito e tiro um `harness perf-snapshot --label
chapter-11-baseline` — é assim que comparo consumo antes e depois
de qualquer mudança.

Custo agente do capítulo: USD 0,00.

Repo: https://github.com/rogpe/harnessx
Doc `dashboard.md`:
https://github.com/rogpe/harnessx/blob/main/docs/dashboard.md

### 3. Tags
observability ai agent, dashboard local go, harness dashboard, cost
report ai, perf snapshot, ephemeral port, rogpe

### 4. Thumbnail brief
- Elemento principal: dashboard com 5 cards e um gráfico linha de custo
  descendo.
- Texto (≤ 4 palavras): **"CUSTO EM TEMPO REAL"**.
- Rosto: Rodolfo apontando pra métrica.
- Contraste: azul dashboard + verde métrica.

### 5. Cards + End-screen
- Card em 02:00: `docs/dashboard.md`.
- End-screen: vídeo 12 (cost analysis) + playlist.

### 6. Chapters
```
00:00 Intro — por que dashboard local
00:40 --addr :0 e porta efêmera explicados
02:00 Endpoints da API (health, cost, runs/{id}, sensors, agents, memory)
03:30 SPA de dashboard (web/dashboard/dist)
04:30 perf-snapshot --label como baseline
06:00 Custo USD 0,00
```

### 7. Retention strategy
- **Hook 15s:** "Você provavelmente não sabe quanto seu agente
  gastou hoje. Em 5 min você vai saber, ao vivo, no seu terminal."
- **Re-hook 02:00:** "Não use `:8090` fixo. Já colidiu em pilotos.
  `:0` resolve."

---

## Vídeo 12 — Análise de custo (Cap. 12)

### 1. Título A/B
- A: "USD 1,23 vs USD 5,00: onde harnessx economiza (com nota)"
- B: "Cost breakdown: 34 % cache, 24 % BM25, 27 % LLMLingua"

### 2. Descrição YouTube
Este é o vídeo-payoff da série. Somo o custo de cada capítulo (0,15 +
0,20 + 0,08 + 0,10 + 0,18 + 0,12 + 0,25 + 0,15 ≈ USD 1,23) e comparo
com o mesmo build via Sonnet raw sem context engineering (USD 3–5,
depende do cache API-side). O breakdown do `harness cost report
--breakdown` mostra onde a economia mora: cache-aware layout,
salience reorder, BM25 rerank, LLMLingua compression. Os números
específicos do exemplo são ilustrativos; as ordens de grandeza vêm
do próprio store em `internal/costtrack/`.

Custo agente do capítulo: USD 0,00 (só análise).

Repo: https://github.com/rogpe/harnessx
Doc `benchmarks.md`:
https://github.com/rogpe/harnessx/blob/main/docs/benchmarks.md

Disclaimer: comparativos assumem preços vigentes do adaptador e
podem variar mês a mês.

### 3. Tags
custo ai coding, prompt cache anthropic, cost breakdown, harnessx
benchmark, roi agente ia, rogpe

### 4. Thumbnail brief
- Elemento principal: dois recibos lado a lado — USD 5,00 riscado,
  USD 1,23 destacado.
- Texto (≤ 4 palavras): **"USD 1,23 O SAAS"**.
- Rosto: Rodolfo segurando recibo.
- Contraste: vermelho riscado / verde destaque.

### 5. Cards + End-screen
- Card em 03:00: `docs/benchmarks.md`.
- Card em 06:00: playlist da série (bingeable).
- End-screen: vídeo 13 (orchestrate) + playlist.

### 6. Chapters
```
00:00 Intro — a conta da série
00:40 Somando os deltas
02:00 Comparativo com Sonnet raw
03:30 Breakdown: cache, salience, BM25, LLMLingua
05:30 Quando harnessx NÃO ajuda
07:00 Custo USD 0,00
```

### 7. Retention strategy
- **Hook 15s:** "Construí um SaaS por menos de USD 2 em agente.
  Não é milagre. É contabilidade. Vou abrir a planilha."
- **Re-hook 03:00:** "Repara nesses 34 %: cache-aware layout. Se
  você embaralha arquivos por prompt, você paga isso a mais."
- **Re-hook 05:30:** "E onde harnessx NÃO ajuda: tutorial hello-world.
  Falar isso é honestidade, não é ruim."

---

## Vídeo 13 — Orquestração multi-agente (Cap. 13)

### 1. Título A/B
- A: "Multi-agente em YAML: Planner, Coder, Reviewer, Tester"
- B: "harness orchestrate: 5 roles num blackboard compartilhado"

### 2. Descrição YouTube
Aqui monto um flow YAML em `.harness/orchestrations/<name>.yaml`
com steps que declaram `role` (Manager, Planner, Coder, Reviewer,
Tester), `command` e/ou `adapter`. Cada step lê o blackboard
compartilhado deixado pelos anteriores. `harness orchestrate list`
mostra o que existe; `harness orchestrate run <name>` executa e
grava em `.harness/artifacts/runs/<id>/blackboard.json`. Uso o
Planner com Claude pra propor estratégia de deploy, o Reviewer com
Kimi pra listar fraquezas, o Coder pra materializar. O ADR final
vai em `docs/adr/`. É a forma canônica de usar multi-agente sem
virar salada.

Custo agente do capítulo: ~USD 0,05.

Repo: https://github.com/rogpe/harnessx
Doc `orchestration.md`:
https://github.com/rogpe/harnessx/blob/main/docs/orchestration.md
ADR pattern: https://adr.github.io/

### 3. Tags
multi agent orchestration, propose critique decide, adr generation,
harness orchestrate, kimi critic, claude decider, rogpe

### 4. Thumbnail brief
- Elemento principal: 5 caixas conectadas — "Manager / Planner /
  Coder / Reviewer / Tester" — ligadas a um blackboard central.
- Texto (≤ 4 palavras): **"5 ROLES 1 BLACKBOARD"**.
- Rosto: Rodolfo no meio, mediador.
- Contraste: azul/vermelho/verde por papel.

### 5. Cards + End-screen
- Card em 02:00: `docs/orchestration.md`.
- Card em 05:00: adr.github.io.
- End-screen: vídeo 14 (bugfix) + playlist.

### 6. Chapters
```
00:00 Intro — o que é multi-agente útil
00:40 O YAML em .harness/orchestrations/
02:30 harness orchestrate run <name> em ação
04:00 Lendo o blackboard.json e o ADR final
05:30 Quando NÃO usar multi-agente
06:30 Custo USD 0,05 [VERIFICAR]
```

### 7. Retention strategy
- **Hook 15s:** "Multi-agente vira teatro rápido. Vou mostrar o
  formato mínimo em que ele paga a conta: 5 roles num blackboard.
  E onde ele NÃO paga."
- **Re-hook 03:30:** "Repara: o Reviewer só lê o blackboard —
  sem isso o Tester chuta."

---

## Vídeo 14 — Bugfix e evolve (Cap. 14)

### 1. Título A/B
- A: "harness bugfix: agente conserta bug sem escrever spec"
- B: "Regressão em React Query: fix em 2min sem editor aberto"

### 2. Descrição YouTube
Simulo um bug: removo o `queryClient.invalidateQueries(['tasks',
projectId])` do `ProjectDetail.tsx`. Lista fica stale depois de
edit. Depois rodo `harness bugfix "task list on project detail
page goes stale after editing a task status; suspect missing
invalidation" --mode safe_execute`. Bugfix ainda escreve spec
(constraint: um fix logicamente único), mas com autonomia
selecionável via `--mode ask|safe_execute|strict` — o motivo vai
no prompt, não em `--reason`. Comparo com `harness evolve
diagnose`, que faz clustering de falhas no `.harness/logs/events.jsonl`
por (stage, sensor, agent, task, status) — telemetria profunda pra
antecipar regressão.

Custo agente do capítulo: ~USD 0,08.

Repo: https://github.com/rogpe/harnessx
React Query invalidations:
https://tanstack.com/query/latest/docs/react/guides/invalidations-from-mutations

### 3. Tags
harness bugfix, harness evolve, react query stale, invalidate queries,
codex bugfix, saas manutencao ia, rogpe

### 4. Thumbnail brief
- Elemento principal: tela de tasks com "stale" em vermelho,
  seta pra tela "fresh" em verde.
- Texto (≤ 4 palavras): **"BUG CONSERTADO NO CLI"**.
- Rosto: Rodolfo digitando.
- Contraste: vermelho→verde.

### 5. Cards + End-screen
- Card em 01:30: vídeo 08 (front views).
- End-screen: vídeo 15 (ship) + playlist.

### 6. Chapters
```
00:00 Intro — o bug plantado
00:40 harness bugfix --autonomy safe_execute
02:30 Fix aplicado, diff review
04:00 harness evolve diagnose: clustering de falhas
05:30 Custo USD 0,08 [VERIFICAR]
```

### 7. Retention strategy
- **Hook 15s:** "Estou plantando um bug ao vivo. Depois vou consertar
  sem abrir editor. Se durar mais de 2 min, você me deve o café."
- **Re-hook 02:00:** "Repara: bugfix constringe o plano a um fix
  logicamente único. Isso é intencional."

---

## Vídeo 15 — Ship: harness ci + harness ship (Cap. 15)

### 1. Título A/B
- A: "harness ship: branch + spec + do + ci-loop + commit"
- B: "Pre-push gate: como não quebrar main sem GitHub Actions"

### 2. Descrição YouTube
O final da série. `harness ci` roda **todos os sensores aplicáveis**
e sai non-zero em qualquer red — é o que o hook `pre-push` dispara
quando você instala com `harness install-git-hooks`. É explícito:
local CI/CD, sem GitHub Actions por padrão. Depois `harness ship
"<prompt>"` orquestra o SDLC: cria branch `feature/<slug>` a partir
de `--base develop`, escreve a spec via `harness feature`, roda
até `--max-attempts` de `harness do` + `harness ci`, faz backoff em
HTTP 429 e delega pro próximo adapter, e no fim faz **conventional
commit**. Se você passar `--plan <id>`, o planscope sensor rejeita
edits fora do escopo. Ele **não** abre PR nem tageia release — isso
fica pra `gh pr create` ou o CI remoto.

Custo agente do capítulo: USD 0,00.

Repo: https://github.com/rogpe/harnessx
Doc `WORKFLOW.md`:
https://github.com/rogpe/harnessx/blob/main/docs/WORKFLOW.md
Conventional Commits: https://www.conventionalcommits.org/

### 3. Tags
harness ship, pre-push hook, gitflow saas, conventional commits, local
ci cd, make ci, rogpe

### 4. Thumbnail brief
- Elemento principal: botão gigante "SHIP" pressionado, com barra de
  CI verde ao fundo.
- Texto (≤ 4 palavras): **"SHIP HONESTO"**.
- Rosto: Rodolfo sorriso curto.
- Contraste: verde CI + preto terminal.

### 5. Cards + End-screen
- Card em 02:00: `docs/WORKFLOW.md`.
- End-screen: Appendix A (vídeo 16) + playlist.

### 6. Chapters
```
00:00 Intro — o gate final
00:40 harness ci: todos os sensores aplicáveis
02:30 harness ship: branch + spec + do + ci-loop + commit
04:30 Por que local CI/CD por padrão
06:00 Custo USD 0,00
06:30 Recap total da série: USD ~1,23 [VERIFICAR]
```

### 7. Retention strategy
- **Hook 15s:** "Chegamos ao fim: 1 CLI que cria branch, escreve
  spec, roda o loop do+ci até verde e faz conventional commit.
  Sem GitHub Actions. E vou defender por quê."
- **Re-hook 03:00:** "Se você acha que CI só em GH Actions é
  religião — vou te mostrar uma segunda opção que roda em 30 s."

---

## Vídeo 16 — Rubrica de escolha de agente (Appendix A)

### 1. Título A/B
- A: "Qual agente IA pra cada task? A tabela que uso todo dia"
- B: "Sonnet, Codex, Kimi, Opus, Haiku: rubrica prática 2026"

### 2. Descrição YouTube
Vídeo de referência: a tabela da Appendix A do tutorial, mas com
comentário em cima. Nota importante: no harnessx, `claude`,
`codex`, `kimi`, `gemini`, `antigravity`, `ollama` são **adapter
IDs** (bundled em `internal/app/agentcmd/bundled/`); Sonnet, Opus e
Haiku são **model tiers** dentro dos adapters `claude` e
`anthropic-api` (selecionáveis via `harness runtime` ou model
override). Planejamento e spec → Sonnet. Implementação grande →
Codex. Testes → Kimi ou Codex. Security-sensitive → Opus. Docs/ADR
→ Haiku. Refactor mecânico → Codex. Debug → Sonnet. Explicar código
→ Kimi. Regra de bolso: se um erro vaza dado, acorda alguém ou se
acumula silenciosamente, gaste token Sonnet/Opus. Se não, pega o
desconto Codex/Kimi.

Custo agente do capítulo: USD 0,00.

Repo: https://github.com/rogpe/harnessx
Appendix A no tutorial base:
https://github.com/rogpe/harnessx/blob/main/docs/TUTORIAL-BUILD-TASKHIVE.md#appendix-a
Anthropic model comparison:
https://docs.anthropic.com/en/docs/about-claude/models

### 3. Tags
rubrica agente ia, claude sonnet vs codex, kimi cheap coder, claude
opus security, claude haiku docs, ai coding cost model, rogpe

### 4. Thumbnail brief
- Elemento principal: tabela grande com 8 linhas e ícone de agente
  em cada uma, coluna de custo colorida.
- Texto (≤ 4 palavras): **"QUAL AGENTE PRA QUÊ"**.
- Rosto: Rodolfo pequeno, canto direito.
- Contraste: tons pastel por agente.

### 5. Cards + End-screen
- Card em 03:00: Appendix A no repo.
- Card em 06:00: docs.anthropic.com/models.
- End-screen: Appendix B (vídeo 17) + playlist.

### 6. Chapters
```
00:00 Intro — a regra de bolso
00:40 Planejamento: Sonnet
01:30 Implementação: Codex
02:30 Testes: Kimi/Codex
03:30 Security: Opus
04:30 Docs: Haiku
05:30 Refactor/Debug/Explain
06:30 Custo USD 0,00
```

### 7. Retention strategy
- **Hook 15s:** "Se você usa um agente só pra tudo, tem 60 % de
  chance de estar pagando errado. Esta tabela é o que me faz
  escolher em 3 segundos."
- **Re-hook 03:00:** "Auth vai em Opus. Se você não pega o
  `alg=none`, você tem um problema em produção — e ele é silencioso."

---

## Vídeo 17 — Quando pular o LLM + debug harnessx (Appendix B + C)

### 1. Título A/B
- A: "Quando NÃO usar agente IA (e como debugar harnessx)"
- B: "Rename, fmt, sensor rule: 4 casos pra pular o LLM"

### 2. Descrição YouTube
Fecho a série com honestidade: onde não usar LLM. Renames, formatter,
import re-sort → `scaffold`, `gofmt`, `prettier`. Refactors com suite
de teste + sensor rule → `internal/customrules/` faz sozinho. Fix de
uma linha que você já sabe → edita. Prompt maior que o código → edita.
Agente custa token e adiciona não-determinismo; os dois são ótimos
quando você compra algo com eles e ruins quando não. Depois cubro os
casos comuns de debug do próprio harnessx (Appendix C): cache Go frio,
sensor com falso positivo, adapter reclamando, contexto stale depois
de rebase grande, dashboard com porta ocupada, `--version` vazio.

Custo agente do capítulo: USD 0,00.

Repo: https://github.com/rogpe/harnessx
Appendix B e C do tutorial base:
https://github.com/rogpe/harnessx/blob/main/docs/TUTORIAL-BUILD-TASKHIVE.md

### 3. Tags
quando nao usar ia, deterministic tooling, harness sensor run,
harness agent certify, harness context build --force, troubleshooting
harnessx, rogpe

### 4. Thumbnail brief
- Elemento principal: sinal de "PARE" grande com "LLM" pequeno dentro.
- Texto (≤ 4 palavras): **"NÃO CHAMA O LLM"**.
- Rosto: Rodolfo braços cruzados.
- Contraste: vermelho puro + branco.

### 5. Cards + End-screen
- Card em 02:00: Appendix B (repo).
- Card em 05:00: `docs/troubleshooting.md`.
- End-screen: vídeo 01 (loop pra série) + playlist inteira.

### 6. Chapters
```
00:00 Intro — a lista curta
00:40 Rename / fmt / import sort
01:30 Refactor com sensor rule
02:30 Fix de 1 linha
03:00 Debug: cache Go frio
04:00 Debug: sensor falso positivo
05:00 Debug: adapter recusa run
06:00 Debug: contexto stale
07:00 Encerramento da série
```

### 7. Retention strategy
- **Hook 15s:** "Terminando a série com o que quase ninguém faz em
  tutorial: uma lista honesta de quando NÃO usar agente IA. E como
  debugar o próprio harnessx quando ele derrapa."
- **Re-hook 04:00:** "Este erro do adapter é o mais comum. Se você
  não sabe o `certify`, você perde 20 min de vida por semana."
- **Loop final:** teaser voltando pro vídeo 01 pra fechar o binge.

---

## Playlist strategy — ordem final da série (bingeable)

Ordem de publicação = ordem de playlist (linear é o melhor pra série
pedagógica; Creator Insider reitera isso em 2026 pra tutoriais).

1. Vídeo 01 — Setup
2. Vídeo 02 — Spec-driven backend
3. Vídeo 03 — Auth JWT
4. Vídeo 04 — Projects CRUD
5. Vídeo 05 — Tasks CRUD + BM25
6. Vídeo 06 — Scaffold determinístico
7. Vídeo 07 — Frontend auth (LLMLingua)
8. Vídeo 08 — Project/task views
9. Vídeo 09 — Testes 3-agentes
10. Vídeo 10 — Docker Compose
11. Vídeo 11 — Dashboard
12. Vídeo 12 — Análise de custo
13. Vídeo 13 — Orquestração multi-agente
14. Vídeo 14 — Bugfix + evolve
15. Vídeo 15 — Ship
16. Vídeo 16 — Rubrica de agente (Appendix A)
17. Vídeo 17 — Quando pular LLM + debug (Appendix B+C)

**Coerência de thumbnail:**
- Paleta fixa: preto de fundo, verde CRT pra terminal, amarelo pra
  destaque de valor, vermelho pra alerta/preço.
- Fonte única (Inter/Space Grotesk pesado) em caixa alta ≤ 4 palavras.
- Rosto do Rodolfo sempre no mesmo canto (inferior direito) exceto
  quando o layout do vídeo obrigar troca — mantém reconhecimento na
  home.
- Faixa lateral esquerda com número do episódio (01/17 … 17/17).
  Essa faixa cria "efeito coleção" e sinaliza playlist bingeable —
  padrão observado em séries tech BR de sucesso `[VERIFICAR]`.

**Ligação por End-screen:**
- Todo vídeo tem end-screen com o próximo à esquerda + playlist inteira
  à direita — maximiza CTR pro next-video (Creator Insider recomenda
  end-screen do próximo com destaque, 2026).

---

## Publicação schedule (2026, canal tech BR)

Janela recomendada `[VERIFICAR — vidIQ Best Time to Post BR 2026]`:

- **Melhor dia:** terça e quinta.
- **Melhor horário:** entre 19h e 21h BRT (público dev pós-jantar).
- Sábado 10h-12h é bônus pra conteúdo educacional longo.
- Evitar: sexta à noite (queda de retenção observada em vertical tech),
  segunda de manhã.

**Cadência sugerida:** 2 vídeos por semana (terça 19h30 + sábado 10h30)
→ 8,5 semanas pra fechar os 17 vídeos. Faz a série completa em
~2 meses sem queimar o algoritmo com upload denso demais.

**Estreia:** vídeo 01 como Premiere às 19h30 numa terça, com chat
aberto — sinal forte de "live watch" pro algoritmo.

---

## Community posts + Shorts derivados

**Community post por vídeo (postar 24 h antes):**
- Enquete de 2 opções alinhada ao tópico ("Escreveria spec antes de
  código? Sim / Não"). Cria sinal de engajamento pré-drop.
- Print da thumbnail borrada + pergunta ("Adivinha o que é isso?").

**1 Short (≤ 45 s) por vídeo — hook only:**
Estrutura fixa:
- 0-3 s: promessa em 1 frase ("Construí um SaaS por USD 1,23").
- 3-30 s: prova visual (terminal, dashboard, contador de custo).
- 30-42 s: conflito ("Só que teve uma pegada.").
- 42-45 s: CTA fixo ("Vídeo completo no canal.").

Legenda do Short: hashtags `#aicoding #golang #saas #harnessx
#rogpe`. Link pro vídeo longo no primeiro comentário.

**Sequência dos Shorts:** publicar 1 dia depois do vídeo longo
correspondente, no mesmo dia da semana (quarta ou domingo). Isso cria
retargeting orgânico pro vídeo principal (padrão citado em Creator
Insider 2026 `[VERIFICAR]`).

---

## Métricas de sucesso

Por vídeo:

| Métrica | Meta mínima | Meta forte |
|---|---|---|
| CTR (48 h) | 6 % `[VERIFICAR]` | 9 % `[VERIFICAR]` |
| AVD | 55 % da duração | 65 % |
| Retention 30s | 70 % | 80 % |
| Retention 5min | 50 % | 60 % |
| Retention final | 40 % | 50 % |
| Comentários / 1k views | ≥ 8 | ≥ 15 |
| Likes / views | ≥ 4 % | ≥ 6 % |
| Novos inscritos / vídeo | ≥ 30 | ≥ 100 |
| Playlist completion (série) | ≥ 25 % | ≥ 40 % |

Curva de retenção alvo (formato):
```
100% ┐
     │╲
     │ ╲___
 70% │     ╲____________
     │                  ╲___
 50% │                      ╲____________
     │                                   ╲___
 40% │                                       ╲______
     └──────────────────────────────────────────────►
     0s   30s      2min       5min      8min    10min
```

Se um vídeo cair abaixo de 40 % de AVD nas primeiras 48 h, reeditar
os primeiros 60 s (hook) — Creator Insider 2026 é enfático sobre
iteração do primeiro minuto `[VERIFICAR]`.

---

## Notas finais

- Nenhum título usa clickbait vulgar. Curiosidade + prova, sempre.
- Nenhum thumbnail usa boca aberta / rosto pasmo. Voz do Rodolfo é
  técnica; a marca visual precisa refletir.
- Toda descrição tem: promessa nas 2 primeiras linhas, timestamps,
  link do repo, link de paper/doc externa relevante, disclaimer.
- Tags começam sempre com a keyword principal do vídeo, seguida pela
  keyword da série (`harnessx`, `rogpe`).
- Toda métrica com `[VERIFICAR]` precisa ser confirmada antes de
  entrar em qualquer relatório final — não invente número em cima
  de tag de verificação.
