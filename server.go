package main

import (
	"log"
	"net"
	"net/rpc"
	"shared" //Path to the package contains shared struct
)

type MatrixMult int

func (t *MatrixMult) Multiply(args *shared.MatrixArgs, reply *[][]int) error {
	*reply, _ = MultiplyMatrices(args.M1, args.M2)
	return nil
}

// Matrix multiplication algorithm from:
// https://rosettacode.org/wiki/Matrix_multiplication#Library_go.matrix
func MultiplyMatrices(m1, m2 [][]int) (m3 [][]int, ok bool) {
	rows, cols, extra := len(m1), len(m2[0]), len(m2)
	if len(m1[0]) != extra {
		return nil, false
	}
	m3 = make([][]int, rows)
	for i := 0; i < rows; i++ {
		m3[i] = make([]int, cols)
		for j := 0; j < cols; j++ {
			for k := 0; k < extra; k++ {
				m3[i][j] += m1[i][k] * m2[k][j]
			}
		}
	}
	return m3, true
}

func registerArith(server *rpc.Server, arith shared.MatrixMult) {
	// registers Arith interface by name of `Arithmetic`.
	// If you want this name to be same as the type name, you
	// can use server.Register instead.
	server.RegisterName("MatrixMultiply", arith)
}

func main() {
	//Creating an instance of struct which implement MatrixMult interface
	arith := new(MatrixMult)

	var portNums = []string{":1234", ":1235", ":1236", ":1237", ":1238", ":1239", ":1240"}
	serverListener := make(map[*rpc.Server]net.Listener)
	for i := 0; i < len(portNums); i++ {
		server := rpc.NewServer()
		// Register a new rpc server (In most cases, you will use default server only)
		// And register struct we created above by name "Arith"
		// The wrapper method here ensures that only structs which implement Arith interface
		// are allowed to register themselves.
		registerArith(server, arith)
		// Listen for incoming tcp packets on specified port.
		l, e := net.Listen("tcp", portNums[i])
		if e != nil {
			log.Fatal("listen error:", e)
		}
		// append to server/listener map
		serverListener[server] = l
	}
	// start all servers listening, except last
	for server := range serverListener{
		// This statement links rpc server to the socket, and allows rpc server to accept
		// rpc request coming from that socket.
		go func() {
			server.Accept(serverListener[server])
		}()
	}

	//start main server
	server := rpc.NewServer()
	registerArith(server, arith)
	// Listen for incoming tcp packets on specified port.
	l1, e := net.Listen("tcp", ":1242")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	// This statement links rpc server to the socket, and allows rpc server to accept
	// rpc request coming from that socket.
	server.Accept(l1)
}
