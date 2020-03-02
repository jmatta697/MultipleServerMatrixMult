package main

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	"log"
	"net"
	"net/rpc"
	"os"
	"shared" //Path to the package contains shared struct
	"strconv"
	"strings"
	"sync"
)

type tomlConfig struct {
	Title   string
	Owner   ownerInfo
	Servers map[string]server
}

type server struct {
	IP string
	Port string
}

type ownerInfo struct {
	Name string
	Org  string `toml:"organization"`
	Bio  string
}

// Struct used to abstract rpc client
type MatrixMultRPC struct {
	client *rpc.Client
}

// struct to hold matrix data structure (2d slice)
type Matrix struct {
	matrixArray [][]int
}

type ResultMatrixPriority struct {
	order int
	resultMatrix Matrix
}

// remote client function call
func (t *MatrixMultRPC) MultiplyMatrix(matrix1, matrix2 [][]int) [][]int {
	// use shared struct to pass matrix arguments to server
	args := &shared.MatrixArgs{M1: matrix1, M2: matrix2}
	// declare reply matrix variable
	var reply [][]int
	// client call to Registered MatrixMultiply server using Multiply method
	err := t.client.Call("MatrixMultiply.Multiply", args, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	// reply from server
	return reply
}

// gets matrix size from user - with input error checking
func getMatrixSizeFromUser() (int64, error) {
	reader := bufio.NewReader(os.Stdin)
	var intStr string
	fmt.Println("What are the size of the matrices? \n " +
		"(For example, if two 6 X 6 matrices are desired, enter '6'.)")
	// input error checking loop
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		intStr = strings.Replace(text, "\n", "", -1)

		//check if input is an int
		i, err := strconv.ParseInt(intStr, 10, 32)
		if err != nil {
			fmt.Println("Enter a valid INTEGER!")
		} else {
			//check if number is greater than zero
			if i <= 0 {
				fmt.Println("Enter an integer GREATER THAN ZERO.")
			} else {
				break
			}
		}
	}
	return strconv.ParseInt(intStr, 10, 32)
}

// get stand alone integer from user - with input error checking
func getIntegerFromUser() (int64, error) {
	reader := bufio.NewReader(os.Stdin)
	var userInt string
	// input error checking loop
	for {
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		userInt = strings.Replace(text, "\n", "", -1)
		//check if input is an int
		_, err := strconv.ParseInt(userInt, 10, 32)
		if err != nil {
			fmt.Print("ERROR: Enter a valid INTEGER! > ")
		} else {
			break
		}
	}
	return strconv.ParseInt(userInt, 10, 32)
}

// this function builds the 2d slice (matrix) manually from user input
// user will input each element of the matrix one by one
func buildMatrixFromUserInput(matricesSize int) [][]int {
	var matrix [][]int

	for i := 0; i < matricesSize; i++ {
		fmt.Printf("*** ROW %d ***\n", i+1)
		var tempRow []int
		for j := 0; j < matricesSize; j++ {
			fmt.Print("Enter an integer: ")
			inputInt, _ := getIntegerFromUser()
			tempRow = append(tempRow, int(inputInt))
		}
		matrix = append(matrix, tempRow)
	}
	return matrix
}

// Matrix print scheme from:
// https://rosettacode.org/wiki/Matrix_multiplication#Library_go.matrix
// modified to fit the needs of this program
func (m Matrix) toString() string {
	rows := len(m.matrixArray)
	cols := len(m.matrixArray[0])
	out := ""
	for r := 0; r < rows; r++ {
		if r > 0 {
			out += "\n"
		}
		for c := 0; c < cols; c++ {
			out += fmt.Sprintf("%7d", m.matrixArray[r][c])
		}
	}
	return out
}

func setUpServerConfig(config *tomlConfig) {
	if _, err := toml.DecodeFile("serverConfig.toml", config); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("::Available Servers::")
	for serverName, server := range config.Servers {
		fmt.Printf("Server: %s (%s, %s)\n", serverName, server.IP, server.Port)
	}
}

func (c tomlConfig) establishServerConnections(num int) []net.Conn {
	var connArray []net.Conn
	// set up connection array based on how many connections are needed
	var count = 0	//this will count up to the number of connection that are needed
	for i := range c.Servers {
		// establish connection with server
		// FIRST get server address string
		serverAddress := c.Servers[i].IP + ":" + c.Servers[i].Port
		conn, err := net.Dial("tcp", serverAddress)
		if err != nil {
			log.Fatal("Connecting:", err)
		}
		// add connection to conn array
		connArray = append(connArray, conn)
		// increment count
		count++
		// check if count matches number of needed connections
		if count >= num || count >= len(c.Servers) {
			//if so stop making connections
			break
		}
	}
	return connArray
}

func main() {
	var wg sync.WaitGroup
	// process config toml file
	var config tomlConfig
	// use serverConfig.toml to set up tomlConfig server configurations
	setUpServerConfig(&config)
	// get size of matrices from user
	matricesSize, _ := getMatrixSizeFromUser()
	fmt.Println(matricesSize)
	fmt.Println(fmt.Sprintf("Matrices Size: %d", matricesSize))

	// user manually builds matrices
	fmt.Println("\n----- MATRIX 1 -----")
	firstMatrix := buildMatrixFromUserInput(int(matricesSize))
	fmt.Println("\n----- MATRIX 2 -----")
	secondMatrix := buildMatrixFromUserInput(int(matricesSize))

	// wrap each matrix in Matrix struct
	matrix1 := Matrix{matrixArray: firstMatrix}
	matrix2 := Matrix{matrixArray: secondMatrix}

	// print matrices
	fmt.Println("\n----- MATRIX 1 -----")
	fmt.Println(matrix1.toString())
	fmt.Println("\n----- MATRIX 2 -----")
	fmt.Println(matrix2.toString())

	fmt.Println(matricesSize - 1)
	// set up connection array based on how many connections are needed
	var connSliceArray = config.establishServerConnections(int(matricesSize) - 1)
	//fmt.Println(connSliceArray)

	// this will hold all the results coming back form the server
	//resultMatrices := make([]ResultMatrixPriority, matricesSize)
	// channel that will receive the results
	resultsCh := make(chan ResultMatrixPriority, matricesSize)

	// connect to a server port by default to connect to the main server
	wg.Add(1)
	go func(m [][]int, m1 [][]int, wg *sync.WaitGroup, inChan chan<- ResultMatrixPriority) {
		fmt.Println("Calling remote server to multiply")
		defer wg.Done()
		conn, err := net.Dial("tcp", "localhost:1242")
		if err != nil {
			log.Fatal("Connecting:", err)
		}
		matrixMultiply := &MatrixMultRPC{client: rpc.NewClient(conn)}
		partialM1 := make([][]int, 1)
		partialM1[0] = m[0]
		var multiplicationResult = matrixMultiply.MultiplyMatrix(partialM1, m1)
		resultMatrix := Matrix{matrixArray: multiplicationResult}
		inChan <- ResultMatrixPriority{0, resultMatrix}
	}(firstMatrix, secondMatrix, &wg, resultsCh)


	fmt.Print("Length of connArray: ")
	fmt.Println(len(connSliceArray))
	// now send stuff to the servers to be calculated
	for j := 0; j < len(connSliceArray); j++ {
		fmt.Println("entered the j loop")
		// Create a struct, that mimics all methods provided by interface.
		// It is not compulsory, we are doing it here, just to simulate a traditional method call.
		wg.Add(1)
		go func(index int, mx [][]int, m2 [][]int, wg *sync.WaitGroup, inChan chan<- ResultMatrixPriority) {
			fmt.Println("Calling remote server to multiply - J LOOP")
			defer wg.Done()
			matrixMultiply := &MatrixMultRPC{client: rpc.NewClient(connSliceArray[j-1])}
			partialM1 := make([][]int, 1)
			partialM1[0] = mx[j]
			//fmt.Println("Calling remote server to multiply " + string(j))
			var multiplicationResult = matrixMultiply.MultiplyMatrix(partialM1, m2)
			resultMatrix := Matrix{matrixArray: multiplicationResult}
			inChan <- ResultMatrixPriority{j, resultMatrix}
		}(j, firstMatrix, secondMatrix, &wg, resultsCh)
	}

	wg.Wait()

	close(resultsCh)

	for {
		v := <- resultsCh
		fmt.Println(v)
	}

	//for rr := range resultMatrices {
	//	fmt.Println(resultMatrices[rr])
	//}

	//
	//matrixMultiply1 := &MatrixMultRPC{client: rpc.NewClient(conn1)}
	//

	//
	//multiplicationResult1 := matrixMultiply1.MultiplyMatrix(firstMatrix, secondMatrix)
	//

	//resultMatrix1 := Matrix{matrixArray: multiplicationResult1}
	//fmt.Println("\n----- RESULT MATRIX -----")
	//fmt.Println(resultMatrix.toString())
	//
	//fmt.Println("\n----- RESULT MATRIX 2 -----")
	//fmt.Println(resultMatrix1.toString())
}
