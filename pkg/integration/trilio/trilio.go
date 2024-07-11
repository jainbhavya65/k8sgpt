package trilio

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/viper"
)

const (
	TargetValidate = "TargetAnalyzer"
	BackupValidate = "BackupAnalyzer"
)

type Trilio struct{}

func NewTrilio() *Trilio {
	return &Trilio{}
}

func (p *Trilio) Deploy(namespace string) error {
	// no-op
	color.Green("Activating trilio integration...")
	// TODO(pintohutch): add timeout or inherit an upstream context
	// for better signal management.
	ctx := context.Background()
	kubecontext := viper.GetString("kubecontext")
	kubeconfig := viper.GetString("kubeconfig")
	client, err := kubernetes.NewClient(kubecontext, kubeconfig)
	if err != nil {
		color.Red("Error initialising kubernetes client: %v", err)
		os.Exit(1)
	}

	trilioVaultConfig, err := findTrilioInstallation(ctx, client.GetClient(), namespace)
	if err != nil {
		color.Red("Error discovering trilio workloads: %v", err)
		os.Exit(1)
	}
	if trilioVaultConfig == nil {
		color.Yellow(fmt.Sprintf(`Trilio installation not found in namespace: %s.
		Please ensure Trilio is deployed to analyze.`, namespace))
		return errors.New("no Trilio installation found")
	}
	// Prime state of the analyzer so
	color.Green("Found existing installation")
	return nil
}

func (p *Trilio) UnDeploy(_ string) error {
	// no-op
	// We just care about existing deployments.
	color.Yellow("Integration will leave Trilio resources deployed. This is an effective no-op in the cluster.")
	return nil
}

func (p *Trilio) AddAnalyzer(mergedMap *map[string]common.IAnalyzer) {
	(*mergedMap)[TargetValidate] = &TargetAnalyzer{}
	(*mergedMap)[BackupValidate] = &BackupAnalyzer{}
}

func (p *Trilio) GetAnalyzerName() []string {
	return []string{TargetValidate,BackupValidate}
}

func (p *Trilio) GetNamespace() (string, error) {
	return "", nil
}

func (p *Trilio) OwnsAnalyzer(analyzer string) bool {
	return (analyzer == TargetValidate) || (analyzer == BackupValidate)
}

func (t *Trilio) IsActivate() bool {
	activeFilters := viper.GetStringSlice("active_filters")

	for _, filter := range t.GetAnalyzerName() {
		for _, af := range activeFilters {
			if af == filter {
				return true
			}
		}
	}

	return false
}
