package main

import (
	"bytes"
	"os"
	"os/exec"
)

func main() {
	const configName = "config.yml"
	file, err := os.ReadFile(configName)
	if err != nil {
		panic(err)
	}
	qq := os.Getenv("QQ")
	password := os.Getenv("PASSWORD")
	file = bytes.ReplaceAll(file, []byte("!!!QQ!!!"), []byte(qq))
	file = bytes.ReplaceAll(file, []byte("!!!PASSWORD!!!"), []byte(password))
	err = os.WriteFile(configName, file, 0666)
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("/home/go-cqhttp")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
