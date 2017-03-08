package main;

import (
	"fmt"
	"github.com/legolord208/stdutil"
	"github.com/bwmarrin/discordgo"
	"strings"
	"github.com/legolord208/gtable"
	"io/ioutil"
	"os"
	"strconv"
	"sort"
	"errors"
	"encoding/json"
	"time"
)

var RELATIONSHIP_TYPES = map[int]string{
	1: "Friend",
	2: "Blocked",
	3: "Incoming request",
	4: "Sent request",
};

type location struct{
	guildID string
	channelID string
}

var loc location;
var lastMsg location;
var lastLoc location;

var lastUsedMsg string;
var lastUsedRole string;

var cacheGuilds = make(map[string]string, 0);
var cacheChannels = make(map[string]string, 0);
var cacheRead *discordgo.Message;

var messages bool;
var intercept = true;

func command(session *discordgo.Session, cmd string) (returnVal string){
	if(cmd == ""){
		return;
	}
	parts := strings.Fields(cmd);

	cmd = strings.ToLower(parts[0]);
	args := parts[1:];
	nargs := len(args);

	if(cmd == "help"){
		search := strings.Join(args, " ");
		printHelp(search);
	} else if(cmd == "exit"){
		exit(session);
	} else if(cmd == "exec" || cmd == "execute"){
		if(nargs < 1){
			stdutil.PrintErr("exec <command>", nil);
			return;
		}

		cmd := strings.Join(args, " ");

		var err error;
		if(WINDOWS){
			err = execute("cmd", "/c", cmd);
		} else {
			err = execute("sh", "-c", cmd);
		}
		if(err != nil){
			stdutil.PrintErr("Could not execute", err);
		}
	} else if(cmd == "run"){
		if(nargs < 1){
			stdutil.PrintErr("run <lua script>", nil);
			return;
		}
		var script string;
		var scriptArgs []string;

		scriptName := true;
		for i, arg := range args{
			if(scriptName){
				if(i != 0){
					script += " ";
				}
				if(strings.HasSuffix(arg, ":")){
					scriptName = false;
					arg = arg[:len(arg) - 1];
				}
				script += arg;
			} else {
				scriptArgs = append(scriptArgs, arg);
			}
		}

		err := RunLua(session, script, scriptArgs...);
		if(err != nil){
			stdutil.PrintErr("Could not run lua", err);
		}
	} else if(cmd == "guilds"){
		guilds, err := session.UserGuilds();
		if(err != nil){
			stdutil.PrintErr("Could not get guilds", err);
			return;
		}

		cacheGuilds = make(map[string]string, 0);

		table := gtable.NewStringTable();
		table.AddStrings("ID", "Name")

		for _, guild := range guilds{
			table.AddRow();
			table.AddStrings(guild.ID, guild.Name);
			cacheGuilds[strings.ToLower(guild.Name)] = guild.ID;
		}

		printTable(table);
	} else if(cmd == "guild"){
		if(nargs < 1){
			stdutil.PrintErr("guild <id>", nil);
			return;
		}

		lastLoc = loc;

		var ok bool;
		loc.guildID, ok = cacheGuilds[strings.ToLower(strings.Join(args, " "))];

		if(!ok){
			loc.guildID = args[0];
		}

		loc.channelID = loc.guildID;
		clearPointerCache();
	} else if(cmd == "channels"){
		channels(session, "text");
	} else if(cmd == "channel"){
		if(nargs < 1){
			stdutil.PrintErr("channel <id>", nil);
			return;
		}

		var channelID string;
		var ok bool;
		channelID, ok = cacheChannels[strings.ToLower(strings.Join(args, " "))];
		if(!ok){
			channelID = args[0];
		}

		channel, err := session.Channel(channelID);
		if(err != nil){
			stdutil.PrintErr("Could not get channel ", err);
			return;
		}
		lastLoc = loc;

		loc.channelID = channelID;
		loc.guildID = channel.GuildID;
		clearPointerCache();
	} else if(cmd == "say"){
		if(nargs < 1){
			stdutil.PrintErr("say <stuff>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		msg, err := session.ChannelMessageSend(loc.channelID, strings.Join(args, " "));
		if(err != nil){
			stdutil.PrintErr("Could not send", err);
			return;
		}
		fmt.Println("Created message with ID " + msg.ID);
		lastUsedMsg = msg.ID;
		returnVal = msg.ID;
	} else if(cmd == "edit"){
		if(nargs < 2){
			stdutil.PrintErr("edit <message id> <stuff>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		msg, err := session.ChannelMessageEdit(loc.channelID, args[0], strings.Join(args[1:], " "));
		if(err != nil){
			stdutil.PrintErr("Could not edit", err);
			return;
		}
		fmt.Println("Edited " + msg.ID + "!");
		lastUsedMsg = msg.ID;
	} else if(cmd == "del"){
		if(nargs < 1){
			stdutil.PrintErr("del <message id>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		err := session.ChannelMessageDelete(loc.channelID, args[0]);
		if(err != nil){
			stdutil.PrintErr("Couldn't delete", err);
			return;
		}
	} else if(cmd == "log"){
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		directly := nargs < 1;

		limit := 100;
		if(directly){
			limit = 10;
		}

		msgs, err := session.ChannelMessages(loc.channelID, limit, "", "");
		if(err != nil){
			stdutil.PrintErr("Could not get messages", err);
			return;
		}
		s := "";

		for i := len(msgs) - 1; i >= 0; i--{
			msg := msgs[i];
			if(msg.Author == nil){
				return;
			}
			if(directly){
				s += "(ID " + msg.ID + ") ";
			}
			s += msg.Author.Username + ": " + msg.Content + "\n";
		}

		if(directly){
			fmt.Print(s);
			returnVal = s;
			return;
		}

		name := strings.Join(args, " ");
		err = ioutil.WriteFile(name, []byte(s), 0666);
		if(err != nil){
			stdutil.PrintErr("Could not write log file", err);
			return;
		}
		fmt.Println("Wrote chat log to '" + name + "'.");
	} else if(cmd == "playing"){
		if(USER){
			stdutil.PrintErr("This only works for bots.", nil);
			return;
		}
		err := session.UpdateStatus(0, strings.Join(args, " "));
		if(err != nil){
			stdutil.PrintErr("Couldn't update status", err);
		}
	} else if(cmd == "streaming"){
		if(USER){
			stdutil.PrintErr("This only works for bots.", nil);
			return;
		}
		var err error;
		if(nargs <= 0){
			err = session.UpdateStreamingStatus(0, "", "");
		} else if(nargs < 2){
			err = session.UpdateStreamingStatus(0, strings.Join(args[1:], " "), "");
		} else {
			err = session.UpdateStreamingStatus(0, strings.Join(args[1:], " "), args[0]);
		}
		if(err != nil){
			stdutil.PrintErr("Couldn't update status", err);
		}
	} else if(cmd == "typing"){
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected.", nil);
			return;
		}
		err := session.ChannelTyping(loc.channelID);
		if(err != nil){
			stdutil.PrintErr("Couldn't start typing", err);
		}
	} else if(cmd == "pchannels"){
		channels, err := session.UserChannels();
		if(err != nil){
			stdutil.PrintErr("Could not get private channels", err);
			return;
		}

		table := gtable.NewStringTable();
		table.AddStrings("ID", "Recipient");

		for _, channel := range channels{
			table.AddRow();
			table.AddStrings(channel.ID, channel.Recipient.Username);
		}
		printTable(table);
	} else if(cmd == "dm"){
		if(nargs < 1){
			stdutil.PrintErr("dm <user id>", nil);
			return;
		}
		channel, err := session.UserChannelCreate(args[0]);
		if(err != nil){
			stdutil.PrintErr("Could not create DM", err);
			return;
		}
		lastLoc, loc = loc, lastLoc;

		loc.channelID = channel.ID;
		loc.guildID = "";
		clearPointerCache();

		fmt.Println("Selected DM with channel ID " + channel.ID + ".");
	} else if(cmd == "delall"){
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected.", nil);
			return;
		}
		since := "";
		if(nargs >= 1){
			since = args[0];
		}
		messages, err := session.ChannelMessages(loc.channelID, 100, "", since);
		if(err != nil){
			stdutil.PrintErr("Could not get messages", err);
			return;
		}

		ids := make([]string, len(messages));
		for i, msg := range messages{
			ids[i] = msg.ID;
		}

		err = session.ChannelMessagesBulkDelete(loc.channelID, ids);
		if(err != nil){
			stdutil.PrintErr("Could not delete messages", err);
			return;
		}
		returnVal := strconv.Itoa(len(ids));
		fmt.Println("Deleted " + returnVal + " messages!");
	} else if(cmd == "members"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected", nil);
			return;
		}

		members, err := session.GuildMembers(loc.guildID, "", 100);
		if(err != nil){
			stdutil.PrintErr("Could not list members", err);
			return;
		}

		table := gtable.NewStringTable();
		table.AddStrings("ID", "Name", "Nick",);

		for _, member := range members{
			table.AddRow();
			table.AddStrings(member.User.ID, member.User.Username, member.Nick);
		}
		printTable(table);
	} else if(cmd == "invite"){
		if(nargs >= 1){
			if(!USER){
				stdutil.PrintErr("This only works for users.", nil);
				return;
			}

			invite, err := session.InviteAccept(args[0]);
			if(err != nil){
				stdutil.PrintErr("Could not accept invite", err);
				return;
			}
			fmt.Println("Accepted invite.");

			lastLoc, loc = loc, lastLoc;
			loc.guildID = invite.Guild.ID;
			loc.channelID = invite.Channel.ID;
			clearPointerCache();
		} else {
			if(loc.channelID == ""){
				stdutil.PrintErr("No channel selected", nil);
				return;
			}
			invite, err := session.ChannelInviteCreate(loc.channelID, discordgo.Invite{});
			if(err != nil){
				stdutil.PrintErr("Invite could not be created", err);
				return;
			}
			fmt.Println("Created invite with code " + invite.Code);
			returnVal = invite.Code;
		}
	} else if(cmd == "file"){
		if(nargs < 1){
			stdutil.PrintErr("file <file>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected", nil);
			return;
		}
		name := strings.Join(args, " ");
		file, err := os.Open(name);
		if(err != nil){
			stdutil.PrintErr("Couldn't open file", nil);
			return;
		}
		defer file.Close();

		msg, err := session.ChannelFileSend(loc.channelID, name, file);
		if(err != nil){
			stdutil.PrintErr("Could not send file", err);
			return;
		}
		fmt.Println("Sent '" + name + "' with message ID " + msg.ID + ".");
		return msg.ID;
	} else if(cmd == "roles"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected", nil);
			return;
		}

		roles, err := session.GuildRoles(loc.guildID);
		if(err != nil){
			stdutil.PrintErr("Could not get roles", err);
			return;
		}
		sort.Slice(roles, func(i, j int) bool{
			return roles[i].Position > roles[j].Position;
		});

		table := gtable.NewStringTable();
		table.AddStrings("ID", "Name", "Permissions");

		for _, role := range roles{
			table.AddRow();
			table.AddStrings(role.ID, role.Name, strconv.Itoa(role.Permissions));
		}

		printTable(table);
	} else if(cmd == "roleadd" || cmd == "roledel"){
		if(nargs < 2){
			stdutil.PrintErr("roleadd/del <user id> <role id>", nil);
			return;
		}
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected", nil);
			return;
		}

		var err error;
		if(cmd == "roleadd"){
			err = session.GuildMemberRoleAdd(loc.guildID, args[0], args[1]);
		} else {
			err = session.GuildMemberRoleRemove(loc.guildID, args[0], args[1]);
		}

		if(err != nil){
			stdutil.PrintErr("Could not add/remove role", err);
		}
	} else if(cmd == "nick"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected.", nil);
			return;
		}
		err := session.GuildMemberNickname(loc.guildID, "@me/nick", strings.Join(args, " "));
		if(err != nil){
			stdutil.PrintErr("Could not set nickname", err);
		}
	} else if(cmd == "enablemessages"){ messages = true; fmt.Println("Messages will now be intercepted.");
	} else if(cmd == "disablemessages"){ messages = false; fmt.Println("Messages will no longer be intercepted.");
	} else if(cmd == "enableintercept"){ intercept = true; fmt.Println("'console.' commands will now be intercepted.");
	} else if(cmd == "disableintercept"){ intercept = false; fmt.Println("'console.' commands will no longer be intercepted.");
	} else if(cmd == "reply"){
		lastLoc = loc;

		loc.guildID = lastMsg.guildID;
		loc.channelID = lastMsg.channelID;
		clearPointerCache();
	} else if(cmd == "back"){
		loc, lastLoc = lastLoc, loc;
		clearPointerCache();
	} else if(cmd == "rolecreate"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}

		role, err := session.GuildRoleCreate(loc.guildID);
		if(err != nil){
			stdutil.PrintErr("Could not create role", err);
			return;
		}
		fmt.Println("Created role with ID " + role.ID + ".");
		lastUsedRole = role.ID;
		returnVal = role.ID;
	} else if(cmd == "roleedit"){
		if(nargs < 3){
			stdutil.PrintErr("roleedit <roleid> <flag> <value>", nil);
			return;
		}
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}

		value := strings.Join(args[2:], " ");

		roles, err := session.GuildRoles(loc.guildID);
		if(err != nil){
			stdutil.PrintErr("Could not get roles", err);
			return;
		}

		var role *discordgo.Role;
		for _, r := range roles{
			if(r.ID == args[0]){
				role = r;
				break;
			}
		}
		if(role == nil){
			stdutil.PrintErr("Role does not exist with that ID", nil);
			return;
		}

		name := role.Name;
		color := int64(role.Color);
		hoist := role.Hoist;
		perms := role.Permissions;
		mention := role.Mentionable;

		switch(strings.ToLower(args[1])){
			case "name":
				name = value;
			case "color":
				value = strings.TrimPrefix(value, "#");
				color, err = strconv.ParseInt(value, 16, 0);
				if(err != nil){
					stdutil.PrintErr("Not a number", nil);
					return;
				}
			case "separate":
				hoist, err = parseBool(value);
				if(err != nil){
					stdutil.PrintErr(err.Error(), nil);
					return;
				}
			case "perms":
				perms, err = strconv.Atoi(value);
				if(err != nil){
					stdutil.PrintErr("Not a number", nil);
					return;
				}
			case "mention":
				mention, err = parseBool(value);
				if(err != nil){
					stdutil.PrintErr(err.Error(), nil);
					return;
				}
			default:
				stdutil.PrintErr("No such property", nil);
				return;
		}

		role, err = session.GuildRoleEdit(loc.guildID, args[0], name, int(color), hoist, perms, mention);
		if(err != nil){
			stdutil.PrintErr("Could not edit role", err);
			return;
		}
		lastUsedRole = role.ID;
		fmt.Println("Edited role " + role.ID + ".");
	} else if(cmd == "roledelete"){
		if(nargs < 1){
			stdutil.PrintErr("roledelete <roleid>", nil);
			return;
		}
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}

		err := session.GuildRoleDelete(loc.guildID, args[0]);
		if(err != nil){
			fmt.Println("Could not delete role!", err);
		}
	} else if(cmd == "ban"){
		if(nargs < 1){
			stdutil.PrintErr("ban <user id>", nil);
			return;
		}
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}

		err := session.GuildBanCreate(loc.guildID, args[0], 0);
		if(err != nil){
			stdutil.PrintErr("Could not ban user", err);
		}
	} else if(cmd == "unban"){
		if(nargs < 1){
			stdutil.PrintErr("unban <user id>", nil);
			return;
		}
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}

		err := session.GuildBanDelete(loc.guildID, args[0]);
		if(err != nil){
			stdutil.PrintErr("Could not unban user", err);
		}
	} else if(cmd == "kick"){
		if(nargs < 1){
			stdutil.PrintErr("kick <user id>", nil);
			return;
		}
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}

		err := session.GuildMemberDelete(loc.guildID, args[0]);
		if(err != nil){
			stdutil.PrintErr("Could not kick user", err);
		}
	} else if(cmd == "leave"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}

		err := session.GuildLeave(loc.guildID);
		if(err != nil){
			stdutil.PrintErr("Could not leave", err);
			return;
		}

		loc = location{};
		clearPointerCache();
	} else if(cmd == "bans"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}

		bans, err := session.GuildBans(loc.guildID);
		if(err != nil){
			stdutil.PrintErr("Could not list bans", err);
			return;
		}

		table := gtable.NewStringTable();
		table.AddStrings("User", "Reason");

		for _, ban := range bans{
			table.AddRow();
			table.AddStrings(ban.User.Username, ban.Reason);
		}

		printTable(table);
	} else if(cmd == "nickall"){
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}

		members, err := session.GuildMembers(loc.guildID, "", 100);
		if(err != nil){
			stdutil.PrintErr("Could not get members", err);
			return;
		}

		nick := strings.Join(args, " ");

		for _, member := range members{
			err := session.GuildMemberNickname(loc.guildID, member.User.ID, nick);
			if(err != nil){
				stdutil.PrintErr("Could not nickname", err);
			}
		}
	} else if(cmd == "embed"){
		if(nargs < 1){
			stdutil.PrintErr("embed <embed json>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		jsonstr := strings.Join(args, " ");
		var embed = &discordgo.MessageEmbed{};

		err := json.Unmarshal([]byte(jsonstr), embed);
		if(err != nil){
			stdutil.PrintErr("Could not parse json", err);
			return;
		}

		msg, err := session.ChannelMessageSendEmbed(loc.channelID, embed);
		if(err != nil){
			stdutil.PrintErr("Could not send embed", err);
			return;
		}
		fmt.Println("Created message with ID " + msg.ID + ".");
		lastUsedMsg = msg.ID;
		returnVal = msg.ID;
	} else if(cmd == "read"){
		if(nargs < 1){
			stdutil.PrintErr("read <message id> [property]", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}
		msgID := args[0];

		var msg *discordgo.Message;
		var err error;
		if(strings.EqualFold(msgID, "cache")){
			if(cacheRead == nil){
				stdutil.PrintErr("No cache!", nil);
				return;
			}

			msg = cacheRead;
		} else {
			msg, err = getMessage(session, loc.channelID, msgID);
		}
		if(err != nil){
			stdutil.PrintErr("Could not get message", err);
			return;
		}

		property := "";
		if(len(args) >= 2){
			property = strings.ToLower(args[1]);
		}
		switch(property){
			case "":                printMessage(session, msg, false, nil);
			case "cache":           cacheRead = msg; fmt.Println("Message cached!");
			case "text":            returnVal = msg.Content;
			case "channel":         returnVal = msg.ChannelID;
			case "timestamp":
				sent, err := msg.Timestamp.Parse();
				if(err != nil){
					stdutil.PrintErr("Could not parse timestamp!", err);
					return;
				}
				returnVal = sent.Format(time.ANSIC);
				if(msg.EditedTimestamp != ""){
					returnVal += "*";
				}
			case "author":          returnVal = msg.Author.ID;
			case "author_name":     returnVal = msg.Author.Username;
			case "author_avatar":   returnVal = msg.Author.Avatar;
			case "author_bot":      returnVal = strconv.FormatBool(msg.Author.Bot);
			default:                stdutil.PrintErr("Invalid property", nil);
		}

		lastUsedMsg = msg.ID;
		if(returnVal != ""){
			fmt.Println(returnVal);
		}
	} else if(cmd == "cinfo"){
		if(nargs < 1){
			stdutil.PrintErr("cinfo <property>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		channel, err := session.Channel(loc.channelID);
		if(err != nil){
			stdutil.PrintErr("Could not get channel", err);
			return;
		}

		switch(strings.ToLower(args[0])){
			case "guild":
				fmt.Println(channel.GuildID);
				returnVal = channel.GuildID;
			case "name":
				fmt.Println(channel.Name);
				returnVal = channel.Name;
			case "topic":
				fmt.Println(channel.Topic);
				returnVal = channel.Topic;
			case "type":
				fmt.Println(channel.Type);
				returnVal = channel.Type;
			default:
				stdutil.PrintErr("No such property!", nil);
		}
	} else if(cmd == "vchannels"){
		channels(session, "voice");
	} else if(cmd == "play"){
		if(USER){
			stdutil.PrintErr("This command only works for bot users.", nil);
			return;
		}
		if(nargs < 1){
			stdutil.PrintErr("play <dca audio file>", nil);
			return;
		}
		if(loc.guildID == ""){
			stdutil.PrintErr("No guild selected!", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}
		if(playing != ""){
			stdutil.PrintErr("Already playing something", nil);
			return;
		}

		file := strings.Join(args, " ");
		playing = file;

		fmt.Println("Loading...");

		var buffer [][]byte;
		err := load(file, &buffer);
		if(err != nil){
			stdutil.PrintErr("Could not load file.", err);
			playing = "";
			return;
		}

		fmt.Println("Loaded!");
		fmt.Println("Playing!");

		go func(buffer [][]byte, session *discordgo.Session, guild, channel string){
			play(buffer, session, guild, channel);
			playing = "";
		}(buffer, session, loc.guildID, loc.channelID);
	} else if(cmd == "stop"){
		if(USER){
			stdutil.PrintErr("This command only works for bot users.", nil);
			return;
		}
		playing = "";
	} else if(cmd == "reactadd" || cmd == "reactdel"){
		if(nargs < 2){
			stdutil.PrintErr("reactadd/reactdel <message id> <emoji unicode/id>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("No channel selected!", nil);
			return;
		}

		var err error;
		if(cmd == "reactadd"){
			err = session.MessageReactionAdd(loc.channelID, args[0], args[1]);
		} else {
			err = session.MessageReactionRemove(loc.channelID, args[0], args[1], "@me");
		}
		if(err != nil){
			stdutil.PrintErr("Could not react", err);
			return;
		}
	} else if(cmd == "quote"){
		if(nargs < 1){
			stdutil.PrintErr("quote <message id>", nil);
			return;
		}
		if(loc.channelID == ""){
			stdutil.PrintErr("You're not in a channel!", nil);
			return;
		}

		msg, err := getMessage(session, loc.channelID, args[0]);
		if(err != nil){
			stdutil.PrintErr("Could not get message", err);
			return;
		}

		timestamp, err := msg.Timestamp.Parse();
		if(err != nil){
			stdutil.PrintErr("Could not parse timestamp", err);
			return;
		}
		t := timestamp.Format(time.ANSIC);
		if(msg.EditedTimestamp != ""){
			t += "*";
		}

		msg, err = session.ChannelMessageSendEmbed(loc.channelID, &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name: msg.Author.Username,
				IconURL: "https://cdn.discordapp.com/avatars/" + msg.Author.ID + "/" + msg.Author.Avatar,
			},
			Description: msg.Content,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Sent " + t,
			},
		});
		if(err != nil){
			stdutil.PrintErr("Could not send quote", err);
			return;
		}
		fmt.Println("Created message with ID " + msg.ID + ".");
		lastUsedMsg = msg.ID;
		returnVal = msg.ID;
	} else if(cmd == "block"){
		if(nargs < 1){
			stdutil.PrintErr("block <user id>", nil);
			return;
		}
		if(!USER){
			stdutil.PrintErr("Only users can use this.", nil);
			return;
		}
		err := session.RelationshipUserBlock(args[0]);
		if(err != nil){
			stdutil.PrintErr("Couldn't block user", err);
			return;
		}
	} else if(cmd == "friends"){
		if(!USER){
			stdutil.PrintErr("Only users can use this.", nil);
			return;
		}
		relations, err := session.RelationshipsGet();
		if(err != nil){
			stdutil.PrintErr("Couldn't block user", err);
			return;
		}

		table := gtable.NewStringTable();
		table.AddStrings("ID", "Type", "Name");

		for _, relation := range relations{
			table.AddRow();
			table.AddStrings(relation.ID, RELATIONSHIP_TYPES[relation.Type], relation.User.Username);
		}

		printTable(table);
	} else {
		stdutil.PrintErr("Unknown command. Do 'help' for help", nil);
	}
	return;
}

func channels(session *discordgo.Session, kind string){
	if(loc.guildID == ""){
		stdutil.PrintErr("No guild selected!", nil);
		return;
	}
	channels, err := session.GuildChannels(loc.guildID);
	if(err != nil){
		stdutil.PrintErr("Could not get channels", nil);
		return;
	}

	cacheChannels = make(map[string]string);

	table := gtable.NewStringTable();
	table.AddStrings("ID", "Name");

	for _, channel := range channels{
		if(channel.Type != kind){
			continue;
		}
		table.AddRow();
		table.AddStrings(channel.ID, channel.Name);
		cacheChannels[strings.ToLower(channel.Name)] = channel.ID;
	}

	printTable(table);
}

func parseBool(str string) (bool, error){
	if(str == "yes" || str == "true"){
		return true, nil;
	} else if(str == "no" || str == "false"){
		return false, nil;
	}
	return false, errors.New("Please use yes or no");
}

func printTable(table gtable.StringTable){
	table.Each(func(ti *gtable.TableItem){
		ti.Padding(1);
	});
	fmt.Println(table.String());
}
