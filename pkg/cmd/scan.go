package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	cmderrors "github.com/cloudskiff/driftctl/pkg/cmd/errors"
	"github.com/cloudskiff/driftctl/pkg/cmd/scan/output"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/cloudskiff/driftctl/pkg/iac/supplier"
	"github.com/cloudskiff/driftctl/pkg/iac/terraform/state/backend"
	"github.com/cloudskiff/driftctl/pkg/remote"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type ScanOptions struct {
	Coverage bool
	Detect   bool
	From     []config.SupplierConfig
	To       string
	Output   output.OutputConfig
	Filter   *jmespath.JMESPath
}

func NewScanCmd() *cobra.Command {
	opts := &ScanOptions{}

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan",
		Long:  "Scan",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			from, _ := cmd.Flags().GetStringSlice("from")

			iacSource, err := parseFromFlag(from)
			if err != nil {
				return err
			}

			opts.From = iacSource

			to, _ := cmd.Flags().GetString("to")
			if !remote.IsSupported(to) {
				return errors.Errorf(
					"unsupported cloud provider '%s'\nValid values are: %s",
					to,
					strings.Join(remote.GetSupportedRemotes(), ","),
				)
			}

			outputFlag, _ := cmd.Flags().GetString("output")
			out, err := parseOutputFlag(outputFlag)
			if err != nil {
				return err
			}
			opts.Output = *out

			filterFlag, _ := cmd.Flags().GetString("filter")
			if filterFlag != "" {
				expr, err := filter.BuildExpression(filterFlag)
				if err != nil {
					return errors.Wrap(err, "unable to parse filter expression")
				}
				opts.Filter = expr
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return scanRun(opts)
		},
	}

	fl := cmd.Flags()
	fl.StringP(
		"filter",
		"",
		"",
		"JMESPath expression to filter on\n"+
			"Examples : \n"+
			"  - Type == 'aws_s3_bucket' (will filter only s3 buckets)\n"+
			"  - Type =='aws_s3_bucket && Id != 'my_bucket' (excludes s3 bucket 'my_bucket')\n"+
			"  - Attr.Tags.Terraform == 'true' (include only resources that have Tag Terraform equal to 'true')\n",
	)
	fl.StringP(
		"output",
		"o",
		output.Example(output.ConsoleOutputType),
		"Output format, by default it will write to the console\n"+
			"Accepted formats are: "+strings.Join(output.SupportedOutputsExample(), ",")+"\n",
	)
	fl.StringSliceP(
		"from",
		"f",
		[]string{"tfstate://terraform.tfstate"},
		"IaC sources, by default try to find local terraform.tfstate file\n"+
			"Accepted schemes are: "+strings.Join(supplier.GetSupportedSchemes(), ",")+"\n",
	)
	supportedRemotes := remote.GetSupportedRemotes()
	fl.StringVarP(
		&opts.To,
		"to",
		"t",
		supportedRemotes[0],
		"Cloud provider source\n"+
			"Accepted values are: "+strings.Join(supportedRemotes, ",")+"\n",
	)

	return cmd
}

func scanRun(opts *ScanOptions) error {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	alerter := alerter.NewAlerter()
	providerLibrary := terraform.NewProviderLibrary()
	supplierLibrary := resource.NewSupplierLibrary()

	err := remote.Activate(opts.To, alerter, providerLibrary, supplierLibrary)
	if err != nil {
		return err
	}

	// Teardown
	defer func() {
		logrus.Trace("Exiting scan cmd")
		providerLibrary.Cleanup()
		logrus.Trace("Exited")
	}()

	scanner := pkg.NewScanner(supplierLibrary.Suppliers(), alerter)

	iacSupplier, err := supplier.GetIACSupplier(opts.From, providerLibrary)
	if err != nil {
		return err
	}
	ctl := pkg.NewDriftCTL(scanner, iacSupplier, opts.Filter, alerter)

	go func() {
		<-c
		logrus.Warn("Detected interrupt, cleanup ...")
		ctl.Stop()
	}()

	analysis, err := ctl.Run()

	if err != nil {
		return err
	}

	err = output.GetOutput(opts.Output).Write(analysis)
	if err != nil {
		return err
	}

	if !analysis.IsSync() {
		return cmderrors.InfrastructureNotInSync{}
	}

	return nil
}

func parseFromFlag(from []string) ([]config.SupplierConfig, error) {

	configs := make([]config.SupplierConfig, 0, len(from))

	for _, flag := range from {
		schemePath := strings.Split(flag, "://")
		if len(schemePath) != 2 || schemePath[1] == "" || schemePath[0] == "" {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nAccepted schemes are: %s",
						strings.Join(supplier.GetSupportedSchemes(), ","),
					),
				),
				"Unable to parse from flag '%s'",
				flag,
			)
		}

		scheme := schemePath[0]
		path := schemePath[1]
		supplierBackend := strings.Split(scheme, "+")
		if len(supplierBackend) > 2 {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(fmt.Sprintf(
					"\nAccepted schemes are: %s",
					strings.Join(supplier.GetSupportedSchemes(), ","),
				),
				),
				"Unable to parse from scheme '%s'",
				scheme,
			)
		}

		supplierKey := supplierBackend[0]
		if !supplier.IsSupplierSupported(supplierKey) {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nAccepted values are: %s",
						strings.Join(supplier.GetSupportedSuppliers(), ","),
					),
				),
				"Unsupported IaC source '%s'",
				supplierKey,
			)
		}

		backendString := ""
		if len(supplierBackend) == 2 {
			backendString = supplierBackend[1]
			if !backend.IsSupported(backendString) {
				return nil, errors.Wrapf(
					cmderrors.NewUsageError(
						fmt.Sprintf(
							"\nAccepted values are: %s",
							strings.Join(backend.GetSupportedBackends(), ","),
						),
					),
					"Unsupported IaC backend '%s'",
					backendString,
				)
			}
		}

		configs = append(configs, config.SupplierConfig{
			Key:     supplierKey,
			Backend: backendString,
			Path:    path,
		})
	}

	return configs, nil
}

func parseOutputFlag(out string) (*output.OutputConfig, error) {
	schemeOpts := strings.Split(out, "://")
	if len(schemeOpts) < 2 || schemeOpts[0] == "" {
		return nil, errors.Wrapf(
			cmderrors.NewUsageError(
				fmt.Sprintf(
					"\nAccepted formats are: %s",
					strings.Join(output.SupportedOutputsExample(), ","),
				),
			),
			"Unable to parse output flag '%s'",
			out,
		)
	}

	o := schemeOpts[0]
	if !output.IsSupported(o) {
		return nil, errors.Wrapf(
			cmderrors.NewUsageError(
				fmt.Sprintf(
					"\nValid formats are: %s",
					strings.Join(output.SupportedOutputsExample(), ","),
				),
			),
			"Unsupported output '%s'",
			o,
		)
	}

	opts := schemeOpts[1:]
	options := map[string]string{}

	switch o {
	case output.JSONOutputType:
		if len(opts) != 1 || opts[0] == "" {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nMust be of kind: %s",
						output.Example(output.JSONOutputType),
					),
				),
				"Invalid json output '%s'",
				out,
			)
		}
		options["path"] = opts[0]
	}

	return &output.OutputConfig{
		Key:     o,
		Options: options,
	}, nil
}
