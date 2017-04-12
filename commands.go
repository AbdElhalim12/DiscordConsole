package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/gtable"
	"github.com/legolord208/stdutil"
)

var TypeRelationships = map[int]string{
	1: "Friend",
	2: "Blocked",
	3: "Incoming request",
	4: "Sent request",
}
var TypeVerifications = map[discordgo.VerificationLevel]string{
	discordgo.VerificationLevelNone:   "None",
	discordgo.VerificationLevelLow:    "Low",
	discordgo.VerificationLevelMedium: "Medium",
	discordgo.VerificationLevelHigh:   "High",
}
var TypeMessages = map[string]int{
	"all":      MessagesAll,
	"mentions": MessagesMentions,
	"private":  MessagesPrivate,
	"current":  MessagesCurrent,
}
var TypeStatuses = map[string]discordgo.Status{
	"online":    discordgo.StatusOnline,
	"idle":      discordgo.StatusIdle,
	"dnd":       discordgo.StatusDoNotDisturb,
	"invisible": discordgo.StatusInvisible,
	"offline":   discordgo.StatusOffline,
}

type location struct {
	guild   *discordgo.Guild
	channel *discordgo.Channel
}

func (loc *location) push(guild *discordgo.Guild, channel *discordgo.Channel) {
	if loc.guild == guild && loc.channel == channel {
		return
	}
	if guild != nil && channel != nil && loc.guild != nil && loc.channel != nil &&
		loc.guild.ID == guild.ID && loc.channel.ID == channel.ID {
		return
	}
	lastLoc = *loc

	loc.guild = guild
	loc.channel = channel
	pointerCache = ""
}

var loc location
var lastLoc location
var lastMsg location

var lastUsedMsg string
var lastUsedRole string

var cacheGuilds = make(map[string]string)
var cacheChannels = make(map[string]string)
var cacheRead *discordgo.Message

const (
	MessagesNone = iota
	MessagesCurrent
	MessagesPrivate
	MessagesMentions
	MessagesAll
)

var messages = MessagesNone
var intercept = true

var webhookCommands = []string{"big", "say", "sayfile", "embed", "name", "avatar", "exit", "exec", "run"}

func command(session *discordgo.Session, cmd string) (returnVal string) {
	if cmd == "" {
		return
	}
	parts := strings.FieldsFunc(cmd, func(c rune) bool {
		return c != '\n' && unicode.IsSpace(c)
	})

	cmd = strings.ToLower(parts[0])
	args := parts[1:]
	nargs := len(args)

	if UserType == TypeWebhook {
		allowed := false
		for _, allow := range webhookCommands {
			if cmd == allow {
				allowed = true
			}
		}

		if !allowed {
			stdutil.PrintErr(tl("invalid.webhook.command"), nil)
			return
		}
	}

	switch cmd {
	case "help":
		search := strings.Join(args, " ")
		printHelp(search)
	case "exit":
		exit(session)
	case "exec":
		if nargs < 1 {
			stdutil.PrintErr("exec <command>", nil)
			return
		}

		cmd := strings.Join(args, " ")

		err := execute(SH, C, cmd)
		if err != nil {
			stdutil.PrintErr(tl("failed.exec"), err)
		}
	case "run":
		if nargs < 1 {
			stdutil.PrintErr("run <lua script>", nil)
			return
		}
		var script string
		var scriptArgs []string

		scriptName := true
		for i, arg := range args {
			if scriptName {
				if i != 0 {
					script += " "
				}
				if strings.HasSuffix(arg, ":") {
					scriptName = false
					arg = arg[:len(arg)-1]
				}
				script += arg
			} else {
				scriptArgs = append(scriptArgs, arg)
			}
		}

		err := fixPath(&script)
		if err != nil {
			stdutil.PrintErr(tl("failed.fixpath"), err)
		}

		err = RunLua(session, script, scriptArgs...)
		if err != nil {
			stdutil.PrintErr(tl("failed.lua.run"), err)
		}
	case "guilds":
		fallthrough
	case "guild":
		fallthrough
	case "channels":
		fallthrough
	case "channel":
		fallthrough
	case "dm":
		fallthrough
	case "pchannels":
		fallthrough
	case "bookmarks":
		fallthrough
	case "bookmark":
		fallthrough
	case "go":
		returnVal = commands_navigate(session, cmd, args, nargs)
	case "say":
		fallthrough
	case "tts":
		fallthrough
	case "embed":
		fallthrough
	case "big":
		fallthrough
	case "file":
		fallthrough
	case "sayfile":
		returnVal = commands_say(session, cmd, args, nargs)
	case "edit":
		if nargs < 2 {
			stdutil.PrintErr("edit <message id> <stuff>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		msg, err := session.ChannelMessageEdit(loc.channel.ID, args[0], strings.Join(args[1:], " "))
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.edit"), err)
			return
		}
		lastUsedMsg = msg.ID
	case "del":
		if nargs < 1 {
			stdutil.PrintErr("del <message id>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		err := session.ChannelMessageDelete(loc.channel.ID, args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.delete"), err)
			return
		}
	case "log":
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		directly := nargs < 1

		limit := 100
		if directly {
			limit = 10
		}

		msgs, err := session.ChannelMessages(loc.channel.ID, limit, "", "", "")
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.query"), err)
			return
		}
		s := ""

		for i := len(msgs) - 1; i >= 0; i-- {
			msg := msgs[i]
			if msg.Author == nil {
				return
			}
			if directly {
				s += "(ID " + msg.ID + ") "
			}
			s += msg.Author.Username + ": " + msg.Content + "\n"
		}

		if directly {
			fmt.Print(s)
			returnVal = s
			return
		}

		name := strings.Join(args, " ")
		err = fixPath(&name)
		if err != nil {
			stdutil.PrintErr(tl("failed.fixpath"), err)
		}

		err = ioutil.WriteFile(name, []byte(s), 0666)
		if err != nil {
			stdutil.PrintErr(tl("failed.file.write"), err)
			return
		}
		fmt.Println("Wrote chat log to '" + name + "'.")
	case "delall":
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}
		since := ""
		if nargs >= 1 {
			since = args[0]
		}
		messages, err := session.ChannelMessages(loc.channel.ID, 100, "", since, "")
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.query"), err)
			return
		}

		ids := make([]string, len(messages))
		for i, msg := range messages {
			ids[i] = msg.ID
		}

		err = session.ChannelMessagesBulkDelete(loc.channel.ID, ids)
		if err != nil {
			stdutil.PrintErr(tl("failed.msg.query"), err)
			return
		}
		returnVal := strconv.Itoa(len(ids))
		fmt.Println("Deleted " + returnVal + " messages!")
	case "members":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		members, err := session.GuildMembers(loc.guild.ID, "", 100)
		if err != nil {
			stdutil.PrintErr(tl("failed.members"), err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Name", "Nick")

		for _, member := range members {
			table.AddRow()
			table.AddStrings(member.User.ID, member.User.Username, member.Nick)
		}
		printTable(table)
	case "invite":
		if nargs >= 1 {
			if UserType != TypeUser {
				stdutil.PrintErr(tl("invalid.onlyfor.users"), nil)
				return
			}

			invite, err := session.InviteAccept(args[0])
			if err != nil {
				stdutil.PrintErr(tl("failed.invite.accept"), err)
				return
			}
			fmt.Println(tl("status.invite.accept"))

			loc.push(invite.Guild, invite.Channel)
		} else {
			if loc.channel == nil {
				stdutil.PrintErr(tl("failed.channel"), nil)
				return
			}
			invite, err := session.ChannelInviteCreate(loc.channel.ID, discordgo.Invite{})
			if err != nil {
				stdutil.PrintErr(tl("failed.invite.create"), err)
				return
			}
			fmt.Println(tl("status.invite.create") + " " + invite.Code)
			returnVal = invite.Code
		}
	case "roles":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		roles, err := session.GuildRoles(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.roles"), err)
			return
		}
		sort.Slice(roles, func(i, j int) bool {
			return roles[i].Position > roles[j].Position
		})

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Name", "Permissions")

		for _, role := range roles {
			table.AddRow()
			table.AddStrings(role.ID, role.Name, strconv.Itoa(role.Permissions))
		}

		printTable(table)
	case "roleadd":
		fallthrough
	case "roledel":
		if nargs < 2 {
			stdutil.PrintErr("roleadd/del <user id> <role id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		var err error
		if cmd == "roleadd" {
			err = session.GuildMemberRoleAdd(loc.guild.ID, args[0], args[1])
		} else {
			err = session.GuildMemberRoleRemove(loc.guild.ID, args[0], args[1])
		}

		if err != nil {
			stdutil.PrintErr(tl("failed.role.change"), err)
		}
	case "enablemessages":
		if len(args) < 1 {
			messages = MessagesCurrent
			return
		}

		val, ok := TypeMessages[strings.ToLower(args[0])]
		if !ok {
			stdutil.PrintErr(tl("invalid.value"), nil)
			return
		}
		messages = val
		fmt.Println(tl("status.msg.intercept"))
	case "disablemessages":
		messages = MessagesNone
		fmt.Println(tl("status.msg.nointercept"))
	case "enableintercept":
		intercept = true
		fmt.Println(tl("status.cmd.intercept"))
	case "disableintercept":
		intercept = false
		fmt.Println(tl("status.cmd.nointercept"))
	case "reply":
		loc.push(lastMsg.guild, lastMsg.channel)
	case "back":
		loc, lastLoc = lastLoc, loc
		pointerCache = ""
	case "rolecreate":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		role, err := session.GuildRoleCreate(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.role.create"), err)
			return
		}
		fmt.Println("Created role with ID " + role.ID)
		lastUsedRole = role.ID
		returnVal = role.ID
	case "roleedit":
		if nargs < 3 {
			stdutil.PrintErr("roleedit <roleid> <flag> <value>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		value := strings.Join(args[2:], " ")

		roles, err := session.GuildRoles(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.roles"), err)
			return
		}

		var role *discordgo.Role
		for _, r := range roles {
			if r.ID == args[0] {
				role = r
				break
			}
		}
		if role == nil {
			stdutil.PrintErr(tl("invalid.role"), nil)
			return
		}

		name := role.Name
		color := int64(role.Color)
		hoist := role.Hoist
		perms := role.Permissions
		mention := role.Mentionable

		switch strings.ToLower(args[1]) {
		case "name":
			name = value
		case "color":
			value = strings.TrimPrefix(value, "#")
			color, err = strconv.ParseInt(value, 16, 0)
			if err != nil {
				stdutil.PrintErr(tl("invalid.number"), nil)
				return
			}
		case "separate":
			hoist, err = parseBool(value)
			if err != nil {
				stdutil.PrintErr(err.Error(), nil)
				return
			}
		case "perms":
			perms, err = strconv.Atoi(value)
			if err != nil {
				stdutil.PrintErr(tl("invalid.number"), nil)
				return
			}
		case "mention":
			mention, err = parseBool(value)
			if err != nil {
				stdutil.PrintErr(err.Error(), nil)
				return
			}
		default:
			stdutil.PrintErr(tl("invalid.value"), nil)
			return
		}

		role, err = session.GuildRoleEdit(loc.guild.ID, args[0], name, int(color), hoist, perms, mention)
		if err != nil {
			stdutil.PrintErr(tl("failed.role.edit"), err)
			return
		}
		lastUsedRole = role.ID
		fmt.Println("Edited role " + role.ID)
	case "roledelete":
		if nargs < 1 {
			stdutil.PrintErr("roledelete <roleid>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		err := session.GuildRoleDelete(loc.guild.ID, args[0])
		if err != nil {
			fmt.Println(tl("failed.role.delete"), err)
		}
	case "ban":
		if nargs < 1 {
			stdutil.PrintErr("ban <user id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		err := session.GuildBanCreate(loc.guild.ID, args[0], 0)
		if err != nil {
			stdutil.PrintErr(tl("failed.ban.create"), err)
		}
	case "unban":
		if nargs < 1 {
			stdutil.PrintErr("unban <user id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		err := session.GuildBanDelete(loc.guild.ID, args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.ban.delete"), err)
		}
	case "kick":
		if nargs < 1 {
			stdutil.PrintErr("kick <user id>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		err := session.GuildMemberDelete(loc.guild.ID, args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.kick"), err)
		}
	case "leave":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		err := session.GuildLeave(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.leave"), err)
			return
		}

		loc.push(nil, nil)
	case "bans":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		bans, err := session.GuildBans(loc.guild.ID)
		if err != nil {
			stdutil.PrintErr(tl("failed.ban.list"), err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("User ID", "Username", "Reason")

		for _, ban := range bans {
			table.AddRow()
			table.AddStrings(ban.User.ID, ban.User.Username, ban.Reason)
		}

		printTable(table)
	case "nickall":
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}

		members, err := session.GuildMembers(loc.guild.ID, "", 100)
		if err != nil {
			stdutil.PrintErr(tl("failed.members"), err)
			return
		}

		nick := strings.Join(args, " ")

		for _, member := range members {
			err := session.GuildMemberNickname(loc.guild.ID, member.User.ID, nick)
			if err != nil {
				stdutil.PrintErr(tl("failed.nick"), err)
			}
		}
	case "vchannels":
		channels(session, "voice")
	case "play":
		if UserType != TypeBot {
			stdutil.PrintErr(tl("invalid.onlyfor.bots"), nil)
			return
		}
		if nargs < 1 {
			stdutil.PrintErr("play <dca audio file>", nil)
			return
		}
		if loc.guild == nil {
			stdutil.PrintErr(tl("invalid.guild"), nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}
		if playing != "" {
			stdutil.PrintErr(tl("invalid.music.playing"), nil)
			return
		}

		file := strings.Join(args, " ")
		err := fixPath(&file)
		if err != nil {
			stdutil.PrintErr(tl("failed.fixpath"), err)
		}

		playing = file

		fmt.Println(tl("status.loading"))

		var buffer [][]byte
		err = loadAudio(file, &buffer)
		if err != nil {
			stdutil.PrintErr(tl("failed.file.load"), err)
			playing = ""
			return
		}

		fmt.Println("Loaded!")
		fmt.Println("Playing!")

		go func(buffer [][]byte, session *discordgo.Session, guild, channel string) {
			play(buffer, session, guild, channel)
			playing = ""
		}(buffer, session, loc.guild.ID, loc.channel.ID)
	case "stop":
		if UserType != TypeBot {
			stdutil.PrintErr(tl("invalid.onlyfor.bots"), nil)
			return
		}
		playing = ""
	case "reactadd":
		fallthrough
	case "reactdel":
		if nargs < 2 {
			stdutil.PrintErr("reactadd/reactdel <message id> <emoji unicode/id>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		var err error
		if cmd == "reactadd" {
			err = session.MessageReactionAdd(loc.channel.ID, args[0], args[1])
		} else {
			err = session.MessageReactionRemove(loc.channel.ID, args[0], args[1], "@me")
		}
		if err != nil {
			stdutil.PrintErr(tl("failed.react"), err)
			return
		}
	case "block":
		if nargs < 1 {
			stdutil.PrintErr("block <user id>", nil)
			return
		}
		if UserType != TypeUser {
			stdutil.PrintErr(tl("invalid.onlyfor.users"), nil)
			return
		}
		err := session.RelationshipUserBlock(args[0])
		if err != nil {
			stdutil.PrintErr(tl("failed.block"), err)
			return
		}
	case "friends":
		if UserType != TypeUser {
			stdutil.PrintErr(tl("invalid.onlyfor.users"), nil)
			return
		}
		relations, err := session.RelationshipsGet()
		if err != nil {
			stdutil.PrintErr(tl("failed.friends"), err)
			return
		}

		table := gtable.NewStringTable()
		table.AddStrings("ID", "Type", "Name")

		for _, relation := range relations {
			table.AddRow()
			table.AddStrings(relation.ID, TypeRelationships[relation.Type], relation.User.Username)
		}

		printTable(table)
	case "reactbig":
		if nargs < 2 {
			stdutil.PrintErr("reactbig <message id> <text>", nil)
			return
		}
		if loc.channel == nil {
			stdutil.PrintErr(tl("invalid.channel"), nil)
			return
		}

		used := ""

		for _, c := range strings.Join(args[1:], " ") {
			str := string(toEmoji(c))

			if strings.Contains(used, str) {
				fmt.Println(tl("failed.react.used"))
				continue
			}
			used += str

			err := session.MessageReactionAdd(loc.channel.ID, args[0], str)
			if err != nil {
				stdutil.PrintErr(tl("failed.react"), err)
			}
		}
	case "rl":
		full := nargs >= 1 && strings.EqualFold(args[0], "full")

		var err error
		if full {
			fmt.Println(tl("restarting.session"))
			err = session.Close()
			if err != nil {
				stdutil.PrintErr(tl("failed.session.close"), err)
				return
			}
			err = session.Open()
			if err != nil {
				stdutil.PrintErr(tl("failed.session.start"), err)
			}
		}

		fmt.Println(tl("restarting.cache.loc"))
		var guild *discordgo.Guild
		var channel *discordgo.Channel

		if loc.guild != nil {
			guild, err = session.Guild(loc.guild.ID)

			if err != nil {
				stdutil.PrintErr(tl("failed.guild"), err)
				return
			}
		}

		if loc.channel != nil {
			channel, err = session.Channel(loc.channel.ID)

			if err != nil {
				stdutil.PrintErr(tl("failed.channel"), err)
				return
			}
		}

		loc.guild = guild
		loc.channel = channel
		pointerCache = ""

		fmt.Println(tl("restarting.cache.vars"))
		cacheGuilds = make(map[string]string)
		cacheChannels = make(map[string]string)
		cacheAudio = make(map[string][][]byte)

		lastLoc = location{}
		lastMsg = location{}
		lastUsedMsg = ""
		lastUsedRole = ""

		cacheRead = nil
	case "status":
		if nargs < 1 {
			stdutil.PrintErr("status <value>", nil)
			return
		}
		status, ok := TypeStatuses[strings.ToLower(args[0])]
		if !ok {
			stdutil.PrintErr(tl("invalid.value"), nil)
			return
		}

		if status == discordgo.StatusOffline {
			stdutil.PrintErr(tl("invalid.status.offline"), nil)
			return
		}

		_, err := session.UserUpdateStatus(status)
		if err != nil {
			stdutil.PrintErr(tl("failed.status"), err)
			return
		}
		fmt.Println(tl("status.status"))
	case "avatar":
		fallthrough
	case "name":
		fallthrough
	case "playing":
		fallthrough
	case "streaming":
		fallthrough
	case "typing":
		fallthrough
	case "nick":
		returnVal = commands_usermod(session, cmd, args, nargs)
	case "read":
		fallthrough
	case "cinfo":
		fallthrough
	case "ginfo":
		fallthrough
	case "uinfo":
		returnVal = commands_query(session, cmd, args, nargs)
	default:
		stdutil.PrintErr(tl("invalid.command"), nil)
	}
	return
}

func parseBool(str string) (bool, error) {
	if str == "yes" || str == "true" || str == "y" {
		return true, nil
	} else if str == "no" || str == "false" || str == "n" {
		return false, nil
	}
	return false, errors.New(tl("invalid.yn"))
}

func printTable(table gtable.StringTable) {
	table.Each(func(ti *gtable.TableItem) {
		ti.Padding(1)
	})
	fmt.Println(table.String())
}
