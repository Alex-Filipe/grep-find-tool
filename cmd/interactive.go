package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/user/grep-tool/internal/matcher"
	"github.com/user/grep-tool/internal/output"
	"github.com/user/grep-tool/internal/search"
	"github.com/user/grep-tool/internal/walker"
)

func IsInterrupt(err error) bool {
	return errors.Is(err, terminal.InterruptErr)
}

const (
	menuGrep = iota
	menuFind
	menuExit
)

func runInteractive() error {
	formatter := output.NewFormatter(output.ColorAlways)
	sep := "\033[38;5;208m─────────────────────────────────────\033[0m"

	for {
		fmt.Println()
		fmt.Println(sep)
		fmt.Println("  \033[1;38;5;208m🔍 grep-tool  —  Modo Interativo\033[0m")
		fmt.Println(sep)
		fmt.Println()

		action := menuExit
		prompt := &survey.Select{
			Message: "\033[1mO que você quer fazer?\033[0m",
			Options: []string{
				"🔎  Buscar palavra em arquivos",
				"📁  Buscar arquivo por nome",
				"❌  Sair",
			},
			PageSize: 10,
		}
		if err := ask(prompt, &action); err != nil {
			return nil
		}

		switch action {
		case menuGrep:
			runInteractiveGrep(formatter, sep)
		case menuFind:
			runInteractiveFind(sep)
		case menuExit:
			fmt.Println("\n  \033[1;36mAté logo!\033[0m")
			return nil
		}
	}
}

func ask(p survey.Prompt, result interface{}) error {
	if err := survey.AskOne(p, result); err != nil {
		if IsInterrupt(err) {
			return err
		}
	}
	return nil
}

func askAdvanced() (bool, bool) {
	ignoreCase := false
	useRegex := false

	advanced := false
	promptAdv := &survey.Confirm{
		Message: "\033[1m⚙️  Opções avançadas?\033[0m",
		Default: false,
	}
	if err := ask(promptAdv, &advanced); err != nil || !advanced {
		return ignoreCase, useRegex
	}

	promptCase := &survey.Confirm{
		Message: "  🔤 Ignorar maiúscula/minúscula?",
		Default: false,
	}
	ask(promptCase, &ignoreCase)

	promptRegex := &survey.Confirm{
		Message: "  📐 Usar expressão regular?",
		Default: false,
	}
	ask(promptRegex, &useRegex)

	return ignoreCase, useRegex
}

func runInteractiveGrep(formatter *output.Formatter, sep string) {
	pattern := ""
	promptPattern := &survey.Input{
		Message: "\033[1m📝 Digite a palavra para buscar:\033[0m",
	}
	if err := ask(promptPattern, &pattern); err != nil || pattern == "" {
		return
	}

	dir := "."
	promptDir := &survey.Input{
		Message: "\033[1m📂 Diretório:\033[0m",
		Default: ".",
	}
	if err := ask(promptDir, &dir); err != nil {
		return
	}

	ignoreCase, useRegex := askAdvanced()

	fmt.Println()
	fmt.Println(sep)
	fmt.Printf("  \033[1;33m🔎 Buscando\033[0m \033[1m%s\033[0m em \033[1m%s\033[0m...\n", pattern, dir)
	fmt.Println(sep)
	fmt.Println()

	var match matcher.MatchFunc
	if useRegex {
		var err error
		match, err = matcher.NewRegex(pattern, ignoreCase)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\033[1;31mErro:\033[0m regex inválida: %v\n", err)
			return
		}
	} else {
		match = matcher.NewLiteral(pattern, ignoreCase)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	results, err := search.Search(ctx, []string{dir}, match, workers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[1;31mErro:\033[0m %v\n", err)
		return
	}

	var all []search.Result
	hasErr := false
	for r := range results {
		if r.Err != nil {
			fmt.Fprintln(os.Stderr, formatter.FormatResult(r))
			hasErr = true
			continue
		}
		all = append(all, r)
	}

	if len(all) == 0 && !hasErr {
		fmt.Println("  \033[1;33mNenhum resultado encontrado.\033[0m")
	} else if len(all) > 0 {
		fmt.Print(formatter.FormatGrouped(all))
	}

	fmt.Println()
}

func runInteractiveFind(sep string) {
	name := ""
	promptName := &survey.Input{
		Message: "\033[1m📝 Nome do arquivo (use * como curinga):\033[0m",
	}
	if err := ask(promptName, &name); err != nil || name == "" {
		return
	}

	if _, err := filepath.Match(name, "x"); err != nil {
		fmt.Fprintf(os.Stderr, "\033[1;31mErro:\033[0m padrão inválido: %v\n", err)
		return
	}

	dir := "."
	promptDir := &survey.Input{
		Message: "\033[1m📂 Diretório:\033[0m",
		Default: ".",
	}
	if err := ask(promptDir, &dir); err != nil {
		return
	}

	fmt.Println()
	fmt.Println(sep)
	fmt.Printf("  \033[1;33m📁 Buscando\033[0m \033[1m%s\033[0m em \033[1m%s\033[0m...\n", name, dir)
	fmt.Println(sep)
	fmt.Println()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ch, err := walker.Walk(ctx, dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[1;31mErro:\033[0m %v\n", err)
		return
	}

	found := 0
	for path := range ch {
		if matched, _ := filepath.Match(name, filepath.Base(path)); matched {
			fmt.Printf("  📄 %s\n", path)
			found++
		}
	}

	if found == 0 {
		fmt.Println("  \033[1;33mNenhum arquivo encontrado.\033[0m")
	} else {
		fmt.Printf("\n  \033[1;36m%d arquivo(s) encontrado(s).\033[0m\n", found)
	}

	fmt.Println()
}
