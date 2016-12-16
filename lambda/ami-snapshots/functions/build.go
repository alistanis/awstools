package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	files, err := ioutil.ReadDir(pwd)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	gopath := os.Getenv("GOPATH")
	wg := sync.WaitGroup{}
	for _, fi := range files {
		if fi.IsDir() {
			wg.Add(1)
			path := pwd + string(os.PathSeparator) + fi.Name()
			go func(path string) {
				err := os.Chdir(path)
				if err != nil {
					fmt.Println(err)
				}

				// docker run --rm -v $GOPATH:/go -v $PWD:/tmp eawsy/aws-lambda-go
				cmd := exec.Command("docker", strings.Split(fmt.Sprintf("run --rm -v %s:/go -v %s:/tmp eawsy/aws-lambda-go", gopath, path), " ")...)
				cmd.Env = os.Environ()
				_, err = cmd.Output()
				if err != nil {
					fmt.Println(err)
					e := err.(*exec.ExitError)
					fmt.Println(string(e.Stderr))
				}
				wg.Done()
			}(path)
		}
	}
	wg.Wait()
}
