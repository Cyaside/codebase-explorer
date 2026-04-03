package app

type AnalyzeRequest struct {
	RepoPath             string
	OutputRoot           string
	DeterministicOnly    bool
	ExtraIgnorePatterns  []string
	OptionalSupportFiles []string
}

type AnalyzeResult struct {
	OutputPath      string
	ProjectName     string
	ProjectType     string
	TotalFiles      int
	TotalLines      int
	EntryPoints     []string
	PrimaryLanguage string
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
