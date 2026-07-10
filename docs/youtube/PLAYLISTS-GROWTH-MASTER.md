# PLAYLISTS + GROWTH MASTER — Canal Rodolfo Peixoto (harnessx / rogpe.tech / skill.dev)

> Arquitetura de 5 playlists (65 vídeos totais) cobrindo o projeto
> harnessx ponta a ponta. Funil de crescimento, ordem de gravação por
> ROI, cross-linking, brand system, cadência 2026, métricas alvo,
> viralização, colabs e roadmap 12 meses.
>
> Base factual: `HARNESSX-MASTER-PLAN.md` (10 fases), `docs/PAPER-MAPPING.md`
> (paper Ning et al. arXiv 2605.18747v1) + `docs/PAPER-IMPLEMENTATION.md`
> (Context Engineering 2.0, arXiv 2510.26493), `docs/FEATURES.md`
> (86 comandos top-level), `docs/TUTORIAL-BUILD-TASKHIVE.md`
> (15 caps + 3 apêndices), série v1 já roteirizada em
> `docs/youtube/ROTEIROS-TECNICOS.md`, `SERIE-PAPERS-NA-PRATICA.md`
> (10 episódios), `SERIE-DENTRO-DO-HARNESSX.md` (18 episódios),
> `DIDATICA-CONCEITOS.md`, `SEO-RETENCAO.md`.
>
> Voz: Rodolfo Peixoto — PT-BR direto, técnico, sem "galera",
> "basicamente" ou "hoje vou te ensinar". Cada afirmação sobre
> algoritmo/retenção tem fonte 2026 ou `[VERIFICAR]`.

---

## 0. Notas de fonte 2026

Toda tática de crescimento aqui sai de uma destas fontes. Número não
confirmado → marcado `[VERIFICAR]`. Nunca invento.

- YouTube Creator Insider (canal oficial):
  <https://www.youtube.com/@YouTubeCreatorInsider>
- YouTube Creator Academy 2026:
  <https://creatoracademy.youtube.com/>
- Covington et al., "Deep Neural Networks for YouTube Recommendations"
  (Google, RecSys 2016, ainda base do ranker em 2026):
  <https://research.google/pubs/pub45530/>
- Guo, Kim & Rubin, L@S 2014 — retenção em vídeo educacional
  ≤ 6 min: <https://dl.acm.org/doi/10.1145/2556325.2566239>
- TubeBuddy Trends Report 2026: `[VERIFICAR — URL]`
- vidIQ blog "State of YouTube 2026": `[VERIFICAR]`
- Fireship, ThePrimeagen, Theo (public analytics via Social Blade):
  <https://socialblade.com/> — bandas de referência para nicho dev EN
- MrBeast public workshop deck 2024 revalidado 2026 (thumb + CTR):
  `[VERIFICAR link oficial]`
- Nielsen Norman Group — thumbnail attention:
  <https://www.nngroup.com/articles/thumbnail-attention/> `[VERIFICAR 2026]`

---

## 1. Mapa das 5 playlists

Total: **65 vídeos** (17 + 10 + 18 + 10 + 10). Contagens PL2 e PL3
alinhadas a `SERIE-PAPERS-NA-PRATICA.md` (10 eps) e
`SERIE-DENTRO-DO-HARNESSX.md` (18 eps) — fontes canônicas.

Legenda coluna **Camada funil**: TOF = topo (hook amplo), MOF = meio
(educacional aprofundado), BOF = fundo (hardcore técnico).

---

### PL1 — "Construindo SaaS com Agentes IA" (17 vídeos)

Série TaskHive. Roteiros v1 já existem em `ROTEIROS-TECNICOS.md`.
Base 1-para-1 com `docs/TUTORIAL-BUILD-TASKHIVE.md`.

| # | Título | Duração | Camada | Prereq |
|---|---|---|---|---|
| 00 | TaskHive: SaaS Go+React por US$ 1,50 em LLM | 6 min | TOF | nenhum |
| 01 | harness init em 5 min: vault, 8 índices, doctor | 7 min | MOF | 00 |
| 02 | Do zero ao spec: contrato antes do código | 8 min | MOF | 01 |
| 03 | Primeiro agente escreve Go: `harness ship` | 9 min | MOF | 02 |
| 04 | JWT sem drama: sensors travam vazamento | 9 min | MOF | 03 |
| 05 | CRUD com plan-check: escopo ou nada | 9 min | MOF | 04 |
| 06 | React 19 + Vite: agente monta o front | 10 min | MOF | 05 |
| 07 | Playwright rodando em Docker headless | 10 min | MOF | 06 |
| 08 | Multi-agent: Manager → Coder → Reviewer → Tester | 10 min | MOF | 07 |
| 09 | Memory SQLite: por que o agente lembra | 8 min | MOF | 08 |
| 10 | Evolve: agente refatora o próprio harness | 10 min | BOF | 09 |
| 11 | Custo real: `harness cost` e budget routes | 6 min | TOF | qualquer |
| 12 | Deploy: Docker Compose + nginx | 9 min | MOF | 07 |
| 13 | CI local: `make ci` mata GitHub Actions | 7 min | TOF | 01 |
| 14 | Release v1: tag, changelog, homebrew | 8 min | MOF | 13 |
| A1 | Apêndice: trocar Claude por Kimi sem quebrar | 6 min | TOF | 03 |
| A2 | Apêndice: rodar tudo offline com `fake` | 6 min | MOF | 03 |

---

### PL2 — "Papers na Prática" (10 vídeos)

1 paper por vídeo, código Go real do harnessx apontado por
`docs/PAPER-MAPPING.md` e `docs/PAPER-IMPLEMENTATION.md`. Formato:
2 min contexto + 5 min paper + 8 min código + 2 min recap.

Lista canônica em `SERIE-PAPERS-NA-PRATICA.md`:

| # | Título | Duração | Camada | Prereq |
|---|---|---|---|---|
| E1 | O paper que economizou 60% do custo com Claude (LLMLingua-2) | 10 min | MOF | nenhum |
| E2 | Lost in the Middle: por que Claude ignora o meio do prompt | 10 min | BOF | E1 |
| E3 | Reflexion: o agente que se critica e melhora | 10 min | BOF | E1 |
| E4 | BM25: o algoritmo de 1994 que ainda bate embedding em 2026 | 10 min | BOF | E2 |
| E5 | Prompt Caching Anthropic: 90% de desconto que 90% ignora | 10 min | MOF | E1 |
| E6 | Context Engineering: o termo do Karpathy virou disciplina | 10 min | MOF | E4 |
| E7 | Voyager: skills que evoluem sozinhos com gate de benchmark | 10 min | BOF | E3 |
| E8 | CAMEL + Multi-Agent Debate: 5 papéis num blackboard.json | 10 min | BOF | E6 |
| E9 | Constitutional AI: sensor gating como versão determinística | 10 min | BOF | E3 |
| E10 | MCP: o USB-C dos LLMs (com scanner de segurança) | 10 min | MOF | nenhum |

Citação-base do harnessx: Ning et al. arXiv 2605.18747v1 (mai/2026,
102 pp) — mapeada em `PAPER-MAPPING.md`. Cada episódio referencia o
paper original + arquivo Go que implementa (ver
`SERIE-PAPERS-NA-PRATICA.md` §referências).

---

### PL3 — "Dentro do harnessx" (18 vídeos)

Code walkthrough de cada pacote `internal/`. Formato: 1 min contexto
+ 8 min código + 1 min teste rodando. Lista canônica em
`SERIE-DENTRO-DO-HARNESSX.md`.

| # | Título | Duração | Camada | Prereq |
|---|---|---|---|---|
| EP01 | Arquitetura em 10min: por que Clean Arch importa | 10 min | MOF | nenhum |
| EP02 | Cobra + lipgloss: como montar CLI em 60 arquivos | 10 min | BOF | EP01 |
| EP03 | Sensors: o firewall determinístico | 10 min | BOF | EP01 |
| EP04 | Secrets Sensor: regex `AKIA…` vs .env commitado | 10 min | BOF | EP03 |
| EP05 | Budget Sensor: token counting sem surpresa | 10 min | BOF | EP03 |
| EP06 | Context Engineering: git + rg + LSP + rerank BM25 | 10 min | BOF | EP01 |
| EP07 | Adapter HTTP: falando com Claude/OpenAI/Gemini | 10 min | BOF | EP01 |
| EP08 | Fake Adapter: E2E sem gastar 1 centavo | 10 min | BOF | EP07 |
| EP09 | Router: primary/fallback + decisão explicável | 10 min | BOF | EP07 |
| EP10 | Memory SQLite: evidence + confidence floor 0.4 | 10 min | BOF | EP01 |
| EP11 | Spec-driven: markdown template + gate | 10 min | BOF | EP01 |
| EP12 | Plan Contract: escopo declarado, escopo forçado | 10 min | BOF | EP11 |
| EP13 | Skills: versionadas + gate por benchmark | 10 min | BOF | EP10 |
| EP14 | TUI Bubble Tea: workflow view em tempo real | 10 min | BOF | EP02 |
| EP15 | Dashboard React: métricas via `/api` | 10 min | BOF | EP01 |
| EP16 | MCP: templates YAML pra 1-click install | 10 min | BOF | EP07 |
| EP17 | Design-to-Product: ZIP → feature-map.json | 10 min | BOF | EP01 |
| EP18 | Evolve: harness que se conserta sob HITL | 10 min | BOF | EP03 |

---

### PL4 — "Agentes IA em 2026" (10 vídeos)

Conceituais, comparativos, sem exigir Go. Vitrine do canal para
audiência ampla. Cada vídeo termina com CTA cruzada para PL1.

| # | Título | Duração | Camada | Prereq |
|---|---|---|---|---|
| A00 | Vibe coding é mentira: o que agente de verdade faz | 8 min | TOF | nenhum |
| A01 | Claude vs Codex vs Kimi vs Gemini 3 em Go real | 10 min | TOF | nenhum |
| A02 | MCP em 2026: pra que serve, quando não usar | 9 min | TOF | nenhum |
| A03 | Claude Code vs Cursor vs Windsurf vs Antigravity | 10 min | TOF | nenhum |
| A04 | O custo real de agente por mês (com nota fiscal) | 8 min | TOF | nenhum |
| A05 | Spec-driven dev: o que muda quando LLM escreve | 9 min | MOF | A00 |
| A06 | Sandbox de agente: por que Docker não basta | 9 min | MOF | A00 |
| A07 | Multi-agent é hype ou o futuro real | 10 min | TOF | A00 |
| A08 | Alucinação de código: como sensor mata | 9 min | MOF | A05 |
| A09 | Rodolfo responde: 20 perguntas sobre agente em 2026 | 12 min | TOF | qualquer |

---

### PL5 — "Do Zero ao SaaS Deployado" (10 vídeos)

Bootstrap real: do domínio ao Stripe. Complementar a PL1 (PL1 = código;
PL5 = negócio). Público misto dev + fundador solo.

| # | Título | Duração | Camada | Prereq |
|---|---|---|---|---|
| S00 | Fiz um SaaS por US$ 12 no primeiro mês | 8 min | TOF | nenhum |
| S01 | Comprar domínio, DNS, e-mail: `rogpe.tech` na prática | 7 min | TOF | S00 |
| S02 | VPS ou Fly.io ou Railway em 2026 | 9 min | MOF | S00 |
| S03 | HTTPS, nginx, certificados sem drama | 8 min | MOF | S02 |
| S04 | Stripe: checkout, webhook, refund em Go | 12 min | MOF | S02 |
| S05 | Landing que converte: skill.dev por dentro | 10 min | TOF | S04 |
| S06 | Analytics sem espião: Plausible + Clarity | 8 min | MOF | S05 |
| S07 | Suporte solo: Crisp, e-mail, SLA de 1 dev | 8 min | MOF | S05 |
| S08 | LGPD sem advogado (mínimo viável) | 10 min | MOF | S04 |
| S09 | Primeiros 100 usuários: canais que funcionaram | 9 min | TOF | S05 |

---

## 2. Funil de crescimento por playlist

Modelo TOF → MOF → BOF. Cada playlist mistura camadas, mas a
proporção muda:

| Playlist | TOF | MOF | BOF | Papel no funil |
|---|---|---|---|---|
| PL1 TaskHive | 24 % | 65 % | 12 % | Aquisição + retenção |
| PL2 Papers | 8 % | 8 % | 84 % | Autoridade, "moat" |
| PL3 Dentro do harnessx | 6 % | 6 % | 88 % | Nicho profundo, subs qualificados |
| PL4 Agentes IA 2026 | 70 % | 30 % | 0 % | Máquina de descoberta |
| PL5 Zero ao Deploy | 40 % | 60 % | 0 % | Ponte dev → founder |

### Fluxo de tráfego pretendido

```
        [YouTube search / Home]
                │
        ┌───────┴───────┐
        ▼               ▼
     PL4 A00         PL5 S00           ← TOF, hooks amplos
    (vibe é         (US$ 12 no
     mentira)        1º mês)
        │               │
        └───────┬───────┘
                ▼
             PL1 v00 (série TaskHive)      ← MOF, série longa
                │
        ┌───────┼───────┐
        ▼       ▼       ▼
      PL2     PL3    PL5 S04          ← BOF, retém super-fã
    Papers  Interno  Stripe
```

### CTAs cruzadas (obrigatórias no roteiro)

- **Fim de PL4 A00** → card + end-screen: "Se aceita a tese, o
  próximo passo é ver isso rodando: PL1 v00."
- **Fim de PL4 A01 (Claude vs Codex vs Kimi)** → PL1 A1 (trocar
  Claude por Kimi).
- **Fim de qualquer PL1** → PL2 do paper correspondente ao capítulo
  (mapa abaixo em §4).
- **Fim de PL2 E6 (Context Engineering)** → PL3 EP06 (rerank + BM25).
- **Fim de PL3** → PL1 v08 (multi-agent aplicado).
- **Fim de PL5 S04** → PL1 v11 (custo real) + PL1 v12 (deploy).

---

## 3. Ordem de gravação priorizada por ROI

Critério: menor custo de produção (roteiro já existe / é curto / não
exige B-roll pesado) × maior potencial de views (hook amplo, keyword
com volume, evergreen).

### Os 10 primeiros a gravar

| Ordem | Vídeo | Por quê |
|---|---|---|
| 1 | PL4 A00 "Vibe coding é mentira" | Roteiro conceitual, sem tela, hook viral, evergreen. Bootstrap do canal. |
| 2 | PL5 S00 "SaaS por US$ 12 no 1º mês" | Prova social, número no título, rosto+objeto na thumb. |
| 3 | PL1 v00 "TaskHive: SaaS por US$ 1,50 em LLM" | Roteiro pronto em `ROTEIROS-TECNICOS.md`. Ancoragem da série. |
| 4 | PL4 A01 "Claude vs Codex vs Kimi vs Gemini 3" | Comparativo = alto CTR. Reaproveita `harness use` de todos. |
| 5 | PL1 v11 "Custo real: `harness cost`" | 6 min, standalone, prova concreta, alimenta 1, 2, 3. |
| 6 | PL1 v13 "CI local mata GitHub Actions" | Polêmica leve + evergreen dev BR. |
| 7 | PL4 A04 "Custo real de agente por mês" | Extensão do #5 para audiência não-Go. |
| 8 | PL2 E1 "O paper que economizou 60% do custo (LLMLingua-2)" | Autoridade, número no título, dor real. |
| 9 | PL1 v01 "harness init em 5 min" | Já roteirizado. Fundação da série. |
| 10 | PL4 A02 "MCP em 2026: pra que serve" | Keyword em alta, competição BR baixa. |

Justificativa geral: os 5 primeiros são **standalone** (não exigem
pré-req nem contexto de série), com hook amplo e rosto+número em
thumb — bootstrap saudável antes da série longa começar.

---

## 4. Cross-linking matrix

Mapa canônico. Cada vídeo referencia 2–4 outros via **card no minuto**
específico, **end-screen** e **descrição**. Não usar mais que 4
referências no mesmo vídeo (Creator Insider — dilui CTR).

### PL1 → PL2 (série → paper)

| PL1 | Referencia PL2 |
|---|---|
| v02 (spec) | E9 (Constitutional AI / sensor gating) |
| v03 (`ship`) | E6 (Context Engineering) |
| v04 (JWT + sensors) | E9 (Constitutional AI) |
| v05 (plan-check) | E9 (gating determinístico) |
| v08 (multi-agent) | E8 (CAMEL + Multi-Agent Debate) |
| v09 (memory) | E1 (LLMLingua-2), E2 (Lost in the Middle) |
| v10 (evolve) | E7 (Voyager — skills evolutivas) |
| v11 (cost) | E5 (Prompt Caching), E1 (LLMLingua-2) |

### PL1 → PL3 (série → walkthrough interno)

| PL1 | Referencia PL3 |
|---|---|
| v01 (init) | EP01 (arquitetura), EP02 (Cobra) |
| v03 (ship) | EP11 (spec), EP12 (plan contract) |
| v04 (sensors) | EP03 (sensors), EP04 (secrets) |
| v08 (multi-agent) | EP09 (router) |
| v09 (memory) | EP10 (memory) |
| v10 (evolve) | EP18 (evolve) |
| v11 (cost) | EP05 (budget) |

### PL4 → PL1 (topo → série)

| PL4 | Referencia PL1 |
|---|---|
| A00 | v00, v03 |
| A01 | A1 (trocar adapter) |
| A02 | v08 |
| A03 | A1 |
| A04 | v11 |
| A05 | v02 |
| A06 | v10 |
| A08 | v04 |

### PL5 → PL1 e PL4

| PL5 | Referencia |
|---|---|
| S00 | PL1 v00, PL4 A04 |
| S02 | PL1 v12 |
| S04 | PL1 v11 |
| S05 | PL4 A00 |
| S09 | PL4 A09 |

### Regra do "primeiro card"

Todo vídeo põe o **primeiro card no ponto de tensão** (não no início).
Ponto de tensão = momento em que a tela mostra o resultado que o
título prometeu. Fonte: Creator Insider — cards ganham CTR quando
disparam depois do payoff, não antes.

---

## 5. Brand system da série

### Nome oficial do canal

**Rodolfo Peixoto** (pessoa; identidade principal).
**Selo técnico do conteúdo:** `rogpe.tech`.
**SaaS/produto associado:** `skill.dev`.

### Tagline

> "Agente escreve. Sensor aprova. Você fica com o merge."

Alternativas para A/B em thumb/descrição:

- "Código com agente. Sem alucinação."
- "IA que compila. E passa no teste."

### Paleta oficial

Canônica de `docs/youtube/BRAND-THUMB-TITLE-MATRIX.md`. Contraste
WCAG AA sobre fundo `#0B0B0F` `[VERIFICAR com WebAIM contrast
checker antes de fechar brand book v1]`.

| Uso | Nome | Hex |
|---|---|---|
| Fundo dark (canônico) | Vault Black `#0B0B0F` | `#0B0B0F` |
| Texto principal | Terminal White | `#E6E8EC` |
| Marca / prompt CRT | CRT Green | `#00E38A` |
| Highlight primário | Sensor Yellow | `#F5C518` |
| Alerta / erro | Alert Red | `#FF3B3B` |
| Link/CTA secundário | Route Cyan | `#39C5CF` |
| Acento sutil | Slate | `#4E5A65` |

Regra: thumbnail usa **no máximo 3 cores** (Vault Black `#0B0B0F` + Terminal
White + 1 acento por playlist).

Acento por playlist:

- PL1 → Route Cyan (série)
- PL2 → Sensor Yellow (papers, autoridade)
- PL3 → CRT Green (código real, "verde de teste passa")
- PL4 → Alert Red (opinião, tensão)
- PL5 → mistura Cyan + Yellow (dev + founder)

### Tipografia

- **Títulos on-screen:** Inter Bold 900, tracking -2 %.
- **Terminal/código:** JetBrains Mono 18 pt (já herdado dos roteiros
  v1).
- **Thumbnail:** Inter Black, no máximo 4 palavras, tamanho mínimo
  120 px de altura no 1280×720.

### Logo / mascote sugerido

Marca-d'água canto inferior direito: cifrão `$` estilizado como cursor
de terminal `$_` piscando. Não usar mascote antropomórfico —
diluiria posicionamento sério.

Avatar do canal: monograma `rp` em Vault Black `#0B0B0F` com underline Sensor
Yellow.

### Template intro (5 s, obrigatória em todo vídeo)

```
0.0s  Vault Black `#0B0B0F`
0.3s  Cursor $_ aparece, sound: keypress único
1.0s  Título do vídeo digita em Inter Bold
3.0s  Highlight Sensor Yellow varre o título
4.0s  Fade para primeiro frame de conteúdo
```

Sem música. Silêncio + keypress. Fonte: canais como Fireship e
ThePrimeagen mantêm intros ≤ 3 s e AVD dos primeiros 30 s ≥ 85 %
`[VERIFICAR — Social Blade / declaração pública]`.

### Template outro (10 s, obrigatória)

```
0–3s   Recap em 3 bullets on-screen (não repetir tudo)
3–6s   Card end-screen: próximo vídeo da mesma playlist
6–10s  Card end-screen: playlist relacionada + subscribe pulsando
```

Voz: "Próximo vídeo: <X>. Playlist completa: <Y>. Inscreve."
Nada de "curte, comenta, compartilha" — dilui o CTA principal
(Creator Insider Q4 2025 `[VERIFICAR]`).

---

## 6. Frequência de publicação 2026

Cadência **semanal**, um vídeo por semana, alternando playlists por
regra fixa para que o feed não fique monotemático.

### Rotação semanal (rolling 4 semanas)

| Semana | Playlist | Racional |
|---|---|---|
| S1 | PL4 (TOF) | Aquisição de descoberta |
| S2 | PL1 (MOF) | Retenção da série |
| S3 | PL2 ou PL3 (BOF) | Fidelização super-fã |
| S4 | PL5 (TOF/MOF) | Ponte founder |

12 vídeos por trimestre → 48 vídeos/ano. Buffer de 18 vídeos para o
segundo ano ou para gravar "em lote" (block-shooting) e sobreviver a
semanas ocupadas.

### Dia + horário

- **Quinta-feira 19:00 BRT** (horário-âncora).
- Fonte: vidIQ 2026 recomenda meio de semana à noite para dev BR
  `[VERIFICAR]`. TubeBuddy 2026 confirma pico quinta 18h–22h para
  audiência tech LatAm `[VERIFICAR]`.
- Shorts: 3 por semana (seg/qua/sex 12:00 BRT), cortes dos vídeos
  longos + trecho de terminal com highlight.

### Ancoragem em eventos

| Evento | Ação |
|---|---|
| Release novo do harnessx (`vX.Y.0`) | Shorts no mesmo dia + vídeo PL1 "o que mudou" na semana |
| Paper novo relevante (arXiv agents) | Vídeo PL2 dentro de 10 dias |
| Drama IA (novo modelo, polêmica) | Short reativo em 24 h + PL4 na semana seguinte |
| Aniversário harnessx | Vídeo PL3 "1 ano do repo" |

Regra: nunca deixar release do harnessx passar sem shorts. É a
principal fonte de authority signal do canal.

---

## 7. Métricas alvo por playlist

Bandas metas. Todas medidas nas **primeiras 48 h** (CTR) e no
**vitalício** (AVD, watch-time).

| Playlist | CTR 48h | AVD | Watch-time cumulativo | Subs/1000 views |
|---|---|---|---|---|
| PL1 TaskHive | 6–9 % | ≥ 55 % | ≥ 5 min | 12–20 |
| PL2 Papers | 4–6 % | ≥ 60 % | ≥ 8 min | 25–40 |
| PL3 Dentro do harnessx | 3–5 % | ≥ 60 % | ≥ 7 min | 30–50 |
| PL4 Agentes 2026 | 8–12 % | ≥ 45 % | ≥ 3 min | 6–12 |
| PL5 Zero ao Deploy | 6–9 % | ≥ 50 % | ≥ 4 min | 10–18 |

Fontes: Creator Insider Q1 2026 para bandas gerais `[VERIFICAR]`;
comparativos com nichos dev EN via Social Blade
<https://socialblade.com/> (Fireship, ThePrimeagen, Theo). Nichos
técnicos BR historicamente entregam **subs/1000 mais alto e CTR mais
baixo** que o médio EN — banda ajustada.

### Métrica-chave por camada

- **TOF (PL4, PL5):** maximizar **impressões × CTR** — é a boca do
  funil.
- **MOF (PL1):** maximizar **AVD absoluto em segundos** — é o que o
  ranker do YouTube mais pondera (Covington et al., §3.2).
- **BOF (PL2, PL3):** maximizar **subs/1000 views** e comentários —
  sinal de comunidade, não de alcance.

---

## 8. Estratégia de viralização por vídeo âncora

Vídeos âncora são os 3 primeiros a gravar (ordem §3) mais os que
tocam release. Aplico a todos o mesmo playbook:

### 8.1 Título — "hook + tensão + prova"

Fórmula: **[quem/o quê]** + **[tensão]** + **[prova concreta]**.

Exemplos aprovados:

- "Vibe coding é mentira. Provei em 8 minutos de terminal."
- "Fiz um SaaS por US\$ 1,50 em LLM. Aqui tá a nota."
- "Claude vs Codex vs Kimi: quem escreve Go que compila?"

Regra: **um número no título quando houver**. Fonte: MrBeast
workshop 2024 revalidado 2026 `[VERIFICAR link]`.

### 8.2 Thumbnail — contraste rosto + número + objeto

Formato:

- Lado esquerdo: rosto do Rodolfo com expressão específica (curioso,
  duvidoso — nunca boca aberta genérica).
- Lado direito: número gigante (`$1.50`, `8min`, `86 comandos`) ou
  objeto físico (terminal com stack trace vermelho).
- Contraste Vault Black `#0B0B0F` + acento da playlist (§5).
- Zero fundo genérico, zero setinha vermelha.

3 variações por thumb, teste A/B via YouTube Studio Test & Compare
por 7 dias no mínimo.

### 8.3 Primeiro comentário fixado (micro-thread)

Fixar comentário próprio nos primeiros 60 s de publicação. Formato:

```
🧵 Sobre o vídeo:
1/ <ponto que ficou de fora do corte>
2/ <link para PL relacionada>
3/ <pergunta aberta pro público>
```

O ponto 3 é obrigatoriamente **pergunta aberta** — engagement rate
sobe [VERIFICAR banda] quando o creator devolve pergunta em vez de
fechar tópico.

### 8.4 Reply engagement — primeiros 30 min

Janela: 0–30 min após publicar. Ação: responder **todos** os
comentários com resposta ≥ 2 frases. Depois de 30 min, responder
apenas os que agregam.

Racional: Covington et al. §3.2 — sinais de engajamento nas
primeiras horas pesam desproporcionalmente na home feed.

### 8.5 Cross-post adaptado

| Plataforma | Formato | Prazo |
|---|---|---|
| LinkedIn | Post técnico 800 chars + 1 screenshot terminal + link | 2 h após publicar |
| Twitter/X | Thread 5 tweets: hook, prova, código, custo, link | Imediato |
| BlueSky | Versão da thread X + 1 GIF de 6 s do terminal | Imediato |
| Newsletter rogpe.tech | Resumo 400 palavras + link + próximo vídeo | Sexta seguinte |
| Reddit r/golang, r/brdev | Só quando relevante e sem link no título | 24 h após |

Nunca fazer TikTok/Reels de vídeo técnico completo — corte de 45 s
com terminal + código apenas.

---

## 9. Colabs e community

### Canais BR para colab (short-list)

| Canal | Formato sugerido | Ponto de contato |
|---|---|---|
| Filipe Deschamps | Reacts a "SaaS por US\$ 1,50" ou entrevista | público |
| Rocketseat / Diego Fernandes | Aula convidada sobre agentes IA em Go | parceria com Rocketseat |
| Fabio Akita | Debate técnico "vibe coding é mentira" | público |
| Attekita Dev | Live coding pareado no harnessx | público |
| Código Fonte TV | Explicador conceitual PL4 | agência |
| Lucas Montano | Debate sobre custo real de IA no dev | público |

### Canais EN para colab (fase 2, 10k+ subs)

| Canal | Formato |
|---|---|
| ThePrimeagen | React ao paper Ning et al. |
| Theo (t3.gg) | Debate MCP vs adapter |
| Fireship | Sugestão de tema para 100-seconds |
| Beyond Fireship | Tutorial harnessx |

Regra dura: **nenhuma colab antes do canal ter 20 vídeos publicados
e 5 vídeos com > 5k views**. Colab prematuro dilui posicionamento.

### Community posts (semanais, terça)

Rotação:

- Semana 1: enquete técnica ("Você já usou MCP em produção?")
- Semana 2: screenshot de log/terminal + pergunta
- Semana 3: trecho de paper com comentário curto
- Semana 4: preview do próximo vídeo (3 frames)

---

## 10. Roadmap 12 meses

Cortes trimestrais. Cada trimestre = 12 vídeos + 36 shorts.

### Q1 2026 — Bootstrap (0 → 1 000 subs)

**Foco:** provar conceito, ganhar sinal do algoritmo.

Vídeos entregues (ordem §3):

1. PL4 A00 · 2. PL5 S00 · 3. PL1 v00 · 4. PL4 A01 · 5. PL1 v11
6. PL1 v13 · 7. PL4 A04 · 8. PL2 E1 · 9. PL1 v01 · 10. PL4 A02
11. PL1 v02 · 12. PL5 S01

Meta subs: **1 000**.
Monetização: **zero** (canal ainda não elegível para monetização
YPP em maioria dos casos com < 1 000 subs / 4 000 h).

### Q2 2026 — Aprofundamento (1 k → 10 k subs)

**Foco:** completar série TaskHive + primeiros papers.

Playlists ativas: PL1 (v03–v10), PL2 (E2–E5), PL4 (A03, A05).

Meta subs: **10 000**.
Monetização:

- YouTube Partner Program ativado (adsense, ~US\$ 200–800/mês
  `[VERIFICAR CPM Br dev 2026]`).
- Membership canal — **R\$ 9,90/mês**, oferece:
  - Acesso ao `.harness/` de referência
  - Discord privado
- Curso pago 1 em `skill.dev`: "harnessx do zero" — R\$ 297.

### Q3 2026 — Autoridade (10 k → 30 k subs)

**Foco:** PL2 completa + PL3 lançada. Deixar de ser "só série".

Playlists ativas: PL2 (E6–E10), PL3 (EP01–EP09), PL1 (finalizar
v12–v14 + apêndices), PL4 (A06–A08).

Meta subs: **30 000**.
Monetização:

- Patrocínio: 1 vídeo/mês (regra: só ferramentas que Rodolfo usa —
  Anthropic, Kimi, Fly.io, Turso, etc.).
- Curso 2 em `skill.dev`: "Papers de agentes IA na prática" —
  R\$ 497.

### Q4 2026 — Escala (30 k → 50 k subs)

**Foco:** PL5 completa, PL3 finalizada, primeira colab EN.

Playlists ativas: PL5 (S02–S09), PL3 (EP10–EP18), PL4 (A09).

Meta subs: **50 000**.
Monetização:

- Membership tier 2 (R\$ 29,90/mês): call mensal + review de repo
- Curso 3: "Do zero ao SaaS deployado" — R\$ 897.
- Bundle 3-em-1: R\$ 1 497.

### Sinais de aborto/re-plano

Se aos 6 meses:

- CTR médio < 3 % → refazer thumbnails em bloco.
- AVD < 40 % em PL1 → cortar duração alvo para 6 min (Guo/Kim/Rubin).
- Subs/1000 < 5 em PL2/PL3 → problema de posicionamento, não de
  conteúdo — revisar títulos.

---

## Apêndice A — Naming e slugs de arquivo

Padrão de arquivo local para vídeos gravados:

```
docs/youtube/videos/<PL>-<NUM>-<slug-kebab>/
  ├── script.md         (roteiro final)
  ├── thumb-a.png
  ├── thumb-b.png
  ├── description.md    (descrição YouTube, 2 parágrafos + timestamps)
  ├── tags.txt
  └── state.yaml        (data publicação, métricas 48h, 7d, 30d)
```

Exemplo: `docs/youtube/videos/PL1-00-taskhive-saas-1-50/`.

---

## Apêndice B — Checklist pré-publicação (cada vídeo)

- [ ] Título com hook + tensão + prova (≤ 60 chars)
- [ ] 2 thumbs A/B, ≤ 3 cores, número ou rosto grande
- [ ] Descrição: 2 parágrafos + keyword primária + timestamps
- [ ] Tags ≤ 500 chars
- [ ] Chapter "00:00 Intro"
- [ ] End-screen: 1 vídeo (mesma PL) + 1 playlist (outra PL)
- [ ] Cards no ponto de tensão (não no início)
- [ ] Comentário fixado draftado
- [ ] LinkedIn + X + BlueSky drafts prontos
- [ ] `state.yaml` criado
- [ ] Publicação agendada 19:00 BRT quinta
