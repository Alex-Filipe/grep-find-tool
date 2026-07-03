# Plano de implementação

Ordem bottom-up: folhas puras primeiro, wiring por último. Cada passo é
testável e commitável isoladamente.

```
matcher ─┐
extract ─┼─▶ search ─▶ output ─▶ cmd
walker  ─┘
```

## 1. matcher ✅
- `NewLiteral` (literal, `ignoreCase` via `strings.ToLower` fora do closure).
- `NewRegex` (`regexp.Compile`, prefixo `(?i)` para case-insensitive).
- Nota de perf: `ToLower(line)` aloca por linha — otimizar só após medir com `bench`.

## 2. extract ✅
- `textExtractor`: abre uma vez, `bufio.Reader.Peek(512)`, pula binário (NUL byte).
- Contrato: binário → `(nil, nil)` (pular); erro de I/O → `(nil, err)` (propagar).
- `.pdf` é stub (`TODO`); `For(ext)` nil → search pula.

## 3. walker
- `filepath.WalkDir` → canal, envio bloqueante (backpressure entra aqui).
- Respeita `ctx.Done()`. Pula `.git/` e diretórios ocultos.
- Erro de permissão: `continue`, não abortar.
- `.gitignore` fica como `TODO` (adiar; se for implementar, usar lib).

## 4. search (núcleo)
- Worker pool sobre todos os `roots`, fan-out walk → extract → match.
- Canal buffered limitado (`make(chan Result, workers*2)`) = memória plana.
- `bufio.Scanner` linha-a-linha (com `scanner.Buffer` p/ linhas grandes).
- Erros por-arquivo embutidos em `Result.Err`; canal fecha ao fim.
- Saída **não-ordenada** (consequência do streaming/backpressure).

## 5. output — UX de CLI
- `FormatResult`: cor no path, `linha:coluna` e no trecho que casou (highlight do match).
- **TTY-aware**: colorido no terminal, texto puro em pipe/arquivo.
  Flag `--color=auto/always/never`, `auto` como default.
- Formato `caminho:linha:texto`.
- Agrupamento por arquivo (path no cabeçalho, linhas indentadas) é **opcional**
  via flag `--sort` — exige bufferizar, então NÃO é o default (quebraria a
  memória plana). Trade-off "rápido" vs "organizado" fica com o usuário.

## 6. cmd (grep, find) — wiring + UX
- Validar `args`: pattern obrigatório, paths default `["."]` (múltiplos aceitos).
- Flag regex vs literal (literal é default; regex opt-in).
- `cmd` traduz flags → `matcher.MatchFunc` e passa para `Search`.
- Consumir o canal; imprimir `Result.Err` em `stderr` (sem sujar stdout).
- `find` reusa `walker` + `filepath.Match` (não usa `search`).
- Ctrl-C via `signal.NotifyContext` alimentando o `ctx`.
- `--help` vem formatado do cobra.

## Marco intermediário
Após o passo 4 já há busca ponta-a-ponta testável, mesmo sem CLI.

## Pendências fora do fluxo
- `go.mod` precisa declarar `require github.com/spf13/cobra` (+ `go.sum`).
- `.gitignore` do próprio repo; workflow de CI (lint/bench já no Makefile).
