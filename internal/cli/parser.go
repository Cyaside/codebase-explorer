package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Cyaside/codebase-explorer/internal/app"
)

type Service interface {
	Analyze(context.Context, app.AnalyzeRequest) (app.AnalyzeResult, error)
	Doctor(context.Context, app.DoctorRequest) (app.DoctorResult, error)
	Open(context.Context, app.OpenRequest) (app.OpenResult, error)
	Export(context.Context, app.ExportRequest) (app.ExportResult, error)
	ClearCache(context.Context, app.CacheClearRequest) (app.CacheClearResult, error)
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
		command.analyzeRequest.Progress = composeAnalyzeProgress(command.analyzeRequest.Progress, stderr)
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
	case "cache-clear":
		result, runErr := service.ClearCache(ctx, app.CacheClearRequest{})
		if runErr != nil {
			return 1, runErr
		}
		printCacheClearResult(stdout, result)
		return 0, nil
	case "open":
		result, runErr := service.Open(ctx, command.openRequest)
		if runErr != nil {
			return 1, runErr
		}

		var openErr error
		if !command.openRequest.NoBrowser {
			openErr = openViewer(ctx, result.ViewerPath)
		}
		printOpenResult(stdout, result, openErr, command.openRequest.NoBrowser)
		return 0, nil
	case "export":
		result, runErr := service.Export(ctx, command.exportRequest)
		if runErr != nil {
			return 1, runErr
		}
		printExportResult(stdout, result)
		return 0, nil
	default:
		return 2, fmt.Errorf("unknown command %q", command.name)
	}
}

type parsedCommand struct {
	name           string
	analyzeRequest app.AnalyzeRequest
	openRequest    app.OpenRequest
	exportRequest  app.ExportRequest
}

var errUsage = errors.New("usage requested")

func parse(args []string) (parsedCommand, error) {
	if len(args) == 0 {
		return parsedCommand{}, errUsage
	}

	switch args[0] {
	case "-h", "--help", "help":
		return parsedCommand{}, errUsage
	case "cache":
		if len(args) == 2 && args[1] == "clear" {
			return parsedCommand{name: "cache-clear"}, nil
		}
		return parsedCommand{}, fmt.Errorf("unknown cache command")
	case "doctor":
		if len(args) > 1 {
			return parsedCommand{}, fmt.Errorf("doctor does not accept additional arguments")
		}
		return parsedCommand{name: "doctor"}, nil
	case "open":
		return parseOpen(args[1:])
	case "export":
		return parseExport(args[1:])
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
		case current == "--support" || current == "--issues" || current == "--changelog":
			index++
			if index >= len(args) {
				return parsedCommand{}, fmt.Errorf("%s requires a value", current)
			}
			request.OptionalSupportFiles = append(request.OptionalSupportFiles, args[index])
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

func parseOpen(args []string) (parsedCommand, error) {
	request := app.OpenRequest{}

	for index := 0; index < len(args); index++ {
		current := args[index]
		switch {
		case current == "--no-browser":
			request.NoBrowser = true
		case strings.HasPrefix(current, "--"):
			return parsedCommand{}, fmt.Errorf("unknown open flag %q", current)
		case request.BundlePath == "":
			request.BundlePath = current
		default:
			return parsedCommand{}, fmt.Errorf("open accepts at most one bundle path")
		}
	}

	return parsedCommand{
		name:        "open",
		openRequest: request,
	}, nil
}

func parseExport(args []string) (parsedCommand, error) {
	request := app.ExportRequest{}

	for index := 0; index < len(args); index++ {
		current := args[index]
		switch {
		case current == "--output":
			index++
			if index >= len(args) {
				return parsedCommand{}, fmt.Errorf("--output requires a value")
			}
			request.OutputPath = args[index]
		case strings.HasPrefix(current, "--"):
			return parsedCommand{}, fmt.Errorf("unknown export flag %q", current)
		case request.BundlePath == "":
			request.BundlePath = current
		default:
			return parsedCommand{}, fmt.Errorf("export accepts at most one bundle path")
		}
	}

	return parsedCommand{
		name:          "export",
		exportRequest: request,
	}, nil
}

func openViewer(ctx context.Context, viewerPath string) error {
	absolutePath, err := filepath.Abs(viewerPath)
	if err != nil {
		return fmt.Errorf("resolve viewer path %q: %w", viewerPath, err)
	}

	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", absolutePath)
	case "darwin":
		command = exec.CommandContext(ctx, "open", absolutePath)
	default:
		command = exec.CommandContext(ctx, "xdg-open", absolutePath)
	}

	if err := command.Start(); err != nil {
		return fmt.Errorf("open viewer %q: %w", absolutePath, err)
	}

	return nil
}
