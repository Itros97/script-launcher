package domain

import "strings"

// EnvStatus says how ready a Python project's environment is.
type EnvStatus int

const (
	NoProject  EnvStatus = iota // no folder chosen, or it is gone
	NoVenv                      // the folder has no .venv
	Incomplete                  // .venv exists but the project's module cannot be imported
	Ready                       // the module imports: everything needed to run is there
)

const (
	VenvDir          = ".venv"            // the environment's folder inside the project
	VenvPython       = ".venv/bin/python" // the environment's interpreter, relative to the project
	RequirementsFile = "requirements.txt" // the dependencies pip installs
)

// SetupStep is one step of preparing an environment: a label to show and, unless it is only a note,
// the command to run in the project folder.
type SetupStep struct {
	Label   string
	Program string
	Args    []string
}

// SetupPlan lists the steps that prepare a project's environment with the given Python and install
// the requirements file named; when it is missing ("") or lists no packages there is nothing to
// install, and the plan says so.
func SetupPlan(python, requirements string, hasPackages bool) []SetupStep {
	steps := []SetupStep{{Label: "Creando el entorno (" + VenvDir + ")", Program: python, Args: []string{"-m", "venv", VenvDir}}}
	if requirements == "" {
		return append(steps, SetupStep{Label: "Nada que instalar: no hay " + RequirementsFile})
	}
	if !hasPackages {
		return append(steps, SetupStep{Label: "Nada que instalar en " + requirements})
	}
	return append(steps,
		SetupStep{Label: "Actualizando pip", Program: VenvPython, Args: []string{"-m", "pip", "install", "--upgrade", "pip"}},
		SetupStep{Label: "Instalando dependencias (" + requirements + ")", Program: VenvPython, Args: []string{"-m", "pip", "install", "-r", requirements}},
	)
}

// SetupScript turns steps into one bash script that announces each step, stops at the first that
// fails and ends saying the environment is ready.
func SetupScript(steps []SetupStep) string {
	return InstallScript(steps) + "echo " + shellQuote("✓ Entorno preparado") + "\n"
}

// InstallScript turns steps into one bash script that announces each step and stops at the first that fails.
func InstallScript(steps []SetupStep) string {
	var script strings.Builder
	script.WriteString("set -e\n")
	for _, step := range steps {
		script.WriteString("echo " + shellQuote("▸ "+step.Label) + "\n")
		if step.Program == "" {
			continue
		}
		words := []string{shellQuote(step.Program)}
		for _, arg := range step.Args {
			words = append(words, shellQuote(arg))
		}
		script.WriteString(strings.Join(words, " ") + "\n")
	}
	return script.String()
}

// ListsPackages tells whether a requirements file's content asks for anything: a line that is not
// blank or a comment.
func ListsPackages(requirements string) bool {
	for _, line := range strings.Split(requirements, "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			return true
		}
	}
	return false
}

// shellQuote wraps a word in single quotes so bash reads it literally.
func shellQuote(word string) string {
	return "'" + strings.ReplaceAll(word, "'", `'\''`) + "'"
}
