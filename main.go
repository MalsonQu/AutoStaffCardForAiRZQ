package main

import (
	_ "github.com/MalsonQu/AutoStaffCardForAiRZQ/Email"
	"github.com/MalsonQu/AutoStaffCardForAiRZQ/Engine"
	_ "github.com/MalsonQu/AutoStaffCardForAiRZQ/Model"
)

func main() {
	Engine.Run()
}
