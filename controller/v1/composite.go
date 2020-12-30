package v1

import (
	"fmt"

	"github.com/go-logr/logr"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CleanupStrategy is the resource cleanup strategy used by the reconciler.
type CleanupStrategy int

const (
	// OwnerReferenceCleanup depends on k8s garbage collector. All the child
	// objects of a parent are added with a reference of the parent object.
	// When the parent object gets deleted, all the child objects are garbage
	// collected.
	OwnerReferenceCleanup CleanupStrategy = iota
	// FinalizerCleanup allows using custom cleanup logic. When this strategy
	// is set, a finalizer is added to the parent object to avoid accidental
	// deletion of the object. When the object is marked for deletion with a
	// deletion timestamp, the custom cleanup code is executed to delete all
	// the child objects. Once all custom cleanup code finished, the finalizer
	// from the parent object is removed and the parent object is allowed to be
	// deleted.
	FinalizerCleanup
)

// CompositeReconciler defines a composite reconciler.
type CompositeReconciler struct {
	name            string
	initCondition   conditionsv1.Condition
	finalizerName   string
	cleanupStrategy CleanupStrategy
	log             logr.Logger
	ctrlr           Controller
	prototype       client.Object
	client          client.Client
	scheme          *runtime.Scheme
}

// CompositeReconcilerOptions is used to configure CompositeReconciler.
type CompositeReconcilerOptions func(*CompositeReconciler)

// WithName sets the name of the CompositeReconciler.
func WithName(name string) CompositeReconcilerOptions {
	return func(c *CompositeReconciler) {
		c.name = name
	}
}

// WithClient sets the k8s client in the reconciler.
func WithClient(cli client.Client) CompositeReconcilerOptions {
	return func(c *CompositeReconciler) {
		c.client = cli
	}
}

// WithPrototype sets a prototype of the object that's reconciled.
func WithPrototype(obj client.Object) CompositeReconcilerOptions {
	return func(c *CompositeReconciler) {
		c.prototype = obj
	}
}

// WithLogger sets the Logger in a CompositeReconciler.
func WithLogger(log logr.Logger) CompositeReconcilerOptions {
	return func(c *CompositeReconciler) {
		c.log = log
	}
}

// WithController sets the Controller in a CompositeReconciler.
func WithController(ctrlr Controller) CompositeReconcilerOptions {
	return func(c *CompositeReconciler) {
		c.ctrlr = ctrlr
	}
}

// WithInitCondition sets the initial status Condition to be used by the
// CompositeReconciler on a resource object.
func WithInitCondition(cndn conditionsv1.Condition) CompositeReconcilerOptions {
	return func(c *CompositeReconciler) {
		c.initCondition = cndn
	}
}

// WithFinalizer sets the name of the finalizer used by the
// CompositeReconciler.
func WithFinalizer(finalizer string) CompositeReconcilerOptions {
	return func(c *CompositeReconciler) {
		c.finalizerName = finalizer
	}
}

// WithCleanupStrategy sets the CleanupStrategy of the CompositeReconciler.
func WithCleanupStrategy(cleanupStrat CleanupStrategy) CompositeReconcilerOptions {
	return func(c *CompositeReconciler) {
		c.cleanupStrategy = cleanupStrat
	}
}

// WithScheme sets the runtime Scheme of the CompositeReconciler.
func WithScheme(scheme *runtime.Scheme) CompositeReconcilerOptions {
	return func(c *CompositeReconciler) {
		c.scheme = scheme
	}
}

// Init initializes the CompositeReconciler for a given Object with the given
// options.
func (c *CompositeReconciler) Init(mgr ctrl.Manager, prototype client.Object, opts ...CompositeReconcilerOptions) error {
	// Use manager if provided. This is helpful in tests to provide explicit
	// client and scheme without a manager.
	if mgr != nil {
		c.client = mgr.GetClient()
		c.scheme = mgr.GetScheme()
	}

	// Use prototype if provided.
	if prototype != nil {
		c.prototype = prototype
	}

	// Add defaults.
	c.log = ctrl.Log
	c.initCondition = DefaultInitCondition
	c.cleanupStrategy = OwnerReferenceCleanup

	// Run the options to override the defaults.
	for _, opt := range opts {
		opt(c)
	}

	// If a name is set, log it as the reconciler name.
	if c.name != "" {
		c.log = c.log.WithValues("reconciler", c.name)
	}

	// Perform validation.
	if c.ctrlr == nil {
		return fmt.Errorf("must provide a Controller to the CompositeReconciler")
	}

	return nil
}

// DefaultInitCondition is the default init condition used by the composite
// reconciler to add to the status of a new resource.
var DefaultInitCondition conditionsv1.Condition = conditionsv1.Condition{
	Type:    conditionsv1.ConditionProgressing,
	Status:  corev1.ConditionTrue,
	Reason:  "Initializing",
	Message: "Component initializing",
}
