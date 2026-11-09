package domain

import (
	"os/exec"
	"strings"
	"testing"
)

func TestSetupPlanInstallsRequirementsWhenThereAreAny(t *testing.T) {
	steps := SetupPlan("python3", "requirements-macos.txt", true)
	if len(steps) != 3 || steps[0].Program != "python3" || steps[2].Args[len(steps[2].Args)-1] != "requirements-macos.txt" {
		t.Errorf("got %+v", steps)
	}
}

func TestSetupPlanOnlyCreatesTheVenvWithoutRequirements(t *testing.T) {
	for _, requirements := range []string{"", RequirementsFile} {
		steps := SetupPlan("python3", requirements, false)
		if len(steps) != 2 || steps[1].Program != "" || !strings.Contains(steps[1].Label, RequirementsFile) {
			t.Errorf("%q: got %+v", requirements, steps)
		}
	}
}

func TestSetupScriptStopsAtTheFirstFailingStep(t *testing.T) {
	script := SetupScript([]SetupStep{
		{Label: "fails", Program: "false"},
		{Label: "never", Program: "echo", Args: []string{"ran anyway"}},
	})
	output, err := exec.Command("bash", "-c", script).CombinedOutput()
	if err == nil || strings.Contains(string(output), "ran anyway") {
		t.Errorf("got %q, %v", output, err)
	}
}

func TestSetupScriptKeepsArgumentsLiteral(t *testing.T) {
	script := SetupScript([]SetupStep{{Label: "it's quoted", Program: "echo", Args: []string{"$HOME", "a b", "it's"}}})
	output, err := exec.Command("bash", "-c", script).Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), "▸ it's quoted\n$HOME a b it's\n") {
		t.Errorf("got %q", output)
	}
}

func TestListsPackagesIgnoresBlankLinesAndComments(t *testing.T) {
	if ListsPackages("") || ListsPackages("# nada\n\n   # tampoco\n") {
		t.Error("an empty requirements file asks for packages")
	}
	if !ListsPackages("# cámara\nnumpy>=1.24\n") {
		t.Error("numpy was not seen")
	}
}
