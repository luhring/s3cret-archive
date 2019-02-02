package main

import "github.com/luhring/s3cret/cmd"

const keyHex = "6368616e676520746869732070617373776f726420746f206120736563726574"
const plaintextFilePath = "/Users/Dan/Desktop/test.txt"
const encryptedFilePath = "/Users/Dan/Desktop/test.txt.encrypted"

func main() {
	cmd.Execute()
}
