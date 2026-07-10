# BRAND / THUMB / TITLE MATRIX — Canal Rodolfo Peixoto

> Direção de arte + copy de título/thumb para canal técnico BR.
> Foco: harnessx (runtime Go local pra agentes IA), papers na prática,
> SaaS com IA. Público-alvo: dev BR intermediário/sênior.
> Voz: direta, técnica, sem enchimento. Zero "galera", zero "vamos".
> Entregável irmão de `ROTEIROS-TECNICOS.md`, `SEO-RETENCAO.md`,
> `DIDATICA-CONCEITOS.md`.

---

## 1. Sistema visual do canal

### 1.1 Nome oficial da série (3 propostas)

| Proposta | Prós | Contras |
|---|---|---|
| **A. "Construindo com Agentes"** | Curto, cabe em faixa lateral (18 chars), abraça as 5 playlists (não só TaskHive). Casa com o subtítulo `rogpe.tech`. | Genérico demais isolado — depende do canal ter tração. |
| **B. "harnessx na prática"** | Amarra a marca do runtime. Bom pra SEO long-tail. Alinha com `skill.dev`. | Limita: se harnessx pivotar ou perder tração, o nome fica preso. |
| **C. "Custo × Contexto"** | Nome-hook: já entrega a tese (contexto engineering derruba custo). Ótimo pra série de análise/papers. | Menos claro pra iniciante que ainda não conhece "context engineering". |

**Recomendação:** **A** como guarda-chuva do canal, **B** só na playlist
PL1 (TaskHive) como subtítulo. `[VERIFICAR: disponibilidade dos handles
@construindocomagentes no YT/IG antes de fixar]`.

### 1.2 Logo — brief visual (sem gerar imagem)

- **Símbolo:** monograma `Rp` estilizado em serifa geométrica moderna,
  com o `p` cortado por uma linha horizontal fina que sugere prompt (`>`)
  do terminal. Peso 700+, altura 1:1.
- **Wordmark:** "rogpe" em lowercase, JetBrains Mono Bold, kerning
  levemente aberto (+20). Nunca usar caps.
- **Variação avatar 800×800:** só o monograma `Rp` sobre fundo preto
  puro (`#0B0B0F`), com filete verde CRT (`#00E38A`) 4px na base.
- **Banner canal 2560×1440:** monograma à esquerda, wordmark central,
  tagline "agentes IA, custo real, código público" à direita em Inter
  Medium 32pt. Safe zone TV: manter tudo dentro de 1546×423 central.
- **Regras rígidas:** logo nunca esticada, nunca com sombra, nunca com
  gradiente. Aplicar sempre em preto puro ou fundo terminal.

### 1.3 Paleta oficial (5 cores)

| Papel | Hex | Uso |
|---|---|---|
| Primária (fundo) | `#0B0B0F` | Fundo terminal, thumb base, background dashboard |
| Secundária (marca) | `#00E38A` | Prompt CRT, wordmark accent, highlight de sucesso |
| Alerta / preço alto | `#FF3B3B` | Riscados, erros, "não faça isso", preço competidor |
| Sucesso / preço baixo | `#7CF74A` | Deltas positivos, ✓, custo economizado |
| Neutro base (texto) | `#F4F4F5` | Texto em thumb, subtítulos, corpo em card |

Acento auxiliar opcional: **amarelo destaque** `#FFD400` — usar só em
título de thumbnail (≤ 4 palavras). Nunca misturar amarelo + verde
CRT no mesmo eixo horizontal (competem por atenção).

**Fonte da paleta:** derivada da convenção já em uso nos briefs de
`SEO-RETENCAO.md` (preto + verde CRT + amarelo destaque + vermelho
alerta). Cores conferidas pra contraste WCAG AA sobre `#0B0B0F`
(mínimo 4.5:1 para texto normal — `[VERIFICAR com contrast checker
oficial WebAIM antes de fechar brand book]`).

### 1.4 Tipografia

Ambas as opções são **Google Fonts, licença SIL OFL 1.1** (livre pra
uso comercial e embed) — `[VERIFICAR licenças em fonts.google.com
antes de qualquer contrato de patrocínio 2026]`.

**Opção 1 (recomendada):**
- **Display / thumbnail:** JetBrains Mono — pesos 700 (Bold) e 800
  (ExtraBold). Fonte oficial: https://www.jetbrains.com/lp/mono/
  (licença OFL, confirmada 2024, `[VERIFICAR 2026]`).
- **Body / lower thirds / descrição em vídeo:** Inter — pesos 400
  (Regular), 600 (Semibold), 800 (Black). Fonte oficial:
  https://rsms.me/inter/ (licença OFL, confirmada 2024,
  `[VERIFICAR 2026]`).

**Opção 2 (alternativa mais "editorial"):**
- **Display:** Space Grotesk 700 — mais suave, funciona bem em card
  de paper. Licença OFL, `[VERIFICAR]`.
- **Body:** IBM Plex Sans 500 — pareamento IBM/OFL, `[VERIFICAR]`.

**Regra:** nunca misturar Opção 1 com Opção 2 na mesma temporada.

### 1.5 Grid thumbnail 1280×720

Zonas seguras (16:9, considerando corte mobile YouTube ~5 % nas bordas):

```
+------------------------------------------------------+  ← 720
|  ┌─────────────────────────────────────────────┐    |
|  │  SAFE ZONE MOBILE (1216 × 684)              │    |
|  │  ┌───────────────────────┬───────────────┐  │    |
|  │  │                       │  TEXT ZONE    │  │    |
|  │  │  ELEMENTO DOMINANTE   │  ≤ 4 palavras │  │    |
|  │  │  (terminal / diagrama)│  (canto sup   │  │    |
|  │  │  ~60% da largura      │   direito)    │  │    |
|  │  │                       │  ~35%         │  │    |
|  │  │                       ├───────────────┤  │    |
|  │  │                       │  ROSTO        │  │    |
|  │  │                       │  Rodolfo      │  │    |
|  │  │                       │  (canto inf   │  │    |
|  │  │                       │   direito)    │  │    |
|  │  │                       │  ≥ 220×220 px │  │    |
|  │  └───────────────────────┴───────────────┘  │    |
|  │  FAIXA LATERAL EPISÓDIO (64px, esquerda)    │    |
|  └─────────────────────────────────────────────┘    |
+------------------------------------------------------+
 ↑ 1280
```

**Regras hard:**
- Texto acima de 96pt. Contorno externo escuro 4px pra sobreviver
  em modo escuro do app.
- Rosto ocupa **no mínimo 220×220 px** (senão vira ponto na home
  mobile — erro clássico BR).
- Faixa lateral esquerda 64px com número do episódio (`PL1·03`, etc)
  em JetBrains Mono 32pt sobre `#00E38A`.
- Nenhum texto pode passar sobre o rosto. Ponto.
- Máximo 3 cores fortes por thumb (preto + 2 do sistema).

### 1.6 Intro 5s — template

**Roteiro (leitura sobre imagem):**
> "rogpe. Agentes IA na prática."

**Visual (5 keyframes):**
1. `0.0s` — Fundo `#0B0B0F`, monograma `Rp` fade-in centro (0→100 em 500ms).
2. `1.0s` — Terminal fake datilografa `> harness init` em verde CRT.
3. `2.5s` — Cursor pisca, wordmark "rogpe" aparece embaixo.
4. `3.5s` — Título do vídeo entra em Inter Black caps ("EPISÓDIO XX").
5. `5.0s` — Corte seco pro hook. Zero fade-out.

Áudio: single tick de mecânico (Cherry MX) + baixo sub grave 55Hz
0.3s. Sem música.

### 1.7 Outro 10s — template

**Elementos obrigatórios:**
- **0–2s:** frase de recap ("Custo: USD X. Repositório: link").
- **2–7s:** end-screen padrão YouTube — próximo vídeo à ESQUERDA
  (60 % de peso), playlist à DIREITA (30 %), botão inscrever no
  canto inferior esquerdo (10 %).
- **7–10s:** logo estático + call `rogpe.tech / skill.dev`, silêncio.

**CTA falado:** "Próximo vídeo. Playlist. Inscreve." — três frases
curtas, sem "não esquece de", "por favor", "se gostou".

**Regra CLAUDE.md:** end-screen nunca cobre código exibido nos
últimos 10s. Reservar tela limpa.

---

## 2. Fórmulas de título (10 fórmulas testadas 2026)

Base: consolidação de padrões observados em Fireship / ThePrimeagen /
Filipe Deschamps + Creator Insider 2026 `[VERIFICAR]`. Sempre ≤ 60
chars (corte mobile YouTube).

| # | Fórmula | Exemplo |
|---|---|---|
| F1 | `[Número] + [subst forte] + em [tempo]` | "10 sensors que salvam PRs em 30s" |
| F2 | `Como fiz [X] por [$Y]` | "Como fiz um SaaS por US$ 1,23" |
| F3 | `[Ferramenta A] vs [Ferramenta B] em [tarefa]` | "Sonnet vs Codex em CRUD Go" |
| F4 | `Não use [X] antes de ver isso` | "Não use JWT antes de ver isso" |
| F5 | `O paper que [resultado inesperado]` | "O paper que cortou 40% do meu prompt" |
| F6 | `Eu quebrei [X] usando [Y]` | "Quebrei o Codex com um sensor rule" |
| F7 | `Por que [afirmação polêmica]` | "Por que scaffold é melhor que LLM" |
| F8 | `Setup completo [stack] em [tempo]` | "Setup harnessx + Go em 5min" |
| F9 | `[Termo técnico] explicado por [analogia]` | "BM25 explicado por bibliotecário" |
| F10 | `Isso mudou meu jeito de [ação]` | "Isso mudou meu jeito de escolher agente" |

**Regras de aplicação:**
- Testar A/B três variantes por vídeo (uma sempre com número, uma
  sempre com $, uma sempre com verbo forte).
- Nunca começar com "Como" duas vezes seguidas na playlist.
- Se o vídeo é sobre paper → F5. Se é sobre custo → F2 ou F3.
  Se é sobre erro/decisão → F4 ou F6.

---

## 3. Fórmulas de thumbnail (7 fórmulas)

| # | Fórmula | Quando usar |
|---|---|---|
| T1 | **Rosto + expressão + número gigante** | Vídeos de resultado/prova (custo, benchmark, "eu fiz X"). |
| T2 | **Split comparativo (esq vs dir)** | Ferramenta A vs B, antes vs depois de compressão. |
| T3 | **Terminal + highlight vermelho** | Setup, comando, feature. Base do canal. |
| T4 | **Diagrama simples + rosto canto** | Papers, arquitetura, orquestração multi-agente. |
| T5 | **Antes/depois** | Bug fix, refactor, otimização, migração. |
| T6 | **Objeto real + código atrás** | SaaS físico/hardware (R36S), livro de paper, keyboard. |
| T7 | **Só texto** | Só quando a frase é boa demais pra dividir espaço. Ex.: "USD 1,23". Usar no máximo 1x por playlist. |

**Regras hard:**
- T1, T2, T3 = 70 % dos vídeos do canal (padrão previsível ajuda CTR
  em série `[VERIFICAR — Creator Insider 2026]`).
- Nunca boca aberta / expressão de choque. Voz Rodolfo é técnica.
- Nunca seta amarela random. Setas só em T2 (split comparativo).
- Nunca thumbnail com screenshot puro do VSCode — vira mancha na home.

---

## 4. Matriz 60 vídeos

Formato: `PLx-NN | Título A/B/C + 2 thumbs briefs + hipótese CTR`.

### PL1 — TaskHive com harnessx (17 vídeos)

Títulos-base já em `SEO-RETENCAO.md`. Aqui expando pra A/B/C com fórmulas
diferentes.

---

**PL1-01 — Setup + primeiro sensor pass**
- **A (F8):** "Setup harnessx em 5min: init, 8 mapas, doctor"
- **B (F1):** "8 mapas JSON que alimentam o context engine"
  *(profile, commands, dependencies, architecture, test-map, api-map,
  design-system, performance-budget)*
- **C (F4):** "Não use `harness init` sem rodar `harness project index`"
- **Thumb 1 (T3):** terminal `harness init` em `#00E38A`, retângulo
  vermelho `#FF3B3B` sobre linha `doctor: ok`. Texto: **"SETUP DO ZERO"**
  em amarelo `#FFD400`. Rosto: Rodolfo focado, canto inf. direito.
  Cor dominante: preto + verde CRT.
- **Thumb 2 (T4):** diagrama 8 quadrados JSON com setas convergindo
  pra "context.Build". Texto: **"8 ÍNDICES"** em branco. Rosto pequeno
  canto sup. direito. Cor dominante: preto + branco.
- **Hipótese CTR:** T3 tende a ganhar em audiência já de terminal;
  T4 ganha em quem chega via long-tail "context engineering".
  Testar A (busca) vs B (curiosidade).

---

**PL1-02 — Spec-driven backend init**
- **A (F7):** "Por que escrevo spec antes de código Go"
- **B (F2):** "Backend Go inteiro por US$ 0,15"
- **C (F10):** "Isso mudou meu jeito de começar API"
- **Thumb 1 (T2):** split — spec.md à esquerda (fundo escuro), diff Go
  à direita (fundo escuro). Seta `#00E38A` ligando. Texto:
  **"SPEC → CÓDIGO"** vermelho `#FF3B3B` centro superior. Rosto canto
  inf. esq., expressão apontando. Cor dominante: preto + vermelho.
- **Thumb 2 (T1):** rosto Rodolfo dominante (300×400), número gigante
  **"$0,15"** em `#7CF74A`. Cor dominante: verde-sucesso + preto.
- **Hipótese:** B (preço concreto) tende a superar em CTR de feed,
  mas A (polêmica) ganha em CTR de search "spec driven".

---

**PL1-03 — Auth JWT com Claude Sonnet**
- **A (F4):** "Não use JWT antes de ver isso"
- **B (F3):** "Sonnet vs Codex em JWT: o que sobra"
- **C (F6):** "Quebrei o Codex com `alg=none`"
- **Thumb 1 (T5):** antes = cadeado quebrado vermelho, depois = cadeado
  verde CRT. Texto: **"JWT SEM FUROS"** em branco. Rosto canto inf.
  direito, expressão séria. Cor dominante: preto + vermelho → verde.
- **Thumb 2 (T3):** terminal com `alg=none` destacado em `#FF3B3B`,
  cursor no linha bcrypt. Texto: **"ALG=NONE"** em vermelho. Cor
  dominante: preto + vermelho.
- **Hipótese:** C (F6, "quebrei") maximiza curiosidade + é honesto —
  aposta forte pra teste 48h.

---

**PL1-04 — Projects CRUD com Codex**
- **A (F3):** "Codex vs Sonnet: 1/3 do preço em CRUD Go"
- **B (F2):** "CRUD Go inteiro por US$ 0,08"
- **C (F7):** "Por que não uso Sonnet pra CRUD"
- **Thumb 1 (T2):** balança — Sonnet lado esq. com sacola $ pesada,
  Codex lado dir. leve. Texto: **"1/3 DO PREÇO"** em `#7CF74A`. Cor
  dominante: verde + preto.
- **Thumb 2 (T1):** rosto Rodolfo apontando pra cima, número
  **"US$ 0,08"** em `#7CF74A`. Cor dominante: verde-sucesso.
- **Hipótese:** A ganha (F3 comparativo é o formato de maior CTR em
  tech BR 2026 `[VERIFICAR]`).

---

**PL1-05 — Tasks CRUD + BM25 rerank**
- **A (F9):** "BM25 explicado por bibliotecário burro"
- **B (F5):** "O paper que impediu meu agente de duplicar código"
- **C (F1):** "1 linha de log que corta 40% de duplicação"
- **Thumb 1 (T4):** diagrama simples "pack antes / pack depois"
  com uma barra `+0.71` destacada em `#00E38A`. Texto: **"+0.71"**
  gigante. Rosto pequeno canto sup. dir. Cor: preto + verde.
- **Thumb 2 (T3):** terminal com log `bm25 rerank: promoted ...
  (+0.71)` destacado. Texto: **"NÃO DUPLICOU"**. Cor: preto + verde.
- **Hipótese:** A (analogia) puxa iniciante; B (paper) puxa sênior.
  Testar B primeiro no público-alvo desse canal.

---

**PL1-06 — Scaffold determinístico Vite+React**
- **A (F7):** "Por que scaffold é melhor que LLM aqui"
- **B (F8):** "Setup Vite+React 19 em 30s (zero token)"
- **C (F4):** "Não chame LLM pra scaffold"
- **Thumb 1 (T7):** só texto — **"SEM IA. DE PROPÓSITO."** em Inter
  Black caps sobre `#0B0B0F`, wordmark `rogpe` abaixo. Único thumb T7
  da PL1. Cor: preto + branco + verde CRT no wordmark.
- **Thumb 2 (T3):** terminal `harness new ... --stack` com badge
  **"$0,00"** em `#7CF74A`. Cor: preto + verde.
- **Hipótese:** T7 gera curiosidade forte pela ruptura visual da
  playlist. Risco: pode confundir. Usar como teste A/B controlado.

---

**PL1-07 — Frontend auth + compressão LLMLingua-style**
- **A (F5):** "O paper que corta prose do meu pack React"
- **B (F1):** "~35% menos byte no pack React (código verbatim)"
- **C (F10):** "Isso mudou meu jeito de mandar React pro LLM"
- **Thumb 1 (T5):** barra pack antes/depois, corte visível vermelho.
  Texto: **"-35% PROSE"** em `#FFD400` gigante. Rosto canto inf. dir.
  satisfeito (não sorridente — leve aceno). Cor: amarelo + preto.
- **Thumb 2 (T4):** diagrama paper arXiv + seta pra pack React
  comprimido. Texto: **"LLMLINGUA-STYLE"** em branco. Cor: preto + branco.
- **Nota:** compressão é a implementação determinística em
  `internal/context/compress`, não o modelo LLMLingua-2 direto.
- **Hipótese:** A (F5 paper) é a assinatura da PL2 vazando pra PL1 —
  reforça marca "papers na prática".

---

**PL1-08 — Views projeto/tarefa (optimistic update)**
- **A (F9):** "Optimistic update explicado por garçom"
- **B (F6):** "Quebrei optimistic update em 2min (e consertei)"
- **C (F4):** "Não use optimistic update sem rollback"
- **Thumb 1 (T3):** browser DevTools + card sendo arrastado com check
  verde antes do PATCH responder. Texto: **"OTIMISTA"** em branco.
  Cor: preto + verde CRT.
- **Thumb 2 (T5):** antes = "check" verde, depois = check somindo
  vermelho (rollback). Texto: **"E SE FALHAR?"**. Cor: verde→vermelho.
- **Hipótese:** B (drama pessoal) supera C (regra) em CTR mobile.

---

**PL1-09 — Testes 3 agentes**
- **A (F1):** "3 agentes pra 3 testes: US$ 0,25 no total"
- **B (F3):** "Codex vs Kimi vs Sonnet em teste — qual vence"
- **C (F10):** "Isso mudou meu jeito de escrever teste"
- **Thumb 1 (T2):** split em 3 colunas verticais — Codex / Kimi /
  Sonnet cada uma com ícone de tipo de teste. Texto: **"3 x 3"** em
  branco centro. Cor: preto + 3 acentos discretos.
- **Thumb 2 (T1):** rosto Rodolfo + número **"$0,25"** gigante em
  `#7CF74A`. Cor: verde + preto.
- **Hipótese:** A ganha em CTR de série (número + custo).

---

**PL1-10 — Docker Compose mem_limit**
- **A (F6):** "Quebrei meu Mac 2x sem `mem_limit`"
- **B (F4):** "Não suba Compose sem `mem_limit`"
- **C (F8):** "Setup Docker + healthcheck em 8min"
- **Thumb 1 (T5):** baleia Docker com barra RAM vermelha lotada →
  baleia com barra travada verde 512m. Texto: **"MEM LIMIT"** em
  branco. Cor: vermelho→verde.
- **Thumb 2 (T1):** rosto Rodolfo alarmado + gráfico RAM disparando.
  Texto: **"TRAVOU O MAC"**. Cor: preto + vermelho.
- **Hipótese:** A (drama pessoal + F6) é ouro pra retenção — hook
  vira o vídeo inteiro. Aposta forte.

---

**PL1-11 — Dashboard local**
- **A (F1):** "5 endpoints, 1 dashboard, US$ 0,00"
- **B (F8):** "Observabilidade de agente IA em 5min"
- **C (F10):** "Isso mudou meu jeito de auditar custo IA"
- **Thumb 1 (T3):** dashboard React aberto no browser, gráfico de
  custo descendo. Texto: **"CUSTO AO VIVO"** em `#7CF74A`. Cor:
  azul dashboard + verde.
- **Thumb 2 (T1):** rosto apontando pra gráfico + número **"$0"**.
  Cor: preto + verde.
- **Hipótese:** B (setup) supera A em CTR de search "observability".

---

**PL1-12 — Análise de custo (payoff)**
- **A (F2):** "SaaS inteiro por US$ 1,23 — quebrei a conta"
- **B (F1):** "34% cache + 24% BM25 + 27% LLMLingua = US$ 1,23"
- **C (F5):** "O paper que virou meu breakdown de custo"
- **Thumb 1 (T2):** dois recibos lado a lado — US$ 5,00 riscado
  vermelho, US$ 1,23 destacado verde. Texto: **"US$ 1,23"** gigante
  em `#7CF74A`. Cor: verde + vermelho + preto.
- **Thumb 2 (T1):** rosto Rodolfo segurando recibo físico. Texto:
  **"USD 1,23 O SAAS"**. Cor: preto + verde.
- **Hipótese:** T2 é o thumb-assinatura do canal — potencial de virar
  vídeo carro-chefe da PL1.

---

**PL1-13 — Orquestração multi-agente**
- **A (F1):** "3 agentes, 1 ADR, US$ 0,05"
- **B (F9):** "Debate multi-agente explicado por júri"
- **C (F7):** "Por que multi-agente vira teatro (e como evitar)"
- **Thumb 1 (T4):** 3 caixas conectadas propor/criticar/decidir com
  setas. Texto: **"3 → 1 ADR"** em branco. Cor: preto + 3 acentos.
- **Thumb 2 (T3):** terminal com `harness orchestrate run flow.yaml`
  + highlight amarelo em `depends_on`. Texto: **"YAML"** grande.
- **Hipótese:** C (polêmica F7) diferencia num nicho já saturado
  de "multi-agente é o futuro".

---

**PL1-14 — Bugfix + evolve**
- **A (F6):** "Plantei um bug no ar. Consertei em 2min"
- **B (F3):** "harness bugfix vs feature: qual usar quando"
- **C (F1):** "1 comando, 1 diff, US$ 0,08"
- **Thumb 1 (T5):** tela "stale" vermelho → tela "fresh" verde.
  Texto: **"BUG → FIX"** branco. Cor: vermelho→verde.
- **Thumb 2 (T3):** terminal `harness bugfix "..."` + diff highlight
  amarelo. Texto: **"NO CLI"**. Cor: preto + amarelo.
- **Hipótese:** A (drama F6) domina; B ganha em search "bugfix".

---

**PL1-15 — Ship (harness ci + ship)**
- **A (F7):** "Por que não uso GitHub Actions"
- **B (F8):** "Setup CI/CD local em 3min"
- **C (F4):** "Não faça deploy sem `pre-push` hook"
- **Thumb 1 (T3):** terminal `harness ship` + barra CI verde no
  fundo. Texto: **"SHIP"** gigante. Cor: preto + verde.
- **Thumb 2 (T7):** só texto — **"CI LOCAL. SEM ACTIONS."**. Cor:
  preto + branco.
- **Hipótese:** A (polêmica) é a única forma de vídeo "de fechamento"
  não morrer em CTR.

---

**PL1-16 — Rubrica de agente (Appendix A)**
- **A (F1):** "8 tarefas, 8 agentes: a tabela que uso todo dia"
- **B (F3):** "Sonnet, Codex, Kimi, Opus, Haiku: qual pra quê"
- **C (F10):** "Isso mudou meu jeito de escolher agente"
- **Thumb 1 (T4):** tabela 8 linhas com ícone por agente + coluna
  custo colorida. Texto: **"QUAL PRA QUÊ"** branco. Cor: preto + 5
  acentos pastel.
- **Thumb 2 (T1):** rosto Rodolfo + número **"8×"** grande. Cor:
  preto + amarelo.
- **Hipótese:** A ganha (F1 número + promessa de referência).

---

**PL1-17 — Quando pular LLM + debug (Appendix B+C)**
- **A (F4):** "4 casos pra NÃO chamar o LLM"
- **B (F7):** "Por que agente IA às vezes é a pior escolha"
- **C (F1):** "6 debugs de harnessx que te salvam 20min/semana"
- **Thumb 1 (T7):** sinal de PARE gigante com "LLM" dentro. Texto:
  **"PARE ANTES"**. Cor: vermelho + branco.
- **Thumb 2 (T1):** rosto Rodolfo braços cruzados + **"NÃO"** grande.
  Cor: preto + vermelho.
- **Hipótese:** A (F4 lista + negação) é o formato-mãe de retenção
  em vídeo de "quando não usar".

---

### PL2 — Papers na prática (10 vídeos)

Playlist canônica do canal para tradução paper → código harnessx.
Formulas dominantes: F5, F9, F10.

**PL2-01 — LLMLingua-style deterministic compression (adaptação
Pan et al., 2024)**
- **A (F5):** "O paper que corta ~35% do prose sem tocar código"
- **B (F1):** "~35% menos byte em README/docs, código verbatim"
- **C (F9):** "Compressão de contexto explicada por editor de revista"
- **Thumb 1 (T4):** capa arXiv + gráfico -35%. Texto: **"-35%"**.
  Cor: preto + amarelo.
- **Thumb 2 (T5):** pack antes/depois. Texto: **"PROSE CORTA"**.
- **Nota:** harness roda versão determinística LLMLingua-*style* em
  `internal/context/compress`, não o modelo LLMLingua-2. Combinado
  (rerank+compress+reorder) chega em 50–70% no pack — ver
  `docs/PAPER-IMPLEMENTATION.md §5`.
- **Hipótese:** A é a assinatura da PL2.

**PL2-02 — BM25 (Robertson & Zaragoza, 2009)**
- **A (F9):** "BM25 explicado por bibliotecário"
- **B (F5):** "O paper de 2009 que ainda me economiza R$/token"
- **C (F1):** "1 rerank BM25, 0.71 de score, -30% custo"
- **Thumb 1 (T4):** gráfico barras + `+0.71` destacado. Cor: ciano.
- **Thumb 2 (T1):** rosto + **"+0.71"**. Cor: preto + ciano.
- **Hipótese:** A ganha CTR iniciante; B ganha sênior.

**PL2-03 — Prompt Caching (Anthropic)**
- **A (F5):** "O truque de cache que me economiza 30%"
- **B (F1):** "30% economia em 1 flag: prompt cache"
- **C (F10):** "Isso mudou meu jeito de montar pack de contexto"
- **Thumb 1 (T3):** terminal com log `cache_read_input_tokens: 8421`
  destacado verde. Texto: **"CACHE HIT"**. Cor: preto + verde.
- **Thumb 2 (T4):** diagrama prefix estável / suffix dinâmico. Texto:
  **"PREFIX"**. Cor: preto + branco.
- **Hipótese:** A (F5) domina.

**PL2-04 — Salience reorder (context engineering)**
- **A (F9):** "Salience explicado por curador de museu"
- **B (F1):** "3 arquivos reordenados = 210 tokens economizados"
- **C (F7):** "Por que ordem lexicográfica te custa dinheiro"
- **Thumb 1 (T4):** pack antes/depois com setas de reorder. Texto:
  **"REORDER"**. Cor: preto + verde.
- **Thumb 2 (T3):** log `salience reorder: +3 +2`. Texto: **"+210 TOK"**.
- **Hipótese:** C (F7 polêmica) ganha CTR.

**PL2-05 — Chain-of-Thought vs Direct answer**
- **A (F3):** "CoT vs resposta direta: quando vale o custo"
- **B (F5):** "O paper que provou que CoT nem sempre paga"
- **C (F10):** "Isso mudou meu jeito de mandar prompt de debug"
- **Thumb 1 (T2):** split — CoT longo esquerda, direct curto direita,
  balança custo/qualidade. Texto: **"CoT?"**. Cor: preto + amarelo.
- **Thumb 2 (T1):** rosto pensando + **"NEM SEMPRE"**.
- **Hipótese:** A (F3) domina em search técnico.

**PL2-06 — RAG vs Long context**
- **A (F3):** "RAG vs Long Context em 2026: qual sobrou"
- **B (F5):** "O paper que matou meu setup RAG"
- **C (F7):** "Por que dropei RAG pra long context"
- **Thumb 1 (T2):** RAG (blocos DB) vs long-context (rio de tokens).
  Texto: **"RAG × LONG"**. Cor: preto + azul + vermelho.
- **Thumb 2 (T7):** só texto — **"RAG MORREU?"**. Cor: preto + vermelho.
- **Hipótese:** T7 (polêmica pura) é aposta CTR alta.

**PL2-07 — Speculative decoding**
- **A (F9):** "Speculative decoding explicado por corretor de prova"
- **B (F1):** "2x mais rápido, mesma qualidade: como funciona"
- **C (F5):** "O paper que faz seu LLM correr o dobro"
- **Thumb 1 (T4):** diagrama draft-model + verify. Texto: **"2×"**
  gigante. Cor: preto + verde.
- **Thumb 2 (T1):** rosto + **"2× RÁPIDO"**.
- **Hipótese:** B (F1 número claro) ganha.

**PL2-08 — Reflexion / self-critique agents**
- **A (F5):** "O paper que fez meu agente pensar duas vezes"
- **B (F7):** "Por que self-critique é caro e vale"
- **C (F10):** "Isso mudou meu flow.yaml"
- **Thumb 1 (T4):** loop propor→criticar→revisar. Texto: **"2x PENSA"**.
- **Thumb 2 (T2):** antes/depois — resposta ruim / resposta refinada.
- **Hipótese:** A (F5 assinatura PL2).

**PL2-09 — Structured output / JSON mode**
- **A (F4):** "Não use JSON mode antes de ler isso"
- **B (F1):** "3 formas de forçar JSON que não quebram"
- **C (F3):** "JSON mode vs function calling vs schema"
- **Thumb 1 (T3):** terminal com JSON válido/verde e inválido/vermelho.
  Texto: **"JSON OU DIE"**. Cor: preto + verde/vermelho.
- **Thumb 2 (T1):** rosto + **"JSON"** gigante.
- **Hipótese:** A (F4 negação) domina.

**PL2-10 — Constitutional AI / guardrails**
- **A (F5):** "O paper Constitutional AI aplicado em agente CLI"
- **B (F9):** "Guardrail explicado por professor de trânsito"
- **C (F10):** "Isso mudou meu jeito de dar prompt de auth"
- **Thumb 1 (T4):** diagrama guardrail em torno de LLM. Texto:
  **"REGRAS"**. Cor: preto + amarelo.
- **Thumb 2 (T5):** antes = LLM sem freio (vermelho), depois =
  LLM com trilho (verde).
- **Hipótese:** A (F5).

---

### PL3 — Dentro do harnessx (18 vídeos)

Foco: architecture deep dive. Público sênior. Formulas dominantes:
F1, F3, F8, F9.

**PL3-01 — Arquitetura em camadas (domain/app/adapters)**
- **A (F9):** "Clean Architecture explicada por planta de casa"
- **B (F1):** "3 camadas, 0 import cruzado, 1 binário"
- **C (F8):** "Setup Clean Architecture Go em 10min"
- **Thumb 1 (T4):** 3 camadas concêntricas. Texto: **"3 CAMADAS"**.
- **Thumb 2 (T3):** terminal `go list -deps` filtrado. Texto: **"ZERO CGO"**.
- **Hipótese:** A ganha iniciante; B ganha sênior.

**PL3-02 — Context Engine (internal/context)**
- **A (F1):** "4 alavancas de contexto que valem US$ 3,50"
- **B (F9):** "Context engine explicado por chef"
- **C (F10):** "Isso mudou meu jeito de mandar código pro LLM"
- **Thumb 1 (T4):** diagrama 4 alavancas → pack. Texto: **"4×"**.
- **Thumb 2 (T1):** rosto + **"CONTEXT"** grande.
- **Hipótese:** A domina.

**PL3-03 — Sensores (internal/sensors)**
- **A (F1):** "10 sensores que salvam PRs em 30s"
- **B (F4):** "Não faça PR sem esses 10 sensores"
- **C (F7):** "Por que sensor > linter em 2026"
- **Thumb 1 (T3):** terminal `harness sensor run` + 10 checks verdes.
  Texto: **"10 SENSORES"**.
- **Thumb 2 (T5):** PR ruim (vermelho) → PR bloqueado (verde).
- **Hipótese:** A (F1 número forte).

**PL3-04 — Router de agentes (routes.yaml)**
- **A (F8):** "Setup routes.yaml em 5min: budget + fallback"
- **B (F3):** "routes.yaml vs prompt manual"
- **C (F1):** "1 YAML, 14 adapters, 5 dólares de teto" *(contagem em
  `docs/FEATURES.md` §Phase 3 — CLI+local+HTTP+test doubles)*
- **Thumb 1 (T3):** terminal + YAML highlight `budget_usd: 5.00`.
  Texto: **"$5 TETO"**.
- **Thumb 2 (T4):** diagrama intent → agent → fallback.
- **Hipótese:** C (F1).

**PL3-05 — Cost tracker (internal/costtrack)**
- **A (F1):** "1 tabela, 3 dimensões, todo dólar rastreado"
- **B (F9):** "Cost tracker explicado por caixa de padaria"
- **C (F10):** "Isso mudou meu jeito de auditar bill de LLM"
- **Thumb 1 (T3):** terminal `harness cost-compare` + `harness analytics` + tabela agent×intent.
  Texto: **"TODO $"**.
- **Thumb 2 (T1):** rosto + calculadora + **"$/TOKEN"**.
- **Hipótese:** A.

**PL3-06 — Adapters (13 provedores)**
- **A (F1):** "14 adapters, 1 interface Go" *(FEATURES.md §Phase 3)*
- **B (F3):** "Claude vs Codex vs Kimi: por baixo do capô"
- **C (F9):** "Adapter pattern explicado por tomada universal"
- **Thumb 1 (T4):** 14 logos convergindo pra 1 interface. Texto: **"14→1"**.
- **Thumb 2 (T1):** rosto + **"14 ADAPTERS"**.
- **Hipótese:** A.

**PL3-07 — Orquestrador (flow.yaml)**
- **A (F1):** "3 papéis, 1 YAML, US$ 0,05"
- **B (F9):** "Orquestrador explicado por regente de orquestra"
- **C (F7):** "Por que YAML de agente > código Python"
- **Thumb 1 (T3):** terminal `harness orchestrate run` + YAML flow.
  Texto: **"FLOW.YAML"**.
- **Thumb 2 (T4):** diagrama 3 nós + setas `depends_on`.
- **Hipótese:** C (F7 polêmica) ganha.

**PL3-08 — Sensor rules customizadas**
- **A (F1):** "5 regras customizadas que bloqueiam bug antes do PR"
- **B (F6):** "Escrevi uma regra que quebrou 3 PRs em 1 dia"
- **C (F8):** "Setup custom rule em 3min"
- **Thumb 1 (T3):** terminal + rule YAML + PR bloqueado vermelho.
  Texto: **"BLOCK PR"**.
- **Thumb 2 (T1):** rosto + **"5 REGRAS"**.
- **Hipótese:** B (drama F6).

**PL3-09 — Modo debate multi-agente**
- **A (F9):** "Debate multi-agente explicado por júri de tribunal"
- **B (F1):** "3 agentes, 1 decisão, US$ 0,08"
- **C (F5):** "O paper Constitutional AI virou meu flow.yaml"
- **Thumb 1 (T4):** diagrama debate. Texto: **"DEBATE"**.
- **Thumb 2 (T1):** rosto mediador + **"3 VOZES"**.
- **Hipótese:** B.

**PL3-10 — Dashboard interno (internal/dashboardapi)**
- **A (F8):** "Setup dashboard local em 2min"
- **B (F1):** "5 endpoints REST, 0 dependência externa"
- **C (F10):** "Isso mudou meu jeito de auditar agente"
- **Thumb 1 (T3):** browser dashboard + curva de custo. Texto: **"5 API"**.
- **Thumb 2 (T1):** rosto + gráfico + **"$/DIA"**.
- **Hipótese:** B.

**PL3-11 — Perf snapshot**
- **A (F1):** "1 comando, 1 baseline, 0 regressão silenciosa"
- **B (F6):** "Peguei um regressor de 300ms com perf-snapshot"
- **C (F8):** "Setup perf baseline em 30s"
- **Thumb 1 (T5):** gráfico antes/depois com pico vermelho.
- **Thumb 2 (T3):** terminal `harness perf-snapshot`.
- **Hipótese:** B (F6 drama).

**PL3-12 — modernc.org/sqlite (sem CGO)**
- **A (F7):** "Por que troquei mattn/sqlite por modernc"
- **B (F3):** "modernc vs mattn: benchmark real"
- **C (F1):** "1 flag, 0 CGO, 1 binário"
- **Thumb 1 (T2):** split — CGO vermelho, puro Go verde.
- **Thumb 2 (T1):** rosto + **"ZERO CGO"**.
- **Hipótese:** A (F7).

**PL3-13 — Embed FS (go:embed)**
- **A (F1):** "1 diretiva, 0 arquivo solto no /etc"
- **B (F9):** "go:embed explicado por selo em envelope"
- **C (F4):** "Não use `..` em go:embed (aqui é bloqueado)"
- **Thumb 1 (T3):** terminal `//go:embed all:templates` + binário.
- **Thumb 2 (T7):** só texto — **"1 BINÁRIO"**.
- **Hipótese:** C (F4 negação técnica).

**PL3-14 — slog estruturado**
- **A (F8):** "Setup slog estruturado em 5min"
- **B (F3):** "slog vs zap vs zerolog em 2026"
- **C (F10):** "Isso mudou meu jeito de logar"
- **Thumb 1 (T2):** split — slog verde, zap/zerolog cinza.
- **Thumb 2 (T3):** terminal JSON log estruturado.
- **Hipótese:** B (F3 comparativo).

**PL3-15 — Testes de sensor**
- **A (F1):** "1 sensor, 20 casos de teste, 0 falso positivo"
- **B (F6):** "Deletei um teste e o sensor bloqueou o PR"
- **C (F9):** "Sensor explicado por detector de metal"
- **Thumb 1 (T3):** terminal `go test ./internal/sensors` verde.
- **Thumb 2 (T1):** rosto + **"20 CASOS"**.
- **Hipótese:** B (F6).

**PL3-16 — Skills (harnessx skills system)**
- **A (F1):** "5 skills que rodam antes de qualquer prompt"
- **B (F9):** "Skill explicada por perfil de app"
- **C (F10):** "Isso mudou meu jeito de compor prompt longo"
- **Thumb 1 (T4):** diagrama skill → prompt pipeline.
- **Thumb 2 (T1):** rosto + **"5 SKILLS"**.
- **Hipótese:** A.

**PL3-17 — Memória persistente (internal/memory)**
- **A (F1):** "1 store, 3 tabelas, memória entre sessões"
- **B (F5):** "O paper que virou minha memória de agente"
- **C (F4):** "Não use memória sem `evidence_run_id`"
- **Thumb 1 (T4):** diagrama memória + tags.
- **Thumb 2 (T3):** terminal `harness memory list`.
- **Hipótese:** C (F4 técnico).

**PL3-18 — GitFlow + pre-push hook**
- **A (F7):** "Por que uso GitFlow em 2026"
- **B (F8):** "Setup pre-push que roda `make ci` em 3min"
- **C (F4):** "Não faça `--no-verify` nem em hotfix"
- **Thumb 1 (T3):** terminal git graph + pre-push verde.
- **Thumb 2 (T7):** só texto — **"NUNCA --no-verify"**.
- **Hipótese:** C (F4 comando específico).

---

### PL4 — Agentes IA 2026 (10 vídeos)

Foco: estado da arte + análise crítica. Formulas dominantes: F3, F5,
F7, F10.

**PL4-01 — Estado dos agentes IA em 2026 (overview)**
- **A (F1):** "14 agentes IA em 2026: qual vale seu token" *(alinha
  com adapters bundled do harness; ajustar se pool comparado for outro)*
- **B (F3):** "Claude vs Codex vs Kimi vs Gemini vs Grok"
- **C (F10):** "Isso mudou meu jeito de escolher LLM em 2026"
- **Thumb 1 (T4):** grade de logos + rating custo (número = pool real
  comparado no vídeo).
- **Thumb 2 (T1):** rosto + **"14 IA"**.
- **Hipótese:** B (F3).

**PL4-02 — MCP protocol na prática**
- **A (F5):** "O protocolo que virou padrão de tool-calling"
- **B (F8):** "Setup MCP server Go em 10min"
- **C (F9):** "MCP explicado por USB"
- **Thumb 1 (T4):** diagrama MCP.
- **Thumb 2 (T3):** terminal MCP handshake.
- **Hipótese:** C (F9 analogia clássica).

**PL4-03 — Tool calling estável**
- **A (F1):** "3 padrões de tool calling que não quebram"
- **B (F4):** "Não use function calling sem retry"
- **C (F6):** "Quebrei um agente com 1 JSON malformado"
- **Thumb 1 (T3):** terminal JSON válido/inválido.
- **Thumb 2 (T5):** antes/depois retry.
- **Hipótese:** C (F6).

**PL4-04 — Custo real de agentes (2026)**
- **A (F2):** "Como gastei US$ 12/mês em agente IA sério"
- **B (F1):** "3 apps, 1 dashboard, US$ 12 no mês"
- **C (F7):** "Por que preço de token é ilusão"
- **Thumb 1 (T1):** rosto + **"$12/MÊS"**.
- **Thumb 2 (T2):** recibo bruto vs recibo com context engine.
- **Hipótese:** A (F2).

**PL4-05 — Agentic loops (ReAct, Reflexion)**
- **A (F5):** "O paper ReAct explicado sem hype"
- **B (F9):** "Agentic loop explicado por chef corrigindo receita"
- **C (F7):** "Por que a maior parte de agentic é teatro"
- **Thumb 1 (T4):** diagrama ReAct.
- **Thumb 2 (T7):** só texto — **"TEATRO OU FIX?"**.
- **Hipótese:** C (F7 polêmica).

**PL4-06 — Local LLM (Ollama, llama.cpp)**
- **A (F3):** "Ollama vs API paga em 2026: quando vale local"
- **B (F5):** "O paper que provou que 8B chega no 70B em tarefa X"
- **C (F1):** "1 modelo local, 0 API bill, 40% qualidade"
- **Thumb 1 (T2):** local (verde) vs cloud (amarelo) + custo.
- **Thumb 2 (T1):** rosto + **"US$ 0/MÊS"**.
- **Hipótese:** A.

**PL4-07 — Prompt injection e defesa**
- **A (F4):** "Não deploy agente sem filtro de injection"
- **B (F6):** "Quebrei meu próprio agente com 1 comentário Markdown"
- **C (F5):** "O paper de indirect injection na prática"
- **Thumb 1 (T5):** prompt normal → prompt injetado (vermelho).
- **Thumb 2 (T7):** só texto — **"INJECTED"**.
- **Hipótese:** B (F6 drama).

**PL4-08 — Multi-modal (vision + code)**
- **A (F1):** "3 tarefas que só multi-modal resolve"
- **B (F3):** "Sonnet vs GPT-4o vs Gemini em screenshot debug"
- **C (F10):** "Isso mudou meu jeito de debug UI"
- **Thumb 1 (T2):** split imagens + código.
- **Thumb 2 (T1):** rosto + screenshot destacado.
- **Hipótese:** B.

**PL4-09 — Evaluation de agente (evals)**
- **A (F5):** "O paper que me fez parar de fingir que testei agente"
- **B (F8):** "Setup eval harness em 15min"
- **C (F4):** "Não deploy agente sem eval suite"
- **Thumb 1 (T3):** terminal `evals run` + gráfico pass@k.
- **Thumb 2 (T1):** rosto + **"EVAL"** gigante.
- **Hipótese:** A (F5).

**PL4-10 — Escolher stack de agente em 2026 (rubrica)**
- **A (F1):** "5 critérios pra escolher stack de agente em 2026"
- **B (F10):** "Isso mudou meu jeito de recomendar stack IA"
- **C (F3):** "harnessx vs LangGraph vs CrewAI vs raw SDK"
- **Thumb 1 (T4):** tabela 4 stacks × 5 critérios.
- **Thumb 2 (T2):** balança 4 stacks.
- **Hipótese:** C (F3).

---

### PL5 — Zero ao SaaS (10 vídeos)

Foco: empreendedor dev. Público mais amplo. Fórmulas: F2, F7, F10.

**PL5-01 — Do zero ao SaaS: mapa em 10min**
- **A (F2):** "Como fiz meu SaaS por US$ 12/mês"
- **B (F1):** "10 passos, 1 SaaS, 0 investidor"
- **C (F7):** "Por que solo-founder deveria escolher Go+React"
- **Thumb 1 (T1):** rosto + **"US$ 12"**.
- **Thumb 2 (T4):** mapa 10 passos.
- **Hipótese:** A.

**PL5-02 — Escolher nicho de SaaS**
- **A (F7):** "Por que 90% dos SaaS morrem no nicho errado"
- **B (F1):** "3 filtros de nicho que uso antes de codar"
- **C (F10):** "Isso mudou meu jeito de escolher problema"
- **Thumb 1 (T7):** só texto — **"NICHO ERRADO"**.
- **Thumb 2 (T1):** rosto + **"3 FILTROS"**.
- **Hipótese:** A (F7 polêmica).

**PL5-03 — Validar antes de codar**
- **A (F4):** "Não abra editor sem validar essas 3 coisas"
- **B (F6):** "Codei 3 SaaS que ninguém queria"
- **C (F2):** "Como validei um SaaS por US$ 0"
- **Thumb 1 (T5):** landing → 0 signups vermelho.
- **Thumb 2 (T1):** rosto + **"VALIDA"**.
- **Hipótese:** B (F6 confissão).

**PL5-04 — Stack mínima de SaaS solo (2026)**
- **A (F3):** "Go+React vs Rails vs Next.js pra solo-founder"
- **B (F1):** "3 serviços, US$ 12/mês, 1000 usuários"
- **C (F10):** "Isso mudou meu jeito de escolher stack"
- **Thumb 1 (T2):** 3 stacks lado a lado.
- **Thumb 2 (T1):** rosto + **"$12/MÊS"**.
- **Hipótese:** A.

**PL5-05 — Landing page em 1 dia**
- **A (F8):** "Setup landing convertendo em 4h"
- **B (F1):** "5 seções que 100% das landings top têm"
- **C (F7):** "Por que dropei Framer por Astro"
- **Thumb 1 (T3):** browser landing + CTA highlight.
- **Thumb 2 (T5):** landing feia → landing convertendo.
- **Hipótese:** B (F1).

**PL5-06 — Autenticação e billing (Stripe + JWT)**
- **A (F8):** "Setup Stripe + JWT em 2h"
- **B (F4):** "Não use Stripe sem esses 3 webhooks"
- **C (F1):** "3 webhooks obrigatórios de Stripe SaaS"
- **Thumb 1 (T3):** terminal Stripe CLI + JWT.
- **Thumb 2 (T4):** diagrama webhook → server → JWT.
- **Hipótese:** B.

**PL5-07 — Analytics honesto (GA4 + Clarity)**
- **A (F3):** "GA4 vs Plausible vs Umami pra SaaS solo"
- **B (F1):** "3 métricas que decidem se seu SaaS vive"
- **C (F10):** "Isso mudou meu jeito de olhar dashboard"
- **Thumb 1 (T4):** dashboard 3 métricas.
- **Thumb 2 (T1):** rosto + **"3 KPI"**.
- **Hipótese:** B.

**PL5-08 — Ads pagos com margem controlada**
- **A (F4):** "Não escale Meta Ads sem cobrir CAC"
- **B (F2):** "Como gastei R$ 5k em Meta e sobrevivi"
- **C (F1):** "3 portões antes de dobrar orçamento"
- **Thumb 1 (T5):** gráfico ROAS ruim → bom.
- **Thumb 2 (T1):** rosto + **"CAC"**.
- **Hipótese:** B (F2 número real).

**PL5-09 — SEO técnico para SaaS**
- **A (F8):** "Setup SEO técnico em 1 dia"
- **B (F1):** "5 páginas que travam ranking do seu SaaS"
- **C (F7):** "Por que blog não é mais o vetor SEO em 2026"
- **Thumb 1 (T3):** DevTools Lighthouse + score verde.
- **Thumb 2 (T7):** só texto — **"BLOG MORREU?"**.
- **Hipótese:** C (F7).

**PL5-10 — Escalar ou fechar: decisão de sócio-fundador**
- **A (F7):** "Por que fechei meu último SaaS"
- **B (F10):** "Isso mudou meu jeito de decidir escalar"
- **C (F1):** "3 sinais pra escalar, 3 sinais pra fechar"
- **Thumb 1 (T7):** só texto — **"ESCALAR OU FECHAR?"**.
- **Thumb 2 (T1):** rosto + balança.
- **Hipótese:** A (F7 confissão pessoal, ouro em CTR).

---

## 5. Protocolo de teste A/B (48h)

### 5.1 Ferramentas 2026

- **YouTube Studio — Test & Compare (nativo).** Testa até 3 thumbnails
  na mesma URL. Recomendado como default. Documentação:
  https://support.google.com/youtube/ `[VERIFICAR URL exata 2026]`.
- **thumbnailTest.com** — teste externo com painel de terceiros
  (útil pré-publicação). `[VERIFICAR — verificar preço/tier 2026]`.
- **TubeBuddy A/B Testing** — plugin browser. `[VERIFICAR — plano
  necessário 2026]`.
- **vidIQ Boost** — sugestões de título/thumb baseadas em CTR de
  vídeos comparáveis. `[VERIFICAR 2026]`.

**Regra:** rodar Test & Compare nativo do YouTube em 100 % dos vídeos
da série. Ferramenta externa só se estagnação.

### 5.2 Métrica que decide

Nunca decidir por CTR isolado. **CTR × AVD** é a única métrica válida
(CTR alto sem AVD = clickbait, o algoritmo pune em 7 dias — Creator
Insider 2026 `[VERIFICAR]`).

Fórmula prática:
```
score = CTR (%) × AVD (%)
Ex.: CTR 8% × AVD 55% = 4.4
     CTR 12% × AVD 30% = 3.6  → perde apesar do CTR maior
```

**Metas por vídeo (da PL1, coerente com `SEO-RETENCAO.md`):**

| Métrica | Mínimo | Bom | Excelente |
|---|---|---|---|
| CTR 48h | 5 % | 7 % | 9 %+ |
| AVD | 50 % | 55 % | 65 %+ |
| Score CTR×AVD | 2.75 | 3.85 | 5.85+ |

### 5.3 Quando pivotar

- **< 500 impressões:** não decide nada, apenas coleta.
- **500–2000 impressões:** primeira leitura, ainda sem trocar.
- **2000+ impressões:** decisão final. Se score < 2.75, trocar thumb
  (não título — título tem custo SEO, thumb é livre).
- **Após 7 dias:** se CTR ainda abaixo do mínimo, editar hook dos
  primeiros 60 s do vídeo (Creator Insider 2026: iteração de hook é o
  ganho #1 de retenção pós-publicação `[VERIFICAR]`).

---

## 6. Erros comuns (título/thumb técnicos BR)

1. **Título > 60 chars** — corta em mobile no meio da palavra. Regra:
   contar caracteres, testar em preview mobile antes de publicar.
2. **Thumb com rosto < 200×200 px** — vira ponto na home mobile.
   Solução: rosto ocupa 25–35 % da área útil.
3. **Contraste ruim em modo escuro** — texto cinza claro sobre fundo
   escuro do app YouTube some. Solução: contorno externo 4px preto
   ou branco, testar sobre `#0F0F0F` (fundo YT dark).
4. **Texto sobrepondo rosto** — divide atenção, cognitivamente pesado.
   Regra: 0 % de overlap texto-rosto.
5. **Fórmula genérica** ("Aprenda X hoje", "Guia completo de Y") —
   0 diferenciação, CTR abaixo da média BR tech 2026 `[VERIFICAR]`.
6. **Emoji em thumb técnico** — quebra a marca "voz técnica". Não usar.
7. **Screenshot puro de VSCode** — mancha ilegível em mobile. Sempre
   editar: zoom + highlight amarelo em ≤ 3 linhas.
8. **Texto em inglês num canal BR** — perde CTR mobile.
   Regra: se o termo tem tradução BR corrente (ex.: "compressão" pra
   "compression"), usar BR; se não tem (ex.: "prompt", "cache"),
   manter em inglês.
9. **Rosto no lado errado (esquerda)** — YouTube coloca duração do
   vídeo no canto inferior direito. Rosto no inferior direito
   compete com duração — mover pro superior direito nesse layout, OU
   colocar duração como texto queimado na thumb.
10. **Mudar sistema visual entre vídeos da série** — quebra "efeito
    coleção" na home. Regra: paleta e faixa lateral fixas em 100 %
    da série.

---

## 7. Referências de canais (6 canais)

### Internacionais

**Fireship** — https://www.youtube.com/@Fireship
- **Copiar:** ritmo de corte (< 3s por plano), meme como reforço técnico,
  título com número + tempo ("X in 100 seconds").
- **Evitar:** trilha sonora densa (voz Rodolfo é técnica, não meme).

**ThePrimeagen** — https://www.youtube.com/@ThePrimeagen
- **Copiar:** rosto dominante em thumb, terminal como cenário canônico,
  reações honestas a código real.
- **Evitar:** clickbait tipo "he broke Neovim". Rodolfo = seriedade.

**Web Dev Simplified** — https://www.youtube.com/@WebDevSimplified
- **Copiar:** título limpo direto ("Learn X in Y Minutes"), thumb com
  contraste alto e ≤ 3 palavras.
- **Evitar:** repetição excessiva do mesmo layout (perde diferenciação
  a longo prazo).

### BR

**Filipe Deschamps** — https://www.youtube.com/@FilipeDeschamps
- **Copiar:** honestidade + humildade + "gastei X, aprendi Y",
  storytelling de decisão pessoal.
- **Evitar:** vídeos > 25 min sem chapters — retenção BR cai.

**Lucas Montano** — https://www.youtube.com/@LucasMontano
- **Copiar:** vídeo de opinião com base técnica ("Por que eu não uso X").
- **Evitar:** thumb com muita cor + emoji (não combina com voz Rodolfo).

**Rocketseat** — https://www.youtube.com/@rocketseat
- **Copiar:** consistência de brand (paleta roxo/preto reconhecível
  em qualquer thumb da home), playlist bingeable.
- **Evitar:** thumb com rosto muito pequeno (erro comum em canal-empresa
  vs canal-pessoa).

---

## Anexo A — Checklist pré-upload

- [ ] Título ≤ 60 chars e usa 1 das 10 fórmulas.
- [ ] 3 variantes A/B/C prontas (fórmulas distintas).
- [ ] 2 thumbs 1280×720, cada uma em 1 fórmula T1–T7 diferente.
- [ ] Contraste testado no preview mobile YT dark.
- [ ] Rosto ≥ 220×220 px.
- [ ] Faixa lateral esquerda com `PLx·NN`.
- [ ] Texto ≤ 4 palavras.
- [ ] Nenhum overlap texto-rosto.
- [ ] Paleta oficial (só cores do sistema).
- [ ] Fonte oficial (JetBrains Mono ou Inter, licença OFL).
- [ ] Test & Compare do Studio ativado.
- [ ] Score CTR×AVD alvo definido antes de publicar.

## Anexo B — Nota sobre `[VERIFICAR]`

Marcadores usados neste doc, todos precisam ser confirmados antes de
virar afirmação pública:

1. Licença OFL das fontes JetBrains Mono e Inter em 2026.
2. Contraste WCAG AA da paleta sobre `#0B0B0F`.
3. Disponibilidade de handle `@construindocomagentes` (nome série A).
4. Documentação atual do YouTube Test & Compare 2026.
5. Preço/tier de thumbnailTest, TubeBuddy, vidIQ em 2026.
6. Métricas de banda de CTR de tech BR 2026 (Creator Insider).
7. Recomendação atual sobre iteração de hook pós-publicação.
8. Padrão "faixa lateral do episódio" em séries tech BR 2026.
