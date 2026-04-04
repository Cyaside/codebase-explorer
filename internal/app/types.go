package app

type AnalyzeRequest struct {
	RepoPath             string
	OutputRoot           string
	DeterministicOnly    bool
	ExtraIgnorePatterns  []string
	OptionalSupportFiles []string
	Progress             AnalyzeProgressReporter
}

type AnalyzeResult struct {
	OutputPath      string
	ProjectName     string
	ProjectType     string
	TotalFiles      int
	TotalLines      int
	EntryPoints     []string
	PrimaryLanguage string
	AI              AnalyzeAISummary
	Output          AnalyzeOutputSummary
}

type AnalyzeProgressEvent struct {
	Stage  string
	Status string
	Detail string
}

type AnalyzeProgressReporter func(AnalyzeProgressEvent)

type AnalyzeAISummary struct {
	Status           string
	Provider         string
	Model            string
	Used             bool
	Note             string
	ContextSummary   string
	ContextTruncated bool
}

type AnalyzeOutputSummary struct {
	RetentionLimit int
	PrunedBundles  int
}

type OpenRequest struct {
	BundlePath string
	NoBrowser  bool
}

type OpenResult struct {
	BundlePath     string
	ViewerPath     string
	ResolvedLatest bool
}

type DoctorRequest struct{}

type DoctorCheck struct {
	Name   string
	Status string
	Detail string
}

type DoctorResult struct {
	ConfigSource string
	OutputRoot   string
	Checks       []DoctorCheck
}
