Joe Matta
Distributed Systems
Task 3 [Bonus]
2/29/2020

This program multiplies two square matrices that have the same dimensions.
The first matrix is divided along its rows.
Each row from the first matrix is sent to a separate server along with the entire second matrix.
Each server concurrently multiplies the single row from the first matrix with the entire second matrix to produce
one row of the product matrix.
Each server sends their product row back to the client and all product rows are assembled in order to make the entire
resulting product matrix.

Instructions:

- The shared folder must be placed in the following directory on the server and client machines:

    C:\Go\src

    This folder contains:

    interface.go
    shared_structs.go

1) run server.go:
    $ go run server.go

    Eight concurrent server threads will start. Information about each server thread will be displayed.

2) run client.go:
    $ go run client.go
3) Enter the size of the matrices you would like to multiply.
    *** Note: Matrix size is limited to 1x1 up to and including 8x8. ***
    Both matrices will have the same dimensions.
4) Manually enter integer values row by row.
    Both matrices will be displayed.
    The size of the matrices will determine the number of server connections made. (One connection per 'first matrix' row.)
    Connection confirmations will be displayed.
    Product rows will be displayed in the order that they are returned.
5) The product matrix is displayed

----- Notes -----
Developed in GoLand 2019.3.2
             Build #GO-193.6015.58, built on February 3, 2020
             Runtime version: 11.0.4+10-b520.11 amd64
             VM: OpenJDK 64-Bit Server VM by JetBrains s.r.o
             Windows 10 10.0

             using:
             go version go1.13.7 windows/amd64

The following external packages must be installed:

    github.com/BurntSushi/toml
    $ go get github.com/BurntSushi/toml

    github.com/bradfitz/slice
    $ go get github.com/bradfitz/slice