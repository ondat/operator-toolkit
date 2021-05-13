package v1

import (
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/darkowlzz/operator-toolkit/constant"
	"github.com/darkowlzz/operator-toolkit/telemetry"
)

// Name of the tracer.
const tracerName = constant.LibraryName + "/controller/sync"

// Reconciler defines a sync reconciler.
type Reconciler struct {
	Name          string
	Ctrlr         Controller
	Prototype     client.Object
	PrototypeList client.ObjectList
	Client        client.Client
	Scheme        *runtime.Scheme
	Log           logr.Logger
	SyncFuncs     []SyncFunc
	Inst          *telemetry.Instrumentation
}

// ReconcilerOption is used to configure Reconciler.
type ReconcilerOption func(*Reconciler)

// WithName sets the name of the Reconciler.
func WithName(name string) ReconcilerOption {
	return func(s *Reconciler) {
		s.Name = name
	}
}

// WithClient sets the k8s client in the reconciler.
func WithClient(cli client.Client) ReconcilerOption {
	return func(s *Reconciler) {
		s.Client = cli
	}
}

// WithPrototype sets a prototype of the object that's reconciled.
func WithPrototype(obj client.Object) ReconcilerOption {
	return func(s *Reconciler) {
		s.Prototype = obj
	}
}

// WithLogger sets the Logger in a Reconciler.
func WithLogger(log logr.Logger) ReconcilerOption {
	return func(s *Reconciler) {
		s.Log = log
	}
}

// WithScheme sets the runtime Scheme of the Reconciler.
func WithScheme(scheme *runtime.Scheme) ReconcilerOption {
	return func(s *Reconciler) {
		s.Scheme = scheme
	}
}

// WithSyncFuncs sets the syncFuncs of the Reconciler.
func WithSyncFuncs(sf []SyncFunc) ReconcilerOption {
	return func(s *Reconciler) {
		s.SyncFuncs = sf
	}
}

// WithInstrumentation configures the instrumentation of the Reconciler.
func WithInstrumentation(tp trace.TracerProvider, mp metric.MeterProvider) ReconcilerOption {
	return func(s *Reconciler) {
		s.Inst = telemetry.NewInstrumentation(tracerName, tp, mp)
	}
}

// Init initializes the Reconciler for a given Object with the given
// options.
func (s *Reconciler) Init(mgr ctrl.Manager, ctrlr Controller, prototype client.Object, prototypeList client.ObjectList, opts ...ReconcilerOption) error {
	s.Ctrlr = ctrlr

	// Use manager if provided. This is helpful in tests to provide explicit
	// client and scheme without a manager.
	if mgr != nil {
		s.Client = mgr.GetClient()
		s.Scheme = mgr.GetScheme()
	}

	// Use prototype and prototypeList if provided.
	if prototype != nil {
		s.Prototype = prototype
	}
	if prototypeList != nil {
		s.PrototypeList = prototypeList
	}

	// Add defaults.
	s.Log = ctrl.Log

	// Run the options to override the defaults.
	for _, opt := range opts {
		opt(s)
	}

	// If a name is set, log it as the reconciler name.
	if s.Name != "" {
		s.Log = s.Log.WithValues("reconciler", s.Name)
	}

	// If instrumentation is nil, create a new instrumentation with default
	// providers.
	if s.Inst == nil {
		s.Inst = telemetry.NewInstrumentation(tracerName, nil, nil)
	}

	// Run the sync functions.
	s.RunSyncFuncs()

	return nil
}

// RunSyncFuncs runs all the SyncFuncs in go routines.
func (s *Reconciler) RunSyncFuncs() {
	for _, sf := range s.SyncFuncs {
		go sf.Run()
	}
}
