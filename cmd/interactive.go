package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/user/grep-tool/internal/output"
	"github.com/user/grep-tool/internal/search"
	"github.com/user/grep-tool/internal/walker"
)

func isInterrupt(err error) bool {
	return errors.Is(err, terminal.InterruptErr)
}

// ask routes every prompt through a single seam.
func ask(p survey.Prompt, result interface{}) error {
	return survey.AskOne(p, result)
}

// --- UI helpers (shared palette from the output package) ---

func hr() string { return output.Accent + output.Separator + output.Reset }

func printBanner() {
	fmt.Println()
	fmt.Println(hr())
	fmt.Printf("  %s🔍 grep-tool  —  Modo Interativo%s\n", output.Bold+output.Accent, output.Reset)
	fmt.Println(hr())
	fmt.Println()
}

func printSearching(icon, term, dir string) {
	fmt.Println()
	fmt.Println(hr())
	fmt.Printf("  %s%s Buscando%s %s%s%s em %s%s%s...\n",
		output.Yellow, icon, output.Reset, output.Bold, term, output.Reset, output.Bold, dir, output.Reset)
	fmt.Println(hr())
	fmt.Println()
}

func printErr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "  %sErro:%s %s\n", output.Red, output.Reset, fmt.Sprintf(format, args...))
}

func printInfo(msg string) {
	fmt.Printf("  %s%s%s\n", output.Cyan, msg, output.Reset)
}

func printWarn(msg string) {
	fmt.Printf("  %s%s%s\n", output.Yellow, msg, output.Reset)
}

// Output destinations for a search.
const (
	destTerminal = iota
	destFile
)

// chooseDestination asks where the results should go, before the search runs.
func chooseDestination() (int, error) {
	dest := destTerminal
	prompt := &survey.Select{
		Message: output.Bold + "📤 Onde mostrar o resultado?" + output.Reset,
		Options: []string{
			"🖥️   No terminal",
			"💾  Salvar em Downloads",
		},
	}
	err := ask(prompt, &dest)
	return dest, err
}

// saveAndReport writes content to a report file and prints a short summary
// (used by the "salvar em Downloads" destination).
func saveAndReport(prefix, summary, content string) {
	path, err := saveReport(prefix, content)
	if err != nil {
		printErr("não foi possível salvar: %v", err)
		return
	}
	printInfo("✅ " + summary)
	printInfo("📄 Salvo em: " + path)
	fmt.Println()
}

// --- menu ---

const (
	menuGrep = iota
	menuFind
	menuExit
)

func runInteractive() error {
	formatter := output.NewFormatter(output.ColorAlways)

	lastPattern := ""
	lastDir := "."
	lastFindName := ""

	for {
		printBanner()

		action := menuExit
		prompt := &survey.Select{
			Message: output.Bold + "O que você quer fazer?" + output.Reset,
			Options: []string{
				"🔎  Buscar palavra em arquivos",
				"📁  Buscar arquivo por nome",
				"❌  Sair",
			},
			PageSize: 10,
		}
		if err := ask(prompt, &action); err != nil {
			return nil // Ctrl-C / EOF no menu → sair limpo
		}

		switch action {
		case menuGrep:
			p, d, err := runInteractiveGrep(formatter, lastPattern, lastDir)
			if isInterrupt(err) {
				return nil
			}
			if err == nil && p != "" {
				lastPattern, lastDir = p, d
			}

		case menuFind:
			n, d, err := runInteractiveFind(lastFindName, lastDir)
			if isInterrupt(err) {
				return nil
			}
			if err == nil && n != "" {
				lastFindName, lastDir = n, d
			}

		case menuExit:
			fmt.Printf("\n  %sAté logo!%s\n", output.Cyan, output.Reset)
			return nil
		}
	}
}

func runInteractiveGrep(formatter *output.Formatter, lastPattern, lastDir string) (string, string, error) {
	pattern := ""
	promptPattern := &survey.Input{
		Message: output.Bold + "📝 Palavra para buscar" + output.Reset + " (Enter vazio volta ao menu):",
		Default: lastPattern,
	}
	if err := ask(promptPattern, &pattern); err != nil {
		return "", "", err
	}
	if pattern == "" {
		return "", "", nil
	}

	dir := orDot(lastDir)
	ignoreCase := false
	useRegex := false

	advanced := false
	if err := ask(&survey.Confirm{Message: output.Bold + "⚙️  Opções avançadas?" + output.Reset}, &advanced); err != nil {
		return "", "", err
	}
	if advanced {
		d, err := pickDirectory(dir)
		if err != nil {
			return "", "", err
		}
		dir = d
		if err := ask(&survey.Confirm{Message: "  🔤 Ignorar maiúscula/minúscula?"}, &ignoreCase); err != nil {
			return "", "", err
		}
		if err := ask(&survey.Confirm{Message: "  📐 Usar expressão regular?"}, &useRegex); err != nil {
			return "", "", err
		}
	}

	dest, err := chooseDestination()
	if err != nil {
		return "", "", err
	}

	printSearching("🔎", pattern, dir)
	runGrepSearch(formatter, pattern, dir, ignoreCase, useRegex, dest == destFile)
	return pattern, dir, nil
}

func runInteractiveFind(lastName, lastDir string) (string, string, error) {
	name := ""
	promptName := &survey.Input{
		Message: output.Bold + "📝 Nome do arquivo" + output.Reset + " (use * como curinga; Enter vazio volta ao menu):",
		Default: lastName,
	}
	if err := ask(promptName, &name); err != nil {
		return "", "", err
	}
	if name == "" {
		return "", "", nil
	}

	if _, err := filepath.Match(name, "x"); err != nil {
		printErr("padrão inválido: %v", err)
		return "", "", nil
	}

	dir := orDot(lastDir)
	advanced := false
	if err := ask(&survey.Confirm{Message: output.Bold + "📂  Escolher outra pasta?" + output.Reset}, &advanced); err != nil {
		return "", "", err
	}
	if advanced {
		d, err := pickDirectory(dir)
		if err != nil {
			return "", "", err
		}
		dir = d
	}

	dest, err := chooseDestination()
	if err != nil {
		return "", "", err
	}
	toFile := dest == destFile

	printSearching("📁", name, dir)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ch, err := walker.Walk(ctx, dir)
	if err != nil {
		printErr("%v", err)
		return name, dir, nil
	}

	var matches []string
	for path := range ch {
		if matched, _ := filepath.Match(name, filepath.Base(path)); matched {
			if !toFile {
				fmt.Printf("  📄 %s\n", path)
			}
			matches = append(matches, path)
		}
	}

	if len(matches) == 0 {
		printWarn("Nenhum arquivo encontrado.")
		fmt.Println()
		return name, dir, nil
	}

	summary := fmt.Sprintf("%d arquivo(s) encontrado(s).", len(matches))

	if toFile {
		saveAndReport("arquivos", summary, buildFindReport(name, dir, matches))
		return name, dir, nil
	}

	fmt.Println()
	printInfo(summary)
	fmt.Println()
	return name, dir, nil
}

// buildFindReport renders a plain-text report of matched file paths.
func buildFindReport(name, dir string, matches []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "grep-tool — Relatório de busca de arquivos\n")
	fmt.Fprintf(&b, "Padrão:   %s\n", name)
	fmt.Fprintf(&b, "Pasta:    %s\n", dir)
	fmt.Fprintf(&b, "Data:     %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "Total:    %d arquivo(s)\n\n", len(matches))
	for _, p := range matches {
		b.WriteString(p + "\n")
	}
	return b.String()
}

func runGrepSearch(formatter *output.Formatter, pattern, dir string, ignoreCase, useRegex, toFile bool) {
	match, err := buildMatcher(pattern, ignoreCase, useRegex)
	if err != nil {
		printErr("%v", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	results, err := search.Search(ctx, []string{dir}, match, workers)
	if err != nil {
		printErr("%v", err)
		return
	}

	var all []search.Result
	files := map[string]struct{}{}
	hasErr := false
	for r := range results {
		if r.Err != nil {
			fmt.Fprintln(os.Stderr, formatter.FormatResult(r))
			hasErr = true
			continue
		}
		all = append(all, r)
		files[r.Path] = struct{}{}
	}

	if len(all) == 0 {
		if !hasErr {
			printWarn("Nenhum resultado encontrado.")
			fmt.Println()
		}
		return
	}

	summary := fmt.Sprintf("%d resultado(s) em %d arquivo(s).", len(all), len(files))

	if toFile {
		saveAndReport("busca", summary, buildGrepReport(pattern, dir, all, len(files)))
		return
	}

	fmt.Print(formatter.FormatGrouped(all))
	fmt.Println()
	printInfo(summary)
	fmt.Println()
}

// buildGrepReport renders a plain-text (no ANSI) report of grep results.
func buildGrepReport(pattern, dir string, all []search.Result, fileCount int) string {
	plain := output.NewFormatter(output.ColorNever)
	var b strings.Builder
	fmt.Fprintf(&b, "grep-tool — Relatório de busca\n")
	fmt.Fprintf(&b, "Palavra:  %s\n", pattern)
	fmt.Fprintf(&b, "Pasta:    %s\n", dir)
	fmt.Fprintf(&b, "Data:     %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "Total:    %d resultado(s) em %d arquivo(s)\n\n", len(all), fileCount)
	b.WriteString(plain.FormatGrouped(all))
	return b.String()
}

// orDot returns "." when s is empty.
func orDot(s string) string {
	if s == "" {
		return "."
	}
	return s
}

const dirPrefix = "📂  "

// pickDirectory lets the user navigate the filesystem with the arrow keys and
// choose a directory — no path typing required. Returns the chosen directory,
// or start unchanged if the user cancels. err is non-nil only on interrupt.
func pickDirectory(start string) (string, error) {
	cur, err := filepath.Abs(orDot(start))
	if err != nil {
		cur = orDot(start)
	}

	const (
		optHere = "✅  Buscar nesta pasta"
		optUp   = "⬆️   .. (subir um nível)"
		optBack = "⬅️   Cancelar"
	)

	for {
		subdirs, rderr := subdirsOf(cur)
		if rderr != nil {
			printErr("%v", rderr)
			parent := filepath.Dir(cur)
			if parent == cur {
				return start, nil
			}
			cur = parent
			continue
		}

		opts := []string{optHere, optUp}
		for _, d := range subdirs {
			opts = append(opts, dirPrefix+d)
		}
		opts = append(opts, optBack)

		choice := ""
		prompt := &survey.Select{
			Message:  fmt.Sprintf("%s📂 Pasta: %s%s", output.Bold, cur, output.Reset),
			Options:  opts,
			PageSize: 12,
		}
		if err := ask(prompt, &choice); err != nil {
			return start, err
		}

		switch {
		case choice == optHere:
			return cur, nil
		case choice == optUp:
			cur = filepath.Dir(cur)
		case choice == optBack:
			return start, nil
		default:
			cur = filepath.Join(cur, strings.TrimPrefix(choice, dirPrefix))
		}
	}
}

// subdirsOf returns the visible (non-hidden) subdirectory names of dir.
func subdirsOf(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs, nil
}
