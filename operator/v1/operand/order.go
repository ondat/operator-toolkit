package operand

import (
	"fmt"
	"sort"
	"strings"
)

// OperandOrder stores the operands in order of their execution. The first
// dimension of the slice depicts the execution step and the second dimention
// contains the operands that can be run in parallel.
type OperandOrder [][]Operand

// BlockingOperands stores the names of operands that are considered as blocking
// operands. A blocking operand is an operand that must be executed successully
// to proceed to the next step in the OperandOrder. If every operand in step n+1
// 'requires' an operand in step n, then that operand is deemed a blocking operand
// to the completion of step n.
type BlockingOperands map[string]bool

// String implements the Stringer interface for OperandOrder.
// Example string result:
// [
//  0: [ A B ]
//  1: [ C ]
//  2: [ D F ]
//  3: [ E ]
// ]
func (o OperandOrder) String() string {
	var result strings.Builder
	result.WriteString("[\n")

	for i, s := range o {
		// Sort the items for deterministic results.
		items := []string{}
		for _, op := range s {
			items = append(items, op.Name())
		}
		sort.Strings(items)
		itemsStr := strings.Join(items, " ")
		line := fmt.Sprintf("  %d: [ %s ]\n", i, itemsStr)
		result.WriteString(line)
	}
	result.WriteString("]")
	return result.String()
}

// Reverse returns the OperandOrder in reverse order.
func (o OperandOrder) Reverse() OperandOrder {
	// Refer: https://github.com/golang/go/wiki/SliceTricks#reversing
	for left, right := 0, len(o)-1; left < right; left, right = left+1, right-1 {
		o[left], o[right] = o[right], o[left]
	}
	return o
}

// Blockers returns the BlockingOperands of the order.
func (o OperandOrder) Blockers() BlockingOperands {
	blockers := make(map[string]bool)
	if len(o) == 1 {
		// If there's only a single step, then there can be no blockers.
		return blockers
	}

	for _, step := range o {
		for _, blocker := range blockersInStep(step) {
			blockers[blocker] = true
		}
	}
	return blockers
}

// blockersInStep returns a slice of operand names from a single step
// that are considered blockers.
func blockersInStep(operands []Operand) []string {
	if len(operands) == 0 {
		return nil
	} else if len(operands) == 1 {
		// If there is only a single operand in a step,
		// then return that step's required operands.
		return operands[0].Requires()
	}

	blockers := operands[0].Requires()
	for _, operand := range operands[1:] {
		blockers = commonStringsInSlices(blockers, operand.Requires())
	}
	return blockers
}

func commonStringsInSlices(sliceA, sliceB []string) []string {
	commonStrings := make([]string, 0)
	for _, sliceAString := range sliceA {
		for _, sliceBString := range sliceB {
			if sliceAString == sliceBString {
				commonStrings = append(commonStrings, sliceAString)
			}
		}
	}
	return commonStrings
}

// StepRequeueStrategy returns the requeue strategy of a step. By default, the
// operands are requeued on error. Since the operands in a step run
// concurrently, if an operand has RequeueAlways strategy, the whole step gets
// RequeueAlways strategy.
func StepRequeueStrategy(step []Operand) RequeueStrategy {
	strategy := RequeueOnError
	for _, o := range step {
		if o.RequeueStrategy() == RequeueAlways {
			strategy = RequeueAlways
			break
		}
	}
	return strategy
}
