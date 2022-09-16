package helper

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/apply"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/prune"
)

// ApplyFlagsToApplyOptions converts flags to option.
// Copied from kubectl package, because of removed API.
func ApplyFlagsToApplyOptions(flags *apply.ApplyFlags) (*apply.ApplyOptions, error) {
	serverSideApply := false
	forceConflicts := true
	dryRunStrategy := cmdutil.DryRunNone

	dynamicClient, err := flags.Factory.DynamicClient()
	if err != nil {
		return nil, err
	}

	dryRunVerifier := resource.NewQueryParamVerifier(dynamicClient, flags.Factory.OpenAPIGetter(), resource.QueryParamDryRun)
	fieldValidationVerifier := resource.NewQueryParamVerifier(dynamicClient, flags.Factory.OpenAPIGetter(), resource.QueryParamFieldValidation)
	fieldManager := apply.FieldManagerClientSideApply

	// allow for a success message operation to be specified at print time
	toPrinter := func(operation string) (printers.ResourcePrinter, error) {
		flags.PrintFlags.NamePrintFlags.Operation = operation
		cmdutil.PrintFlagsWithDryRunStrategy(flags.PrintFlags, dryRunStrategy)
		return flags.PrintFlags.ToPrinter()
	}

	recorder, err := flags.RecordFlags.ToRecorder()
	if err != nil {
		return nil, err
	}

	deleteOptions, err := flags.DeleteFlags.ToOptions(dynamicClient, flags.IOStreams)
	if err != nil {
		return nil, err
	}

	// We don't need this validation
	// err = deleteOptions.FilenameOptions.RequireFilenameOrKustomize()
	// if err != nil {
	// 	return nil, err
	// }

	openAPISchema, err := flags.Factory.OpenAPISchema()
	if err != nil {
		println(err)
	}

	validationDirective := "Warn"

	validator, err := flags.Factory.Validator(validationDirective, fieldValidationVerifier)
	if err != nil {
		return nil, err
	}
	builder := flags.Factory.NewBuilder()
	mapper, err := flags.Factory.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	namespace, enforceNamespace, err := flags.Factory.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return nil, err
	}

	if flags.Prune {
		flags.PruneResources, err = prune.ParseResources(mapper, flags.PruneWhitelist)
		if err != nil {
			return nil, err
		}
	}

	o := &apply.ApplyOptions{
		PrintFlags: flags.PrintFlags,

		DeleteOptions:   deleteOptions,
		ToPrinter:       toPrinter,
		ServerSideApply: serverSideApply,
		ForceConflicts:  forceConflicts,
		FieldManager:    fieldManager,
		Selector:        flags.Selector,
		DryRunStrategy:  dryRunStrategy,
		DryRunVerifier:  dryRunVerifier,
		Prune:           flags.Prune,
		PruneResources:  flags.PruneResources,
		All:             flags.All,
		Overwrite:       flags.Overwrite,
		OpenAPIPatch:    flags.OpenAPIPatch,
		PruneWhitelist:  flags.PruneWhitelist,

		Recorder:            recorder,
		Namespace:           namespace,
		EnforceNamespace:    enforceNamespace,
		Validator:           validator,
		ValidationDirective: validationDirective,
		Builder:             builder,
		Mapper:              mapper,
		DynamicClient:       dynamicClient,
		OpenAPISchema:       openAPISchema,

		IOStreams: flags.IOStreams,

		VisitedUids:       sets.NewString(),
		VisitedNamespaces: sets.NewString(),
	}

	o.PostProcessorFn = o.PrintAndPrunePostProcessor()

	return o, nil
}

// NewApplyOptions creates an apply options.
func NewApplyOptions(manifest string, stream genericclioptions.IOStreams) (*apply.ApplyOptions, error) {
	restClient := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()

	b := resource.NewBuilder(restClient)
	res := b.Unstructured().Stream(strings.NewReader(manifest), "manifestString").Do()
	infos, err := res.Infos()
	if err != nil {
		return nil, err
	}

	f := cmdutil.NewFactory(restClient)

	applyFlags := apply.NewApplyFlags(f, stream)

	applyOpts, err := ApplyFlagsToApplyOptions(applyFlags)
	if err != nil {
		return nil, err
	}

	applyOpts.SetObjects(infos)

	return applyOpts, nil
}
