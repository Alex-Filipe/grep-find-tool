package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/user/grep-tool/internal/matcher"
	"github.com/user/grep-tool/internal/output"
	"github.com/user/grep-tool/internal/search"
	"github.com/user/grep-tool/internal/walker"
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

		action := ""
		prompt := &survey.Select{
			Message: "\033[1mO que você quer fazer?\033[0m",
			Options: []string{
				"🔎  Buscar palavra em arquivos",
				"📁  Buscar arquivo por nome",
				"❌  Sair",
			},
			PageSize: 10,
		}
		survey.AskOne(prompt, &action)

		switch {
		case strings.Contains(action, "Sair"):
			fmt.Println("\n  \033[1;36mAté logo!\033[0m")
			return nil

		case strings.Contains(action, "palavra"):
			if err := runInteractiveGrep(formatter, sep); err != nil {
				return err
			}

		case strings.Contains(action, "nome"):
			if err := runInteractiveFind(sep); err != nil {
				return err
			}
		}
	}
}

func runInteractiveGrep(formatter *output.Formatter, sep string) error {
	pattern := ""
	promptPattern := &survey.Input{
		Message: "\033[1m📝 Digite a palavra para buscar:\033[0m",
		Help:    "A palavra ou expressão que será procurada nos arquivos",
	}
	if err := survey.AskOne(promptPattern, &pattern, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	dir := "."
	promptDir := &survey.Input{
		Message: "\033[1m📂 Diretório para busca:\033[0m",
		Default: ".",
		Help:    "Deixe . para buscar na pasta atual",
	}
	survey.AskOne(promptDir, &dir)

	ignoreCase := false
	promptCase := &survey.Confirm{
		Message: "\033[1m🔤 Ignorar maiúscula/minúscula?\033[0m",
		Default: false,
	}
	survey.AskOne(promptCase, &ignoreCase)

	useRegex := false
	promptRegex := &survey.Confirm{
		Message: "\033[1m📐 Usar expressão regular?\033[0m",
		Default: false,
	}
	survey.AskOne(promptRegex, &useRegex)

	workers := 4
	promptWorkers := &survey.Input{
		Message: "\033[1m⚡ Trabalhadores paralelos:\033[0m",
		Default: "4",
	}
	workersStr := "4"
	survey.AskOne(promptWorkers, &workersStr)
	fmt.Sscanf(workersStr, "%d", &workers)

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
			return fmt.Errorf("regex inválida: %w", err)
		}
	} else {
		match = matcher.NewLiteral(pattern, ignoreCase)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	roots := strings.Fields(dir)
	results, err := search.Search(ctx, roots, match, workers)
	if err != nil {
		return err
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
	cont := false
	promptCont := &survey.Confirm{
		Message: "\033[1m🔄 Fazer outra busca?\033[0m",
		Default: true,
	}
	survey.AskOne(promptCont, &cont)
	if !cont {
		fmt.Println("\n  \033[1;36mAté logo!\033[0m")
		os.Exit(0)
	}

	return nil
}

func runInteractiveFind(sep string) error {
	name := ""
	promptName := &survey.Input{
		Message: "\033[1m📝 Digite o nome do arquivo (use * como curinga):\033[0m",
		Help:    "Ex: *.go, test_*, relatorio*.pdf",
	}
	if err := survey.AskOne(promptName, &name, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	dir := "."
	promptDir := &survey.Input{
		Message: "\033[1m📂 Diretório para busca:\033[0m",
		Default: ".",
	}
	survey.AskOne(promptDir, &dir)

	if _, err := filepath.Match(name, "x"); err != nil {
		return fmt.Errorf("padrão inválido %q: %w", name, err)
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
		return err
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
	cont := false
	promptCont := &survey.Confirm{
		Message: "\033[1m🔄 Fazer outra busca?\033[0m",
		Default: true,
	}
	survey.AskOne(promptCont, &cont)
	if !cont {
		fmt.Println("\n  \033[1;36mAté logo!\033[0m")
		os.Exit(0)
	}

	return nil
}
