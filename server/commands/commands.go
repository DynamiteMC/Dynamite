package commands

import (
	"slices"

	"github.com/aimjel/minecraft/protocol/types"

	pk "github.com/aimjel/minecraft/packet"
)

func cond[T any](c bool, t T, f T) T {
	if c {
		return t
	} else {
		return f
	}
}

const (
	ChatModeEnabled = iota
	ChatModeCommandsOnly
	ChatModeHidden
)

type Command struct {
	Name                string
	Arguments           []Argument
	Aliases             []string
	Execute             func(ctx CommandContext)
	RequiredPermissions []string
}

type Parser struct {
	ID         int32
	Properties types.CommandProperties
}

type Argument struct {
	Name        string
	Suggest     func(ctx SuggestionsContext)
	Parser      Parser
	Alternative *Argument
}

type Graph struct {
	Commands []*Command
}

func (graph *Graph) AddCommands(commands ...*Command) *Graph {
	graph.Commands = append(graph.Commands, commands...)
	return graph
}

func (command *Command) AddArguments(arguments ...Argument) *Command {
	command.Arguments = append(command.Arguments, arguments...)
	return command
}

func (graph Graph) Data() *pk.DeclareCommands {
	packet := &pk.DeclareCommands{}
	packet.Nodes = append(packet.Nodes, types.CommandNode{
		Flags: 0,
	})
	commands := graph.Commands
	rootChildren := []int32{}
	for _, command := range commands {
		if command == nil {
			continue
		}
		for _, alias := range command.Aliases {
			commands = append(commands, &Command{
				Name:      alias,
				Arguments: command.Arguments,
			})
		}
	}
	for _, command := range commands {
		if command == nil {
			continue
		}
		rootChildren = append(rootChildren, int32(len(packet.Nodes)))
		packet.Nodes = append(packet.Nodes, types.CommandNode{
			Name:  command.Name,
			Flags: 1 | 0x04,
		})
		for _, argument := range command.Arguments {
			parent := len(packet.Nodes) - 1
			packet.Nodes[parent].Children = append(packet.Nodes[parent].Children, int32(len(packet.Nodes)))
			node := types.CommandNode{Flags: 2, Name: argument.Name, Properties: argument.Parser.Properties, ParserID: argument.Parser.ID}
			if argument.Suggest != nil {
				node.Flags |= 0x10
				node.SuggestionsType = "minecraft:ask_server"
			}
			packet.Nodes = append(packet.Nodes, node)
			if argument.Alternative != nil {
				argument = *argument.Alternative
				packet.Nodes[parent].Children = append(packet.Nodes[parent].Children, int32(len(packet.Nodes)))
				node := types.CommandNode{Flags: 2, Name: argument.Name, Properties: argument.Parser.Properties, ParserID: argument.Parser.ID}
				if argument.Suggest != nil {
					node.Flags |= 0x10
					node.SuggestionsType = "minecraft:ask_server"
				}
				packet.Nodes = append(packet.Nodes, node)
			}
		}
	}
	packet.Nodes[0].Children = rootChildren
	return packet
}

func RegisterCommands(commands ...*Command) *pk.DeclareCommands {
	return Graph{Commands: commands}.Data()
}

func (graph *Graph) FindCommand(name string) (cmd *Command) {
	for _, c := range graph.Commands {
		if c == nil {
			continue
		}
		if c.Name == name {
			cmd = c
			return
		}

		for _, a := range c.Aliases {
			if a == name {
				cmd = c
				return
			}
		}
	}
	return
}

func (graph *Graph) DeleteCommand(name string) (found bool) {
	for i, c := range graph.Commands {
		if c == nil {
			continue
		}
		if c.Name == name {
			graph.Commands = slices.Delete(graph.Commands, i, i+1)
			return true
		}

		for _, a := range c.Aliases {
			if a == name {
				graph.Commands = slices.Delete(graph.Commands, i, i+1)
				return true
			}
		}
	}
	return false
}
