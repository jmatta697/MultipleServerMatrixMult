package shared

type MatrixMult interface {
	Multiply(args *MatrixArgs, reply *[][]int) error
}
