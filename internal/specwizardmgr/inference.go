package specwizardmgr

import (
	"errors"
	"fmt"
	"strings"

	specwizardmodel "ada-love-ide/internal/config/specwizard"
)

// dependencyError is returned when a required field is missing.
type dependencyError struct {
	Target     string
	Dependency string
}

func (e *dependencyError) Error() string {
	return "[BLOQUEIO] Não é possível inferir '" + e.Target + "' sem preencher o campo: '" + e.Dependency + "'"
}

// ValidarInferencia checks that all prerequisites for the given target field
// are present in the SpecWizardConfig. It returns nil if inference can proceed.
func ValidarInferencia(target string, cfg specwizardmodel.SpecWizardConfig) error {
	target = strings.ToUpper(target)

	fmt.Printf("[inference] ValidarInferencia target=%s expert=%q architecture=%q persistence=%q\n",
		target, cfg.ExpertLanguagePlugin, cfg.Architecture, cfg.Persistence)

	switch target {
	case "PRD":
		if strings.TrimSpace(cfg.ExpertLanguagePlugin) == "" {
			return &dependencyError{Target: target, Dependency: "Expert Language Plugin"}
		}
		fmt.Printf("[inference] PRD validation OK\n")

	case "FUNCTIONAL":
		if strings.TrimSpace(cfg.ExpertLanguagePlugin) == "" {
			return &dependencyError{Target: target, Dependency: "Expert Language Plugin"}
		}
		if strings.TrimSpace(cfg.PRD) == "" {
			return &dependencyError{Target: target, Dependency: "PRD"}
		}
		fmt.Printf("[inference] FUNCTIONAL validation OK (PRD len=%d)\n", len(cfg.PRD))

	case "NON_FUNCTIONAL":
		if strings.TrimSpace(cfg.ExpertLanguagePlugin) == "" {
			return &dependencyError{Target: target, Dependency: "Expert Language Plugin"}
		}
		if strings.TrimSpace(cfg.PRD) == "" {
			return &dependencyError{Target: target, Dependency: "PRD"}
		}
		if len(cfg.FunctionalRequirements) == 0 {
			return &dependencyError{Target: target, Dependency: "Functional Requirements"}
		}
		fmt.Printf("[inference] NON_FUNCTIONAL validation OK\n")

	case "API_CONTRACT":
		if err := checkArchitectureBlock(cfg); err != nil {
			return err
		}
		fmt.Printf("[inference] API_CONTRACT validation OK\n")

	case "CUSTOMIZATION":
		if err := checkArchitectureBlock(cfg); err != nil {
			return err
		}
		fmt.Printf("[inference] CUSTOMIZATION validation OK\n")

	case "FINAL_ADJUSTMENTS":
		if err := checkArchitectureBlock(cfg); err != nil {
			return err
		}
		if strings.TrimSpace(cfg.Business.APIContract) == "" {
			return &dependencyError{Target: target, Dependency: "API Contract"}
		}
		if strings.TrimSpace(cfg.Business.CustomizationDetails) == "" {
			return &dependencyError{Target: target, Dependency: "Customization Details"}
		}
		fmt.Printf("[inference] FINAL_ADJUSTMENTS validation OK\n")

	default:
		return errors.New("[BLOQUEIO] Campo de inferência desconhecido: " + target)
	}

	return nil
}

func checkArchitectureBlock(cfg specwizardmodel.SpecWizardConfig) error {
	fmt.Printf("[inference] checkArchitectureBlock: expert=%q arch=%q persist=%q philos=%d designs=%d data=%d deps=%d\n",
		cfg.ExpertLanguagePlugin, cfg.Architecture, cfg.Persistence,
		len(cfg.EngineeringPhilosophies), len(cfg.DesignPatterns), len(cfg.DataPatterns), len(cfg.DependencyManifest))

	if strings.TrimSpace(cfg.ExpertLanguagePlugin) == "" {
		return &dependencyError{Target: "bloco de arquitetura", Dependency: "Expert Language Plugin"}
	}
	if strings.TrimSpace(cfg.Architecture) == "" {
		return &dependencyError{Target: "bloco de arquitetura", Dependency: "Select Base Architecture"}
	}
	if strings.TrimSpace(cfg.Persistence) == "" {
		return &dependencyError{Target: "bloco de arquitetura", Dependency: "Persistence Strategy"}
	}
	if len(cfg.EngineeringPhilosophies) == 0 {
		return &dependencyError{Target: "bloco de arquitetura", Dependency: "Engineering Philosophies"}
	}
	if len(cfg.DesignPatterns) == 0 {
		return &dependencyError{Target: "bloco de arquitetura", Dependency: "Design Patterns"}
	}
	if len(cfg.DataPatterns) == 0 {
		return &dependencyError{Target: "bloco de arquitetura", Dependency: "Data Patterns"}
	}
	if len(cfg.DependencyManifest) == 0 {
		return &dependencyError{Target: "bloco de arquitetura", Dependency: "Dependency Manifest"}
	}
	return nil
}


