package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

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

	printSearching("🔎", pattern, dir)
	runGrepSearch(formatter, pattern, dir, ignoreCase, useRegex)
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

	printSearching("📁", name, dir)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ch, err := walker.Walk(ctx, dir)
	if err != nil {
		printErr("%v", err)
		return name, dir, nil
	}

	found := 0
	for path := range ch {
		if matched, _ := filepath.Match(name, filepath.Base(path)); matched {
			fmt.Printf("  📄 %s\n", path)
			found++
		}
	}

	if found == 0 {
		printWarn("Nenhum arquivo encontrado.")
	} else {
		fmt.Println()
		printInfo(fmt.Sprintf("%d arquivo(s) encontrado(s).", found))
	}
	fmt.Println()
	return name, dir, nil
}

func runGrepSearch(formatter *output.Formatter, pattern, dir string, ignoreCase, useRegex bool) {
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

	fmt.Print(formatter.FormatGrouped(all))
	fmt.Println()
	printInfo(fmt.Sprintf("%d resultado(s) em %d arquivo(s).", len(all), len(files)))
	fmt.Println()
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
