package order

import (
	"github.com/ondat/operator-toolkit/operator/v1/operand"
)

// BlockingOperands stores the names of operands that are considered as blocking
// operands. A blocking operand is an operand that must be executed successully
// to proceed to the next step in the OperandOrder. If every operand in step n+1
// 'requires' an operand in step n, then that operand is deemed a blocking operand
// to the completion of step n.
type BlockingOperands map[string]bool

// Blockers returns the BlockingOperands of the order.
func Blockers(order OperandOrder, requiredOperands RequiredOperands) BlockingOperands {
	blockers := make(map[string]bool)
	if len(order) == 1 {
		// If there's only a single step, then there can be no blockers.
		return blockers
	}

	for _, step := range order {
		for _, blocker := range blockersInStep(step, requiredOperands) {
			blockers[blocker] = true
		}
	}
	return blockers
}

// blockersInStep returns a slice of operand names from a single step
// that are considered blockers.
func blockersInStep(operands []operand.Operand, requiredOperands RequiredOperands) []string {
	if len(operands) == 0 {
		return nil
	} else if len(operands) == 1 {
		// If there is only a single operand in a step,
		// then return that step's required operands.
		// This and subsequent map assignments are safe
		// as we have already asserted the map entries
		// during playbook creation.
		return requiredOperands[operands[0].Name()]
	}

	blockers := requiredOperands[operands[0].Name()]
	for _, operand := range operands[1:] {
		blockers = commonStringsInSlices(blockers, requiredOperands[operand.Name()])
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
