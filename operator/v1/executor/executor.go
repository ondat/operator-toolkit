package executor

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ondat/operator-toolkit/constant"
	eventv1 "github.com/ondat/operator-toolkit/event/v1"
	"github.com/ondat/operator-toolkit/operator/v1/operand"
	"github.com/ondat/operator-toolkit/operator/v1/playbook/order"
	"github.com/ondat/operator-toolkit/telemetry"
)

// Name of the instrumentation.
const instrumentationName = constant.LibraryName + "/operator/executor"

// ExecutionStrategy is the operands execution strategy of an operator.
type ExecutionStrategy int

const (
	Parallel ExecutionStrategy = iota
	Serial
)

// Executor is an operand executor. It is used to configure how the operands
// are executed. The event recorder is used to broadcast an event right after
// executing an operand.
type Executor struct {
	execStrategy ExecutionStrategy
	recorder     record.EventRecorder

	inst *telemetry.Instrumentation
}

// NewExecutor initializes and returns an Executor.
func NewExecutor(e ExecutionStrategy, r record.EventRecorder) *Executor {
	return &Executor{
		execStrategy: e,
		recorder:     r,
		inst:         telemetry.NewInstrumentation(instrumentationName),
	}
}

// ExecuteOperands executes operands in a given OperandOrder by calling a given
// OperandRunCall function on each of the operands. The OperandRunCall can be a
// call to Ensure or Delete.
func (exe *Executor) ExecuteOperands(
	operandOrder order.OperandOrder,
	blockers order.BlockingOperands,
	call operand.OperandRunCall,
	ctx context.Context,
	obj client.Object,
	ownerRef metav1.OwnerReference,
) (result ctrl.Result, rerr error) {
	ctx, span, _ := exe.inst.Start(ctx, "execute")
	defer span.End()

	span.SetAttributes(attribute.Int("order-length", len(operandOrder)))
	span.AddEvent("Start operand execution")
	// Iterate through the order steps and run the operands in the steps as per
	// the execution strategy.
	for _, ops := range operandOrder {
		// Error in the current execution step.
		var execErr error

		// res is the Result of the step.
		// TODO: Change the type of res to something that reflects that a
		// change took place. The value of Result is not propagated to the
		// caller.
		var res *ctrl.Result

		// failedOperands are the operands that have failed during this step.
		var failedOperands map[string]bool

		requeueStrategy := order.StepRequeueStrategy(ops)

		span.AddEvent(
			"Execute operands",
			trace.WithAttributes(
				attribute.Int("requeue-strategy", int(requeueStrategy)),
			),
		)

		switch exe.execStrategy {
		case Serial:
			// Run the operands serially.
			res, failedOperands, execErr = exe.serialExec(ops, call, ctx, obj, ownerRef)
		case Parallel:
			// Run the operands concurrently.
			res, failedOperands, execErr = exe.concurrentExec(ops, call, ctx, obj, ownerRef)
		default:
			rerr = fmt.Errorf("unknown operands execution strategy: %v", exe.execStrategy)
			return
		}
		if execErr != nil {
			rerr = kerrors.NewAggregate([]error{rerr, execErr})
			// Check if any failed operands are also blocking operands.
			// If so, we break (ie block the step) instead of proceeding to the next step.
			// Otherwise, continue to the next step in the order.
			// In either case, the result is to requeue.
			result = ctrl.Result{Requeue: true}
			if failuresContainBlockingOperand(failedOperands, blockers) {
				break
			}
			continue
		}

		// If a change was made with a Result received after the execution and
		// the RequeueStrategy is RequeueAlways, set a requeued result.
		if res != nil && requeueStrategy == operand.RequeueAlways {
			result = ctrl.Result{Requeue: true}
			break
		}
	}

	span.AddEvent("Finish operand execution")

	return
}

// failuresContainBlockingOperand returns true if a blocking operand is also a failed operand.
func failuresContainBlockingOperand(failedOperands map[string]bool, blockers order.BlockingOperands) bool {
	for operandName, isBlocker := range blockers {
		if failed, ok := failedOperands[operandName]; ok && failed && isBlocker {
			return true
		}
	}
	return false
}

// serialExec runs the given set of operands serially with the given call
// function. An event is used to know if a change was applied. When an event is
// found, a result object is returned, else nil.
func (exe *Executor) serialExec(
	ops []operand.Operand,
	call operand.OperandRunCall,
	ctx context.Context,
	obj client.Object,
	ownerRef metav1.OwnerReference,
) (result *ctrl.Result, failedOperands map[string]bool, rerr error) {
	ctx, span, _ := exe.inst.Start(ctx, "serial-exec")
	defer span.End()

	result = nil

	failedOperands = make(map[string]bool)

	span.AddEvent(
		"Execute serially",
		trace.WithAttributes(attribute.Int("operand-count", len(ops))),
	)

	for _, op := range ops {
		span.AddEvent(
			"Executing operand",
			trace.WithAttributes(attribute.String("operand-name", op.Name())),
		)
		// Initially operand as not failed.
		failedOperands[op.Name()] = false
		// Call the run call function. Since this is serial execution, return
		// if an error occurs.
		event, err := call(op)(ctx, obj, ownerRef)
		if err != nil {
			rerr = kerrors.NewAggregate([]error{rerr, err})
			failedOperands[op.Name()] = true
			return
		}
		if event != nil {
			event.Record(exe.recorder)
			result = &ctrl.Result{}
		}
	}

	span.AddEvent("Finish serial execution")

	return
}

// concurrentExec runs the operands concurrently, collecting the errors from
// the operand executions and returns them.
func (exe *Executor) concurrentExec(
	ops []operand.Operand,
	call operand.OperandRunCall,
	ctx context.Context,
	obj client.Object,
	ownerRef metav1.OwnerReference,
) (result *ctrl.Result, failedOperands map[string]bool, rerr error) {
	ctx, span, _ := exe.inst.Start(ctx, "concurrent-exec")
	defer span.End()

	result = nil

	failedOperands = make(map[string]bool)

	// Wait group to synchronize the go routines.
	var wg sync.WaitGroup

	totalOperands := len(ops)

	// resultChan is used to collect the result returned from the concurrent
	// execution of the operands.
	var resultChan chan ctrl.Result = make(chan ctrl.Result, totalOperands)

	// Error buffered channel to collect all the errors from the go routines.
	var errChan chan error = make(chan error, totalOperands)

	// String buffered channel to collect the names of operands that have failed
	// to execute.
	var failedOperandChan chan string = make(chan string, totalOperands)

	span.AddEvent(
		"Execute concurrently",
		trace.WithAttributes(attribute.Int("operand-count", len(ops))),
	)

	wg.Add(totalOperands)
	for _, op := range ops {
		span.AddEvent(
			"Executing operand",
			trace.WithAttributes(attribute.String("operand-name", op.Name())),
		)
		// Initially set operand as not failed.
		failedOperands[op.Name()] = false
		go exe.operateWithWaitGroup(&wg, resultChan, failedOperandChan, errChan, call(op), ctx, obj, ownerRef, op.Name())
	}
	wg.Wait()
	close(errChan)
	close(failedOperandChan)

	// Check if any errors were encountered.
	for err := range errChan {
		rerr = kerrors.NewAggregate([]error{rerr, err})
	}

	// Set any failed operands to true.
	for failedOperand := range failedOperandChan {
		failedOperands[failedOperand] = true
	}

	// Check the result channel, if it contains any result, return a result
	// object.
	foundResult := false
	if len(resultChan) > 0 {
		foundResult = true
	}
	if foundResult {
		result = &ctrl.Result{}
	}

	span.AddEvent("Finish concurrent execution")

	return
}

// operateWithWaitGroup runs the given function f and calls done on the wait
// group at the end. This is a goroutine function used for running the operands
// concurrently. The result from events and errors from the execution are
// communicated via the respective channels.
func (exe *Executor) operateWithWaitGroup(
	wg *sync.WaitGroup,
	resultChan chan ctrl.Result,
	failedOperandChan chan string,
	errChan chan error,
	f func(context.Context, client.Object, metav1.OwnerReference) (eventv1.ReconcilerEvent, error),
	ctx context.Context,
	obj client.Object,
	ownerRef metav1.OwnerReference,
	operandName string,
) {
	defer wg.Done()

	event, err := f(ctx, obj, ownerRef)
	if err != nil {
		errChan <- err
		failedOperandChan <- operandName
	}

	// Event is used to determine if a change took place. Send a result to the
	// result channel when an event is received.
	if event != nil {
		event.Record(exe.recorder)
		resultChan <- ctrl.Result{}
	}
}
