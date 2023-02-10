package playbook

import (
	"github.com/ondat/operator-toolkit/operator/v1/operand"
	"github.com/ondat/operator-toolkit/operator/v1/playbook/dag"
	"github.com/ondat/operator-toolkit/operator/v1/playbook/order"
)

// Playbook holds data relative to how operands are executed
// relative to each other.
type Playbook struct {
	dag      *dag.OperandDAG
	order    *order.OperandOrder
	blockers *order.BlockingOperands
}

func NewPlaybook(operands []operand.Operand, operandRunCallName operand.OperandRunCallName) (*Playbook, error) {
	requiredOperands, err := order.Required(operands, operandRunCallName)
	if err != nil {
		return nil, err
	}

	operandDAG, err := dag.NewOperandDAG(operands, requiredOperands)
	if err != nil {
		return nil, err
	}

	opOrder, err := operandDAG.Order()
	if err != nil {
		return nil, err
	}

	blockers := order.Blockers(opOrder, requiredOperands)

	return &Playbook{
		dag:      operandDAG,
		order:    &opOrder,
		blockers: &blockers,
	}, nil
}

func (p *Playbook) DAG() dag.OperandDAG {
	return *p.dag
}

func (p *Playbook) Order() order.OperandOrder {
	return *p.order
}

func (p *Playbook) Blockers() order.BlockingOperands {
	return *p.blockers
}
