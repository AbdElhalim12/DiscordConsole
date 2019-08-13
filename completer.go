/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2019 Mnpn

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"github.com/chzyer/readline"
)

func setCompleter(rl *readline.Instance) {
	rl.Config.AutoComplete = readline.NewPrefixCompleter(
		readline.PcItem("guild", readline.PcItemDynamic(func(name string) []string {
			names := make([]string, len(cacheGuilds))
			for i, g := range cacheGuilds {
				names[i] = g.Name
			}
			return names
		})),
		readline.PcItem("channel", readline.PcItemDynamic(func(name string) []string {
			names := make([]string, len(cacheChannels))
			for i, c := range cacheChannels {
				names[i] = c.Name
			}
			return names
		})),

		readline.PcItem("edit", readline.PcItemDynamic(singleValue(&lastUsedMsg))),
		readline.PcItem("del", readline.PcItemDynamic(singleValue(&lastUsedMsg))),
		readline.PcItem("quote", readline.PcItemDynamic(singleValue(&lastUsedMsg))),
		readline.PcItem("react",
			readline.PcItem("add", readline.PcItemDynamic(singleValue(&lastUsedMsg))),
			readline.PcItem("del", readline.PcItemDynamic(singleValue(&lastUsedMsg))),
		),

		readline.PcItem("role",
			readline.PcItem("add", readline.PcItem(userID, readline.PcItemDynamic(singleValue(&lastUsedRole)))),
			readline.PcItem("rem", readline.PcItem(userID, readline.PcItemDynamic(singleValue(&lastUsedRole)))),
			readline.PcItem("edit", readline.PcItemDynamic(singleValue(&lastUsedRole))),
			readline.PcItem("delete", readline.PcItemDynamic(singleValue(&lastUsedRole))),
		),

		readline.PcItem("bookmark", readline.PcItemDynamic(bookmarkTab)),
		readline.PcItem("go", readline.PcItemDynamic(bookmarkTab)),

		readline.PcItem("read", readline.PcItemDynamic(singleValue(&lastUsedMsg))),
		readline.PcItem("cinfo",
			readline.PcItem("guild"),
			readline.PcItem("name"),
			readline.PcItem("topic"),
			readline.PcItem("type"),
		),
		readline.PcItem("ginfo",
			readline.PcItem("name"),
			readline.PcItem("icon"),
			readline.PcItem("region"),
			readline.PcItem("owner"),
			readline.PcItem("splash"),
			readline.PcItem("members"),
			readline.PcItem("level"),
		),
		readline.PcItem("messages",
			readline.PcItem("all"),
			readline.PcItem("mentions"),
			readline.PcItem("private"),
			readline.PcItem("current"),
			readline.PcItem("none"),
		),
		readline.PcItem("intercept",
			readline.PcItem("true"),
			readline.PcItem("false"),
		),
		readline.PcItem("output",
			readline.PcItem("true"),
			readline.PcItem("false"),
		),
		readline.PcItem("avatar",
			readline.PcItem("link"),
			readline.PcItem("file"),
		),
		readline.PcItem("log",
			readline.PcItem("file"),
			readline.PcItem("directly"),
		),
		readline.PcItem("game",
			readline.PcItem("streaming"),
			readline.PcItem("watching"),
			readline.PcItem("listening"),
		),
		readline.PcItem("friend",
			readline.PcItem("add"),
			readline.PcItem("accept"),
			readline.PcItem("remove"),
			readline.PcItem("list"),
		),
		readline.PcItem("new",
			readline.PcItem("channel"),
			readline.PcItem("vchannel"),
			readline.PcItem("guild"),
		),
		readline.PcItem("delete",
			readline.PcItem("guild"),
			readline.PcItem("channel"),
			readline.PcItem("category"),
		),
	)
}
func bookmarkTab(line string) []string {
	items := make([]string, len(bookmarks))

	i := 0
	for key := range bookmarks {
		items[i] = key
		i++
	}
	return items
}
func singleValue(val *string) func(string) []string {
	return func(line string) []string {
		return []string{*val}
	}
}
