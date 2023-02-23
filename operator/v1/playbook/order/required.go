package order

import (
	"fmt"

	"github.com/ondat/operator-toolkit/operator/v1/operand"
)

// RequiredOperands stores operand name to operands required. The required
// operand values can be those of Requires() or CleanupRequires().
type RequiredOperands map[string][]string

// RequiredOperands creates a map of operand name to required operands.
// This function is used to store this map locally and limit the number of Requires()
// and CleanupRequires() calls to the individual operands. It can be used for either
// cleanup or ensure during playbook creation.
func Required(operands []operand.Operand, operandRunCallName operand.OperandRunCallName) (RequiredOperands, error) {
	requiredOperands := make(RequiredOperands)
	for _, op := range operands {
		switch operandRunCallName {
		case operand.Cleanup:
			requiredOperands[op.Name()] = op.CleanupRequires()
		case operand.Ensure:
			requiredOperands[op.Name()] = op.Requires()
		default:
			return nil, fmt.Errorf("unknown operandRunCallName - must be Ensure or Cleanup")
		}
	}

	// assert that each operand has an entry in the map
	for _, op := range operands {
		if _, ok := requiredOperands[op.Name()]; !ok {
			return nil, fmt.Errorf("error creating map of required operands - no entry for operand %s", op.Name())
		}
	}
	return requiredOperands, nil
}
