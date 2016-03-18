package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Task struct {
	extWriter io.Writer
	newline bool
}

type TaskFailed struct {}

func (task *Task) Write(p []byte) (n int, err error) {
	// TODO handle the case when p[len(np)-1] != "\n"

	np := ""
	if task.newline {
		np = "\t"
		task.newline = false
	}
	np += strings.Replace(string(p), "\n", "\n\t", -1)
	if strings.HasSuffix(np, "\n\t") {
		np = np[:len(np)-1]
		task.newline = true
	}

	n, err = task.extWriter.Write([]byte(np))
	return n-len(np)+len(p), err
}

func (task *Task) Finish() {
	//task.extWriter.Write([]byte("finished\n"))
}

func (task *Task) Assert(test bool, err error) {
	if !test {
		task.Require(err)
	}
}

func (task *Task) Require(err error) {
	if err != nil {
		fmt.Println(task, err)
		panic(TaskFailed{})
	}
}

func (task *Task) RequireAndFinish(err error) {
	defer task.Finish()
	task.Require(err)
}

func (task *Task) RunCmd(cmd *exec.Cmd) error {
	subtask := NewTask(task, cmd.Args[0] + " (" + strings.Join(cmd.Args[1:], ") (") + ")"); defer subtask.Finish()

	cmd.Stdout = subtask
	cmd.Stderr = subtask

	return cmd.Run()
}

func (task *Task) RunCommand(name string, arg ...string) error {
	return task.RunCmd(exec.Command(name, arg...))
}

func NewTask(outer io.Writer, desc string) *Task {
	outer.Write([]byte(desc + "...\n"))
	task := &Task{outer, true}
	return task
}
