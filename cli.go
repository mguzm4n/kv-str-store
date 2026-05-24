package main

import (
	"github.com/abiosoft/ishell/v2"
)

func cli() {
	store, _ := NewStore()

	shell := ishell.New()
	shell.Println("Interactive Shell (type 'help' for commands)")

	shell.AddCmd(&ishell.Cmd{
		Name: "set",
		Help: "Set the application state (e.g., set user1 user1name)",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 2 {
				c.Err(nil)
				c.Println("Error: expects two args")
				return
			}

			store.PutKey(c.Args[0], c.Args[1])
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "Get the application state (e.g., get user1 -> user1name)",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 1 {
				c.Err(nil)
				c.Println("Error: expects one arg")
				return
			}

			val, _ := store.GetKey(c.Args[0])
			c.Println(val)
		},
	})
	shell.Run()
}
