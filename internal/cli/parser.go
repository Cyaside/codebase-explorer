package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/app"
)

type Service interface {
	Analyze(context.Context, app.AnalyzeRequest) (app.AnalyzeResult, error)
	Doctor(context.Context, app.DoctorRequest) (app.DoctorResult, error)
}

func Run(ctx context.Context, args []string, service Service, stdout, stderr io.Writer) (int, error) {
	command, err := parse(args)
	if err != nil {
		if errors.Is(err, errUsage) {
			printUsage(stdout)
			return 0, nil
		}
		printUsage(stderr)
		return 2, err
	}

	switch command.name {
	case "analyze":
		result, runErr := service.Analyze(ctx, command.analyzeRequest)
		if runErr != nil {
			return 1, runErr
		}
		printAnalyzeResult(stdout, result)
		return 0, nil
	case "doctor":
		result, runErr := service.Doctor(ctx, app.DoctorRequest{})
		if runErr != nil {
			return 1, runErr
		}
		printDoctorResult(stdout, result)
		return 0, nil
	default:
		return 2, fmt.Errorf("unknown command %q", command.name)
	}
}

type parsedCommand struct {
	name           string
	analyzeRequest app.AnalyzeRequest
}

var errUsage = errors.New("usage requested")

func parse(args []string) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errUsage
	}

	switch args[0] {
	case "-h", "--help", "help":
		return parsedCommand{}, errUsage
	case "doctor":
		if len(args) > 1 {
			return parsedCommand{}, fmt.Errorf("doctor does not accept additional arguments")
		}
		return parsedCommand{name: "doctor"}, nil
	case "analyze":
		return parseAnalyze(args[1:])
	default:
		return parsedCommand{}, fmt.Errorf("unknown command %q", args[0])
	}
}

func parseAnalyze(args []string) (parsedCommand, error) {
	request := app.AnalyzeRequest{}

	for index := 0; index < len(args); index++ {
		current := args[index]
		switch {
		case current == "--deterministic-only":
			request.DeterministicOnly = true
		case current == "--output":
			index++
			if index >= len(args) {
				return parsedCommand{}, fmt.Errorf("--output requires a value")
			}
			request.OutputRoot = args[index]
		case current == "--ignore":
			index++
			if index >= len(args) {
				return parsedCommand{}, fmt.Errorf("--ignore requires a value")
			}
			request.ExtraIgnorePatterns = append(request.ExtraIgnorePatterns, args[index])
		case strings.HasPrefix(current, "--"):
			return parsedCommand{}, fmt.Errorf("unknown analyze flag %q", current)
		case request.RepoPath == "":
			request.RepoPath = current
		default:
			request.OptionalSupportFiles = append(request.OptionalSupportFiles, current)
		}
	}

	if request.RepoPath == "" {
		return parsedCommand{}, fmt.Errorf("analyze requires a repository path")
	}

	return parsedCommand{
		name:           "analyze",
		analyzeRequest: request,
	}, nil
}
