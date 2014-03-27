package ecmascript

import (
	"fmt"
	"io/ioutil"
	"os"
)

func ScriptTest() {
	buffer, err := ioutil.ReadFile("src/javascript/test.js")
	if err != nil {
		panic(err)
	}

	var p parser
	s, err := p.parseProgram([]byte(buffer))
	if err != nil {
		panic(err)
	}
	fmt.Println(s)

	fmt.Println("Begin")
	s.generate(os.Stdout)
}
